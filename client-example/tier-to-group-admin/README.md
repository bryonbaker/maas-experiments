# Tier-to-Group Admin Tool

A REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project. This tool provides CRUD operations for managing tiers that map Kubernetes groups to user-defined subscription tiers.

## Features

- **Create Tiers**: Add new tiers with name, description, level, and groups
- **List Tiers**: Retrieve all tiers or a specific tier by name
- **Update Tiers**: Modify tier description, level, and groups (name is immutable)
- **Delete Tiers**: Remove tiers from the configuration
- **File-based Storage**: Initial implementation uses YAML file storage
- **Extensible Design**: Storage backend can be swapped to Kubernetes ConfigMap

## Architecture

The tool is built with a clean architecture:

- **Models**: Data structures for Tier and TierConfig
- **Storage Interface**: Abstract interface for persistence (file or Kubernetes)
- **Service Layer**: Business logic for tier management
- **API Layer**: REST API handlers using Gin framework

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

### Start the Server

```bash
go run cmd/server/main.go
```

Or with custom options:
```bash
go run cmd/server/main.go -file=my-config.yaml -port=9090
```

Default options:
- File: `tier-config.yaml`
- Port: `8080`

### API Endpoints

All endpoints are under `/api/v1/tiers`

#### Create a Tier

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

#### List All Tiers

```bash
curl http://localhost:8080/api/v1/tiers
```

#### Get a Specific Tier

```bash
curl http://localhost:8080/api/v1/tiers/free
```

#### Update a Tier

```bash
curl -X PUT http://localhost:8080/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated free tier description",
    "level": 2,
    "groups": ["system:authenticated", "free-users"]
  }'
```

Note: The `name` field cannot be changed. Only `description`, `level`, and `groups` can be updated.

#### Delete a Tier

```bash
curl -X DELETE http://localhost:8080/api/v1/tiers/free
```

#### Add a Group to a Tier

```bash
curl -X POST http://localhost:8080/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "new-group"}'
```

#### Remove a Group from a Tier

```bash
curl -X DELETE http://localhost:8080/api/v1/tiers/free/groups/system:authenticated
```

### Health Check

```bash
curl http://localhost:8080/health
```

### Swagger Documentation

The API includes interactive Swagger documentation. Once the server is running, access it at:

```
http://localhost:8080/swagger/index.html
```

The Swagger UI provides:
- Interactive API documentation
- Try-it-out functionality for all endpoints
- Request/response examples
- Schema definitions

To regenerate Swagger documentation after making changes to API annotations:

```bash
swag init -g cmd/server/main.go -o docs
```

## Configuration File Format

The tool reads and writes YAML files in the ConfigMap format:

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

## Business Rules

1. **Tier Name**: Set at creation time and cannot be changed
2. **Tier Uniqueness**: Tier names must be unique
3. **Required Fields**: Name and description are required
4. **Level**: Must be a non-negative integer
5. **Groups**: Array of Kubernetes group names

## Error Responses

All errors follow this format:

```json
{
  "error": "error message"
}
```

HTTP Status Codes:
- `200 OK`: Success
- `201 Created`: Tier created successfully
- `204 No Content`: Tier deleted successfully
- `400 Bad Request`: Validation error or invalid request
- `404 Not Found`: Tier not found
- `409 Conflict`: Tier already exists
- `500 Internal Server Error`: Server error

## Future Enhancements

- Kubernetes ConfigMap integration
- Authentication and authorization
- Rate limiting
- Logging and metrics
- Configuration file support
- Docker containerization

## Development

### Project Structure

```
tier-to-group-admin/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go      # HTTP request handlers
│   │   └── router.go        # Route configuration
│   ├── models/
│   │   ├── tier.go          # Data models
│   │   └── errors.go        # Error definitions
│   ├── storage/
│   │   ├── interface.go     # Storage interface
│   │   ├── file.go          # File-based storage
│   │   └── k8s.go           # Kubernetes storage (future)
│   └── service/
│       └── tier.go          # Business logic
├── go.mod
├── go.sum
├── README.md
└── PLAN.md
```

### Building

```bash
go build -o tier-admin cmd/server/main.go
```

### Running Tests

#### Unit Tests

```bash
go test ./...
```

#### API Integration Tests

A comprehensive test script is provided to test all API endpoints:

```bash
# Make sure the server is running first
go run cmd/server/main.go

# In another terminal, run the test script
./test-api.sh

# Or with a custom base URL
BASE_URL=http://localhost:9090 ./test-api.sh
```

The test script will:
- Test all CRUD operations (Create, Read, Update, Delete)
- Test group management (Add/Remove groups)
- Test error cases (duplicate tiers, invalid data, not found, etc.)
- Test edge cases (empty groups, validation, etc.)
- Display colored output with pass/fail status
- Provide a summary at the end

The script uses tier names `acme-inc-1`, `acme-inc-2`, and `acme-inc-3` for testing.

## License

This project is part of the Open Data Hub Model as a Service project.

