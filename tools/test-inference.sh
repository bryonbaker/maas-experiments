#! /bin/bash

# Ask if user wants to do a load test
read -rp "Do you want to do a load test? (yes/no): " LOAD_TEST
LOAD_TEST=$(echo "$LOAD_TEST" | tr '[:upper:]' '[:lower:]')

CLUSTER_DOMAIN=$(kubectl get ingresses.config.openshift.io cluster -o jsonpath='{.spec.domain}')
HOST="http://maas.${CLUSTER_DOMAIN}"
echo "Host: $HOST"
echo "Note: TLS is disabled for this test so using http instead of https."

TOKEN_RESPONSE=$(curl -sSk \
  -H "Authorization: Bearer $(oc whoami -t)" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"expiration": "10m"}' \
  "${HOST}/maas-api/v1/tokens")

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r .token)

MODELS=$(curl -sSk "${HOST}/maas-api/v1/models" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" | jq -r .)

echo "$MODELS" | jq .

# Prompt user for index
TOTAL_MODELS=$(echo "$MODELS" | jq '.data | length')
echo "There are ${TOTAL_MODELS} models available."

read -rp "Enter the array index to use (0 - $((TOTAL_MODELS-1))): " INDEX

# Validate numeric input
if ! [[ "$INDEX" =~ ^[0-9]+$ ]]; then
  echo "Invalid input. Must be a number."
  exit 1
fi

# Validate range
if (( INDEX < 0 || INDEX >= TOTAL_MODELS )); then
  echo "Index out of range."
  exit 1
fi

MODEL_NAME=$(echo "$MODELS" | jq -r ".data[$INDEX].id")
MODEL_URL=$(echo "$MODELS" | jq -r ".data[$INDEX].url")

echo "Selected Model Index: $INDEX"
echo "Model Name: $MODEL_NAME"
echo "Model URL:  $MODEL_URL"

if [[ "$LOAD_TEST" == "yes" || "$LOAD_TEST" == "y" ]]; then
  echo ""
  echo "Starting load test... (will continue until non-JSON response)"
  echo ""
  
  REQUEST_COUNT=0
  while true; do
    REQUEST_COUNT=$((REQUEST_COUNT + 1))
    echo "Request #${REQUEST_COUNT}:"
    
    RESPONSE=$(curl -sSk -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"model\": \"${MODEL_NAME}\", \"prompt\": \"Hello\", \"max_tokens\": 50}" \
      "${MODEL_URL}/v1/completions")
    
    # Check if response is valid JSON
    if echo "$RESPONSE" | jq . >/dev/null 2>&1; then
      echo "✓ Valid JSON response received"
      echo "$RESPONSE" | jq .
    else
      echo "✗ Non-JSON response received (likely rate limited):"
      echo "$RESPONSE"
      echo ""
      echo "Load test completed after ${REQUEST_COUNT} requests"
      break
    fi
    
    echo ""
    sleep 0.5  # Small delay between requests
  done
else
  curl -sSk -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"model\": \"${MODEL_NAME}\", \"prompt\": \"Hello\", \"max_tokens\": 50}" \
    "${MODEL_URL}/v1/completions"
fi

