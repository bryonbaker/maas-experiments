# External IDP Portal

A tenant registry and provisioning portal for onboarding corporate customers who authenticate via external Identity Providers (e.g. Azure AD, Okta, Google). The portal captures tenant configuration, then provisions the corresponding Identity Provider and platform groups in Keycloak.

This portal implements the tenant onboarding workflow described in the [Multi-Tenancy Naming Conventions & Integration Guide](Multi-Tenancy%20Naming%20Conventions%20%26%20Integration%20Guide.md).

## What It Does

- **Tenant Registry** - Stores tenant metadata (alias, slug, IdP endpoints, client credentials, group definitions) in a local SQLite database.
- **Keycloak Provisioning** - On activation, creates the OpenID Connect Identity Provider and tenant-scoped platform groups (`tenant:<slug>:administrator`, plus any custom groups) via the Keycloak Admin REST API.
- **Lifecycle Management** - Tenants can be edited, re-activated, or deleted. Deletion cleans up both Keycloak objects and the registry entry, handling missing resources gracefully.

### Tenant Naming

The alias follows the format `{tenant-slug}-{idp-type}` (e.g. `acme-corp-azure`). The tenant slug is derived automatically by stripping the last hyphenated segment, and is used to prefix all platform groups.

### Current Scope

- Human federated identity onboarding
- IdP creation and group provisioning in Keycloak
- IdP mappers (tenant attribute, group membership, username projection) are planned but not yet implemented

## Project Structure

```
external-idp-portal/
├── app/
│   ├── app.py                # Flask application
│   ├── app-config.yaml       # Keycloak connection config
│   ├── requirements.txt      # Python dependencies
│   ├── Containerfile          # Container image build
│   ├── Makefile               # Build/push/run targets
│   └── templates/
│       ├── index.html         # Tenant list dashboard
│       ├── tenant_form.html   # Create/edit tenant
│       └── tenant_detail.html # View tenant details
└── README.md
```

## Prerequisites

- Python 3.12+
- Access to a Keycloak instance with a realm configured (default: `maas-tenants`)
- A Keycloak service account client (`realm-admin-cli`) with `manage-identity-providers` and `manage-realm` roles

## Configuration

### Application Config

Edit `app/app-config.yaml`:

```yaml
keycloak:
  base_url: https://keycloak.apps.your-cluster.example.com
  realm: maas-tenants
  admin_creds_file: ${HOME}/secrets/keycloak/ethan/admin-creds.yaml
```

### Admin Credentials

Create the credentials file referenced above:

```yaml
# $HOME/secrets/keycloak/ethan/admin-creds.yaml
client_id: <your-keycloak-client-id>>
client_secret: <your-keycloak-client-secret>
```

This file must not be committed to version control.

## Running Locally

```bash
cd app

# Create a virtual environment and install dependencies
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Start the development server
python3 app.py
```

The portal will be available at `http://localhost:8080`.

## Building and Running as a Container

```bash
cd app

# Build the image
make build

# Run locally (mounts secrets directory read-only)
make run

# Push to registry
make login
make push
```

See `make help` for all available targets.

## Deploying to OpenShift

The container image can be deployed to OpenShift using standard Deployment, Service, and Route resources. Key considerations:

- Mount the admin credentials file as a Secret volume or pass credentials via environment variables.
- Update `app-config.yaml` with the in-cluster Keycloak URL if Keycloak is co-located.
- The SQLite database is ephemeral inside the container. For production use, mount a PersistentVolumeClaim to the app directory or migrate to an external database.
- The container runs as non-root (UID 1001) and exposes port 8080.

## Usage

1. Open the portal and click **New Tenant**.
2. Enter the **Alias** (e.g. `acme-corp-azure`). The tenant slug and redirect URI update live as you type.
3. Copy the **Redirect URI** and configure it in the customer's Identity Provider.
4. Fill in the remaining fields: Display Name, Authorisation URL, Token URL, Client ID, Client Secret, and optional Groups.
5. Click **Save** to store the tenant as a draft.
6. Click **Activate** to provision the IdP and groups in Keycloak.
