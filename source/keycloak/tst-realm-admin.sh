#!/usr/bin/env bash
set -euo pipefail

# ---- Config ----
KEYCLOAK_URL="${KEYCLOAK_URL:-}"
CLIENT_ID="${CLIENT_ID:-realm-admin-cli}"
KK_CLIENT_SECRET="${KK_CLIENT_SECRET:-}"

REALM="${REALM:-tenant-test-1}"
TENANT_CLIENT_ID="${TENANT_CLIENT_ID:-tenant-admin-cli}"
GROUP_NAME="${GROUP_NAME:-default-users}"

USER_NAME="${USER_NAME:-alice}"
EMAIL="${EMAIL:-alice@wonderland.com}"
FIRST_NAME="${FIRST_NAME:-Alice}"
LAST_NAME="${LAST_NAME:-Wonderland}"

INITIAL_PASSWORD="${INITIAL_PASSWORD:-}"

# SSL verification (set INSECURE_SSL=true only for development/testing)
INSECURE_SSL="${INSECURE_SSL:-false}"
if [[ "$INSECURE_SSL" == "true" ]]; then
  CURL_INSECURE="-k"
else
  CURL_INSECURE=""
fi

# ---- Helpers ----
need() {
  if [[ -z "${2:-}" ]]; then
    echo "ERROR: Missing required value for $1" >&2
    exit 1
  fi
}

http_status() {
  awk 'NR==1 {print $2}'
}

# ---- Preconditions ----
need "KEYCLOAK_URL" "$KEYCLOAK_URL"
need "CLIENT_ID" "$CLIENT_ID"
need "KK_CLIENT_SECRET" "$KK_CLIENT_SECRET"
need "INITIAL_PASSWORD" "$INITIAL_PASSWORD"
command -v jq >/dev/null 2>&1 || { echo "ERROR: jq is required" >&2; exit 1; }

echo "Keycloak URL : $KEYCLOAK_URL"
echo "Client ID   : $CLIENT_ID"
echo "Realm       : $REALM"
echo "User        : $USER_NAME ($FIRST_NAME $LAST_NAME)"
echo

# ---- 1) Get admin token ----
echo "[1/8] Getting access token (client_credentials)..."
TOKEN_JSON="$(curl -s ${CURL_INSECURE} \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${KK_CLIENT_SECRET}" \
  "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token")"

ACCESS_TOKEN="$(jq -r .access_token <<<"$TOKEN_JSON")"

if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
  echo "ERROR: Failed to obtain access token" >&2
  echo "$TOKEN_JSON" | jq . >&2 || echo "$TOKEN_JSON" >&2
  exit 1
fi

echo "Token acquired"
echo

# ---- 2) Create realm ----
echo "[2/8] Creating realm: $REALM"
RESP_HEADERS="$(curl -s ${CURL_INSECURE} -D - -o /dev/null \
  -X POST \
  "${KEYCLOAK_URL}/admin/realms" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"realm\":\"${REALM}\",\"enabled\":true}")"

STATUS="$(echo "$RESP_HEADERS" | http_status)"
echo "HTTP status: $STATUS"
[[ "$STATUS" == "201" || "$STATUS" == "409" ]] || exit 1
echo

# ---- Wait for realm to become available ----
echo "Waiting for realm ${REALM} to become available..."

for i in {1..10}; do
  if curl -s ${CURL_INSECURE} \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    "${KEYCLOAK_URL}/admin/realms/${REALM}" \
    >/dev/null 2>&1; then
    echo "Realm ${REALM} is ready."
    break
  fi

  echo "Realm not ready yet, retrying (${i}/10)..."
  sleep 2
done

# ---- 3) Create tenant admin client ----
echo "[3/8] Creating tenant client: $TENANT_CLIENT_ID"
RESP_HEADERS="$(curl -s ${CURL_INSECURE} -D - -o /dev/null \
  -X POST \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/clients" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"clientId\": \"${TENANT_CLIENT_ID}\",
    \"enabled\": true,
    \"publicClient\": false,
    \"serviceAccountsEnabled\": true,
    \"standardFlowEnabled\": false,
    \"directAccessGrantsEnabled\": true
  }")"

STATUS="$(echo "$RESP_HEADERS" | http_status)"
echo "HTTP status: $STATUS"
[[ "$STATUS" == "201" || "$STATUS" == "409" ]] || exit 1
echo

# ---- 4) Create default group ----
echo "[4/8] Creating group: $GROUP_NAME"
RESP_HEADERS="$(curl -s ${CURL_INSECURE} -D - -o /dev/null \
  -X POST \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/groups" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${GROUP_NAME}\"}")"

STATUS="$(echo "$RESP_HEADERS" | http_status)"
echo "HTTP status: $STATUS"
[[ "$STATUS" == "201" || "$STATUS" == "204" || "$STATUS" == "409" ]] || exit 1
echo

# ---- 5) Create user ----
echo "[5/8] Creating user: $USER_NAME ($FIRST_NAME $LAST_NAME)"
RESP_HEADERS="$(curl -s ${CURL_INSECURE} -D - -o /dev/null \
  -X POST \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/users" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USER_NAME}\",
    \"enabled\": true,
    \"email\": \"${EMAIL}\",
    \"emailVerified\": true,
    \"firstName\": \"${FIRST_NAME}\",
    \"lastName\": \"${LAST_NAME}\"
  }")"

STATUS="$(echo "$RESP_HEADERS" | http_status)"
echo "HTTP status: $STATUS"
[[ "$STATUS" == "201" || "$STATUS" == "409" ]] || exit 1
echo

# ---- 6) Resolve user ID ----
USER_ID="$(curl -s ${CURL_INSECURE} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/users?username=${USER_NAME}" \
  | jq -r '.[0].id')"

[[ -n "$USER_ID" && "$USER_ID" != "null" ]] || { echo "ERROR: Could not resolve USER_ID"; exit 1; }

# ---- 7) Set initial password ----
echo "[6/8] Setting initial password (non-temporary)"
curl -s ${CURL_INSECURE} -i \
  -X PUT \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/users/${USER_ID}/reset-password" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"password\",
    \"value\": \"${INITIAL_PASSWORD}\",
    \"temporary\": false
  }" >/dev/null
echo "Password set"
echo

# ---- 8) Assign admin roles (users + groups only) ----
echo "[7/8] Assigning limited admin roles"

ADMIN_CLIENT_ID="$(curl -s ${CURL_INSECURE} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/clients" \
  | jq -r '.[] | select(.clientId=="red-hat-realm") | .id')"

MANAGE_USERS_ROLE="$(curl -s ${CURL_INSECURE} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${ADMIN_CLIENT_ID}/roles/manage-users")"

QUERY_GROUPS_ROLE="$(curl -s ${CURL_INSECURE} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${ADMIN_CLIENT_ID}/roles/query-groups")"

curl -s ${CURL_INSECURE} -i \
  -X POST \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/users/${USER_ID}/role-mappings/clients/${ADMIN_CLIENT_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "[
    ${MANAGE_USERS_ROLE},
    ${QUERY_GROUPS_ROLE}
  ]" >/dev/null

echo "Admin permissions assigned"
echo

# ---- 9) Add user to default group ----
echo "[8/8] Assigning user to group"

GROUP_ID="$(curl -s ${CURL_INSECURE} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/groups?search=${GROUP_NAME}" \
  | jq -r '.[0].id')"

curl -s ${CURL_INSECURE} -i \
  -X PUT \
  "${KEYCLOAK_URL}/admin/realms/${REALM}/users/${USER_ID}/groups/${GROUP_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" >/dev/null

echo
echo "Bootstrap complete:"
echo "  Realm : $REALM"
echo "  User  : $USER_NAME ($FIRST_NAME $LAST_NAME)"
echo "  Group : $GROUP_NAME"

