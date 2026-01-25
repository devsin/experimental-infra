#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8081}
ACCOUNTS_URL=${ACCOUNTS_URL:-http://localhost:8080}
REQ_ID=${REQ_ID:-$(uuidgen 2>/dev/null || date +%s%N)}
FROM_NAME=${FROM_NAME:-"Smoke From"}
TO_NAME=${TO_NAME:-"Smoke To"}
CURRENCY=${CURRENCY:-"USD"}
AMOUNT_CENTS=${AMOUNT_CENTS:-500}
IDEMPOTENCY_KEY=${IDEMPOTENCY_KEY:-"smoke-$(date +%s)"}

command -v jq >/dev/null 2>&1 || { echo "jq is required (e.g. brew install jq)"; exit 1; }

info() { printf "[smoke] %s\n" "$1"; }

info "Accounts URL: ${ACCOUNTS_URL}"
info "Transfers URL: ${BASE_URL}"
info "Request ID: ${REQ_ID}"

# Create two accounts
from_json=$(curl -s -f \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: ${REQ_ID}" \
  -X POST \
  -d "{\"name\":\"${FROM_NAME}\",\"currency\":\"${CURRENCY}\"}" \
  "${ACCOUNTS_URL}/accounts")
from_id=$(echo "${from_json}" | jq -r '.id')
info "Created from account: ${from_id}"

to_json=$(curl -s -f \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: ${REQ_ID}" \
  -X POST \
  -d "{\"name\":\"${TO_NAME}\",\"currency\":\"${CURRENCY}\"}" \
  "${ACCOUNTS_URL}/accounts")
to_id=$(echo "${to_json}" | jq -r '.id')
info "Created to account: ${to_id}"

# Create transfer
create_json=$(curl -s -f \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: ${REQ_ID}" \
  -X POST \
  -d "{\"from\":\"${from_id}\",\"to\":\"${to_id}\",\"amount_cents\":${AMOUNT_CENTS},\"idempotency_key\":\"${IDEMPOTENCY_KEY}\"}" \
  "${BASE_URL}/transfers")
transfer_id=$(echo "${create_json}" | jq -r '.id')
info "Created transfer: ${transfer_id}"

# Get transfer
get_json=$(curl -s -f \
  -H "X-Request-ID: ${REQ_ID}" \
  "${BASE_URL}/transfers/${transfer_id}")
info "Fetched transfer: ${get_json}"

# List transfers for account
list_json=$(curl -s -f \
  -H "X-Request-ID: ${REQ_ID}" \
  "${BASE_URL}/transfers?accountId=${from_id}&limit=5")
info "Listed transfers (from account): ${list_json}"

info "Smoke test completed"
