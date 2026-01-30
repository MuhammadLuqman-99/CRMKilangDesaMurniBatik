#!/bin/bash
# ==============================================================================
# PostgreSQL Backup Script for CRM Kilang Desa Murni Batik
# ==============================================================================
# This script performs automated backups of PostgreSQL databases with:
# - Compressed custom format dumps
# - S3/MinIO upload support
# - Backup verification
# - Retention policy enforcement
# ==============================================================================

set -euo pipefail

# ==============================================================================
# Configuration
# ==============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/tmp/pg-backups}"
DATE_FORMAT=$(date +%Y%m%d_%H%M%S)
RETENTION_DAILY=${RETENTION_DAILY:-7}
RETENTION_WEEKLY=${RETENTION_WEEKLY:-4}
RETENTION_MONTHLY=${RETENTION_MONTHLY:-12}

# Database connection
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
PG_USER="${PG_USER:-crm_admin}"
PG_PASSWORD="${PG_PASSWORD:-}"
PGPASSWORD="${PG_PASSWORD}"
export PGPASSWORD

# Databases to backup
DATABASES="${DATABASES:-crm_iam crm_sales crm_notification}"

# S3/MinIO settings
S3_ENABLED="${S3_ENABLED:-false}"
S3_BUCKET="${S3_BUCKET:-crm-backups}"
S3_PREFIX="${S3_PREFIX:-postgres}"
S3_ENDPOINT="${S3_ENDPOINT:-}"

# Notification settings
NOTIFICATION_ENABLED="${NOTIFICATION_ENABLED:-false}"
NOTIFICATION_WEBHOOK="${NOTIFICATION_WEBHOOK:-}"

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

send_notification() {
    local status="$1"
    local message="$2"
    
    if [[ "${NOTIFICATION_ENABLED}" == "true" && -n "${NOTIFICATION_WEBHOOK}" ]]; then
        curl -s -X POST "${NOTIFICATION_WEBHOOK}" \
            -H "Content-Type: application/json" \
            -d "{\"status\": \"${status}\", \"message\": \"${message}\", \"timestamp\": \"$(date -Iseconds)\"}" \
            || log_warn "Failed to send notification"
    fi
}

check_dependencies() {
    local deps=(pg_dump pg_restore)
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Required dependency not found: $dep"
            exit 1
        fi
    done
    
    if [[ "${S3_ENABLED}" == "true" ]]; then
        if ! command -v aws &> /dev/null; then
            log_error "AWS CLI required for S3 uploads but not found"
            exit 1
        fi
    fi
}

create_backup_dir() {
    mkdir -p "${BACKUP_DIR}"/{daily,weekly,monthly}
    log_info "Backup directory ensured: ${BACKUP_DIR}"
}

backup_database() {
    local db_name="$1"
    local backup_file="${BACKUP_DIR}/daily/${db_name}_${DATE_FORMAT}.dump"
    
    log_info "Starting backup of database: ${db_name}"
    
    # Create compressed custom format dump
    pg_dump \
        -h "${PG_HOST}" \
        -p "${PG_PORT}" \
        -U "${PG_USER}" \
        -d "${db_name}" \
        -Fc \
        --no-password \
        --verbose \
        -f "${backup_file}" \
        2>&1 | grep -v "^pg_dump:" || true
    
    if [[ -f "${backup_file}" ]]; then
        local size=$(du -h "${backup_file}" | cut -f1)
        log_info "Backup completed: ${backup_file} (${size})"
        echo "${backup_file}"
    else
        log_error "Backup failed for database: ${db_name}"
        return 1
    fi
}

verify_backup() {
    local backup_file="$1"
    
    log_info "Verifying backup: ${backup_file}"
    
    # List contents to verify integrity
    if pg_restore --list "${backup_file}" > /dev/null 2>&1; then
        log_info "Backup verification passed: ${backup_file}"
        return 0
    else
        log_error "Backup verification failed: ${backup_file}"
        return 1
    fi
}

upload_to_s3() {
    local backup_file="$1"
    local filename=$(basename "${backup_file}")
    local s3_path="s3://${S3_BUCKET}/${S3_PREFIX}/${filename}"
    
    log_info "Uploading to S3: ${s3_path}"
    
    local aws_args=("s3" "cp" "${backup_file}" "${s3_path}")
    if [[ -n "${S3_ENDPOINT}" ]]; then
        aws_args=("--endpoint-url" "${S3_ENDPOINT}" "${aws_args[@]}")
    fi
    
    if aws "${aws_args[@]}"; then
        log_info "Upload completed: ${s3_path}"
        return 0
    else
        log_error "Upload failed: ${backup_file}"
        return 1
    fi
}

rotate_backups() {
    log_info "Rotating backups with retention policy"
    
    # Daily backups - keep last N days
    find "${BACKUP_DIR}/daily" -name "*.dump" -mtime +${RETENTION_DAILY} -delete 2>/dev/null || true
    
    # Weekly backups (Sunday) - create symlink or copy
    local day_of_week=$(date +%u)
    if [[ "${day_of_week}" == "7" ]]; then
        for file in "${BACKUP_DIR}"/daily/*_${DATE_FORMAT%%_*}*.dump; do
            if [[ -f "${file}" ]]; then
                cp "${file}" "${BACKUP_DIR}/weekly/"
                log_info "Created weekly backup: ${file}"
            fi
        done
    fi
    
    # Weekly retention
    find "${BACKUP_DIR}/weekly" -name "*.dump" -mtime +$((RETENTION_WEEKLY * 7)) -delete 2>/dev/null || true
    
    # Monthly backups (1st of month)
    local day_of_month=$(date +%d)
    if [[ "${day_of_month}" == "01" ]]; then
        for file in "${BACKUP_DIR}"/daily/*_${DATE_FORMAT%%_*}*.dump; do
            if [[ -f "${file}" ]]; then
                cp "${file}" "${BACKUP_DIR}/monthly/"
                log_info "Created monthly backup: ${file}"
            fi
        done
    fi
    
    # Monthly retention
    find "${BACKUP_DIR}/monthly" -name "*.dump" -mtime +$((RETENTION_MONTHLY * 30)) -delete 2>/dev/null || true
    
    log_info "Backup rotation completed"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    local dry_run=false
    local single_db=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run)
                dry_run=true
                shift
                ;;
            --database)
                single_db="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [--dry-run] [--database <name>]"
                echo ""
                echo "Options:"
                echo "  --dry-run     Show what would be done without executing"
                echo "  --database    Backup only specified database"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    log_info "=========================================="
    log_info "PostgreSQL Backup Starting"
    log_info "=========================================="
    
    if [[ "${dry_run}" == "true" ]]; then
        log_info "DRY RUN MODE - No changes will be made"
    fi
    
    check_dependencies
    create_backup_dir
    
    local dbs_to_backup
    if [[ -n "${single_db}" ]]; then
        dbs_to_backup="${single_db}"
    else
        dbs_to_backup="${DATABASES}"
    fi
    
    local success_count=0
    local fail_count=0
    local backup_files=()
    
    for db in ${dbs_to_backup}; do
        if [[ "${dry_run}" == "true" ]]; then
            log_info "[DRY RUN] Would backup database: ${db}"
            continue
        fi
        
        if backup_file=$(backup_database "${db}"); then
            if verify_backup "${backup_file}"; then
                backup_files+=("${backup_file}")
                
                if [[ "${S3_ENABLED}" == "true" ]]; then
                    if upload_to_s3 "${backup_file}"; then
                        ((success_count++))
                    else
                        ((fail_count++))
                    fi
                else
                    ((success_count++))
                fi
            else
                ((fail_count++))
            fi
        else
            ((fail_count++))
        fi
    done
    
    if [[ "${dry_run}" != "true" ]]; then
        rotate_backups
    fi
    
    log_info "=========================================="
    log_info "Backup Summary"
    log_info "=========================================="
    log_info "Successful: ${success_count}"
    log_info "Failed: ${fail_count}"
    
    if [[ ${fail_count} -gt 0 ]]; then
        send_notification "error" "PostgreSQL backup completed with ${fail_count} failures"
        exit 1
    else
        send_notification "success" "PostgreSQL backup completed successfully (${success_count} databases)"
        exit 0
    fi
}

main "$@"
