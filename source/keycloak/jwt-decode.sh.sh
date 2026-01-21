#!/usr/bin/env bash

set -euo pipefail

JWT="${1:-}"

if [[ -z "$JWT" ]]; then
  echo "Usage: $0 <jwt>"
  exit 1
fi

IFS='.' read -r HEADER PAYLOAD SIGNATURE <<< "$JWT"

if [[ -z "${HEADER:-}" || -z "${PAYLOAD:-}" || -z "${SIGNATURE:-}" ]]; then
  echo "Error: Invalid JWT format"
  exit 1
fi

decode_base64url() {
  local input="$1"
  local pad=$(( (4 - ${#input} % 4) % 4 ))
  input="${input}$(printf '=%.0s' $(seq 1 $pad))"
  echo "$input" | tr '_-' '/+' | base64 --decode 2>/dev/null
}

echo
echo "=== JWT HEADER ==="
decode_base64url "$HEADER" | jq .

echo
echo "=== JWT PAYLOAD ==="
decode_base64url "$PAYLOAD" | jq .

echo
echo "=== JWT SIGNATURE ==="
echo "(signature present, not decoded)"
echo "$SIGNATURE"
