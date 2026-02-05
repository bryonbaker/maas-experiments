# MaaS Toolbox API Interface Documentation

Complete reference for all MaaS Toolbox REST API endpoints, including request formats, response structures, and usage examples.

## Table of Contents

1. [Base URL and Authentication](#base-url-and-authentication)
2. [Tier Management API](#tier-management-api)
3. [Group Query API](#group-query-api)
4. [User Query API](#user-query-api)
5. [Configuration Details](#configuration-details)
6. [Business Rules](#business-rules)
7. [Error Responses](#error-responses)

---

## Base URL and Authentication

### Getting the API URL

After deployment, retrieve the route URL:

```bash
ROUTE_URL=$(oc get route maas-toolbox -n maas-toolbox -o jsonpath='{.spec.host}')
echo "API URL: https://$ROUTE_URL"
```

### Authentication

Currently, MaaS Toolbox relies on OpenShift Route authentication. Future versions will include token-based authentication.

### Health Check

```bash
curl https://$ROUTE_URL/health
```

**Response:**
```json
{
  "status": "ok"
}
```

---

## Tier Management API

Base path: `/api/v1/tiers`

### Create a Tier

**Endpoint:** `POST /api/v1/tiers`

**Description:** Create a new tier with specified name, description, level, and associated groups.

**Request Body:**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

**Request Fields:**
- `name` (string, required): Unique tier identifier. Cannot be changed after creation.
- `description` (string, required): Human-readable description of the tier.
- `level` (integer, optional): Priority level for tier selection. Default: 0. Must be non-negative.
- `groups` (array of strings, required): List of Kubernetes/OpenShift group names that have access to this tier.

**Example Request:**
```bash
curl -X POST https://$ROUTE_URL/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

**Success Response (201 Created):**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

**Error Responses:**
- `400 Bad Request`: Validation error (missing required fields, invalid data)
- `409 Conflict`: Tier with this name already exists

---

### List All Tiers

**Endpoint:** `GET /api/v1/tiers`

**Description:** Retrieve all configured tiers.

**Example Request:**
```bash
curl https://$ROUTE_URL/api/v1/tiers
```

**Success Response (200 OK):**
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
    "description": "Premium tier",
    "level": 10,
    "groups": ["premium-users"]
  }
]
```

---

### Get a Specific Tier

**Endpoint:** `GET /api/v1/tiers/{name}`

**Description:** Retrieve details for a specific tier by name.

**Path Parameters:**
- `name` (string): The tier name

**Example Request:**
```bash
curl https://$ROUTE_URL/api/v1/tiers/free
```

**Success Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

**Error Responses:**
- `404 Not Found`: Tier does not exist

---

### Update a Tier

**Endpoint:** `PUT /api/v1/tiers/{name}`

**Description:** Update an existing tier's description, level, and groups. The tier name cannot be changed.

**Path Parameters:**
- `name` (string): The tier name to update

**Request Body:**
```json
{
  "description": "Updated free tier description",
  "level": 2,
  "groups": ["system:authenticated", "free-users"]
}
```

**Request Fields:**
- `description` (string, optional): Updated description
- `level` (integer, optional): Updated priority level
- `groups` (array of strings, optional): Updated group list

**Note:** The `name` field cannot be changed. Only `description`, `level`, and `groups` can be updated.

**Example Request:**
```bash
curl -X PUT https://$ROUTE_URL/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated free tier description",
    "level": 2,
    "groups": ["system:authenticated", "free-users"]
  }'
```

**Success Response (200 OK):**
```json
{
  "name": "free",
  "description": "Updated free tier description",
  "level": 2,
  "groups": ["system:authenticated", "free-users"]
}
```

**Error Responses:**
- `400 Bad Request`: Validation error
- `404 Not Found`: Tier does not exist

---

### Delete a Tier

**Endpoint:** `DELETE /api/v1/tiers/{name}`

**Description:** Remove a tier from the configuration.

**Path Parameters:**
- `name` (string): The tier name to delete

**Example Request:**
```bash
curl -X DELETE https://$ROUTE_URL/api/v1/tiers/free
```

**Success Response (204 No Content):**
No response body.

**Error Responses:**
- `404 Not Found`: Tier does not exist
- `500 Internal Server Error`: Failed to delete tier

---

### Add a Group to a Tier

**Endpoint:** `POST /api/v1/tiers/{name}/groups`

**Description:** Add a Kubernetes/OpenShift group to an existing tier.

**Path Parameters:**
- `name` (string): The tier name

**Request Body:**
```json
{
  "group": "new-group"
}
```

**Example Request:**
```bash
curl -X POST https://$ROUTE_URL/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "new-group"}'
```

**Success Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated", "new-group"]
}
```

**Error Responses:**
- `400 Bad Request`: Invalid group name or group already exists in tier
- `404 Not Found`: Tier does not exist

---

### Remove a Group from a Tier

**Endpoint:** `DELETE /api/v1/tiers/{name}/groups/{group}`

**Description:** Remove a group from a tier's access list.

**Path Parameters:**
- `name` (string): The tier name
- `group` (string): The group name to remove (URL-encoded if contains special characters)

**Example Request:**
```bash
# For group with special characters like "system:authenticated"
curl -X DELETE https://$ROUTE_URL/api/v1/tiers/free/groups/system:authenticated
```

**Success Response (200 OK):**
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": []
}
```

**Error Responses:**
- `404 Not Found`: Tier or group does not exist in tier

---

## Group Query API

Base path: `/api/v1/groups`

### Get Tiers by Group

**Endpoint:** `GET /api/v1/groups/{group}/tiers`

**Description:** Retrieve all tiers that contain a specific Kubernetes/OpenShift group.

**Path Parameters:**
- `group` (string): The group name (URL-encoded if contains special characters)

**Example Request:**
```bash
curl https://$ROUTE_URL/api/v1/groups/premium-users/tiers
```

**Success Response (200 OK):**
```json
[
  {
    "name": "premium",
    "description": "Premium tier with high priority",
    "level": 10,
    "groups": ["premium-users", "cluster-admins"]
  }
]
```

**Notes:**
- Returns an empty array `[]` if no tiers contain the specified group
- Always returns `200 OK` even if the array is empty

---

## User Query API

Base path: `/api/v1/users`

### Get Tiers for User

**Endpoint:** `GET /api/v1/users/{username}/tiers`

**Description:** Retrieve all tiers that a user has access to based on their group memberships. Results are sorted by tier level (priority) in ascending order.

**Path Parameters:**
- `username` (string): The username to lookup

**Example Request:**
```bash
curl https://$ROUTE_URL/api/v1/users/bryonbaker/tiers
```

**Success Response (200 OK):**
```json
[
  {
    "name": "standard",
    "description": "Standard tier for authenticated users",
    "level": 5,
    "groups": ["system:authenticated"]
  },
  {
    "name": "premium",
    "description": "Premium tier with high priority",
    "level": 10,
    "groups": ["cluster-admins", "premium-users"]
  }
]
```

**Response Details:**
- Each tier includes the complete tier configuration
- Tiers are sorted by `level` (ascending)
- Only includes tiers where the user belongs to at least one of the tier's groups
- Automatically includes the implicit `system:authenticated` group for all authenticated users

**Error Responses:**
- `404 Not Found`: User does not exist in the cluster
- `500 Internal Server Error`: Failed to query user or groups

---

## Configuration Details

### ConfigMap Storage

Tier configuration is stored in a Kubernetes ConfigMap following the [official ODH MaaS format](https://opendatahub-io.github.io/models-as-a-service/latest/configuration-and-management/tier-overview/).

**ConfigMap Structure:**
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

### Group Validation

By default, MaaS Toolbox validates that groups exist in the cluster before creating or updating tiers. This validation can be disabled for environments where groups are managed externally.

**When to disable group validation:**
- Groups are managed in external identity providers (Keycloak, LDAP, Active Directory)
- Groups are synced to OpenShift/Kubernetes but not as Group custom resources
- You want to pre-configure tiers before groups exist in the cluster

**To disable validation:**
Set the `VALIDATE_GROUPS` environment variable to `no` in the deployment configuration.

**Note:** User-to-tier lookups (`/api/v1/users/{username}/tiers`) always use runtime group membership from the Kubernetes API, regardless of this setting.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | `maas-api` | Kubernetes namespace where the ConfigMap is stored |
| `CONFIGMAP_NAME` | `tier-to-group-mapping` | Name of the ConfigMap containing tier configuration |
| `PORT` | `8080` | HTTP server port |
| `VALIDATE_GROUPS` | `yes` | Validate groups exist before creating/updating tiers. Set to `no` when using external identity providers |

---

## Business Rules

### Tier Name Constraints

1. **Immutable**: Tier name is set at creation time and cannot be changed
2. **Unique**: Tier names must be unique across all tiers
3. **Required**: Name field is mandatory for tier creation

### Tier Properties

1. **Description**: Required field, can be updated
2. **Level**: Optional, defaults to 0, must be non-negative integer
3. **Groups**: Required array, must contain at least one group (unless validation is disabled)

### Group Management

1. **Group Existence**: By default, groups must exist in the cluster (can be disabled with `VALIDATE_GROUPS=no`)
2. **Duplicates**: Adding a group that already exists in a tier returns an error
3. **Special Groups**: The group `system:authenticated` is implicitly available to all authenticated users

---

## Error Responses

All error responses follow this JSON format:

```json
{
  "error": "error message describing what went wrong"
}
```

### HTTP Status Codes

| Status Code | Meaning |
|------------|---------|
| `200 OK` | Request successful |
| `201 Created` | Resource created successfully |
| `204 No Content` | Resource deleted successfully (no response body) |
| `400 Bad Request` | Validation error or invalid request data |
| `404 Not Found` | Resource not found |
| `409 Conflict` | Resource already exists (duplicate) |
| `500 Internal Server Error` | Server error during processing |

### Common Error Scenarios

#### 400 Bad Request

**Missing Required Fields:**
```json
{
  "error": "name and description are required"
}
```

**Invalid Group:**
```json
{
  "error": "group 'invalid-group' does not exist in the cluster"
}
```

**Duplicate Group:**
```json
{
  "error": "group already exists in tier"
}
```

#### 404 Not Found

**Tier Not Found:**
```json
{
  "error": "tier not found"
}
```

**User Not Found:**
```json
{
  "error": "user 'unknown-user' not found"
}
```

#### 409 Conflict

**Duplicate Tier:**
```json
{
  "error": "tier already exists"
}
```

#### 500 Internal Server Error

**Storage Errors:**
```json
{
  "error": "failed to save tier configuration"
}
```

---

## Swagger/OpenAPI Documentation

Interactive API documentation is available via Swagger UI:

```
https://$ROUTE_URL/swagger/index.html
```

The Swagger interface provides:
- Complete API documentation
- Interactive "Try it out" functionality
- Request/response examples for all endpoints
- Schema definitions
- Authentication requirements

### Regenerating Swagger Docs

After making changes to API annotations in the code:

```bash
swag init -g cmd/server/main.go -o docs
```

---

## API Integration Testing

A comprehensive test suite is available to validate all API endpoints:

```bash
# Test against deployed cluster
./tests/test-api.sh https://maas-toolbox-maas-toolbox.apps.$BASE_DOMAIN
```

**The test script validates:**
- All CRUD operations (Create, Read, Update, Delete)
- Group management (Add/Remove groups)
- Group query endpoints
- User query endpoints
- Error cases (duplicates, not found, invalid data)
- Edge cases (empty groups, validation, etc.)

**Test output includes:**
- Colored pass/fail status for each test
- Detailed error messages on failures
- Summary statistics at completion

---

## Additional Examples

For complete workflow examples and common use cases, see the [EXAMPLES.md](./EXAMPLES.md) file.

For comparison with direct OpenShift/MaaS API integration (showing the complexity that MaaS Toolbox simplifies), see the [OPENSHIFT-MAAS-API-GUIDE.md](./OPENSHIFT-MAAS-API-GUIDE.md).
