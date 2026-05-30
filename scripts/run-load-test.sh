#!/bin/sh
# Copyright (c) 2026 RoundPenny. All rights reserved.
set -e

BASE_URL="${BASE_URL:-http://localhost:8000}"

echo "=== Running k6 Load Tests ==="
echo "Target: $BASE_URL"

docker run --rm -i \
  --network host \
  -e BASE_URL="$BASE_URL" \
  -v "$(pwd)/scripts:/scripts" \
  grafana/k6:latest run \
  --vus 10 \
  --duration 30s \
  /scripts/load-test.js

echo ""
echo "=== Load test complete ==="
