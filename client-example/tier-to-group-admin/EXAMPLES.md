# API Usage Examples

This document provides example curl commands for testing the Tier-to-Group Admin API.

## Prerequisites

1. Start the server:
   ```bash
   go run cmd/server/main.go
   ```
   Or use the built binary:
   ```bash
   ./tier-admin
   ```

2. The server will run on `http://localhost:8080` by default.

## Example Commands

### 1. Create a Free Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

Expected response (201 Created):
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

### 2. Create a Premium Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "premium",
    "description": "Premium tier",
    "level": 10,
    "groups": ["premium-users"]
  }'
```

### 3. Create an Enterprise Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "enterprise",
    "description": "Enterprise tier",
    "level": 20,
    "groups": ["enterprise-users"]
  }'
```

### 4. List All Tiers

```bash
curl http://localhost:8080/api/v1/tiers
```

Expected response (200 OK):
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
  },
  {
    "name": "enterprise",
    "description": "Enterprise tier",
    "level": 20,
    "groups": ["enterprise-users"]
  }
]
```

### 5. Get a Specific Tier

```bash
curl http://localhost:8080/api/v1/tiers/free
```

Expected response (200 OK):
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

### 6. Update a Tier (Description and Level)

```bash
curl -X PUT http://localhost:8080/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated free tier description",
    "level": 2
  }'
```

Note: The `groups` field is optional. If not provided, it remains unchanged.

### 7. Update a Tier (Add Groups)

```bash
curl -X PUT http://localhost:8080/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated", "free-users", "trial-users"]
  }'
```

### 8. Update a Tier (Remove Groups)

```bash
curl -X PUT http://localhost:8080/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

### 9. Delete a Tier

```bash
curl -X DELETE http://localhost:8080/api/v1/tiers/free
```

Expected response: 204 No Content (empty body)

### 10. Add a Group to a Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "trial-users"}'
```

Expected response (200 OK):
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated", "trial-users"]
}
```

### 11. Remove a Group from a Tier

```bash
curl -X DELETE http://localhost:8080/api/v1/tiers/free/groups/trial-users
```

Expected response (200 OK):
```json
{
  "name": "free",
  "description": "Free tier for basic users",
  "level": 1,
  "groups": ["system:authenticated"]
}
```

### 12. Health Check

```bash
curl http://localhost:8080/health
```

Expected response (200 OK):
```json
{
  "status": "ok"
}
```

## Error Examples

### Attempt to Create Duplicate Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Another free tier",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

Expected response (409 Conflict):
```json
{
  "error": "tier already exists"
}
```

### Attempt to Get Non-Existent Tier

```bash
curl http://localhost:8080/api/v1/tiers/nonexistent
```

Expected response (404 Not Found):
```json
{
  "error": "tier not found"
}
```

### Attempt to Update Tier Name (Immutable)

```bash
curl -X PUT http://localhost:8080/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "name": "new-name",
    "description": "Updated description",
    "level": 2
  }'
```

Expected response (400 Bad Request):
```json
{
  "error": "tier name cannot be changed"
}
```

### Create Tier with Missing Required Fields

```bash
curl -X POST http://localhost:8080/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "incomplete",
    "level": 1
  }'
```

Expected response (400 Bad Request):
```json
{
  "error": "tier description is required"
}
```

### Attempt to Add Duplicate Group

```bash
curl -X POST http://localhost:8080/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "system:authenticated"}'
```

Expected response (409 Conflict):
```json
{
  "error": "group already exists in tier"
}
```

### Attempt to Remove Non-Existent Group

```bash
curl -X DELETE http://localhost:8080/api/v1/tiers/free/groups/nonexistent-group
```

Expected response (404 Not Found):
```json
{
  "error": "group not found in tier"
}
```

## Using with Postman

1. Create a new collection called "Tier Admin API"
2. Set base URL: `http://localhost:8080`
3. Create requests for each endpoint:
   - POST `/api/v1/tiers`
   - GET `/api/v1/tiers`
   - GET `/api/v1/tiers/:name`
   - PUT `/api/v1/tiers/:name`
   - DELETE `/api/v1/tiers/:name`
4. Set Content-Type header to `application/json` for POST and PUT requests
5. Use the JSON body examples above for POST and PUT requests

