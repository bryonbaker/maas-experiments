# Portal OpenShift Integration Guide

## Overview

This guide describes how to integrate a custom portal with OpenShift when users authenticate via Keycloak. The integration enables the portal to call OpenShift APIs on behalf of authenticated users while maintaining proper authorization and security.

### Architecture

```
┌─────────┐      ┌──────────┐      ┌──────────────┐      ┌──────────┐
│  User   │─────>│  Portal  │─────>│   Keycloak   │      │OpenShift │
└─────────┘      └──────────┘      └──────────────┘      └──────────┘
                      │                    │                    │
                      │  1. User Login     │                    │
                      │───────────────────>│                    │
                      │                    │                    │
                      │  2. Keycloak JWT   │                    │
                      │<───────────────────│                    │
                      │                    │                    │
                      │  3. Portal validates JWT                │
                      │     extracts username & groups          │
                      │                                         │
                      │  4. API Call with Impersonation         │
                      │────────────────────────────────────────>│
                      │     Headers:                            │
                      │     - Authorization: Bearer <portal-token>
                      │     - Impersonate-User: username        │
                      │     - Impersonate-Group: group1         │
                      │                                         │
                      │  5. Response (as user)                  │
                      │<────────────────────────────────────────│
```

### Key Components

1. **Keycloak**: Identity provider that issues JWTs for authenticated users
2. **Portal**: Custom application that users interact with
3. **Portal Service Account**: OpenShift service account with impersonation privileges
4. **OpenShift API**: Receives API calls with impersonation headers
5. **OpenShift AI**: Configured to allow specific Keycloak groups

### Authentication Flow

1. **User Authentication**: User logs into the portal, which redirects to Keycloak
2. **JWT Issuance**: Keycloak authenticates the user and issues a JWT containing:
   - User identity (email, username)
   - Group memberships
   - Claims (issuer, expiration, etc.)
3. **JWT Validation**: Portal validates the JWT to ensure authenticity
4. **User Extraction**: Portal extracts username and groups from the JWT
5. **API Impersonation**: Portal makes OpenShift API calls using:
   - Its own service account token for authentication
   - Impersonation headers to act on behalf of the user
6. **Authorization**: OpenShift checks if the impersonated user has permissions
7. **Response**: OpenShift returns data based on user's permissions

### Why This Approach?

- **Security**: Portal never needs user credentials; only validates JWTs
- **Authorization**: Users maintain their own OpenShift permissions and groups
- **Auditability**: OpenShift audit logs show the actual user performing actions
- **Scalability**: Single service account handles all portal requests
- **Flexibility**: Works with any Keycloak-authenticated user without creating OpenShift tokens

---

## Procedure to Configure

### Prerequisites

- OpenShift cluster with admin access
- Keycloak realm configured
- `oc` CLI tool installed and logged in as admin
- `jq` utility for JSON parsing

### Step 1: Configure Keycloak Client

1. **Access Keycloak Admin Console**
   ```
   URL: https://keycloak.apps.<cluster-domain>/admin
   ```

2. **Select or Create a Realm**
   - Navigate to the appropriate realm (e.g., `maas-tenants`)

3. **Configure Client Settings**

   Navigate to: **Clients** → **[your-client]** → **Settings**

   Set the following:
   - **Client ID**: `maas` (or your preferred client ID)
   - **Client Protocol**: `openid-connect`
   - **Access Type**: `confidential`
   - **Valid Redirect URIs**:
     ```
     https://oauth-openshift.apps.<cluster-domain>/oauth2callback/<identity-provider-name>
     https://oauth-openshift.apps.<cluster-domain>/*
     ```
   - **Standard Flow Enabled**: `ON`

4. **Configure Group Mapper**

   Navigate to: **Clients** → **[your-client]** → **Mappers** → **Create**

   Create a mapper with:
   - **Name**: `groups`
   - **Mapper Type**: `Group Membership`
   - **Token Claim Name**: `groups`
   - **Full group path**: `OFF`
   - **Add to ID token**: `ON`
   - **Add to access token**: `ON`
   - **Add to userinfo**: `ON`

5. **Create Groups for OpenShift AI Access**

   Navigate to: **Groups** → **New**

   Create the following group:
   - **Name**: `rhoai-users`

   Add users to this group who should have OpenShift AI access.

6. **Get Client Secret**

   Navigate to: **Clients** → **[your-client]** → **Credentials**

   Copy the **Secret** value - you'll need this for OpenShift configuration.

### Step 2: Configure OpenShift OAuth Integration

1. **Create Client Secret in OpenShift**

   ```bash
   oc create secret generic keycloak-client-secret \
     --from-literal=clientSecret=<YOUR_CLIENT_SECRET> \
     -n openshift-config
   ```

2. **Add Keycloak as Identity Provider**

   ```bash
   oc patch oauth cluster --type=json -p '[
     {
       "op": "add",
       "path": "/spec/identityProviders/-",
       "value": {
         "name": "keycloak",
         "type": "OpenID",
         "mappingMethod": "claim",
         "openID": {
           "clientID": "maas",
           "clientSecret": {
             "name": "keycloak-client-secret"
           },
           "issuer": "https://keycloak.apps.<cluster-domain>/realms/<realm-name>",
           "claims": {
             "preferredUsername": ["preferred_username"],
             "name": ["name"],
             "email": ["email"],
             "groups": ["groups"]
           }
         }
       }
     }
   ]'
   ```

3. **Wait for OAuth Pods to Rollout**

   ```bash
   oc get pods -n openshift-authentication -w
   ```

   Wait until new pods are running.

4. **Verify OAuth Configuration**

   ```bash
   oc get oauth cluster -o yaml
   ```

### Step 3: Configure OpenShift AI to Allow Keycloak Groups

1. **Update OpenShift AI Auth Resource**

   ```bash
   oc patch auth auth --type=merge -p '{
     "spec": {
       "allowedGroups": [
         "system:authenticated",
         "rhoai-users"
       ]
     }
   }'
   ```

2. **Verify Configuration**

   ```bash
   oc get auth auth -o yaml
   ```

   Expected output:
   ```yaml
   spec:
     adminGroups:
     - rhods-admins
     allowedGroups:
     - system:authenticated
     - rhoai-users
   ```

3. **Test User Login via Web Console**

   - Navigate to OpenShift Console: `https://console-openshift-console.apps.<cluster-domain>`
   - Select "keycloak" as login method
   - Log in with a Keycloak user in the `rhoai-users` group
   - Verify user and groups are created:
     ```bash
     oc get users
     oc get groups
     ```

### Step 4: Create Portal Service Account with Impersonation Rights

1. **Create Service Account**

   ```bash
   oc create serviceaccount portal-sa -n default
   ```

2. **Create Impersonation ClusterRole**

   ```bash
   cat <<EOF | oc apply -f -
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: user-impersonator
   rules:
   - apiGroups: [""]
     resources: ["users", "groups"]
     verbs: ["impersonate"]
   - apiGroups: ["user.openshift.io"]
     resources: ["users", "groups"]
     verbs: ["impersonate"]
   - apiGroups: ["authentication.k8s.io"]
     resources: ["userextras"]
     verbs: ["impersonate"]
   EOF
   ```

3. **Grant Impersonation Rights to Service Account**

   ```bash
   oc create clusterrolebinding portal-impersonator \
     --clusterrole=user-impersonator \
     --serviceaccount=default:portal-sa
   ```

4. **Create Long-Lived Service Account Token**

   ```bash
   # Create token valid for 10 years
   oc create token portal-sa -n default --duration=87600h > portal-token.txt

   # Secure the token file
   chmod 600 portal-token.txt
   ```

5. **Verify Service Account Can Impersonate**

   ```bash
   TOKEN=$(cat portal-token.txt)

   oc --token="$TOKEN" \
      --as=<keycloak-username> \
      --as-group=rhoai-users \
      whoami
   ```

   Expected output: `<keycloak-username>`

### Step 5: Configure Portal Application

1. **Store Portal Token Securely**

   Store the token from `portal-token.txt` in your portal's secure configuration (e.g., environment variable, secret manager, encrypted config file).

2. **Configure Portal Constants**

   ```bash
   # In your portal configuration
   OPENSHIFT_API_URL="https://api.<cluster-domain>:6443"
   KEYCLOAK_ISSUER="https://keycloak.apps.<cluster-domain>/realms/<realm-name>"
   PORTAL_SA_TOKEN="<contents-of-portal-token.txt>"
   ```

3. **Install Required Libraries**

   For bash scripts:
   ```bash
   # Ensure jq and curl are installed
   yum install -y jq curl
   ```

   For Python applications:
   ```bash
   pip install requests PyJWT cryptography
   ```

### Step 6: Verify Complete Integration

1. **Test JWT Validation**

   Get a Keycloak JWT for a test user and verify it can be decoded:
   ```bash
   echo "<jwt>" | cut -d'.' -f2 | base64 -d | jq .
   ```

2. **Test Impersonation**

   ```bash
   TOKEN=$(cat portal-token.txt)

   curl -sk \
     -H "Authorization: Bearer $TOKEN" \
     -H "Impersonate-User: <username>" \
     -H "Impersonate-Group: rhoai-users" \
     https://api.<cluster-domain>:6443/apis/user.openshift.io/v1/users/~ \
     | jq .
   ```

3. **Test OpenShift AI Access**

   ```bash
   oc --token="$TOKEN" \
      --as=<username> \
      --as-group=rhoai-users \
      get auths.services.platform.opendatahub.io
   ```

---

## Examples

### Example 1: Bash Script - Complete Portal Integration

**File**: `portal-openshift-integration.sh`

```bash
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
```

**Usage**:

```bash
# Make script executable
chmod +x portal-openshift-integration.sh

# List namespaces for authenticated user
./portal-openshift-integration.sh "eyJhbGc..." list-namespaces

# Get user information
./portal-openshift-integration.sh "eyJhbGc..." get-user

# Create a namespace
./portal-openshift-integration.sh "eyJhbGc..." create-namespace my-project
```

### Example 2: Simple cURL Commands

**List Namespaces**:
```bash
PORTAL_TOKEN="<your-portal-sa-token>"
USERNAME="brbaker@redhat.com"

curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.<cluster-domain>:6443/api/v1/namespaces | jq '.items[].metadata.name'
```

**Get Current User Info**:
```bash
curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.<cluster-domain>:6443/apis/user.openshift.io/v1/users/~ | jq .
```

**Create a Namespace**:
```bash
curl -sk -X POST \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Namespace",
    "metadata": {
      "name": "my-new-project"
    }
  }' \
  https://api.<cluster-domain>:6443/api/v1/namespaces
```

**Access OpenShift AI Resources**:
```bash
curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.<cluster-domain>:6443/apis/services.platform.opendatahub.io/v1alpha1/auths/auth \
  | jq '{allowedGroups: .spec.allowedGroups, adminGroups: .spec.adminGroups}'
```

### Example 3: Using `oc` CLI with Impersonation

**Basic Impersonation**:
```bash
PORTAL_TOKEN=$(cat portal-token.txt)

oc --token="$PORTAL_TOKEN" \
   --as=brbaker@redhat.com \
   --as-group=rhoai-users \
   get namespaces
```

**Create Resources**:
```bash
oc --token="$PORTAL_TOKEN" \
   --as=brbaker@redhat.com \
   --as-group=rhoai-users \
   new-project my-project
```

**Check Permissions**:
```bash
oc --token="$PORTAL_TOKEN" \
   --as=brbaker@redhat.com \
   --as-group=rhoai-users \
   auth can-i create namespaces
```

### Example 4: Python Flask API Endpoint

```python
from flask import Flask, request, jsonify
import requests
import jwt
import os

app = Flask(__name__)

# Configuration
OPENSHIFT_API = os.getenv('OPENSHIFT_API', 'https://api.example.com:6443')
PORTAL_TOKEN = os.getenv('PORTAL_SA_TOKEN')
KEYCLOAK_PUBLIC_KEY = os.getenv('KEYCLOAK_PUBLIC_KEY')
KEYCLOAK_ISSUER = os.getenv('KEYCLOAK_ISSUER')

def validate_jwt(token):
    """Validate Keycloak JWT and return decoded payload"""
    try:
        decoded = jwt.decode(
            token,
            KEYCLOAK_PUBLIC_KEY,
            algorithms=["RS256"],
            audience="account",
            issuer=KEYCLOAK_ISSUER
        )
        return decoded
    except jwt.InvalidTokenError as e:
        raise ValueError(f"Invalid token: {e}")

def call_openshift_api(endpoint, username, groups, method='GET', data=None):
    """Call OpenShift API with user impersonation"""
    headers = {
        'Authorization': f'Bearer {PORTAL_TOKEN}',
        'Impersonate-User': username,
    }

    # Add all groups
    for group in groups:
        headers[f'Impersonate-Group'] = group

    url = f"{OPENSHIFT_API}{endpoint}"

    if method == 'GET':
        response = requests.get(url, headers=headers, verify=False)
    elif method == 'POST':
        headers['Content-Type'] = 'application/json'
        response = requests.post(url, headers=headers, json=data, verify=False)

    return response.json()

@app.route('/api/namespaces', methods=['GET'])
def list_namespaces():
    """List namespaces for authenticated user"""
    auth_header = request.headers.get('Authorization', '')
    if not auth_header.startswith('Bearer '):
        return jsonify({'error': 'Missing or invalid authorization header'}), 401

    keycloak_token = auth_header[7:]  # Remove 'Bearer ' prefix

    try:
        # Validate JWT
        payload = validate_jwt(keycloak_token)
        username = payload['preferred_username']
        groups = payload.get('groups', [])

        # Call OpenShift API
        result = call_openshift_api('/api/v1/namespaces', username, groups)

        # Extract namespace names
        namespaces = [item['metadata']['name'] for item in result.get('items', [])]

        return jsonify({
            'user': username,
            'namespaces': namespaces
        })

    except ValueError as e:
        return jsonify({'error': str(e)}), 401
    except Exception as e:
        return jsonify({'error': f'Internal error: {str(e)}'}), 500

@app.route('/api/namespaces', methods=['POST'])
def create_namespace():
    """Create namespace for authenticated user"""
    auth_header = request.headers.get('Authorization', '')
    if not auth_header.startswith('Bearer '):
        return jsonify({'error': 'Missing or invalid authorization header'}), 401

    keycloak_token = auth_header[7:]
    namespace_name = request.json.get('name')

    if not namespace_name:
        return jsonify({'error': 'Namespace name required'}), 400

    try:
        # Validate JWT
        payload = validate_jwt(keycloak_token)
        username = payload['preferred_username']
        groups = payload.get('groups', [])

        # Prepare namespace object
        namespace_data = {
            'apiVersion': 'v1',
            'kind': 'Namespace',
            'metadata': {
                'name': namespace_name
            }
        }

        # Call OpenShift API
        result = call_openshift_api(
            '/api/v1/namespaces',
            username,
            groups,
            method='POST',
            data=namespace_data
        )

        return jsonify({
            'message': f'Namespace {namespace_name} created',
            'namespace': result
        })

    except ValueError as e:
        return jsonify({'error': str(e)}), 401
    except Exception as e:
        return jsonify({'error': f'Internal error: {str(e)}'}), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
```

**Usage**:
```bash
# Start the Flask app
export PORTAL_SA_TOKEN="<your-portal-token>"
export KEYCLOAK_PUBLIC_KEY="<keycloak-public-key>"
export KEYCLOAK_ISSUER="https://keycloak.apps.example.com/realms/maas-tenants"
export OPENSHIFT_API="https://api.example.com:6443"

python app.py

# Call the API with Keycloak JWT
curl -H "Authorization: Bearer <keycloak-jwt>" \
  http://localhost:5000/api/namespaces

# Create a namespace
curl -X POST \
  -H "Authorization: Bearer <keycloak-jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-new-project"}' \
  http://localhost:5000/api/namespaces
```

### Example 5: Testing and Validation

**Test JWT Validation**:
```bash
#!/bin/bash

KEYCLOAK_JWT="eyJhbGc..."

# Decode payload
PAYLOAD=$(echo "$KEYCLOAK_JWT" | cut -d'.' -f2)

# Fix padding
case $((${#PAYLOAD} % 4)) in
  2) PAYLOAD="${PAYLOAD}==" ;;
  3) PAYLOAD="${PAYLOAD}=" ;;
esac

# Decode and pretty print
echo "$PAYLOAD" | base64 -d | jq .
```

**Test Impersonation**:
```bash
#!/bin/bash

PORTAL_TOKEN=$(cat portal-token.txt)
USERNAME="brbaker@redhat.com"

# Test whoami
echo "=== Testing whoami ==="
curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.example.com:6443/apis/user.openshift.io/v1/users/~ \
  | jq '.metadata.name'

# Test permissions
echo "=== Testing OpenShift AI access ==="
curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.example.com:6443/apis/services.platform.opendatahub.io/v1alpha1/auths \
  | jq '.items[].spec.allowedGroups'
```

**End-to-End Test**:
```bash
#!/bin/bash

set -e

echo "=== Portal OpenShift Integration Test ==="

# 1. Get Keycloak JWT (simulated - normally from user login)
KEYCLOAK_JWT="<your-test-jwt>"

# 2. Validate JWT
echo "Step 1: Validating JWT..."
PAYLOAD=$(echo "$KEYCLOAK_JWT" | cut -d'.' -f2 | base64 -d 2>/dev/null)
USERNAME=$(echo "$PAYLOAD" | jq -r '.preferred_username')
GROUPS=$(echo "$PAYLOAD" | jq -c '.groups')
echo "  User: $USERNAME"
echo "  Groups: $GROUPS"

# 3. Load portal token
echo "Step 2: Loading portal token..."
PORTAL_TOKEN=$(cat portal-token.txt)

# 4. Test API call
echo "Step 3: Testing API call..."
RESULT=$(curl -sk \
  -H "Authorization: Bearer $PORTAL_TOKEN" \
  -H "Impersonate-User: $USERNAME" \
  -H "Impersonate-Group: rhoai-users" \
  https://api.example.com:6443/apis/user.openshift.io/v1/users/~)

echo "  API returned user: $(echo "$RESULT" | jq -r '.metadata.name')"

echo "=== Test Complete ==="
```

---

## Troubleshooting

### Common Issues

**Issue**: "Unauthorized" when calling OpenShift API
- **Cause**: Portal service account token is invalid or expired
- **Solution**: Regenerate the token:
  ```bash
  oc create token portal-sa -n default --duration=87600h > portal-token.txt
  ```

**Issue**: "Forbidden" when impersonating user
- **Cause**: Service account doesn't have impersonation rights
- **Solution**: Verify ClusterRoleBinding exists:
  ```bash
  oc get clusterrolebinding portal-impersonator
  ```

**Issue**: Groups not syncing from Keycloak
- **Cause**: Group mapper not configured correctly
- **Solution**: Verify mapper in Keycloak client configuration

**Issue**: User doesn't have access to OpenShift AI
- **Cause**: User's groups not in Auth allowedGroups
- **Solution**: Update Auth resource:
  ```bash
  oc get auth auth -o yaml
  # Verify rhoai-users is in spec.allowedGroups
  ```

### Debug Commands

```bash
# Check OAuth configuration
oc get oauth cluster -o yaml

# Check if user exists
oc get users

# Check if groups exist
oc get groups

# Check Auth configuration
oc get auth auth -o yaml

# Test impersonation
oc --as=<username> --as-group=rhoai-users whoami

# View portal service account permissions
oc describe clusterrole user-impersonator
```

---

## Security Considerations

1. **Secure Token Storage**: Store the portal service account token in a secure location (encrypted, environment variable, secret manager)

2. **JWT Validation**: Always validate JWTs for:
   - Signature verification
   - Expiration time
   - Issuer
   - Audience

3. **HTTPS Only**: Always use HTTPS for API calls in production

4. **Least Privilege**: The portal service account only has impersonation rights, not direct resource access

5. **Audit Logging**: OpenShift audit logs will show the impersonated user, maintaining accountability

6. **Token Rotation**: Regularly rotate the portal service account token

7. **Rate Limiting**: Implement rate limiting in your portal to prevent abuse

---

## Additional Resources

- [OpenShift User Impersonation](https://docs.openshift.com/container-platform/latest/authentication/impersonation.html)
- [Keycloak Client Configuration](https://www.keycloak.org/docs/latest/server_admin/#_clients)
- [OpenShift OAuth Configuration](https://docs.openshift.com/container-platform/latest/authentication/configuring-internal-oauth.html)
- [OpenShift AI Documentation](https://access.redhat.com/documentation/en-us/red_hat_openshift_ai_self-managed)
