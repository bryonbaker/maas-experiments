#!/bin/bash
#
# Portal OpenShift Integration Script
# This script demonstrates how to validate Keycloak JWTs and call OpenShift APIs
# on behalf of authenticated users using service account impersonation.
#

set -e

# Configuration
OPENSHIFT_API="https://api.ethan-sno-kk.sandbox3469.opentlc.com:6443"
PORTAL_SA_TOKEN_FILE="/tmp/portal-token.txt"
KEYCLOAK_ISSUER="https://keycloak.apps.ethan-sno-kk.sandbox3469.opentlc.com/realms/maas-tenants"

# Read portal service account token
if [[ ! -f "$PORTAL_SA_TOKEN_FILE" ]]; then
    echo "Error: Portal token file not found at $PORTAL_SA_TOKEN_FILE"
    exit 1
fi
PORTAL_TOKEN=$(cat "$PORTAL_SA_TOKEN_FILE")

#
# Function: validate_and_extract_user
# Validates Keycloak JWT and extracts user information
# Args: $1 - Keycloak JWT
# Returns: JSON object with username, email, and groups
#
validate_and_extract_user() {
    local keycloak_jwt="$1"

    # Decode JWT payload (base64 decode the middle part)
    local payload=$(echo "$keycloak_jwt" | cut -d'.' -f2)

    # Add padding if needed for base64 decoding
    local padded_payload="$payload"
    local mod=$((${#payload} % 4))
    if [[ $mod -eq 2 ]]; then
        padded_payload="${payload}=="
    elif [[ $mod -eq 3 ]]; then
        padded_payload="${payload}="
    fi

    # Decode and parse JSON
    local decoded=$(echo "$padded_payload" | base64 -d 2>/dev/null)

    if [[ -z "$decoded" ]]; then
        echo "Error: Failed to decode JWT" >&2
        return 1
    fi

    # Verify issuer
    local issuer=$(echo "$decoded" | jq -r '.iss')
    if [[ "$issuer" != "$KEYCLOAK_ISSUER" ]]; then
        echo "Error: Invalid issuer. Expected $KEYCLOAK_ISSUER, got $issuer" >&2
        return 1
    fi

    # Check expiration
    local exp=$(echo "$decoded" | jq -r '.exp')
    local now=$(date +%s)
    if [[ $exp -lt $now ]]; then
        echo "Error: Token has expired" >&2
        return 1
    fi

    # Extract user information
    local username=$(echo "$decoded" | jq -r '.preferred_username')
    local email=$(echo "$decoded" | jq -r '.email')
    local groups=$(echo "$decoded" | jq -c '.groups // []')

    # Return as JSON
    jq -n \
        --arg username "$username" \
        --arg email "$email" \
        --argjson groups "$groups" \
        '{username: $username, email: $email, groups: $groups}'
}

#
# Function: call_openshift_api
# Calls OpenShift API on behalf of a user using impersonation
# Args: $1 - API endpoint (e.g., /api/v1/namespaces)
#       $2 - Username to impersonate
#       $3 - Groups (JSON array)
#
call_openshift_api() {
    local endpoint="$1"
    local username="$2"
    local groups="$3"

    # Build impersonation headers
    local headers=(-H "Authorization: Bearer $PORTAL_TOKEN")
    headers+=(-H "Impersonate-User: $username")

    # Add all groups
    local group_count=$(echo "$groups" | jq 'length')
    for ((i=0; i<group_count; i++)); do
        local group=$(echo "$groups" | jq -r ".[$i]")
        headers+=(-H "Impersonate-Group: $group")
    done

    # Make API call
    curl -sk "${headers[@]}" "${OPENSHIFT_API}${endpoint}"
}

#
# Function: create_namespace
# Creates a namespace on behalf of a user
# Args: $1 - Namespace name
#       $2 - Username to impersonate
#       $3 - Groups (JSON array)
#
create_namespace() {
    local namespace="$1"
    local username="$2"
    local groups="$3"

    # Build impersonation headers
    local headers=(-H "Authorization: Bearer $PORTAL_TOKEN")
    headers+=(-H "Impersonate-User: $username")
    headers+=(-H "Content-Type: application/json")

    # Add all groups
    local group_count=$(echo "$groups" | jq 'length')
    for ((i=0; i<group_count; i++)); do
        local group=$(echo "$groups" | jq -r ".[$i]")
        headers+=(-H "Impersonate-Group: $group")
    done

    # Create namespace JSON
    local namespace_json=$(jq -n \
        --arg name "$namespace" \
        '{
            "apiVersion": "v1",
            "kind": "Namespace",
            "metadata": {
                "name": $name
            }
        }')

    # Make API call
    curl -sk -X POST \
        "${headers[@]}" \
        -d "$namespace_json" \
        "${OPENSHIFT_API}/api/v1/namespaces"
}

#
# Main function to handle user request
# Args: $1 - Keycloak JWT
#       $2 - Action (e.g., "list-namespaces", "get-user", "create-namespace")
#       $3 - Optional: Additional parameter (e.g., namespace name)
#
handle_user_request() {
    local keycloak_jwt="$1"
    local action="$2"
    local param="$3"

    echo "=== Validating Keycloak JWT ===" >&2

    # Validate JWT and extract user info
    local user_info=$(validate_and_extract_user "$keycloak_jwt")
    if [[ $? -ne 0 ]]; then
        echo "Error: JWT validation failed" >&2
        return 1
    fi

    local username=$(echo "$user_info" | jq -r '.username')
    local groups=$(echo "$user_info" | jq -c '.groups')

    echo "User: $username" >&2
    echo "Groups: $groups" >&2
    echo "" >&2

    # Perform requested action
    case "$action" in
        list-namespaces)
            echo "=== Listing Namespaces ===" >&2
            call_openshift_api "/api/v1/namespaces" "$username" "$groups"
            ;;
        get-user)
            echo "=== Getting User Info ===" >&2
            call_openshift_api "/apis/user.openshift.io/v1/users/~" "$username" "$groups"
            ;;
        get-openshift-ai-auth)
            echo "=== Getting OpenShift AI Auth Config ===" >&2
            call_openshift_api "/apis/services.platform.opendatahub.io/v1alpha1/auths/auth" "$username" "$groups"
            ;;
        create-namespace)
            if [[ -z "$param" ]]; then
                echo "Error: Namespace name required" >&2
                return 1
            fi
            echo "=== Creating Namespace: $param ===" >&2
            create_namespace "$param" "$username" "$groups"
            ;;
        *)
            echo "Error: Unknown action '$action'" >&2
            echo "Valid actions: list-namespaces, get-user, get-openshift-ai-auth, create-namespace" >&2
            return 1
            ;;
    esac
}

# Usage example
if [[ $# -eq 0 ]]; then
    cat <<EOF
Usage: $0 <keycloak-jwt> <action> [param]

Actions:
  list-namespaces          - List all namespaces user can see
  get-user                 - Get current user information
  get-openshift-ai-auth    - Get OpenShift AI auth configuration
  create-namespace <name>  - Create a new namespace

Example:
  $0 "eyJhbG..." list-namespaces
  $0 "eyJhbG..." create-namespace my-new-project

EOF
    exit 1
fi

# Run the main function
handle_user_request "$@"
