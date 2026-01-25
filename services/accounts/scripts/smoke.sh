#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
REQ_ID=${REQ_ID:-$(uuidgen 2>/dev/null || date +%s%N)}
NAME=${NAME:-"Smoke User"}
CURRENCY=${CURRENCY:-"USD"}
UPDATED_NAME=${UPDATED_NAME:-"Smoke User Updated"}

command -v jq >/dev/null 2>&1 || { echo "jq is required (e.g. brew install jq)"; exit 1; }

info() { printf "[smoke] %s\n" "$1"; }

info "Base URL: ${BASE_URL}"
info "Request ID: ${REQ_ID}"

create_json=$(curl -s -f \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: ${REQ_ID}" \
  -X POST \
  -d "{\"name\":\"${NAME}\",\"currency\":\"${CURRENCY}\"}" \
  "${BASE_URL}/accounts")

acct_id=$(echo "${create_json}" | jq -r '.id')
info "Created account: ${acct_id}"

get_json=$(curl -s -f \
  -H "X-Request-ID: ${REQ_ID}" \
  "${BASE_URL}/accounts/${acct_id}")
info "Fetched account: ${get_json}"

update_json=$(curl -s -f \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: ${REQ_ID}" \
  -X PATCH \
  -d "{\"name\":\"${UPDATED_NAME}\"}" \
  "${BASE_URL}/accounts/${acct_id}")
info "Updated account: ${update_json}"

delete_json=$(curl -s -f \
  -H "X-Request-ID: ${REQ_ID}" \
  -X DELETE \
  "${BASE_URL}/accounts/${acct_id}")
info "Deleted account: ${delete_json}"

info "Smoke test completed"
