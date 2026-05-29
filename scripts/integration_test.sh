#!/bin/sh
set -e

BASE_URL="${BASE_URL:-http://kong:8000}"
FAILED=0

echo "=== Roundup Platform Integration Tests ==="

echo "--- Waiting for services ---"
for i in $(seq 1 60); do
  if curl -sf "$BASE_URL/v1/health" > /dev/null 2>&1; then
    echo "Kong gateway ready"
    break
  fi
  if [ "$i" = 60 ]; then
    echo "FAIL: timeout waiting for services"
    exit 1
  fi
  sleep 2
done

EMAIL="integration-$(date +%s)@test.com"

echo "--- 1. Register ($EMAIL) ---"
REG=$(curl -sf -X POST "$BASE_URL/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"password123\",\"full_name\":\"Integration User\"}") || { echo "FAIL: register"; exit 1; }
echo "Register OK"

echo "--- 2. Login ($EMAIL) ---"
LOGIN=$(curl -sf -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"password123\"}") || { echo "FAIL: login"; exit 1; }
TOKEN=$(echo "$LOGIN" | jq -r '.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "FAIL: no access_token"
  exit 1
fi
echo "Login OK (token obtained)"

echo "--- 3. MFA Setup ---"
MFA_SETUP=$(curl -sf -X POST "$BASE_URL/v1/auth/mfa/setup" \
  -H "Authorization: Bearer $TOKEN") || { echo "FAIL: mfa setup"; exit 1; }
MFA_SECRET=$(echo "$MFA_SETUP" | jq -r '.secret')
if [ -z "$MFA_SECRET" ] || [ "$MFA_SECRET" = "null" ]; then
  echo "FAIL: no mfa secret"
  exit 1
fi
echo "MFA Setup OK (secret=$MFA_SECRET)"

echo "--- 4. Create Merchant ---"
MERCHANT=$(curl -sf -X POST "$BASE_URL/v1/merchants" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Integration Merchant","email":"int-merchant@test.com","description":"Test merchant"}') || { echo "FAIL: create merchant"; exit 1; }
MERCHANT_ID=$(echo "$MERCHANT" | jq -r '.id')
if [ -z "$MERCHANT_ID" ] || [ "$MERCHANT_ID" = "null" ]; then
  echo "FAIL: no merchant id"
  exit 1
fi
echo "Create Merchant OK (id=$MERCHANT_ID)"

USER_ID=$(echo "$LOGIN" | jq -r '.user.id')

echo "--- 5. Create Payment (user=$USER_ID) ---"
PAYMENT=$(curl -sf -X POST "$BASE_URL/v1/payments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\",\"amount\":1000,\"currency\":\"USD\",\"payment_method\":\"card\",\"description\":\"Integration test payment\"}") || { echo "FAIL: create payment"; exit 1; }
PAYMENT_ID=$(echo "$PAYMENT" | jq -r '.id')
if [ -z "$PAYMENT_ID" ] || [ "$PAYMENT_ID" = "null" ]; then
  echo "FAIL: no payment id"
  exit 1
fi
echo "Create Payment OK (id=$PAYMENT_ID)"

echo "--- 6. Create Webhook ---"
WEBHOOK=$(curl -sf -X POST "$BASE_URL/v1/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://webhook.example.com/int-test\",\"events\":[\"payment.completed\",\"payment.failed\"],\"user_id\":\"$USER_ID\"}") || { echo "FAIL: create webhook"; exit 1; }
WEBHOOK_ID=$(echo "$WEBHOOK" | jq -r '.id')
if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" = "null" ]; then
  echo "FAIL: no webhook id"
  exit 1
fi
echo "Create Webhook OK (id=$WEBHOOK_ID)"

echo "--- 7. Create Analytics Event ---"
ANALYTICS=$(curl -sf -X POST "$BASE_URL/v1/analytics/events" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"event_type":"integration_test","properties":{"test":"true","iteration":1}}') || { echo "FAIL: analytics event"; exit 1; }
EVENT_ID=$(echo "$ANALYTICS" | jq -r '.id')
if [ -z "$EVENT_ID" ] || [ "$EVENT_ID" = "null" ]; then
  echo "FAIL: no analytics event id"
  exit 1
fi
echo "Create Analytics Event OK (id=$EVENT_ID)"

echo "--- 8. User Profile ---"
PROFILE=$(curl -sf "$BASE_URL/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN") || { echo "FAIL: get profile"; exit 1; }
echo "$PROFILE" | jq -e ".email == \"$EMAIL\"" > /dev/null || { echo "FAIL: profile email mismatch"; exit 1; }
echo "User Profile OK (email verified)"

echo "--- 9. Refresh Token ---"
REFRESH_TOKEN=$(echo "$LOGIN" | jq -r '.refresh_token')
REFRESH=$(curl -sf -X POST "$BASE_URL/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}") || { echo "FAIL: refresh token"; exit 1; }
NEW_TOKEN=$(echo "$REFRESH" | jq -r '.access_token')
if [ -z "$NEW_TOKEN" ] || [ "$NEW_TOKEN" = "null" ]; then
  echo "FAIL: no access_token in refresh"
  exit 1
fi
echo "Refresh Token OK"

echo "--- 10. Logout ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/auth/logout" \
  -H "Authorization: Bearer $TOKEN")
if [ "$HTTP_CODE" != "204" ]; then
  echo "FAIL: logout returned $HTTP_CODE"
  exit 1
fi
echo "Logout OK"

echo ""
echo "=== All integration tests passed ==="
exit 0
