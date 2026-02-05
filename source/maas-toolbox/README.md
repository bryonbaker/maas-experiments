# MaaS Toolbox

A simplified REST API for managing Open Data Hub Models-as-a-Service (ODH MaaS) configurations on OpenShift/Kubernetes clusters.

## Overview

Configuring and managing ODH MaaS directly through OpenShift and Kubernetes APIs requires complex multi-step workflows involving ConfigMaps, custom resources, YAML parsing, and coordination across multiple API endpoints. **MaaS Toolbox provides a simplified, purpose-built REST API that abstracts this complexity**, making it easy to integrate MaaS management into service portals, automation tools, and administrative workflows.

### Why MaaS Toolbox?

**Without MaaS Toolbox**, managing tiers, groups, and rate limits requires:
- Direct manipulation of Kubernetes ConfigMaps with YAML parsing
- Multi-step GET-MODIFY-PUT workflows for every change
- Manual coordination between tier configuration and rate limit policies
- Complex jq/yq scripting to extract and update nested data structures
- Deep knowledge of multiple Kubernetes/OpenShift APIs

See the [OpenShift MaaS API Guide](./OPENSHIFT-MAAS-API-GUIDE.md) for detailed examples of the complexity involved in direct API integration.

**With MaaS Toolbox**, you get:
- Simple REST API with intuitive CRUD operations
- Single-step operations for common management tasks
- Automatic validation and consistency checks
- Clean JSON request/response format
- Built-in error handling and meaningful error messages

## Target Use Case

MaaS Toolbox is designed for Neo/private/sovereign cloud providers building their own **Service Request Portals** on top of ODH MaaS. If you're creating a customer-facing portal or automation system that needs to provision and manage AI model access, MaaS Toolbox provides the management API layer you need.

## Key Capabilities

- **Tier Management**: Create, update, list, and delete subscription tiers with group-based access control
- **Group Association**: Manage which Kubernetes/OpenShift groups have access to each tier
- **User Queries**: Lookup which tiers a user can access based on their group memberships
- **Rate Limit Management**: Configure request and token-based rate limits per tier
- **Model Assignment**: Associate LLM inference services with specific tiers
- **Native Storage**: Stores configuration in Kubernetes ConfigMaps for GitOps compatibility

## Architecture

MaaS Toolbox follows a clean, layered architecture:

```
┌─────────────────────────────────────────┐
│         REST API (Gin)                  │
│  /api/v1/tiers, /groups, /users        │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│         Service Layer                   │
│  Business logic & validation            │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│    Storage Layer (Kubernetes)           │
│  ConfigMaps, CRDs, Group API            │
└─────────────────────────────────────────┘
```

### Components

- **Models**: Data structures for Tier, RateLimitPolicy, TokenRateLimitPolicy
- **Storage**: Kubernetes ConfigMap and CRD-based persistence
- **Service Layer**: Business logic, validation, and group resolution
- **API Layer**: REST endpoints with OpenAPI/Swagger documentation

## Quick Start

### Prerequisites

- OpenShift or Kubernetes cluster
- `oc` or `kubectl` CLI tool
- Cluster admin permissions for RBAC setup

### Deploy to Your Cluster

```bash
# Deploy using Kustomize
oc apply -k yaml/base/

# Verify deployment
oc get pods -n maas-toolbox
```

### Get the API URL

```bash
ROUTE_URL=$(oc get route maas-toolbox -n maas-toolbox -o jsonpath='{.spec.host}')
echo "MaaS Toolbox API: https://$ROUTE_URL"
```

### Create Your First Tier

```bash
curl -X POST https://$ROUTE_URL/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Free tier for authenticated users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

### Explore the API

Interactive Swagger documentation is available at:
```
https://$ROUTE_URL/swagger/index.html
```

## Documentation

- **[Interface Documentation](./INTERFACE_DOCUMENTATION.md)** - Complete API reference with all endpoints, parameters, and examples
- **[OpenShift MaaS API Guide](./OPENSHIFT-MAAS-API-GUIDE.md)** - Detailed guide showing the complex direct API workflows that MaaS Toolbox simplifies
- **[Examples](./EXAMPLES.md)** - Common use cases and workflow examples
- **[Deployment Guide](./yaml/README.md)** - Detailed deployment and configuration instructions

## Configuration

MaaS Toolbox stores tier configuration in a Kubernetes ConfigMap named `tier-to-group-mapping` in the `maas-api` namespace. This follows the [official ODH MaaS tier configuration format](https://opendatahub-io.github.io/models-as-a-service/latest/configuration-and-management/tier-overview/).

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | `maas-api` | Kubernetes namespace for ConfigMap storage |
| `CONFIGMAP_NAME` | `tier-to-group-mapping` | Name of the tier configuration ConfigMap |
| `PORT` | `8080` | Server port |
| `VALIDATE_GROUPS` | `yes` | Validate groups exist before creating/updating tiers. Set to `no` when using external identity providers (Keycloak, LDAP) |

## Development

### Building

```bash
# Build container image
make build

# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Project Structure

```
maas-toolbox/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── models/         # Data models and validation
│   ├── service/        # Business logic
│   └── storage/        # Kubernetes client and storage
├── yaml/               # Deployment manifests
├── tests/              # Integration test scripts
└── docs/               # Generated Swagger docs
```

### API Integration Tests

```bash
# Test all endpoints against a deployed cluster
./tests/test-api.sh https://maas-toolbox-maas-toolbox.apps.your-cluster.com
```

## Roadmap

- [x] Tier management (CRUD operations)
- [x] Group-based access control
- [x] User tier lookup
- [x] Swagger/OpenAPI documentation
- [x] Rate limit policy configuration
- [x] Token rate limit policy configuration
- [ ] Referential-integrity health checks
- [ ] Authentication and authorization
- [ ] Enhanced logging and metrics
- [ ] Multi-namespace support

## Support & Contributing

For issues, questions, or contributions, please refer to the main repository documentation.

## License

This source file includes portions generated or suggested by  artificial intelligence tools and subsequently reviewed,  modified, and validated by human contributors.
 
Human authorship, design decisions, and final responsibility for this code remain with the project contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
