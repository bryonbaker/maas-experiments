# MaaS Toolbox API - Usage Examples

Practical examples for using the MaaS Toolbox API. These examples demonstrate common workflows and use cases.

## Prerequisites

Set the base URL for your environment:

```bash
# Get the route URL from your cluster
ROUTE_URL=$(oc get route maas-toolbox -n maas-toolbox -o jsonpath='{.spec.host}')
export BASE_URL="https://${ROUTE_URL}"

# Or set it directly if you know your cluster domain
# export BASE_DOMAIN=<clustername>.<domain>
# export BASE_URL="https://maas-toolbox-maas-toolbox.apps.$BASE_DOMAIN"
```

---

## Basic Tier Management

### 1. Create a Free Tier

Create a tier accessible to all authenticated users:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

**Response (201 Created):**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

### 2. Create a Premium Tier

Create a tier for premium users:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "premium",
    "description": "Premium tier with enhanced features",
    "level": 10,
    "groups": ["premium-users"]
  }'
```

**Response (201 Created):**
```json
{
  "name": "premium",
  "description": "Premium tier with enhanced features",
  "level": 10,
  "groups": ["premium-users"]
}
```

### 3. Create a Dedicated Tier for a Customer

Create a tier for Acme Inc's production users:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "acme-inc-prod",
    "description": "Dedicated tier for Acme Inc production workloads",
    "level": 50,
    "groups": ["acme-prod-users", "acme-admins"]
  }'
```

**Response (201 Created):**
```json
{
  "name": "acme-inc-prod",
  "description": "Dedicated tier for Acme Inc production workloads",
  "level": 50,
  "groups": ["acme-prod-users", "acme-admins"]
}
```

### 4. List All Tiers

Retrieve all configured tiers:

```bash
curl ${BASE_URL}/api/v1/tiers
```

**Response (200 OK):**
```json
[
  {
    "name": "free",
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  },
  {
    "name": "premium",
    "description": "Premium tier with enhanced features",
    "level": 10,
    "groups": ["premium-users"]
  },
  {
    "name": "acme-inc-prod",
    "description": "Dedicated tier for Acme Inc production workloads",
    "level": 50,
    "groups": ["acme-prod-users", "acme-admins"]
  }
]
```

### 5. Get a Specific Tier

Retrieve details for a single tier:

```bash
curl ${BASE_URL}/api/v1/tiers/premium
```

**Response (200 OK):**
```json
{
  "name": "premium",
  "description": "Premium tier with enhanced features",
  "level": 10,
  "groups": ["premium-users"]
}
```

---

## Updating Tiers

### 6. Update Tier Description and Level

Update a tier's description and priority level:

```bash
curl -X PUT ${BASE_URL}/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Free tier with basic model access",
    "level": 2
  }'
```

**Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier with basic model access",
  "level": 2,
  "groups": ["system:authenticated"]
}
```

**Note:** The `groups` field is optional in the update request. If omitted, groups remain unchanged.

### 7. Replace Groups on a Tier

Replace the groups assigned to a tier:

```bash
curl -X PUT ${BASE_URL}/api/v1/tiers/premium \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Premium tier with enhanced features",
    "level": 10,
    "groups": ["premium-users", "enterprise-users"]
  }'
```

**Response (200 OK):**
```json
{
  "name": "premium",
  "description": "Premium tier with enhanced features",
  "level": 10,
  "groups": ["premium-users", "enterprise-users"]
}
```

**Important:** The PUT operation **replaces** the entire groups array. To add or remove individual groups, use the dedicated group management endpoints (see below).

### 8. Delete a Tier

Remove a tier from the configuration:

```bash
curl -X DELETE ${BASE_URL}/api/v1/tiers/acme-inc-prod
```

**Response:** `204 No Content` (empty response body)

---

## Group Management

### 9. Add a Group to a Tier

Add a single group to an existing tier without affecting other groups:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "trial-users"}'
```

**Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier with basic model access",
  "level": 2,
  "groups": ["system:authenticated", "trial-users"]
}
```

### 10. Remove a Group from a Tier

Remove a specific group from a tier:

```bash
curl -X DELETE ${BASE_URL}/api/v1/tiers/free/groups/trial-users
```

**Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier with basic model access",
  "level": 2,
  "groups": ["system:authenticated"]
}
```

**Note:** For groups with special characters (like `system:authenticated`), the colon is typically handled correctly in the URL path, but you can URL-encode if needed.

---

## Query APIs

### 11. Get All Tiers for a Group

Find which tiers a specific group has access to:

```bash
curl ${BASE_URL}/api/v1/groups/premium-users/tiers
```

**Response (200 OK):**
```json
[
  {
    "name": "premium",
    "description": "Premium tier with enhanced features",
    "level": 10,
    "groups": ["premium-users", "enterprise-users"]
  }
]
```

**Note:** Returns an empty array `[]` if no tiers contain the group.

### 12. Get All Tiers for a User

Find which tiers a user can access based on their group memberships:

```bash
curl ${BASE_URL}/api/v1/users/alice/tiers
```

**Response (200 OK):**
```json
[
  {
    "name": "free",
    "description": "Free tier with basic model access",
    "level": 2,
    "groups": ["system:authenticated"]
  },
  {
    "name": "premium",
    "description": "Premium tier with enhanced features",
    "level": 10,
    "groups": ["premium-users", "enterprise-users"]
  }
]
```

**Notes:**
- Tiers are sorted by level (ascending)
- User must exist in the cluster
- Includes the implicit `system:authenticated` group for authenticated users

---

## Health Check

### 13. Verify API is Running

```bash
curl ${BASE_URL}/health
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

## Error Handling Examples

### Duplicate Tier Creation

Attempt to create a tier that already exists:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Another free tier",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

**Response (409 Conflict):**
```json
{
  "error": "tier already exists"
}
```

### Get Non-Existent Tier

Attempt to retrieve a tier that doesn't exist:

```bash
curl ${BASE_URL}/api/v1/tiers/nonexistent
```

**Response (404 Not Found):**
```json
{
  "error": "tier not found"
}
```

### Update Tier Name (Immutable Field)

Attempt to change a tier's name:

```bash
curl -X PUT ${BASE_URL}/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "name": "new-name",
    "description": "Updated description",
    "level": 2
  }'
```

**Response (400 Bad Request):**
```json
{
  "error": "tier name cannot be changed"
}
```

### Missing Required Fields

Attempt to create a tier without description:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "incomplete",
    "level": 1
  }'
```

**Response (400 Bad Request):**
```json
{
  "error": "tier description is required"
}
```

### Add Duplicate Group

Attempt to add a group that already exists in the tier:

```bash
curl -X POST ${BASE_URL}/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "system:authenticated"}'
```

**Response (409 Conflict):**
```json
{
  "error": "group already exists in tier"
}
```

### Remove Non-Existent Group

Attempt to remove a group that isn't assigned to the tier:

```bash
curl -X DELETE ${BASE_URL}/api/v1/tiers/free/groups/nonexistent-group
```

**Response (404 Not Found):**
```json
{
  "error": "group not found in tier"
}
```

### Query User That Doesn't Exist

Attempt to get tiers for a non-existent user:

```bash
curl ${BASE_URL}/api/v1/users/unknownuser/tiers
```

**Response (404 Not Found):**
```json
{
  "error": "user 'unknownuser' not found"
}
```

---

## Complete Workflow Example

This example demonstrates a complete onboarding workflow for a new customer:

```bash
# 1. Create a dedicated tier for the customer
curl -X POST ${BASE_URL}/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "customer-xyz",
    "description": "Dedicated tier for Customer XYZ",
    "level": 40,
    "groups": ["customer-xyz-users"]
  }'

# 2. Verify the tier was created
curl ${BASE_URL}/api/v1/tiers/customer-xyz

# 3. Add an additional group for their admins
curl -X POST ${BASE_URL}/api/v1/tiers/customer-xyz/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "customer-xyz-admins"}'

# 4. Check which tiers a specific user can access
curl ${BASE_URL}/api/v1/users/john@customer-xyz.com/tiers

# 5. List all tiers that include the customer's user group
curl ${BASE_URL}/api/v1/groups/customer-xyz-users/tiers

# 6. Update the tier's priority level
curl -X PUT ${BASE_URL}/api/v1/tiers/customer-xyz \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Dedicated tier for Customer XYZ (upgraded)",
    "level": 60
  }'

# 7. View all tiers to confirm changes
curl ${BASE_URL}/api/v1/tiers
```

---

## Using with API Tools

### Postman Collection Setup

1. **Create a new collection** called "MaaS Toolbox API"
2. **Set collection variable**: `baseUrl` = your API URL
3. **Create requests** for each endpoint:
   - Tier Management: POST, GET, PUT, DELETE `/api/v1/tiers`
   - Group Management: POST, DELETE `/api/v1/tiers/:tierName/groups`
   - Query APIs: GET `/api/v1/groups/:groupName/tiers`, GET `/api/v1/users/:username/tiers`
4. **Set headers**: `Content-Type: application/json` for POST and PUT requests
5. **Use request body** examples from this document

### cURL Tips

**Save frequently used values:**
```bash
export BASE_URL="https://maas-toolbox-maas-toolbox.apps.your-cluster.com"
export TIER_NAME="free"
export GROUP_NAME="trial-users"
```

**Use variables in requests:**
```bash
curl ${BASE_URL}/api/v1/tiers/${TIER_NAME}
curl ${BASE_URL}/api/v1/groups/${GROUP_NAME}/tiers
```

**Pretty-print JSON responses:**
```bash
curl ${BASE_URL}/api/v1/tiers | jq .
```

**Include HTTP status in output:**
```bash
curl -w "\nHTTP Status: %{http_code}\n" ${BASE_URL}/api/v1/tiers
```

---

## Next Steps

- See [INTERFACE_DOCUMENTATION.md](./INTERFACE_DOCUMENTATION.md) for complete API reference
- See [OPENSHIFT-MAAS-API-GUIDE.md](./OPENSHIFT-MAAS-API-GUIDE.md) for direct API integration details
- Run the integration test suite: `./tests/test-api.sh ${BASE_URL}`
