#!/usr/bin/env bash
set -e

BASE="${BASE_URL:-http://localhost}"
EMAIL="demo-$(date +%s)@roundpenny.com"
PASS="password123"
FAILED=0
PASSED=0

GREEN='\033[0;32m'; RED='\033[0;31m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; NC='\033[0m'

pass() { PASSED=$((PASSED+1)); echo -e "${GREEN}  ✓ $1${NC}"; }
fail() { FAILED=$((FAILED+1)); echo -e "${RED}  ✗ $1${NC}"; }
header() { echo; echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; echo -e "${CYAN}  $1${NC}"; echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; }
step() { echo -e "${YELLOW}  ▶ $1${NC}"; }

echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════╗"
echo "║              RoundPenny  Demo                       ║"
echo "║     Round-Up Micro-Transaction Platform             ║"
echo "╚══════════════════════════════════════════════════════╝"
echo -e "${NC}"

header "1. HEALTH CHECK"
step "Checking all services..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/v1/health" 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
  pass "All services healthy (HTTP $HTTP_CODE)"
else
  fail "Health check failed (HTTP $HTTP_CODE)"
fi

header "2. USER REGISTRATION"
step "Registering: $EMAIL"
REG=$(curl -sf -X POST "$BASE/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"full_name\":\"Demo User\"}") || { fail "Register failed"; exit 1; }
pass "User registered: $(echo "$REG" | tr -d '\n')"

header "3. USER LOGIN"
step "Authenticating..."
LOGIN=$(curl -sf -X POST "$BASE/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}") || { fail "Login failed"; exit 1; }
TOKEN=$(echo "$LOGIN" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
USER_ID=$(echo "$LOGIN" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
  pass "Token obtained ✓"
else
  fail "No token in response"; exit 1
fi

header "4. MERCHANT ONBOARDING"
step "Creating merchant..."
MERCHANT=$(curl -sf -X POST "$BASE/v1/merchants" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"RoundPenny Demo Store","email":"store@roundpenny.com","description":"Demo merchant"}')
MERCHANT_ID=$(echo "$MERCHANT" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
[ -n "$MERCHANT_ID" ] && pass "Merchant created: $MERCHANT_ID" || fail "No merchant ID"

header "5. PAYMENT PROCESSING"
step "Creating payment (10 USD)..."
PAYMENT=$(curl -sf -X POST "$BASE/v1/payments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\",\"amount\":10.00,\"currency\":\"USD\",\"payment_method\":\"card\",\"description\":\"Demo purchase\"}")
PAYMENT_ID=$(echo "$PAYMENT" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
STATUS=$(echo "$PAYMENT" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
[ -n "$PAYMENT_ID" ] && pass "Payment created: $PAYMENT_ID (status: $STATUS)" || fail "No payment ID"

step "Confirming payment..."
CONFIRM=$(curl -sf -X POST "$BASE/v1/payments/$PAYMENT_ID/confirm" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"payment_method_id":"pm_mock_123"}') || true
echo -e "  ${YELLOW}→ Confirm response:$(echo "$CONFIRM" | head -c 200)${NC}"

header "6. WEBHOOK REGISTRATION"
step "Registering webhook endpoint..."
WEBHOOK=$(curl -sf -X POST "$BASE/v1/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\",\"url\":\"https://webhook.example.com/callback\",\"events\":[\"payment.completed\",\"payment.failed\"],\"description\":\"Demo webhook\"}")
WEBHOOK_ID=$(echo "$WEBHOOK" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
[ -n "$WEBHOOK_ID" ] && pass "Webhook registered: $WEBHOOK_ID" || fail "No webhook ID"

header "7. ANALYTICS"
step "Tracking analytics event..."
EVENT=$(curl -sf -X POST "$BASE/v1/analytics/events" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"event_type":"demo_test","properties":{"demo":true,"source":"presentation"}}')
EVENT_ID=$(echo "$EVENT" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
[ -n "$EVENT_ID" ] && pass "Event tracked: $EVENT_ID" || fail "No event ID"

header "8. FRAUD DETECTION"
step "Creating fraud detection rule..."
RULE=$(curl -sf -X POST "$BASE/v1/fraud/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"High amount alert\",\"description\":\"Flag transactions over 1000\",\"rule_type\":\"amount_threshold\",\"threshold\":1000,\"severity\":\"high\",\"enabled\":true}")
RULE_ID=$(echo "$RULE" | grep -o '"id":"[^"]*' | cut -d'"' -f4 || echo "")
[ -n "$RULE_ID" ] && pass "Fraud rule created: $RULE_ID" || fail "No rule ID"

step "Listing fraud alerts..."
ALERTS=$(curl -sf "$BASE/v1/fraud/alerts" \
  -H "Authorization: Bearer $TOKEN")
ALERT_COUNT=$(echo "$ALERTS" | grep -o '"id"' | wc -l || echo "0")
pass "Fraud alerts checked: $ALERT_COUNT found"

header "9. USER PROFILE"
step "Fetching user profile..."
PROFILE=$(curl -sf "$BASE/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN")
PROFILE_EMAIL=$(echo "$PROFILE" | grep -o '"email":"[^"]*' | cut -d'"' -f4)
[ "$PROFILE_EMAIL" = "$EMAIL" ] && pass "Profile verified: $PROFILE_EMAIL" || fail "Email mismatch"

header "10. MONITORING"
step "Checking Grafana dashboards..."
GF_HTTP=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:3000" 2>/dev/null || echo "000")
case "$GF_HTTP" in
  200|302) pass "Grafana: http://localhost:3000 (admin/admin)" ;;
  *) echo -e "  ${YELLOW}→ Grafana unavailable (HTTP $GF_HTTP)${NC}" ;;
esac
PM_HTTP=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:9090" 2>/dev/null || echo "000")
case "$PM_HTTP" in
  200|302) pass "Prometheus: http://localhost:9090" ;;
  *) echo -e "  ${YELLOW}→ Prometheus unavailable (HTTP $PM_HTTP)${NC}" ;;
esac
SW_HTTP=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080" 2>/dev/null || echo "000")
[ "$SW_HTTP" = "200" ] && pass "Swagger UI: http://localhost:8080 (HTTP $SW_HTTP)" || echo -e "  ${YELLOW}→ Swagger UI not reachable (HTTP $SW_HTTP)${NC}"

header "11. TOKEN REFRESH"
step "Refreshing JWT token..."
REFRESH_TOKEN=$(echo "$LOGIN" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
REFRESH=$(curl -sf -X POST "$BASE/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
NEW_TOKEN=$(echo "$REFRESH" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
[ -n "$NEW_TOKEN" ] && pass "Token refreshed (token rotation ✓)" || fail "Refresh failed"

header "12. LOGOUT"
step "Signing out..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/v1/auth/logout" \
  -H "Authorization: Bearer $TOKEN")
[ "$HTTP_CODE" = "204" ] && pass "Logged out (HTTP $HTTP_CODE)" || fail "Logout returned $HTTP_CODE"

echo
echo -e "${CYAN}══════════════════════════════════════════════════════════${NC}"
echo
if [ "$FAILED" = "0" ]; then
  echo -e "${GREEN}  ✅  RoundPenny Demo Complete — All $PASSED steps passed!${NC}"
else
  echo -e "${RED}  ⚠️  Demo completed with $FAILED failure(s), $PASSED passed${NC}"
fi
echo
echo -e "  ${CYAN}Resources:${NC}"
echo -e "  ${YELLOW}• API:${NC}        http://localhost/v1"
echo -e "  ${YELLOW}• Swagger:${NC}    http://localhost:8080"
echo -e "  ${YELLOW}• Grafana:${NC}    http://localhost:3000 (admin/admin)"
echo -e "  ${YELLOW}• Prometheus:${NC} http://localhost:9090"
echo -e "  ${YELLOW}• Kong Admin:${NC} http://localhost:8001"
echo
echo -e "  ${YELLOW}Postman Collection:${NC} docs/RoundPenny.postman_collection.json"
echo -e "  ${YELLOW}API Spec:${NC}          docs/openapi.yaml"
echo
exit $FAILED
