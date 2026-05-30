#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCAN_OUTPUT_FILE="${1:-scan-output.json}"

if [[ ! -f "${SCAN_OUTPUT_FILE}" ]]; then
    echo -e "${RED}Error: Scan output file not found: ${SCAN_OUTPUT_FILE}${NC}"
    echo "Usage: $0 <scan-output-file>"
    exit 1
fi

echo -e "${BLUE}=========================================="
echo "Kextant Agent Scan Results Validator"
echo -e "==========================================${NC}"
echo ""
echo "Analyzing: ${SCAN_OUTPUT_FILE}"
echo ""

# Read the scan output
SCAN_CONTENT=$(cat "${SCAN_OUTPUT_FILE}")

# Function to extract information from scan output
extract_info() {
    local pattern="$1"
    echo "${SCAN_CONTENT}" | grep -oP "${pattern}" | head -1 || echo "Not found"
}

count_occurrences() {
    local pattern="$1"
    echo "${SCAN_CONTENT}" | grep -c "${pattern}" || echo "0"
}

# Extract basic information
echo -e "${CYAN}=== Basic Information ===${NC}"
CLUSTER_NAME=$(extract_info "Cluster:\s+\K.*")
HEALTH_SCORE=$(extract_info "Health Score:\s+\K\d+")
GRADE=$(extract_info "Grade:\s+\K\w+")

echo "Cluster: ${CLUSTER_NAME}"
echo "Health Score: ${HEALTH_SCORE}"
echo "Grade: ${GRADE}"
echo ""

# Count findings by severity
echo -e "${CYAN}=== Findings by Severity ===${NC}"

CRITICAL_COUNT=$(count_occurrences "CRITICAL")
WARNING_COUNT=$(count_occurrences "WARNING")
INFO_COUNT=$(count_occurrences "INFO")

if [[ ${CRITICAL_COUNT} -gt 0 ]]; then
    echo -e "${RED}Critical: ${CRITICAL_COUNT}${NC}"
else
    echo -e "Critical: ${CRITICAL_COUNT}"
fi

if [[ ${WARNING_COUNT} -gt 0 ]]; then
    echo -e "${YELLOW}Warning: ${WARNING_COUNT}${NC}"
else
    echo -e "Warning: ${WARNING_COUNT}"
fi

echo -e "Info: ${INFO_COUNT}"
echo ""

# Check for each category
echo -e "${CYAN}=== Findings by Category ===${NC}"

declare -A CATEGORIES=(
    ["RES"]="Resource Configuration"
    ["PRB"]="Health Probes"
    ["REP"]="Replica Configuration"
    ["IMG"]="Image Configuration"
    ["SEC"]="Security Context"
    ["API"]="Deprecated APIs"
    ["NS"]="Namespace Health"
)

for category_id in "${!CATEGORIES[@]}"; do
    local category_name="${CATEGORIES[$category_id]}"
    local count=$(count_occurrences "\[${category_id}[0-9]{3}\]")

    if [[ ${count} -gt 0 ]]; then
        echo -e "${GREEN}✓${NC} ${category_name} (${category_id}xxx): ${count} finding(s)"
    else
        echo -e "${YELLOW}⚠${NC} ${category_name} (${category_id}xxx): No findings"
    fi
done
echo ""

# Detailed check breakdown
echo -e "${CYAN}=== Detailed Check Breakdown ===${NC}"

declare -A CHECK_DETAILS=(
    # Resource Configuration
    ["RES001"]="Missing CPU requests|Warning"
    ["RES002"]="Missing CPU limits|Warning"
    ["RES003"]="Missing memory requests|Warning"
    ["RES004"]="Missing memory limits|Critical"
    ["RES005"]="CPU limit equals request|Info"

    # Health Probes
    ["PRB001"]="Missing readiness probe|Warning"
    ["PRB002"]="Missing liveness probe|Warning"
    ["PRB003"]="Missing startup probe|Info"
    ["PRB004"]="Identical liveness/readiness probes|Info"

    # Replica Configuration
    ["REP001"]="Single replica deployment|Warning"
    ["REP002"]="No PodDisruptionBudget|Warning"
    ["REP003"]="Mismatched replicas|Warning"

    # Image Configuration
    ["IMG001"]="Using :latest tag|Warning"
    ["IMG002"]="No image tag|Warning"
    ["IMG003"]="ImagePullPolicy not set|Info"
    ["IMG004"]="Always policy with specific tag|Info"

    # Security Context
    ["SEC001"]="Running as root|Warning"
    ["SEC002"]="Privileged container|Critical"
    ["SEC003"]="Writable root filesystem|Warning"
    ["SEC004"]="Privilege escalation allowed|Warning"
    ["SEC005"]="Capabilities not dropped|Info"

    # Namespace Health
    ["NS001"]="No ResourceQuota|Info"
    ["NS002"]="No LimitRange|Info"
    ["NS003"]="No NetworkPolicy|Warning"
)

# Group by severity for display
echo ""
echo -e "${RED}Critical Checks:${NC}"
for check_id in "${!CHECK_DETAILS[@]}"; do
    IFS='|' read -r description severity <<< "${CHECK_DETAILS[$check_id]}"
    if [[ "${severity}" == "Critical" ]]; then
        local count=$(count_occurrences "\[${check_id}\]")
        if [[ ${count} -gt 0 ]]; then
            echo -e "  ${RED}✓${NC} ${check_id}: ${description} - ${count} instance(s)"
        else
            echo -e "  ${YELLOW}○${NC} ${check_id}: ${description} - Not detected"
        fi
    fi
done

echo ""
echo -e "${YELLOW}Warning Checks:${NC}"
for check_id in "${!CHECK_DETAILS[@]}"; do
    IFS='|' read -r description severity <<< "${CHECK_DETAILS[$check_id]}"
    if [[ "${severity}" == "Warning" ]]; then
        local count=$(count_occurrences "\[${check_id}\]")
        if [[ ${count} -gt 0 ]]; then
            echo -e "  ${YELLOW}✓${NC} ${check_id}: ${description} - ${count} instance(s)"
        else
            echo -e "  ${YELLOW}○${NC} ${check_id}: ${description} - Not detected"
        fi
    fi
done

echo ""
echo -e "Info Checks:"
for check_id in "${!CHECK_DETAILS[@]}"; do
    IFS='|' read -r description severity <<< "${CHECK_DETAILS[$check_id]}"
    if [[ "${severity}" == "Info" ]]; then
        local count=$(count_occurrences "\[${check_id}\]")
        if [[ ${count} -gt 0 ]]; then
            echo -e "  ${GREEN}✓${NC} ${check_id}: ${description} - ${count} instance(s)"
        else
            echo -e "  ○ ${check_id}: ${description} - Not detected"
        fi
    fi
done

echo ""

# Affected resources summary
echo -e "${CYAN}=== Sample Affected Resources ===${NC}"

# Extract a few example affected resources
echo "${SCAN_CONTENT}" | grep -A 2 "Affected:" | head -20

echo ""

# Health score analysis
echo -e "${CYAN}=== Health Score Analysis ===${NC}"

if [[ "${HEALTH_SCORE}" != "Not found" ]]; then
    # Calculate expected penalty
    local expected_penalty=$((CRITICAL_COUNT * 10 + WARNING_COUNT * 3 + INFO_COUNT * 1))
    local expected_score=$((100 - expected_penalty))
    if [[ ${expected_score} -lt 0 ]]; then
        expected_score=0
    fi

    echo "Detected findings:"
    echo "  - Critical: ${CRITICAL_COUNT} × 10 points = $((CRITICAL_COUNT * 10)) points"
    echo "  - Warning: ${WARNING_COUNT} × 3 points = $((WARNING_COUNT * 3)) points"
    echo "  - Info: ${INFO_COUNT} × 1 point = ${INFO_COUNT} points"
    echo "  - Total penalty: ${expected_penalty} points"
    echo ""
    echo "Calculated score: 100 - ${expected_penalty} = ${expected_score}"
    echo "Reported score: ${HEALTH_SCORE}"

    if [[ "${HEALTH_SCORE}" -eq "${expected_score}" ]]; then
        echo -e "${GREEN}✓ Health score calculation is correct${NC}"
    else
        echo -e "${YELLOW}⚠ Health score calculation may differ (this could be due to counting method)${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Health score not found in scan output${NC}"
fi

echo ""

# Grade validation
echo -e "${CYAN}=== Grade Validation ===${NC}"

if [[ "${HEALTH_SCORE}" != "Not found" ]]; then
    local expected_grade="Unknown"

    if [[ ${HEALTH_SCORE} -ge 90 ]]; then
        expected_grade="Excellent"
    elif [[ ${HEALTH_SCORE} -ge 75 ]]; then
        expected_grade="Good"
    elif [[ ${HEALTH_SCORE} -ge 60 ]]; then
        expected_grade="Fair"
    elif [[ ${HEALTH_SCORE} -ge 40 ]]; then
        expected_grade="Poor"
    else
        expected_grade="Critical"
    fi

    echo "Expected grade for score ${HEALTH_SCORE}: ${expected_grade}"
    echo "Reported grade: ${GRADE}"

    if [[ "${GRADE}" == *"${expected_grade}"* ]] || [[ "${expected_grade}" == *"${GRADE}"* ]]; then
        echo -e "${GREEN}✓ Grade is correct${NC}"
    else
        echo -e "${YELLOW}⚠ Grade may not match expected value${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Cannot validate grade without health score${NC}"
fi

echo ""

# Test coverage report
echo -e "${CYAN}=== Test Coverage Summary ===${NC}"

local total_checks=${#CHECK_DETAILS[@]}
local detected_checks=0

for check_id in "${!CHECK_DETAILS[@]}"; do
    local count=$(count_occurrences "\[${check_id}\]")
    if [[ ${count} -gt 0 ]]; then
        detected_checks=$((detected_checks + 1))
    fi
done

local coverage_percent=$((detected_checks * 100 / total_checks))

echo "Total check types: ${total_checks}"
echo "Detected check types: ${detected_checks}"
echo "Coverage: ${coverage_percent}%"

if [[ ${coverage_percent} -ge 80 ]]; then
    echo -e "${GREEN}✓ Good test coverage${NC}"
elif [[ ${coverage_percent} -ge 50 ]]; then
    echo -e "${YELLOW}⚠ Moderate test coverage${NC}"
else
    echo -e "${RED}✗ Low test coverage${NC}"
fi

echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}Validation Complete${NC}"
echo -e "${BLUE}==========================================${NC}"
