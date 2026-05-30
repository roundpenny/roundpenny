#!/bin/sh
# Copyright (c) 2026 RoundPenny. All rights reserved.
set -e

BASE_URL="${BASE_URL:-http://kong:8000}"
FAILED=0

echo "=== Roundup Platform E2E Tests ==="

echo "--- Waiting for services health ---"
for i in $(seq 1 30); do
  if curl -sf "$BASE_URL/v1/auth/me" > /dev/null 2>&1; then
    echo "Auth service ready"
    break
  fi
  if [ "$i" = 30 ]; then
    echo "Timeout waiting for services"
    exit 1
  fi
  sleep 2
done

echo "--- Register ---"
REG_RESP=$(curl -sf -X POST "$BASE_URL/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123","name":"E2E User"}') || { echo "FAIL: register"; FAILED=1; exit 1; }
echo "Register OK"

echo "--- Login ---"
LOGIN_RESP=$(curl -sf -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123"}') || { echo "FAIL: login"; FAILED=1; exit 1; }
ACCESS_TOKEN=$(echo "$LOGIN_RESP" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESP" | jq -r '.refresh_token')
if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "FAIL: no access_token in login response"
  exit 1
fi
echo "Login OK"

echo "--- Refresh ---"
REF_RESP=$(curl -sf -X POST "$BASE_URL/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}") || { echo "FAIL: refresh"; FAILED=1; exit 1; }
NEW_ACCESS=$(echo "$REF_RESP" | jq -r '.access_token')
if [ "$NEW_ACCESS" = "null" ] || [ -z "$NEW_ACCESS" ]; then
  echo "FAIL: no access_token in refresh response"
  exit 1
fi
echo "Refresh OK"

echo "--- Me ---"
ME_RESP=$(curl -sf "$BASE_URL/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN") || { echo "FAIL: me"; FAILED=1; exit 1; }
echo "Me OK"

echo "--- Logout ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
if [ "$HTTP_CODE" != "204" ]; then
  echo "FAIL: logout returned $HTTP_CODE"
  exit 1
fi
echo "Logout OK"

echo "--- Update Profile ---"
USER_ACCESS=$(curl -sf -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123"}' | jq -r '.access_token')
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL/v1/users/me/profile" \
  -H "Authorization: Bearer $USER_ACCESS" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"E2E Updated"}')
if [ "$HTTP_CODE" != "204" ]; then
  echo "FAIL: update profile returned $HTTP_CODE"
  exit 1
fi
echo "Update Profile OK"

echo "--- Update Preferences ---"
PREF_RESP=$(curl -sf -X PUT "$BASE_URL/v1/users/me/preferences" \
  -H "Authorization: Bearer $USER_ACCESS" \
  -H "Content-Type: application/json" \
  -d '{"theme":"dark","notifications_enabled":true}') || { echo "FAIL: update preferences"; FAILED=1; exit 1; }
echo "Update Preferences OK"

echo "--- Create Merchant ---"
MERCHANT_RESP=$(curl -sf -X POST "$BASE_URL/v1/merchants" \
  -H "Authorization: Bearer $USER_ACCESS" \
  -H "Content-Type: application/json" \
  -d '{"name":"E2E Merchant","email":"merchant@test.com","description":"Test merchant"}') || { echo "FAIL: create merchant"; FAILED=1; exit 1; }
MERCHANT_ID=$(echo "$MERCHANT_RESP" | jq -r '.id')
if [ "$MERCHANT_ID" = "null" ] || [ -z "$MERCHANT_ID" ]; then
  echo "FAIL: no merchant id"
  exit 1
fi
echo "Create Merchant OK (id=$MERCHANT_ID)"

echo "--- Get Merchant ---"
curl -sf "$BASE_URL/v1/merchants/$MERCHANT_ID" \
  -H "Authorization: Bearer $USER_ACCESS" > /dev/null || { echo "FAIL: get merchant"; FAILED=1; exit 1; }
echo "Get Merchant OK"

echo "--- Update Merchant ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL/v1/merchants/$MERCHANT_ID" \
  -H "Authorization: Bearer $USER_ACCESS" \
  -H "Content-Type: application/json" \
  -d '{"name":"E2E Merchant Updated"}')
if [ "$HTTP_CODE" != "200" ]; then
  echo "FAIL: update merchant returned $HTTP_CODE"
  exit 1
fi
echo "Update Merchant OK"

echo "--- Delete Merchant ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/v1/merchants/$MERCHANT_ID" \
  -H "Authorization: Bearer $USER_ACCESS")
if [ "$HTTP_CODE" != "204" ]; then
  echo "FAIL: delete merchant returned $HTTP_CODE"
  exit 1
fi
echo "Delete Merchant OK"

echo "=== All E2E tests passed ==="
exit 0
