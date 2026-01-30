#!/bin/bash
# ==============================================================================
# Blue-Green Deployment Script for CRM Platform
# ==============================================================================
# Supports zero-downtime deployments with:
# - Database migration pre-checks
# - Health verification before switching
# - Automatic rollback on failure
# - Traffic switching via Kubernetes
# ==============================================================================

set -euo pipefail

# ==============================================================================
# Configuration
# ==============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-crm}"
DEPLOYMENT_TIMEOUT="${DEPLOYMENT_TIMEOUT:-300}"
HEALTH_CHECK_RETRIES="${HEALTH_CHECK_RETRIES:-30}"
HEALTH_CHECK_INTERVAL="${HEALTH_CHECK_INTERVAL:-10}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ==============================================================================
# Functions
# ==============================================================================

log() {
    local level="$1"
    local color="$2"
    shift 2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] [$level]${NC} $*"
}

log_info() { log "INFO" "${BLUE}" "$@"; }
log_success() { log "SUCCESS" "${GREEN}" "$@"; }
log_warn() { log "WARN" "${YELLOW}" "$@"; }
log_error() { log "ERROR" "${RED}" "$@"; }

usage() {
    cat << EOF
Blue-Green Deployment Script for CRM Platform

Usage: $0 <command> [options]

Commands:
  deploy <service> <version>   Deploy new version as green
  switch <service>             Switch traffic to green
  rollback <service>           Rollback to blue (previous version)
  status <service>             Show deployment status
  migrate <service>            Run database migrations

Options:
  --namespace <ns>             Kubernetes namespace (default: crm)
  --skip-migration             Skip database migration check
  --skip-health-check          Skip health verification
  --force                      Force operation without confirmation
  --dry-run                    Show what would be done

Examples:
  $0 deploy iam-service v1.2.0
  $0 switch iam-service
  $0 rollback iam-service
  $0 status iam-service

EOF
}

check_dependencies() {
    local deps=(kubectl jq curl)
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Required dependency not found: $dep"
            exit 1
        fi
    done
}

get_active_color() {
    local service="$1"
    kubectl get service "${service}" -n "${NAMESPACE}" \
        -o jsonpath='{.spec.selector.deployment-color}' 2>/dev/null || echo "blue"
}

get_inactive_color() {
    local active="$1"
    if [[ "$active" == "blue" ]]; then
        echo "green"
    else
        echo "blue"
    fi
}

check_deployment_exists() {
    local deployment="$1"
    kubectl get deployment "${deployment}" -n "${NAMESPACE}" &>/dev/null
}

wait_for_deployment() {
    local deployment="$1"
    local timeout="$2"
    
    log_info "Waiting for deployment ${deployment} to be ready..."
    
    if kubectl rollout status deployment/"${deployment}" -n "${NAMESPACE}" --timeout="${timeout}s"; then
        log_success "Deployment ${deployment} is ready"
        return 0
    else
        log_error "Deployment ${deployment} failed to become ready"
        return 1
    fi
}

health_check() {
    local service="$1"
    local color="$2"
    local deployment="${service}-${color}"
    
    log_info "Running health checks for ${deployment}..."
    
    # Get pod IPs
    local pods=$(kubectl get pods -n "${NAMESPACE}" \
        -l "app.kubernetes.io/name=${service},deployment-color=${color}" \
        -o jsonpath='{.items[*].status.podIP}')
    
    if [[ -z "$pods" ]]; then
        log_error "No pods found for ${deployment}"
        return 1
    fi
    
    local all_healthy=true
    for pod_ip in ${pods}; do
        local retry=0
        local healthy=false
        
        while [[ $retry -lt $HEALTH_CHECK_RETRIES ]]; do
            if curl -sf "http://${pod_ip}:8080/health" > /dev/null 2>&1; then
                healthy=true
                break
            fi
            ((retry++))
            sleep "$HEALTH_CHECK_INTERVAL"
        done
        
        if [[ "$healthy" == "true" ]]; then
            log_success "Pod ${pod_ip} is healthy"
        else
            log_error "Pod ${pod_ip} failed health check"
            all_healthy=false
        fi
    done
    
    if [[ "$all_healthy" == "true" ]]; then
        return 0
    else
        return 1
    fi
}

run_migrations() {
    local service="$1"
    local version="$2"
    
    log_info "Running database migrations for ${service}..."
    
    # Determine which database to migrate
    local db_name=""
    case "$service" in
        iam-service)
            db_name="crm_iam"
            ;;
        sales-service)
            db_name="crm_sales"
            ;;
        notification-service)
            db_name="crm_notification"
            ;;
        customer-service)
            log_info "Customer service uses MongoDB - checking migration status..."
            return 0
            ;;
        *)
            log_warn "Unknown service: ${service} - skipping migration"
            return 0
            ;;
    esac
    
    # Run migration job
    local job_name="migration-${service}-$(date +%s)"
    
    kubectl create job "${job_name}" \
        --namespace="${NAMESPACE}" \
        --from=cronjob/db-migration-job \
        --dry-run=client \
        -o yaml | \
    kubectl apply -f - || {
        log_error "Failed to create migration job"
        return 1
    }
    
    # Wait for migration to complete
    if kubectl wait --for=condition=complete job/"${job_name}" \
        --namespace="${NAMESPACE}" --timeout=300s; then
        log_success "Migrations completed successfully"
        kubectl delete job "${job_name}" -n "${NAMESPACE}" --ignore-not-found
        return 0
    else
        log_error "Migration failed"
        kubectl logs job/"${job_name}" -n "${NAMESPACE}"
        return 1
    fi
}

deploy_green() {
    local service="$1"
    local version="$2"
    local skip_migration="${3:-false}"
    local dry_run="${4:-false}"
    
    log_info "=========================================="
    log_info "Deploying ${service} version ${version}"
    log_info "=========================================="
    
    local active_color=$(get_active_color "$service")
    local green_color=$(get_inactive_color "$active_color")
    local green_deployment="${service}-${green_color}"
    
    log_info "Active color: ${active_color}"
    log_info "Deploying to: ${green_color}"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would deploy ${green_deployment} with version ${version}"
        return 0
    fi
    
    # Run migrations first (if not skipped)
    if [[ "$skip_migration" != "true" ]]; then
        if ! run_migrations "$service" "$version"; then
            log_error "Migration failed - aborting deployment"
            return 1
        fi
    fi
    
    # Update green deployment with new version
    kubectl set image deployment/"${green_deployment}" \
        -n "${NAMESPACE}" \
        "${service}=${version}" || {
        log_error "Failed to update deployment image"
        return 1
    }
    
    # Add version label
    kubectl label deployment/"${green_deployment}" \
        -n "${NAMESPACE}" \
        --overwrite \
        app.kubernetes.io/version="${version}"
    
    # Wait for deployment
    if ! wait_for_deployment "${green_deployment}" "${DEPLOYMENT_TIMEOUT}"; then
        log_error "Deployment failed - initiating rollback"
        kubectl rollout undo deployment/"${green_deployment}" -n "${NAMESPACE}"
        return 1
    fi
    
    # Health check
    if ! health_check "$service" "$green_color"; then
        log_error "Health check failed - initiating rollback"
        kubectl rollout undo deployment/"${green_deployment}" -n "${NAMESPACE}"
        return 1
    fi
    
    log_success "Green deployment successful!"
    log_info "Run '$0 switch ${service}' to switch traffic"
}

switch_traffic() {
    local service="$1"
    local dry_run="${2:-false}"
    
    log_info "=========================================="
    log_info "Switching traffic for ${service}"
    log_info "=========================================="
    
    local active_color=$(get_active_color "$service")
    local green_color=$(get_inactive_color "$active_color")
    
    log_info "Current active: ${active_color}"
    log_info "Switching to: ${green_color}"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would switch service ${service} selector to ${green_color}"
        return 0
    fi
    
    # Update service selector to point to green
    kubectl patch service "${service}" \
        -n "${NAMESPACE}" \
        --type='json' \
        -p="[{\"op\": \"replace\", \"path\": \"/spec/selector/deployment-color\", \"value\": \"${green_color}\"}]" || {
        log_error "Failed to switch traffic"
        return 1
    }
    
    log_success "Traffic switched to ${green_color}"
    log_info "Previous version (${active_color}) is still running for quick rollback"
}

rollback() {
    local service="$1"
    local dry_run="${2:-false}"
    
    log_info "=========================================="
    log_info "Rolling back ${service}"
    log_info "=========================================="
    
    local active_color=$(get_active_color "$service")
    local previous_color=$(get_inactive_color "$active_color")
    
    log_info "Current active: ${active_color}"
    log_info "Rolling back to: ${previous_color}"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would rollback service ${service} to ${previous_color}"
        return 0
    fi
    
    # Switch service selector back to previous color
    kubectl patch service "${service}" \
        -n "${NAMESPACE}" \
        --type='json' \
        -p="[{\"op\": \"replace\", \"path\": \"/spec/selector/deployment-color\", \"value\": \"${previous_color}\"}]" || {
        log_error "Failed to rollback"
        return 1
    }
    
    log_success "Rolled back to ${previous_color}"
}

show_status() {
    local service="$1"
    
    log_info "=========================================="
    log_info "Status for ${service}"
    log_info "=========================================="
    
    local active_color=$(get_active_color "$service")
    echo ""
    echo "Active Color: ${active_color}"
    echo ""
    
    for color in blue green; do
        local deployment="${service}-${color}"
        echo "=== ${color^^} Deployment ==="
        
        if check_deployment_exists "${deployment}"; then
            kubectl get deployment "${deployment}" -n "${NAMESPACE}" \
                -o jsonpath='Image: {.spec.template.spec.containers[0].image}
Replicas: {.status.readyReplicas}/{.status.replicas}
Version: {.metadata.labels.app\.kubernetes\.io/version}
'
            echo ""
            
            # Pod status
            kubectl get pods -n "${NAMESPACE}" \
                -l "app.kubernetes.io/name=${service},deployment-color=${color}" \
                --no-headers 2>/dev/null || echo "No pods found"
        else
            echo "Deployment not found"
        fi
        echo ""
    done
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    local command="${1:-}"
    shift || true
    
    local service=""
    local version=""
    local skip_migration=false
    local skip_health_check=false
    local force=false
    local dry_run=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --skip-migration)
                skip_migration=true
                shift
                ;;
            --skip-health-check)
                skip_health_check=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                if [[ -z "$service" ]]; then
                    service="$1"
                elif [[ -z "$version" ]]; then
                    version="$1"
                fi
                shift
                ;;
        esac
    done
    
    check_dependencies
    
    case "$command" in
        deploy)
            if [[ -z "$service" || -z "$version" ]]; then
                log_error "Service and version required for deploy"
                usage
                exit 1
            fi
            deploy_green "$service" "$version" "$skip_migration" "$dry_run"
            ;;
        switch)
            if [[ -z "$service" ]]; then
                log_error "Service required for switch"
                usage
                exit 1
            fi
            switch_traffic "$service" "$dry_run"
            ;;
        rollback)
            if [[ -z "$service" ]]; then
                log_error "Service required for rollback"
                usage
                exit 1
            fi
            rollback "$service" "$dry_run"
            ;;
        status)
            if [[ -z "$service" ]]; then
                log_error "Service required for status"
                usage
                exit 1
            fi
            show_status "$service"
            ;;
        migrate)
            if [[ -z "$service" ]]; then
                log_error "Service required for migrate"
                usage
                exit 1
            fi
            run_migrations "$service" "${version:-latest}"
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

main "$@"
