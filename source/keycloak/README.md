# Keycloak Admin Scripts

## Security Best Practices

### tst-realm-admin.sh

This script creates a Keycloak realm with users and groups. **All secrets are managed via environment variables** - no hardcoded credentials.

#### Required Environment Variables

```bash
export KEYCLOAK_URL="https://keycloak.example.com"
export CLIENT_ID="realm-admin-cli"
export KK_CLIENT_SECRET="your-client-secret-here"
export INITIAL_PASSWORD="secure-password-here"
```

#### Optional Environment Variables (with defaults)

```bash
export REALM="tenant-test-1"
export TENANT_CLIENT_ID="tenant-admin-cli"
export GROUP_NAME="default-users"
export USER_NAME="alice"
export EMAIL="alice@wonderland.com"
export FIRST_NAME="Alice"
export LAST_NAME="Wonderland"
```

#### SSL Certificate Verification

By default, the script **validates SSL certificates**. For development/testing with self-signed certificates only:

```bash
export INSECURE_SSL="true"
```

**⚠️ WARNING**: Never use `INSECURE_SSL=true` in production!

#### Example Usage

```bash
# Production (with SSL verification)
export KEYCLOAK_URL="https://keycloak.production.com"
export CLIENT_ID="realm-admin-cli"
export KK_CLIENT_SECRET="$(cat /path/to/secret/file)"
export INITIAL_PASSWORD="$(openssl rand -base64 32)"

./tst-realm-admin.sh
```

```bash
# Development (with self-signed certs)
export KEYCLOAK_URL="https://keycloak.dev.local"
export CLIENT_ID="realm-admin-cli"
export KK_CLIENT_SECRET="dev-secret"
export INITIAL_PASSWORD="dev-password"
export INSECURE_SSL="true"

./tst-realm-admin.sh
```

#### Using with Secret Management Systems

**Kubernetes Secrets:**
```bash
kubectl create secret generic keycloak-admin \
  --from-literal=url='https://keycloak.example.com' \
  --from-literal=client-id='realm-admin-cli' \
  --from-literal=client-secret='your-secret' \
  --from-literal=initial-password='user-password'
```

**HashiCorp Vault:**
```bash
export KEYCLOAK_URL="$(vault kv get -field=url secret/keycloak)"
export KK_CLIENT_SECRET="$(vault kv get -field=client-secret secret/keycloak)"
export INITIAL_PASSWORD="$(vault kv get -field=initial-password secret/keycloak)"
```

**AWS Secrets Manager:**
```bash
export KK_CLIENT_SECRET="$(aws secretsmanager get-secret-value \
  --secret-id keycloak/client-secret \
  --query SecretString --output text)"
```

## Changes Made to Fix Security Issues

### 1. ✅ Fixed OAuth2 Parameter Name (Critical Bug)
- **Before**: `-d "KK_CLIENT_SECRET=${KK_CLIENT_SECRET}"`
- **After**: `-d "client_secret=${KK_CLIENT_SECRET}"`

The OAuth2 specification requires the parameter name to be `client_secret`, not the variable name.

### 2. ✅ Removed Hardcoded Password Default
- **Before**: `INITIAL_PASSWORD="${INITIAL_PASSWORD:-password}"`
- **After**: `INITIAL_PASSWORD="${INITIAL_PASSWORD:-}"`

Now requires explicit password via environment variable.

### 3. ✅ SSL Verification Configurable
- **Before**: Always used `-k` flag (insecure)
- **After**: SSL verification enabled by default, opt-in to disable via `INSECURE_SSL=true`

### 4. ✅ All Secrets from Environment
- No hardcoded credentials in the script
- Clear documentation on required vs optional variables
- Works with secret management systems (Vault, K8s Secrets, AWS Secrets Manager, etc.)

