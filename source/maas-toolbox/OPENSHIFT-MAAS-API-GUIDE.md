# OpenShift and MaaS Direct API Integration Guide

This document describes how to interact directly with OpenShift and MaaS APIs to implement the functionality provided by maas-toolbox. This enables customers to implement their own version of the toolbox by understanding the underlying API flows.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Authentication](#authentication)
3. [Tier Management via ConfigMap](#tier-management-via-configmap)
4. [Group Management via OpenShift API](#group-management-via-openshift-api)
5. [User and Group Queries](#user-and-group-queries)
6. [LLMInferenceService Management](#llminferenceservice-management)
7. [TokenRateLimitPolicy Management](#tokenratelimitpolicy-management)
8. [RateLimitPolicy Management](#ratelimitpolicy-management)
9. [Complete Workflow Examples](#complete-workflow-examples)

---

## Prerequisites

### Set Environment Variables

```bash
# OpenShift API Server
export OPENSHIFT_API="https://api.your-cluster.example.com:6443"

# Authentication token (see Authentication section)
export TOKEN="your-bearer-token-here"

# ConfigMap settings for tier management
export TIER_NAMESPACE="maas-api"
export TIER_CONFIGMAP="tier-to-group-mapping"

# Rate limit policy settings
export RATELIMIT_NAMESPACE="openshift-ingress"
export RATELIMIT_POLICY_NAME="gateway-rate-limits"
export TOKEN_RATELIMIT_POLICY_NAME="gateway-token-rate-limits"
```

---

## Authentication

OpenShift uses bearer token authentication. You can obtain a token in several ways:

### Option 1: Use an existing login session

```bash
# Login to OpenShift
oc login https://api.your-cluster.example.com:6443

# Extract the token
export TOKEN=$(oc whoami -t)
```

### Option 2: Service Account Token (for automated systems)

```bash
# Create a service account
oc create serviceaccount maas-admin -n maas-api

# Grant necessary permissions
oc adm policy add-cluster-role-to-user cluster-admin -z maas-admin -n maas-api

# Get the service account token
export TOKEN=$(oc create token maas-admin -n maas-api)
```

### Verify Authentication

```bash
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces"
```

**Expected Response:** A JSON list of namespaces if authentication is successful.

---

## Tier Management via ConfigMap

Tiers are stored in a Kubernetes ConfigMap in the `maas-api` namespace. The ConfigMap contains a YAML list of tier definitions.

### ConfigMap Structure

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tier-to-group-mapping
  namespace: maas-api
data:
  tiers: |
    - name: free
      description: Free tier for basic users
      level: 1
      groups:
        - system:authenticated
    - name: premium
      description: Premium tier
      level: 10
      groups:
        - premium-users
```

### Scenario 1: Create Initial ConfigMap with Tiers

```bash
# Create the ConfigMap with initial tier configuration
curl -k -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps" \
  -d '{
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "tier-to-group-mapping",
      "namespace": "maas-api",
      "labels": {
        "app": "tier-to-group-admin"
      }
    },
    "data": {
      "tiers": "- name: free\n  description: Free tier for basic users\n  level: 1\n  groups:\n    - system:authenticated\n"
    }
  }'
```

**Expected Response (201 Created):** The created ConfigMap object in JSON format.

### Scenario 2: Get Existing Tier Configuration

```bash
# Get the ConfigMap
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}"
```

**Expected Response (200 OK):** ConfigMap object with the tier data.

**Extract and parse the tiers:**

```bash
# Get ConfigMap and extract tiers field
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers'
```

**Output:** The YAML content of the tiers list.

### Scenario 3: Add a New Tier to Existing Configuration

This requires:
1. Get the current ConfigMap
2. Parse the existing tiers
3. Add the new tier to the list
4. Update the ConfigMap

```bash
# Step 1: Get current ConfigMap
CURRENT_CM=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}")

# Step 2: Extract current tiers YAML
CURRENT_TIERS=$(echo "$CURRENT_CM" | jq -r '.data.tiers')

# Step 3: Append new tier to YAML (manual or programmatic)
# For this example, we'll construct new YAML with the additional tier
NEW_TIERS="${CURRENT_TIERS}
- name: premium
  description: Premium tier
  level: 10
  groups:
    - premium-users
"

# Step 4: Update the ConfigMap with PATCH
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

**Expected Response (200 OK):** Updated ConfigMap object.

**Note:** In practice, you would use a YAML parser (like `yq`) to properly manipulate the YAML structure:

```bash
# Using yq to add a tier
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Add new tier using yq
NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval '. += [{"name": "premium", "description": "Premium tier", "level": 10, "groups": ["premium-users"]}]' -)

# Update ConfigMap
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

### Scenario 4: Update a Tier's Properties

```bash
# Get current tiers
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Update tier using yq (example: update "free" tier's level to 2)
NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval '(.[] | select(.name == "free") | .level) = 2' -)

# Update ConfigMap
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

### Scenario 5: Delete a Tier

```bash
# Get current tiers
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Remove tier using yq (example: remove "free" tier)
NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval 'del(.[] | select(.name == "free"))' -)

# Update ConfigMap
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

### Scenario 6: Add a Group to a Tier

```bash
# Get current tiers
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Add group to tier using yq (example: add "trial-users" to "free" tier)
NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval '(.[] | select(.name == "free") | .groups) += ["trial-users"]' -)

# Update ConfigMap
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

### Scenario 7: Remove a Group from a Tier

```bash
# Get current tiers
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Remove group from tier using yq (example: remove "trial-users" from "free" tier)
NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval '(.[] | select(.name == "free") | .groups) -= ["trial-users"]' -)

# Update ConfigMap
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"
```

---

## Group Management via OpenShift API

OpenShift Groups are cluster-scoped resources in the `user.openshift.io/v1` API.

### Scenario 1: List All Groups in the Cluster

```bash
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups"
```

**Expected Response (200 OK):** JSON list of all groups with their members.

**Parse to get group names and users:**

```bash
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  | jq -r '.items[] | "\(.metadata.name): \(.users | join(", "))"'
```

### Scenario 2: Check if a Specific Group Exists

```bash
GROUP_NAME="premium-users"

curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups/${GROUP_NAME}"
```

**Expected Response:**
- **200 OK:** Group exists (returns group object)
- **404 Not Found:** Group does not exist

### Scenario 3: Get Groups for a Specific User

```bash
USERNAME="bryonbaker"

# List all groups and filter for the user
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  | jq -r --arg user "$USERNAME" '.items[] | select(.users and (.users[] == $user)) | .metadata.name'
```

**Output:** List of group names that the user belongs to.

**Note:** The special group `system:authenticated` is always implicitly assigned to authenticated users but won't appear in the API response.

### Scenario 4: Create a New Group

```bash
curl -k -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  -d '{
    "apiVersion": "user.openshift.io/v1",
    "kind": "Group",
    "metadata": {
      "name": "premium-users"
    },
    "users": [
      "bryonbaker",
      "johndoe"
    ]
  }'
```

**Expected Response (201 Created):** The created Group object.

### Scenario 5: Add a User to an Existing Group

This requires getting the group, modifying the users list, and updating:

```bash
GROUP_NAME="premium-users"
NEW_USER="janedoe"

# Step 1: Get current group
CURRENT_GROUP=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups/${GROUP_NAME}")

# Step 2: Add user to the users array
UPDATED_GROUP=$(echo "$CURRENT_GROUP" | jq --arg user "$NEW_USER" '.users += [$user] | .users |= unique')

# Step 3: Update the group
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups/${GROUP_NAME}" \
  -d "$UPDATED_GROUP"
```

**Expected Response (200 OK):** Updated group object.

---

## User and Group Queries

### Scenario 1: Get All Tiers a User Has Access To

This combines getting user groups with tier configuration:

```bash
USERNAME="bryonbaker"

# Step 1: Get all groups the user belongs to
USER_GROUPS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  | jq -r --arg user "$USERNAME" '.items[] | select(.users and (.users[] == $user)) | .metadata.name')

# Add system:authenticated (implicit for all authenticated users)
USER_GROUPS="${USER_GROUPS}
system:authenticated"

# Step 2: Get tier configuration
TIER_CONFIG=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Step 3: Filter tiers where user's groups intersect with tier's groups
# (This requires scripting logic - example with yq and bash)
echo "$TIER_CONFIG" | yq eval -o=json '.' | jq --arg groups "$USER_GROUPS" '
  .[] | select(
    .groups as $tier_groups | 
    ($groups | split("\n")) as $user_groups |
    ($tier_groups | map(. as $g | $user_groups | index($g)) | any)
  )'
```

**Output:** JSON array of tiers the user can access.

### Scenario 2: Get All Tiers Associated with a Specific Group

```bash
GROUP_NAME="premium-users"

# Get tier configuration
TIER_CONFIG=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

# Filter tiers that contain the group
echo "$TIER_CONFIG" | yq eval -o=json '.' | jq --arg group "$GROUP_NAME" '
  .[] | select(.groups | index($group))'
```

**Output:** JSON array of tiers containing the specified group.

---

## LLMInferenceService Management

LLMInferenceService resources are managed via the KServe API (`serving.kserve.io/v1alpha1`).

### Scenario 1: List All LLMInferenceServices Across All Namespaces

```bash
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/llminferenceservices"
```

**Expected Response (200 OK):** JSON list of all LLMInferenceService resources.

### Scenario 2: Get a Specific LLMInferenceService

```bash
NAMESPACE="acme-inc-models"
SERVICE_NAME="acme-dev-model"

curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}"
```

**Expected Response (200 OK):** The LLMInferenceService object.

### Scenario 3: Annotate LLMInferenceService with a Tier

This involves adding/updating the `serving.kserve.io/tiers` annotation:

```bash
NAMESPACE="acme-inc-models"
SERVICE_NAME="acme-dev-model"
TIER_NAME="free"

# Step 1: Get current LLMInferenceService
CURRENT_SERVICE=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}")

# Step 2: Extract current tiers annotation (if exists)
CURRENT_TIERS=$(echo "$CURRENT_SERVICE" | jq -r '.metadata.annotations["serving.kserve.io/tiers"] // "[]"')

# Step 3: Add new tier to the list (avoiding duplicates)
NEW_TIERS=$(echo "$CURRENT_TIERS" | jq --arg tier "$TIER_NAME" '. + [$tier] | unique')

# Step 4: Update the annotation using PATCH
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/merge-patch+json" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}" \
  -d "{
    \"metadata\": {
      \"annotations\": {
        \"serving.kserve.io/tiers\": $(echo "$NEW_TIERS" | jq -c .)
      }
    }
  }"
```

**Expected Response (200 OK):** Updated LLMInferenceService object.

### Scenario 4: Remove a Tier from LLMInferenceService Annotation

```bash
NAMESPACE="acme-inc-models"
SERVICE_NAME="acme-dev-model"
TIER_NAME="free"

# Step 1: Get current service
CURRENT_SERVICE=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}")

# Step 2: Extract and remove tier from annotation
CURRENT_TIERS=$(echo "$CURRENT_SERVICE" | jq -r '.metadata.annotations["serving.kserve.io/tiers"] // "[]"')
NEW_TIERS=$(echo "$CURRENT_TIERS" | jq --arg tier "$TIER_NAME" 'del(.[] | select(. == $tier))')

# Step 3: Update annotation
curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/merge-patch+json" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}" \
  -d "{
    \"metadata\": {
      \"annotations\": {
        \"serving.kserve.io/tiers\": $(echo "$NEW_TIERS" | jq -c .)
      }
    }
  }"
```

### Scenario 5: Find All LLMInferenceServices for a Specific Tier

```bash
TIER_NAME="free"

# Get all LLMInferenceServices
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/llminferenceservices" \
  | jq --arg tier "$TIER_NAME" '.items[] | select(
      .metadata.annotations["serving.kserve.io/tiers"] and 
      (.metadata.annotations["serving.kserve.io/tiers"] | fromjson | index($tier))
    ) | {name: .metadata.name, namespace: .metadata.namespace, tiers: (.metadata.annotations["serving.kserve.io/tiers"] | fromjson)}'
```

**Output:** JSON array of LLMInferenceServices that have the specified tier.

---

## TokenRateLimitPolicy Management

TokenRateLimitPolicy is a Kuadrant CRD (`kuadrant.io/v1alpha1`) that manages token consumption rate limits.

### Policy Structure

```yaml
apiVersion: kuadrant.io/v1alpha1
kind: TokenRateLimitPolicy
metadata:
  name: gateway-token-rate-limits
  namespace: openshift-ingress
spec:
  limits:
    serverless-user-tokens:
      counters:
        - expression: auth.identity.userid
      rates:
        - limit: 100
          window: 5m
      when:
        - predicate: |
            auth.identity.tier == "serverless" && !request.path.endsWith("/v1/models")
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: maas-default-gateway
```

### Scenario 1: Get Current TokenRateLimitPolicy

```bash
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}"
```

**Expected Response (200 OK):** TokenRateLimitPolicy object.

### Scenario 2: Add a New Token Rate Limit Entry

```bash
# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}")

# Step 2: Add new limit to spec.limits map
LIMIT_NAME="free-user-tokens"
TIER_NAME="free"
LIMIT_VALUE=100
WINDOW="1m"

UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" \
  --arg tier "$TIER_NAME" \
  --argjson limit $LIMIT_VALUE \
  --arg window "$WINDOW" \
  '.spec.limits[$name] = {
    "counters": [{"expression": "auth.identity.userid"}],
    "rates": [{"limit": $limit, "window": $window}],
    "when": [{"predicate": "auth.identity.tier == \"" + $tier + "\" && !request.path.endsWith(\"/v1/models\")"}]
  }')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

**Expected Response (200 OK):** Updated TokenRateLimitPolicy object.

### Scenario 3: Update an Existing Token Rate Limit

```bash
LIMIT_NAME="free-user-tokens"
NEW_LIMIT_VALUE=200
NEW_WINDOW="2m"

# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}")

# Step 2: Update the limit values (keep tier from predicate)
UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" \
  --argjson limit $NEW_LIMIT_VALUE \
  --arg window "$NEW_WINDOW" \
  '.spec.limits[$name].rates[0].limit = $limit |
   .spec.limits[$name].rates[0].window = $window')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

### Scenario 4: Delete a Token Rate Limit Entry

```bash
LIMIT_NAME="free-user-tokens"

# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}")

# Step 2: Remove the limit from spec.limits map
UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" 'del(.spec.limits[$name])')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

### Scenario 5: List All Token Rate Limits

```bash
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
  | jq '.spec.limits | to_entries[] | {
      name: .key,
      limit: .value.rates[0].limit,
      window: .value.rates[0].window,
      tier: (.value.when[0].predicate | capture("tier == \"(?<tier>[^\"]+)\"").tier)
    }'
```

**Output:** JSON array of all token rate limits with their settings.

---

## RateLimitPolicy Management

RateLimitPolicy is a Kuadrant CRD (`kuadrant.io/v1`) that manages request rate limits.

### Policy Structure

```yaml
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
metadata:
  name: gateway-rate-limits
  namespace: openshift-ingress
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: maas-default-gateway
  limits:
    serverless:
      rates:
        - limit: 5
          window: 2m
      when:
        - predicate: |
            auth.identity.tier == "serverless"
      counters:
        - expression: auth.identity.userid
```

### Scenario 1: Get Current RateLimitPolicy

```bash
curl -k -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}"
```

**Expected Response (200 OK):** RateLimitPolicy object.

### Scenario 2: Add a New Request Rate Limit Entry

```bash
# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}")

# Step 2: Add new limit to spec.limits map
LIMIT_NAME="free"
TIER_NAME="free"
LIMIT_VALUE=5
WINDOW="2m"

UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" \
  --arg tier "$TIER_NAME" \
  --argjson limit $LIMIT_VALUE \
  --arg window "$WINDOW" \
  '.spec.limits[$name] = {
    "counters": [{"expression": "auth.identity.userid"}],
    "rates": [{"limit": $limit, "window": $window}],
    "when": [{"predicate": "auth.identity.tier == \"" + $tier + "\""}]
  }')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

**Expected Response (200 OK):** Updated RateLimitPolicy object.

### Scenario 3: Update an Existing Request Rate Limit

```bash
LIMIT_NAME="free"
NEW_LIMIT_VALUE=10
NEW_WINDOW="3m"

# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}")

# Step 2: Update the limit values
UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" \
  --argjson limit $NEW_LIMIT_VALUE \
  --arg window "$NEW_WINDOW" \
  '.spec.limits[$name].rates[0].limit = $limit |
   .spec.limits[$name].rates[0].window = $window')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

### Scenario 4: Delete a Request Rate Limit Entry

```bash
LIMIT_NAME="free"

# Step 1: Get current policy
CURRENT_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}")

# Step 2: Remove the limit from spec.limits map
UPDATED_POLICY=$(echo "$CURRENT_POLICY" | jq --arg name "$LIMIT_NAME" 'del(.spec.limits[$name])')

# Step 3: Update the policy
curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_POLICY"
```

### Scenario 5: List All Request Rate Limits

```bash
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  | jq '.spec.limits | to_entries[] | {
      name: .key,
      limit: .value.rates[0].limit,
      window: .value.rates[0].window,
      tier: (.value.when[0].predicate | capture("tier == \"(?<tier>[^\"]+)\"").tier)
    }'
```

**Output:** JSON array of all request rate limits with their settings.

---

## Complete Workflow Examples

### Workflow 1: Onboard a New Customer with Dedicated Tier

This workflow demonstrates creating a complete tier configuration for a new customer:

```bash
# Variables
CUSTOMER_NAME="acme-inc"
CUSTOMER_TIER="${CUSTOMER_NAME}-dedicated"
CUSTOMER_GROUP="${CUSTOMER_NAME}-users"
CUSTOMER_USERS=("alice" "bob" "charlie")

# Step 1: Create OpenShift Group for the customer
echo "Creating group: ${CUSTOMER_GROUP}"
curl -k -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  -d "{
    \"apiVersion\": \"user.openshift.io/v1\",
    \"kind\": \"Group\",
    \"metadata\": {
      \"name\": \"${CUSTOMER_GROUP}\"
    },
    \"users\": $(printf '%s\n' "${CUSTOMER_USERS[@]}" | jq -R . | jq -s .)
  }"

echo "Group created successfully."

# Step 2: Add tier to ConfigMap
echo "Adding tier: ${CUSTOMER_TIER}"
CURRENT_TIERS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

NEW_TIERS=$(echo "$CURRENT_TIERS" | yq eval ". += [{
  \"name\": \"${CUSTOMER_TIER}\",
  \"description\": \"Tier for ${CUSTOMER_NAME}'s dedicated models\",
  \"level\": 50,
  \"groups\": [\"${CUSTOMER_GROUP}\"]
}]" -)

curl -k -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/strategic-merge-patch+json" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  -d "{
    \"data\": {
      \"tiers\": $(echo "$NEW_TIERS" | jq -Rs .)
    }
  }"

echo "Tier added to ConfigMap."

# Step 3: Create TokenRateLimitPolicy entry
echo "Creating token rate limit for ${CUSTOMER_TIER}"
CURRENT_TOKEN_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}")

UPDATED_TOKEN_POLICY=$(echo "$CURRENT_TOKEN_POLICY" | jq \
  --arg name "${CUSTOMER_TIER}-tokens" \
  --arg tier "${CUSTOMER_TIER}" \
  '.spec.limits[$name] = {
    "counters": [{"expression": "auth.identity.userid"}],
    "rates": [{"limit": 10000, "window": "1h"}],
    "when": [{"predicate": "auth.identity.tier == \"" + $tier + "\" && !request.path.endsWith(\"/v1/models\")"}]
  }')

curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_TOKEN_POLICY"

echo "Token rate limit created."

# Step 4: Create RateLimitPolicy entry
echo "Creating request rate limit for ${CUSTOMER_TIER}"
CURRENT_RATE_POLICY=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}")

UPDATED_RATE_POLICY=$(echo "$CURRENT_RATE_POLICY" | jq \
  --arg name "${CUSTOMER_TIER}" \
  --arg tier "${CUSTOMER_TIER}" \
  '.spec.limits[$name] = {
    "counters": [{"expression": "auth.identity.userid"}],
    "rates": [{"limit": 100, "window": "1m"}],
    "when": [{"predicate": "auth.identity.tier == \"" + $tier + "\""}]
  }')

curl -k -X PUT \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  -d "$UPDATED_RATE_POLICY"

echo "Request rate limit created."
echo "Customer onboarding complete for ${CUSTOMER_NAME}!"
```

**Output:** A complete tier configuration with group, tier mapping, and rate limits.

### Workflow 2: Annotate Multiple LLMInferenceServices for a Tier

```bash
TIER_NAME="premium"
SERVICES=(
  "namespace1:service1"
  "namespace1:service2"
  "namespace2:service3"
)

echo "Annotating services with tier: ${TIER_NAME}"

for SERVICE_SPEC in "${SERVICES[@]}"; do
  IFS=':' read -r NAMESPACE SERVICE_NAME <<< "$SERVICE_SPEC"
  
  echo "Processing ${NAMESPACE}/${SERVICE_NAME}..."
  
  # Get current service
  CURRENT_SERVICE=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
    "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}")
  
  # Extract current tiers
  CURRENT_TIERS=$(echo "$CURRENT_SERVICE" | jq -r '.metadata.annotations["serving.kserve.io/tiers"] // "[]"')
  
  # Add tier
  NEW_TIERS=$(echo "$CURRENT_TIERS" | jq --arg tier "$TIER_NAME" '. + [$tier] | unique')
  
  # Update annotation
  curl -k -X PATCH \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/merge-patch+json" \
    "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/namespaces/${NAMESPACE}/llminferenceservices/${SERVICE_NAME}" \
    -d "{
      \"metadata\": {
        \"annotations\": {
          \"serving.kserve.io/tiers\": $(echo "$NEW_TIERS" | jq -c .)
        }
      }
    }"
  
  echo "✓ ${NAMESPACE}/${SERVICE_NAME} annotated"
done

echo "All services annotated with tier ${TIER_NAME}!"
```

### Workflow 3: Get Complete View of User's Access

This workflow demonstrates getting all relevant information for a user:

```bash
USERNAME="alice"

echo "=== User Access Report for ${USERNAME} ==="

# Step 1: Get user's groups
echo -e "\n1. User Groups:"
USER_GROUPS=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/user.openshift.io/v1/groups" \
  | jq -r --arg user "$USERNAME" '.items[] | select(.users and (.users[] == $user)) | .metadata.name')

echo "$USER_GROUPS"
echo "system:authenticated (implicit)"

# Combine groups
ALL_USER_GROUPS="${USER_GROUPS}
system:authenticated"

# Step 2: Get tiers user has access to
echo -e "\n2. Accessible Tiers:"
TIER_CONFIG=$(curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/api/v1/namespaces/${TIER_NAMESPACE}/configmaps/${TIER_CONFIGMAP}" \
  | jq -r '.data.tiers')

USER_TIERS=$(echo "$TIER_CONFIG" | yq eval -o=json '.' | jq --arg groups "$ALL_USER_GROUPS" '
  .[] | select(
    .groups as $tier_groups | 
    ($groups | split("\n")) as $user_groups |
    ($tier_groups | map(. as $g | $user_groups | index($g)) | any)
  )')

echo "$USER_TIERS" | jq -r '"\(.name) (level \(.level)): \(.description)"'

# Step 3: Get rate limits for user's tiers
echo -e "\n3. Rate Limits:"
echo "$USER_TIERS" | jq -r '.name' | while read -r tier; do
  echo "Tier: $tier"
  
  # Token rate limits
  curl -k -s -H "Authorization: Bearer ${TOKEN}" \
    "${OPENSHIFT_API}/apis/kuadrant.io/v1alpha1/namespaces/${RATELIMIT_NAMESPACE}/tokenratelimitpolicies/${TOKEN_RATELIMIT_POLICY_NAME}" \
    | jq --arg tier "$tier" '.spec.limits | to_entries[] | select(.value.when[0].predicate | contains($tier)) | 
        "  Token Limit: \(.value.rates[0].limit) tokens per \(.value.rates[0].window)"' -r
  
  # Request rate limits
  curl -k -s -H "Authorization: Bearer ${TOKEN}" \
    "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
    | jq --arg tier "$tier" '.spec.limits | to_entries[] | select(.value.when[0].predicate | contains($tier)) | 
        "  Request Limit: \(.value.rates[0].limit) requests per \(.value.rates[0].window)"' -r
done

# Step 4: Get accessible LLMInferenceServices
echo -e "\n4. Accessible LLMInferenceServices:"
echo "$USER_TIERS" | jq -r '.name' | while read -r tier; do
  curl -k -s -H "Authorization: Bearer ${TOKEN}" \
    "${OPENSHIFT_API}/apis/serving.kserve.io/v1alpha1/llminferenceservices" \
    | jq --arg tier "$tier" '.items[] | select(
        .metadata.annotations["serving.kserve.io/tiers"] and 
        (.metadata.annotations["serving.kserve.io/tiers"] | fromjson | index($tier))
      ) | "  - \(.metadata.namespace)/\(.metadata.name) (tiers: \(.metadata.annotations["serving.kserve.io/tiers"]))"' -r
done

echo -e "\n=== End of Report ==="
```

---

## Important Notes

### 1. URL Encoding

When using group names with special characters (like `system:authenticated`), you may need URL encoding in some contexts:
- Colon (`:`) → `%3A`
- Space (` `) → `%20`

### 2. YAML vs JSON Handling

- ConfigMap tiers data is stored as YAML string
- Most OpenShift API operations use JSON
- Tools like `yq` and `jq` are essential for conversions

### 3. Error Handling

Always check HTTP response codes:
- `200 OK`: Success
- `201 Created`: Resource created
- `404 Not Found`: Resource doesn't exist
- `409 Conflict`: Resource already exists
- `401 Unauthorized`: Authentication failure
- `403 Forbidden`: Authorization failure

### 4. Rate Limit Policy Resyncing

After modifying RateLimitPolicy or TokenRateLimitPolicy CRDs, the policies may need time to propagate through the system. Monitor the policy status:

```bash
curl -k -s -H "Authorization: Bearer ${TOKEN}" \
  "${OPENSHIFT_API}/apis/kuadrant.io/v1/namespaces/${RATELIMIT_NAMESPACE}/ratelimitpolicies/${RATELIMIT_POLICY_NAME}" \
  | jq '.status'
```

### 5. Best Practices

1. **Idempotency**: Always check if a resource exists before creating
2. **Atomic Updates**: Use GET-MODIFY-PUT pattern for updates
3. **Error Recovery**: Implement retry logic for transient failures
4. **Validation**: Validate data before sending to API
5. **Logging**: Log all API operations for audit trail

---

## Required Tools

- `curl`: HTTP client
- `jq`: JSON processor
- `yq`: YAML processor (github.com/mikefarah/yq)
- `oc` or `kubectl`: For obtaining authentication tokens

---

## Additional Resources

- [OpenShift REST API Documentation](https://docs.openshift.com/container-platform/latest/rest_api/index.html)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [KServe LLMInferenceService API](https://github.com/kserve/kserve/tree/master/docs/apis)
- [Kuadrant Rate Limiting](https://docs.kuadrant.io/kuadrant-operator/doc/rate-limiting/)
- [ODH MaaS Documentation](https://opendatahub-io.github.io/models-as-a-service/)

---

**Last Updated:** February 2026  
**Version:** 1.0  
**Author:** Bryon Baker
