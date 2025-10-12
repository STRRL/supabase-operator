#!/bin/bash
set -e

# Quick test for database initialization Job
# This script helps verify the Job setup works correctly

echo "=== Database Initialization Job Test ==="
echo ""

PROJECT_NAME="${1:-test-project}"
NAMESPACE="${2:-default}"

echo "Testing project: ${PROJECT_NAME} in namespace: ${NAMESPACE}"
echo ""

# Check if Job exists
echo "1. Checking if Job exists..."
if kubectl get job "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" &>/dev/null; then
    echo "   ✓ Job exists"
else
    echo "   ✗ Job not found"
    exit 1
fi

# Check Job status
echo ""
echo "2. Job Status:"
kubectl get job "${PROJECT_NAME}-db-init" -n "${NAMESPACE}"

# Check if Job succeeded
echo ""
echo "3. Checking completion status..."
SUCCEEDED=$(kubectl get job "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" -o jsonpath='{.status.succeeded}')
FAILED=$(kubectl get job "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" -o jsonpath='{.status.failed}')
ACTIVE=$(kubectl get job "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" -o jsonpath='{.status.active}')

if [ "${SUCCEEDED}" = "1" ]; then
    echo "   ✓ Job completed successfully"
elif [ "${FAILED}" -gt "0" ]; then
    echo "   ✗ Job failed (${FAILED} attempts)"
elif [ "${ACTIVE}" -gt "0" ]; then
    echo "   ⏳ Job is running..."
else
    echo "   ? Job status unknown"
fi

# Show logs
echo ""
echo "4. Job Logs:"
echo "---"
kubectl logs job/"${PROJECT_NAME}-db-init" -n "${NAMESPACE}" --tail=50 || echo "No logs available yet"
echo "---"

# Check ConfigMap
echo ""
echo "5. Checking ConfigMap..."
if kubectl get configmap "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" &>/dev/null; then
    echo "   ✓ ConfigMap exists"
    echo ""
    echo "   SQL Script (first 20 lines):"
    kubectl get configmap "${PROJECT_NAME}-db-init" -n "${NAMESPACE}" -o jsonpath='{.data.init\.sql}' | head -20
    echo ""
    echo "   ..."
else
    echo "   ✗ ConfigMap not found"
fi

echo ""
echo "=== Test Complete ==="
echo ""
echo "Useful commands:"
echo "  - View Job details:  kubectl describe job ${PROJECT_NAME}-db-init -n ${NAMESPACE}"
echo "  - View full logs:    kubectl logs job/${PROJECT_NAME}-db-init -n ${NAMESPACE}"
echo "  - Delete Job:        kubectl delete job ${PROJECT_NAME}-db-init -n ${NAMESPACE}"
echo "  - View ConfigMap:    kubectl get configmap ${PROJECT_NAME}-db-init -n ${NAMESPACE} -o yaml"
