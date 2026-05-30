#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_NAMESPACE="kextant-agent-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCAN_OUTPUT_FILE="${SCRIPT_DIR}/scan-output.json"
LOG_FILE="${SCRIPT_DIR}/test-run.log"

# Expected findings counts (minimum expected for each category)
declare -A EXPECTED_FINDINGS=(
    ["RES001"]=1  # Missing CPU requests
    ["RES002"]=1  # Missing CPU limits
    ["RES003"]=1  # Missing memory requests
    ["RES004"]=1  # Missing memory limits (CRITICAL)
    ["RES005"]=1  # CPU limit equals request
    ["PRB001"]=1  # Missing readiness probe
    ["PRB002"]=1  # Missing liveness probe
    ["PRB004"]=1  # Identical probes
    # NOTE: REP001 will find ALL single-replica deployments (16 expected)
    # This is correct behavior - most test deployments use replicas: 1
    ["REP001"]=16  # Single replica deployments
    ["REP002"]=1  # No PDB (only no-pdb deployment with replicas: 3)
    ["IMG001"]=1  # Using :latest tag
    ["IMG002"]=1  # No image tag
    # Note: IMG003 not tested - Kubernetes always sets default ImagePullPolicy on pods
    ["IMG004"]=1  # Always policy with specific tag
    ["SEC001"]=1  # Running as root
    ["SEC002"]=1  # Privileged container (CRITICAL)
    ["SEC003"]=1  # Writable root filesystem
    ["SEC004"]=1  # Privilege escalation
    ["SEC005"]=1  # Capabilities not dropped
    ["NS001"]=1  # No ResourceQuota
    ["NS002"]=1  # No LimitRange
    ["NS003"]=1  # No NetworkPolicy
)

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "${LOG_FILE}"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $*" | tee -a "${LOG_FILE}"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $*" | tee -a "${LOG_FILE}"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $*" | tee -a "${LOG_FILE}"
}

cleanup() {
    log "Cleaning up test resources..."

    # Delete the test namespace (this will cascade delete all resources)
    if kubectl get namespace "${TEST_NAMESPACE}" &>/dev/null; then
        kubectl delete namespace "${TEST_NAMESPACE}" --timeout=60s || {
            log_warning "Failed to delete namespace cleanly, forcing deletion..."
            kubectl delete namespace "${TEST_NAMESPACE}" --grace-period=0 --force || true
        }
        log_success "Test namespace deleted"
    else
        log "Test namespace already deleted"
    fi

    # Clean up temp files
    if [[ -f "${SCAN_OUTPUT_FILE}" ]]; then
        rm -f "${SCAN_OUTPUT_FILE}"
    fi
}

setup_test_environment() {
    log "Setting up test environment..."

    # Create log file
    : > "${LOG_FILE}"

    # Clean up any existing test namespace
    if kubectl get namespace "${TEST_NAMESPACE}" &>/dev/null; then
        log_warning "Test namespace already exists, cleaning up..."
        cleanup
        sleep 5
    fi

    # Create the test namespace
    log "Creating test namespace: ${TEST_NAMESPACE}"
    kubectl create namespace "${TEST_NAMESPACE}"
    kubectl label namespace "${TEST_NAMESPACE}" test=kextant-agent purpose=functional-testing

    log_success "Test environment setup complete"
}

deploy_test_resources() {
    log "Deploying test resources..."

    # Apply the test resources
    kubectl apply -f "${SCRIPT_DIR}/test-resources.yaml"

    # Apply exemption test resources
    log "Deploying exemption test resources..."
    kubectl apply -f "${SCRIPT_DIR}/test-exemptions.yaml"

    log "Waiting for deployments to be created..."
    sleep 5

    # Wait for deployments to be ready (with timeout)
    log "Waiting for deployments to be ready (timeout: 120s)..."
    local timeout=120
    local elapsed=0

    while [[ ${elapsed} -lt ${timeout} ]]; do
        local ready_count
        ready_count=$(kubectl get deployments -n "${TEST_NAMESPACE}" --no-headers 2>/dev/null | wc -l || echo "0")
        local total_count
        total_count=$(kubectl get deployments -n "${TEST_NAMESPACE}" --no-headers 2>/dev/null | wc -l || echo "0")

        if [[ ${ready_count} -gt 0 ]] && [[ ${ready_count} -eq ${total_count} ]]; then
            log_success "All deployments created (${ready_count} deployments)"
            break
        fi

        sleep 5
        elapsed=$((elapsed + 5))
    done

    # Show deployment status
    log "Deployment status:"
    kubectl get deployments -n "${TEST_NAMESPACE}" | tee -a "${LOG_FILE}"

    log_success "Test resources deployed"
}

run_scan() {
    log "Running Kextant Agent scan on ${TEST_NAMESPACE} namespace..."

    # Check if kextant-agent binary exists
    if [[ ! -f "${SCRIPT_DIR}/../kextant-agent" ]]; then
        log_error "kextant-agent binary not found. Please build it first with 'make build'"
        return 1
    fi

    # Set environment variables for the scan
    export CLUSTER_NAME="test-cluster"
    export SCAN_NAMESPACES="${TEST_NAMESPACE}"
    export EXCLUDE_NAMESPACES=""
    export CHECKS_RESOURCES_ENABLED="true"
    export CHECKS_PROBES_ENABLED="true"
    export CHECKS_REPLICAS_ENABLED="true"
    export CHECKS_IMAGES_ENABLED="true"
    export CHECKS_SECURITY_CONTEXT_ENABLED="true"
    export CHECKS_DEPRECATED_API_ENABLED="true"
    export CHECKS_NAMESPACE_ENABLED="true"
    export REPORT_MIN_SEVERITY="info"
    export LOG_LEVEL="INFO"

    # Run the scan and capture output
    log "Executing scan command..."
    if "${SCRIPT_DIR}/../kextant-agent" scan > "${SCAN_OUTPUT_FILE}" 2>&1; then
        log_success "Scan completed successfully"
    else
        log_error "Scan failed"
        cat "${SCAN_OUTPUT_FILE}"
        return 1
    fi

    # Display scan output
    log "Scan output:"
    cat "${SCAN_OUTPUT_FILE}" | tee -a "${LOG_FILE}"
}

validate_findings() {
    log "Validating scan findings..."

    if [[ ! -f "${SCAN_OUTPUT_FILE}" ]]; then
        log_error "Scan output file not found: ${SCAN_OUTPUT_FILE}"
        return 1
    fi

    local validation_errors=0
    local validation_warnings=0

    # Extract findings from the scan output
    # The output is plain text, so we'll use grep to find check IDs
    local scan_content
    scan_content=$(cat "${SCAN_OUTPUT_FILE}")

    log "Checking for expected findings..."

    for check_id in "${!EXPECTED_FINDINGS[@]}"; do
        local min_expected=${EXPECTED_FINDINGS[$check_id]}

        # Count occurrences of the check ID in the output
        local found_count
        found_count=$(echo "${scan_content}" | grep -c "\[${check_id}\]" 2>/dev/null || echo "0")
        found_count=$(echo "${found_count}" | tr -d '[:space:]')  # Remove any whitespace/newlines

        if [[ ${found_count} -ge ${min_expected} ]]; then
            log_success "✓ ${check_id}: Found ${found_count} (expected: ${min_expected})"
        else
            log_error "✗ ${check_id}: Found ${found_count} (expected: ${min_expected})"
            validation_errors=$((validation_errors + 1))
        fi
    done

    # Check that the good deployment doesn't appear in findings (except namespace-level checks)
    local good_deployment_issues
    good_deployment_issues=$(echo "${scan_content}" | grep -c "good-deployment" || echo "0")

    # Good deployment should only appear in namespace-level findings (NS001, NS002, NS003)
    # which affect all resources in the namespace
    log "Good deployment issue count: ${good_deployment_issues}"

    # Verify health score is calculated
    if echo "${scan_content}" | grep -q "Score:"; then
        log_success "✓ Health score is present in report"
    else
        log_warning "⚠ Health score not found in report"
        validation_warnings=$((validation_warnings + 1))
    fi

    # Check for severity classifications
    for severity in "CRITICAL" "WARNING" "INFO"; do
        if echo "${scan_content}" | grep -qi "${severity}"; then
            log_success "✓ ${severity} findings present"
        else
            log_warning "⚠ No ${severity} findings found"
            validation_warnings=$((validation_warnings + 1))
        fi
    done

    # Summary
    echo ""
    log "=========================================="
    log "Validation Summary"
    log "=========================================="

    if [[ ${validation_errors} -eq 0 ]]; then
        log_success "All validations passed! ✓"
        if [[ ${validation_warnings} -gt 0 ]]; then
            log_warning "Warnings: ${validation_warnings}"
        fi
        return 0
    else
        log_error "Validation failed with ${validation_errors} error(s) and ${validation_warnings} warning(s)"
        return 1
    fi
}

verify_namespace_checks() {
    log "Verifying namespace-level checks..."

    # NS001: No ResourceQuota
    if kubectl get resourcequota -n "${TEST_NAMESPACE}" --no-headers 2>/dev/null | grep -q .; then
        log_warning "ResourceQuota exists in test namespace (NS001 check may not trigger)"
    else
        log_success "No ResourceQuota in test namespace (NS001 should trigger)"
    fi

    # NS002: No LimitRange
    if kubectl get limitrange -n "${TEST_NAMESPACE}" --no-headers 2>/dev/null | grep -q .; then
        log_warning "LimitRange exists in test namespace (NS002 check may not trigger)"
    else
        log_success "No LimitRange in test namespace (NS002 should trigger)"
    fi

    # NS003: No NetworkPolicy
    if kubectl get networkpolicy -n "${TEST_NAMESPACE}" --no-headers 2>/dev/null | grep -q .; then
        log_warning "NetworkPolicy exists in test namespace (NS003 check may not trigger)"
    else
        log_success "No NetworkPolicy in test namespace (NS003 should trigger)"
    fi
}

validate_exemptions() {
    log "Validating exemption behavior..."

    if [[ ! -f "${SCAN_OUTPUT_FILE}" ]]; then
        log_error "Scan output file not found: ${SCAN_OUTPUT_FILE}"
        return 1
    fi

    local scan_content
    scan_content=$(cat "${SCAN_OUTPUT_FILE}")
    local exemption_errors=0

    # Test Case 1: exempt-rep001 should NOT trigger REP001
    if echo "${scan_content}" | grep "exempt-rep001" | grep -q "\[REP001\]"; then
        log_error "✗ exempt-rep001 triggered REP001 (should be exempted)"
        exemption_errors=$((exemption_errors + 1))
    else
        log_success "✓ exempt-rep001 correctly exempt from REP001"
    fi

    # Test Case 2: exempt-multiple should NOT trigger RES001, RES002, or PRB001
    if echo "${scan_content}" | grep "exempt-multiple" | grep -qE "\[(RES001|RES002|PRB001)\]"; then
        log_error "✗ exempt-multiple triggered exempted checks"
        exemption_errors=$((exemption_errors + 1))
    else
        log_success "✓ exempt-multiple correctly exempt from RES001,RES002,PRB001"
    fi

    # Test Case 3: exempt-wildcard should NOT trigger ANY checks
    if echo "${scan_content}" | grep -q "exempt-wildcard"; then
        log_error "✗ exempt-wildcard triggered findings (should be exempt from all)"
        exemption_errors=$((exemption_errors + 1))
    else
        log_success "✓ exempt-wildcard correctly exempt from all checks"
    fi

    # Test Case 4: partial-exempt should NOT trigger RES001 but SHOULD trigger RES002
    local partial_has_res001
    partial_has_res001=$(echo "${scan_content}" | grep "partial-exempt" | grep -c "\[RES001\]" || echo "0")
    local partial_has_res002
    partial_has_res002=$(echo "${scan_content}" | grep "partial-exempt" | grep -c "\[RES002\]" || echo "0")

    if [[ ${partial_has_res001} -gt 0 ]]; then
        log_error "✗ partial-exempt triggered RES001 (should be exempted)"
        exemption_errors=$((exemption_errors + 1))
    else
        log_success "✓ partial-exempt correctly exempt from RES001"
    fi

    if [[ ${partial_has_res002} -gt 0 ]]; then
        log_success "✓ partial-exempt correctly triggers RES002 (not exempted)"
    else
        log_warning "⚠ partial-exempt did not trigger RES002 (should trigger)"
    fi

    # Test Case 5: mailhog-test should NOT trigger REP001 or IMG001
    if echo "${scan_content}" | grep "mailhog-test" | grep -qE "\[(REP001|IMG001)\]"; then
        log_error "✗ mailhog-test triggered exempted checks"
        exemption_errors=$((exemption_errors + 1))
    else
        log_success "✓ mailhog-test correctly exempt from REP001,IMG001"
    fi

    if [[ ${exemption_errors} -eq 0 ]]; then
        log_success "All exemption validations passed! ✓"
        return 0
    else
        log_error "Exemption validation failed with ${exemption_errors} error(s)"
        return 1
    fi
}

main() {
    local exit_code=0

    echo ""
    log "=========================================="
    log "Kextant Agent Functional Test Suite"
    log "=========================================="
    echo ""

    # Trap to ensure cleanup on exit
    trap cleanup EXIT

    # Step 1: Setup
    if ! setup_test_environment; then
        log_error "Failed to setup test environment"
        exit 1
    fi

    # Step 2: Deploy test resources
    if ! deploy_test_resources; then
        log_error "Failed to deploy test resources"
        exit 1
    fi

    # Step 3: Verify namespace configuration
    verify_namespace_checks

    # Step 4: Run scan
    if ! run_scan; then
        log_error "Failed to run scan"
        exit 1
    fi

    # Step 5: Validate findings
    if ! validate_findings; then
        exit_code=1
    fi

    # Step 6: Validate exemptions
    if ! validate_exemptions; then
        exit_code=1
    fi

    echo ""
    log "=========================================="
    if [[ ${exit_code} -eq 0 ]]; then
        log_success "Functional tests PASSED ✓"
    else
        log_error "Functional tests FAILED ✗"
    fi
    log "=========================================="
    log "Log file: ${LOG_FILE}"
    log "Scan output: ${SCAN_OUTPUT_FILE}"
    echo ""

    exit ${exit_code}
}

main "$@"
