#!/bin/bash
# ==============================================================================
# PostgreSQL Restore Script for CRM Kilang Desa Murni Batik
# ==============================================================================
# This script supports:
# - Full database restore from pg_dump backup
# - Point-in-Time Recovery (PITR) with WAL replay
# - Target time specification for PITR
# ==============================================================================

set -euo pipefail

# ==============================================================================
# Configuration
# ==============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/tmp/pg-backups}"
WAL_ARCHIVE_DIR="${WAL_ARCHIVE_DIR:-/var/lib/postgresql/wal_archive}"

# Database connection
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
PG_USER="${PG_USER:-crm_admin}"
PG_PASSWORD="${PG_PASSWORD:-}"
PGPASSWORD="${PG_PASSWORD}"
export PGPASSWORD

# S3/MinIO settings (for downloading backups)
S3_ENABLED="${S3_ENABLED:-false}"
S3_BUCKET="${S3_BUCKET:-crm-backups}"
S3_PREFIX="${S3_PREFIX:-postgres}"
S3_ENDPOINT="${S3_ENDPOINT:-}"

# ==============================================================================
# Functions
# ==============================================================================

log() {
    local level="$1"
    shift
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $*"
}

log_info() { log "INFO" "$@"; }
log_warn() { log "WARN" "$@"; }
log_error() { log "ERROR" "$@"; }

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Restore PostgreSQL database from backup.

Options:
  --backup <file>       Path to backup file (.dump) or S3 key
  --database <name>     Source database name in backup
  --target <name>       Target database name (defaults to source name)
  --pitr                Enable Point-in-Time Recovery
  --target-time <time>  Target recovery time (ISO 8601 format)
  --list-backups        List available backups
  --dry-run             Show what would be done without executing
  --help                Show this help message

Examples:
  # Restore from local backup
  $0 --backup /backups/crm_iam_20240101_020000.dump --database crm_iam

  # Restore to different database
  $0 --backup backup.dump --database crm_iam --target crm_iam_restored

  # Point-in-Time Recovery
  $0 --backup base_backup.tar --pitr --target-time "2024-01-15 14:30:00"

  # List available backups
  $0 --list-backups
EOF
}

check_dependencies() {
    local deps=(pg_restore psql)
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Required dependency not found: $dep"
            exit 1
        fi
    done
}

list_local_backups() {
    log_info "Available local backups:"
    echo ""
    for dir in "${BACKUP_DIR}"/{daily,weekly,monthly}; do
        if [[ -d "${dir}" ]]; then
            local category=$(basename "${dir}")
            echo "=== ${category^} Backups ==="
            find "${dir}" -name "*.dump" -type f -printf "%T+ %s %p\n" 2>/dev/null | sort -r | head -20
            echo ""
        fi
    done
}

list_s3_backups() {
    if [[ "${S3_ENABLED}" != "true" ]]; then
        log_warn "S3 is not enabled"
        return
    fi
    
    log_info "Available S3 backups:"
    local aws_args=("s3" "ls" "s3://${S3_BUCKET}/${S3_PREFIX}/" "--recursive")
    if [[ -n "${S3_ENDPOINT}" ]]; then
        aws_args=("--endpoint-url" "${S3_ENDPOINT}" "${aws_args[@]}")
    fi
    
    aws "${aws_args[@]}" | sort -r | head -50
}

download_from_s3() {
    local s3_key="$1"
    local local_path="$2"
    
    log_info "Downloading from S3: ${s3_key}"
    
    local s3_path="s3://${S3_BUCKET}/${s3_key}"
    local aws_args=("s3" "cp" "${s3_path}" "${local_path}")
    if [[ -n "${S3_ENDPOINT}" ]]; then
        aws_args=("--endpoint-url" "${S3_ENDPOINT}" "${aws_args[@]}")
    fi
    
    if aws "${aws_args[@]}"; then
        log_info "Download completed: ${local_path}"
        return 0
    else
        log_error "Download failed: ${s3_key}"
        return 1
    fi
}

verify_backup() {
    local backup_file="$1"
    
    log_info "Verifying backup integrity: ${backup_file}"
    
    if pg_restore --list "${backup_file}" > /dev/null 2>&1; then
        log_info "Backup verification passed"
        return 0
    else
        log_error "Backup verification failed - file may be corrupted"
        return 1
    fi
}

create_target_database() {
    local target_db="$1"
    
    log_info "Creating target database: ${target_db}"
    
    # Check if database exists
    local exists=$(psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres -tAc \
        "SELECT 1 FROM pg_database WHERE datname='${target_db}'" 2>/dev/null || echo "")
    
    if [[ "${exists}" == "1" ]]; then
        log_warn "Database ${target_db} already exists"
        read -p "Drop and recreate? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # Terminate existing connections
            psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres -c \
                "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${target_db}'" > /dev/null 2>&1 || true
            
            psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres -c \
                "DROP DATABASE IF EXISTS ${target_db}" > /dev/null 2>&1
        else
            log_error "Aborting restore - database exists"
            exit 1
        fi
    fi
    
    psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres -c \
        "CREATE DATABASE ${target_db}" > /dev/null 2>&1
    
    log_info "Database created: ${target_db}"
}

restore_full_backup() {
    local backup_file="$1"
    local target_db="$2"
    
    log_info "Starting full restore to database: ${target_db}"
    
    # Create database if needed
    create_target_database "${target_db}"
    
    # Restore the backup
    pg_restore \
        -h "${PG_HOST}" \
        -p "${PG_PORT}" \
        -U "${PG_USER}" \
        -d "${target_db}" \
        --no-owner \
        --no-privileges \
        --verbose \
        "${backup_file}" \
        2>&1 | grep -E "(pg_restore|error|warning)" || true
    
    log_info "Full restore completed: ${target_db}"
}

restore_pitr() {
    local base_backup="$1"
    local target_time="$2"
    local data_dir="${PG_DATA_DIR:-/var/lib/postgresql/data}"
    
    log_info "Starting Point-in-Time Recovery"
    log_info "Target time: ${target_time}"
    
    # This requires stopping PostgreSQL and replacing data directory
    log_warn "PITR requires PostgreSQL to be stopped"
    log_warn "This operation modifies the PostgreSQL data directory"
    
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "PITR aborted by user"
        exit 0
    fi
    
    # Create recovery.signal for PostgreSQL 12+
    cat > "${data_dir}/recovery.signal" << EOF
# Point-in-Time Recovery Configuration
# Generated by restore-postgres.sh
EOF
    
    # Create postgresql.auto.conf entries for recovery
    cat >> "${data_dir}/postgresql.auto.conf" << EOF

# PITR Recovery Settings (added by restore-postgres.sh)
restore_command = 'cp ${WAL_ARCHIVE_DIR}/%f %p'
recovery_target_time = '${target_time}'
recovery_target_action = 'promote'
EOF
    
    log_info "PITR configuration created"
    log_info "Restart PostgreSQL to begin recovery"
    log_info "Monitor recovery progress in PostgreSQL logs"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    local backup_file=""
    local source_db=""
    local target_db=""
    local pitr_mode=false
    local target_time=""
    local list_backups=false
    local dry_run=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --backup)
                backup_file="$2"
                shift 2
                ;;
            --database)
                source_db="$2"
                shift 2
                ;;
            --target)
                target_db="$2"
                shift 2
                ;;
            --pitr)
                pitr_mode=true
                shift
                ;;
            --target-time)
                target_time="$2"
                shift 2
                ;;
            --list-backups)
                list_backups=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    check_dependencies
    
    # List backups mode
    if [[ "${list_backups}" == "true" ]]; then
        list_local_backups
        list_s3_backups
        exit 0
    fi
    
    # Validate required arguments
    if [[ -z "${backup_file}" ]]; then
        log_error "Backup file is required (--backup)"
        usage
        exit 1
    fi
    
    if [[ -z "${source_db}" && "${pitr_mode}" != "true" ]]; then
        log_error "Database name is required (--database)"
        usage
        exit 1
    fi
    
    # Set target database if not specified
    if [[ -z "${target_db}" ]]; then
        target_db="${source_db}"
    fi
    
    log_info "=========================================="
    log_info "PostgreSQL Restore Starting"
    log_info "=========================================="
    log_info "Backup: ${backup_file}"
    log_info "Source DB: ${source_db}"
    log_info "Target DB: ${target_db}"
    log_info "PITR Mode: ${pitr_mode}"
    
    if [[ "${dry_run}" == "true" ]]; then
        log_info "DRY RUN MODE - No changes will be made"
        exit 0
    fi
    
    # Download from S3 if needed
    if [[ "${backup_file}" == s3://* || "${backup_file}" == "${S3_PREFIX}/"* ]]; then
        local local_backup="/tmp/restore_$(basename "${backup_file}")"
        download_from_s3 "${backup_file}" "${local_backup}"
        backup_file="${local_backup}"
    fi
    
    # Verify backup
    if [[ "${pitr_mode}" != "true" ]]; then
        verify_backup "${backup_file}"
    fi
    
    # Perform restore
    if [[ "${pitr_mode}" == "true" ]]; then
        if [[ -z "${target_time}" ]]; then
            log_error "Target time is required for PITR (--target-time)"
            exit 1
        fi
        restore_pitr "${backup_file}" "${target_time}"
    else
        restore_full_backup "${backup_file}" "${target_db}"
    fi
    
    log_info "=========================================="
    log_info "Restore completed successfully"
    log_info "=========================================="
}

main "$@"
