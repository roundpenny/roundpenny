#!/bin/bash
# Production-grade load test
set -euo pipefail

BASE_URL="${1:-http://localhost:8000}"
STAGES="${2:-full}"

echo "=== RoundPenny Load Test ==="
echo "Target: $BASE_URL"
echo "Stages: $STAGES"

# Run k6
docker run --rm -i \
    --network host \
    -e BASE_URL="$BASE_URL" \
    -v "$(pwd)/scripts:/scripts" \
    grafana/k6:latest run \
    --summary-export=/scripts/load-test-results.json \
    /scripts/load-test-prod.js

echo "=== Results ==="
cat scripts/load-test-results.json 2>/dev/null || echo "No results file"
