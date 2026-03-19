import os
import re
import sqlite3
import yaml
import requests
import urllib3
from datetime import datetime
from flask import Flask, render_template, request, redirect, url_for, flash, jsonify

app = Flask(__name__)
app.secret_key = os.urandom(32)

# Disable SSL warnings for self-signed certs in dev
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
VERIFY_SSL = False
ALIAS_PATTERN = re.compile(r'^(.+)-[^-]+$')

# ---------------------------------------------------------------------------
# Configuration loading
# ---------------------------------------------------------------------------

def load_config():
    config_path = os.path.join(os.path.dirname(__file__), 'app-config.yaml')
    with open(config_path) as f:
        config = yaml.safe_load(f)

    kc = config['keycloak']
    creds_path = os.path.expandvars(kc['admin_creds_file'])
    with open(creds_path) as f:
        creds = yaml.safe_load(f)

    return {
        'base_url': kc['base_url'],
        'realm': kc['realm'],
        'admin_client_id': creds['client_id'],
        'admin_client_secret': creds['client_secret'],
    }

CONFIG = load_config()

# ---------------------------------------------------------------------------
# Database
# ---------------------------------------------------------------------------

DB_PATH = os.path.join(os.path.dirname(__file__), 'tenants.db')

def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    conn.execute('PRAGMA journal_mode=WAL')
    return conn


def derive_slug(alias):
    """Derive tenant slug from alias by stripping the last hyphenated segment.

    Uses regex ^(.+)-[^-]+$ to capture everything before the final -suffix.
    e.g. acme-corp-azure -> acme-corp
         widgets-inc-oidc -> widgets-inc
         my-company-with-dashes-saml -> my-company-with-dashes
    """
    m = ALIAS_PATTERN.match(alias)
    if m:
        return m.group(1)
    return alias


def init_db():
    conn = get_db()
    conn.execute('''
        CREATE TABLE IF NOT EXISTS tenants (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            alias TEXT UNIQUE NOT NULL,
            tenant_slug TEXT NOT NULL DEFAULT '',
            display_name TEXT NOT NULL,
            auth_url TEXT NOT NULL,
            token_url TEXT NOT NULL,
            client_id TEXT NOT NULL,
            client_secret TEXT NOT NULL,
            groups TEXT NOT NULL DEFAULT '',
            state TEXT NOT NULL DEFAULT 'draft',
            error_message TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    # Migration: add tenant_slug column if missing (existing databases)
    columns = [row[1] for row in conn.execute('PRAGMA table_info(tenants)').fetchall()]
    if 'tenant_slug' not in columns:
        conn.execute('ALTER TABLE tenants ADD COLUMN tenant_slug TEXT NOT NULL DEFAULT ""')
        # Backfill from existing aliases
        for row in conn.execute('SELECT id, alias FROM tenants').fetchall():
            slug = derive_slug(row[1])
            conn.execute('UPDATE tenants SET tenant_slug = ? WHERE id = ?', (slug, row[0]))
    conn.commit()
    conn.close()

init_db()

# ---------------------------------------------------------------------------
# Keycloak Admin API helpers
# ---------------------------------------------------------------------------

def get_admin_token():
    """Get an admin token for the maas-tenants realm using realm-admin-cli client credentials."""
    url = f"{CONFIG['base_url']}/realms/{CONFIG['realm']}/protocol/openid-connect/token"
    data = {
        'grant_type': 'client_credentials',
        'client_id': CONFIG['admin_client_id'],
        'client_secret': CONFIG['admin_client_secret'],
    }
    resp = requests.post(url, data=data, verify=VERIFY_SSL)
    resp.raise_for_status()
    return resp.json()['access_token']


def kc_admin_request(method, path, token, json_data=None):
    """Make an authenticated request to the Keycloak Admin REST API."""
    url = f"{CONFIG['base_url']}/admin/realms/{CONFIG['realm']}{path}"
    headers = {
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json',
    }
    resp = requests.request(method, url, headers=headers, json=json_data, verify=VERIFY_SSL)
    return resp


def provision_idp(tenant, token):
    """Create the OpenID Connect Identity Provider in Keycloak."""
    idp_data = {
        'alias': tenant['alias'],
        'displayName': tenant['display_name'],
        'providerId': 'oidc',
        'enabled': True,
        'trustEmail': False,
        'storeToken': False,
        'addReadTokenRoleOnCreate': False,
        'firstBrokerLoginFlowAlias': 'first broker login',
        'config': {
            'authorizationUrl': tenant['auth_url'],
            'tokenUrl': tenant['token_url'],
            'clientId': tenant['client_id'],
            'clientSecret': tenant['client_secret'],
            'clientAuthMethod': 'client_secret_post',
            'defaultScope': 'openid email profile',
            'syncMode': 'INHERIT',
            'useJwksUrl': 'true',
        }
    }
    resp = kc_admin_request('POST', '/identity-provider/instances', token, idp_data)
    if resp.status_code == 409:
        raise Exception(f"IdP with alias '{tenant['alias']}' already exists in Keycloak")
    if not resp.ok:
        raise Exception(f"Failed to create IdP: {resp.status_code} {resp.text}")


def provision_groups(tenant, token):
    """Create tenant groups in Keycloak.

    Always creates tenant:<slug>:administrator.
    Creates additional groups from the tenant's groups field.
    """
    slug = tenant['tenant_slug']
    group_names = [f'tenant:{slug}:administrator']

    custom_groups = [g.strip() for g in tenant['groups'].split(',') if g.strip()]
    for g in custom_groups:
        name = f'tenant:{slug}:{g}'
        if name not in group_names:
            group_names.append(name)

    errors = []
    for group_name in group_names:
        resp = kc_admin_request('POST', '/groups', token, {'name': group_name})
        if resp.status_code == 409:
            continue  # Group already exists, that's fine
        if not resp.ok:
            errors.append(f"Failed to create group '{group_name}': {resp.status_code} {resp.text}")

    if errors:
        raise Exception('; '.join(errors))


def provision_mappers(tenant, token):
    """Create IdP mappers for tenant attribution and group mapping.

    TODO: Implement when mapper requirements are finalised.
    """
    # slug = derive_slug(tenant['alias'])
    #
    # Hardcoded tenantSlug attribute mapper:
    # Sets the tenantSlug user attribute to the tenant slug for every user
    # who authenticates through this IdP.
    #
    # mapper_data = {
    #     'name': f'{slug}-tenant-slug',
    #     'identityProviderAlias': tenant['alias'],
    #     'identityProviderMapper': 'hardcoded-attribute-idp-mapper',
    #     'config': {
    #         'syncMode': 'INHERIT',
    #         'attribute': 'tenantSlug',
    #         'attribute.value': slug,
    #     }
    # }
    # resp = kc_admin_request('POST',
    #     f"/identity-providers/instances/{tenant['alias']}/mappers",
    #     token, mapper_data)
    #
    # Group membership mapper:
    # Emits platform groups in the 'groups' token claim.
    #
    # preferred_username mapper:
    # Constructs preferred_username as oidc:<tenantSlug>:<keycloak-uuid>
    # using stored tenantSlug plus user.id at token issuance time.
    #
    # First-login broker flow assignment:
    # Assign the platform's custom first-login broker flow to this IdP
    # so that first login creates the local user, sets tenant attribute,
    # and assigns baseline group.
    pass


def delete_keycloak_idp(alias, token):
    """Delete an IdP from Keycloak. Silently skips if not found."""
    resp = kc_admin_request('DELETE', f'/identity-provider/instances/{alias}', token)
    if resp.status_code == 404:
        return  # Already gone
    if not resp.ok:
        raise Exception(f"Failed to delete IdP '{alias}': {resp.status_code} {resp.text}")


def delete_keycloak_groups(tenant, token):
    """Delete tenant groups from Keycloak. Silently skips missing groups."""
    slug = tenant['tenant_slug']
    group_names = [f'tenant:{slug}:administrator']
    custom_groups = [g.strip() for g in tenant['groups'].split(',') if g.strip()]
    for g in custom_groups:
        name = f'tenant:{slug}:{g}'
        if name not in group_names:
            group_names.append(name)

    errors = []
    for group_name in group_names:
        # Look up group by name to get its ID
        resp = kc_admin_request('GET', f'/groups?search={group_name}&exact=true', token)
        if not resp.ok:
            errors.append(f"Failed to look up group '{group_name}': {resp.status_code}")
            continue
        groups = resp.json()
        if not groups:
            continue  # Already gone
        for group in groups:
            if group.get('name') == group_name:
                del_resp = kc_admin_request('DELETE', f'/groups/{group["id"]}', token)
                if del_resp.status_code == 404:
                    continue  # Already gone
                if not del_resp.ok:
                    errors.append(f"Failed to delete group '{group_name}': {del_resp.status_code}")

    if errors:
        raise Exception('; '.join(errors))


def delete_tenant(tenant_id):
    """Delete a tenant from Keycloak and the registry.

    If Keycloak is unreachable, raises a user-friendly error and preserves the DB record.
    If Keycloak resources are already gone, silently continues and removes the DB record.
    """
    conn = get_db()
    tenant = conn.execute('SELECT * FROM tenants WHERE id = ?', (tenant_id,)).fetchone()
    if not tenant:
        conn.close()
        raise Exception('Tenant not found')

    try:
        token = get_admin_token()
    except (requests.exceptions.ConnectionError, requests.exceptions.SSLError):
        conn.close()
        raise Exception('Keycloak server is unreachable')
    except requests.exceptions.HTTPError as e:
        conn.close()
        raise Exception(f'Failed to authenticate with Keycloak: {e.response.status_code}')
    except Exception:
        conn.close()
        raise Exception('Keycloak server is unreachable')

    try:
        delete_keycloak_groups(tenant, token)
        delete_keycloak_idp(tenant['alias'], token)
        conn.execute('DELETE FROM tenants WHERE id = ?', (tenant_id,))
        conn.commit()
    except Exception:
        conn.close()
        raise
    finally:
        conn.close()


def activate_tenant(tenant_id):
    """Provision all Keycloak objects for a tenant."""
    conn = get_db()
    tenant = conn.execute('SELECT * FROM tenants WHERE id = ?', (tenant_id,)).fetchone()
    if not tenant:
        conn.close()
        raise Exception('Tenant not found')

    try:
        try:
            token = get_admin_token()
        except (requests.exceptions.ConnectionError, requests.exceptions.SSLError):
            raise Exception('Keycloak server is unreachable')
        except requests.exceptions.HTTPError as e:
            raise Exception(f'Failed to authenticate with Keycloak: {e.response.status_code}')
        except Exception:
            raise Exception('Keycloak server is unreachable')

        provision_idp(tenant, token)
        provision_groups(tenant, token)
        provision_mappers(tenant, token)

        conn.execute(
            'UPDATE tenants SET state = ?, error_message = NULL, updated_at = ? WHERE id = ?',
            ('active', datetime.utcnow(), tenant_id))
        conn.commit()
    except Exception as e:
        conn.execute(
            'UPDATE tenants SET state = ?, error_message = ?, updated_at = ? WHERE id = ?',
            ('error', str(e), datetime.utcnow(), tenant_id))
        conn.commit()
        raise
    finally:
        conn.close()

# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@app.route('/')
def index():
    conn = get_db()
    tenants = conn.execute('SELECT * FROM tenants ORDER BY created_at DESC').fetchall()
    conn.close()
    return render_template('index.html', tenants=tenants)


@app.route('/tenant/new')
def tenant_new():
    return render_template('tenant_form.html',
                           tenant=None,
                           config=CONFIG)


@app.route('/tenant/<int:tenant_id>')
def tenant_detail(tenant_id):
    conn = get_db()
    tenant = conn.execute('SELECT * FROM tenants WHERE id = ?', (tenant_id,)).fetchone()
    conn.close()
    if not tenant:
        flash('Tenant not found', 'error')
        return redirect(url_for('index'))
    slug = tenant['tenant_slug']
    group_list = [f'tenant:{slug}:administrator']
    custom = [g.strip() for g in tenant['groups'].split(',') if g.strip()]
    for g in custom:
        group_list.append(f'tenant:{slug}:{g}')
    return render_template('tenant_detail.html',
                           tenant=tenant,
                           slug=slug,
                           group_list=group_list,
                           config=CONFIG)


@app.route('/tenant/<int:tenant_id>/edit')
def tenant_edit(tenant_id):
    conn = get_db()
    tenant = conn.execute('SELECT * FROM tenants WHERE id = ?', (tenant_id,)).fetchone()
    conn.close()
    if not tenant:
        flash('Tenant not found', 'error')
        return redirect(url_for('index'))
    return render_template('tenant_form.html',
                           tenant=tenant,
                           config=CONFIG)


@app.route('/api/tenant', methods=['POST'])
def tenant_create():
    alias = request.form.get('alias', '').strip()
    display_name = request.form.get('display_name', '').strip()
    auth_url = request.form.get('auth_url', '').strip()
    token_url = request.form.get('token_url', '').strip()
    client_id = request.form.get('client_id', '').strip()
    client_secret = request.form.get('client_secret', '').strip()
    groups = request.form.get('groups', '').strip()

    if not all([alias, display_name, auth_url, token_url, client_id, client_secret]):
        flash('All fields except Groups are required', 'error')
        return redirect(url_for('tenant_new'))

    if not ALIAS_PATTERN.match(alias):
        flash('Alias must include an IdP type suffix after a hyphen (e.g. acme-corp-azure)', 'error')
        return redirect(url_for('tenant_new'))

    tenant_slug = derive_slug(alias)

    conn = get_db()
    try:
        conn.execute('''
            INSERT INTO tenants (alias, tenant_slug, display_name, auth_url, token_url, client_id, client_secret, groups)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ''', (alias, tenant_slug, display_name, auth_url, token_url, client_id, client_secret, groups))
        conn.commit()
        tenant_id = conn.execute('SELECT id FROM tenants WHERE alias = ?', (alias,)).fetchone()['id']
        flash('Tenant saved as draft', 'success')
        return redirect(url_for('tenant_detail', tenant_id=tenant_id))
    except sqlite3.IntegrityError:
        flash(f"A tenant with alias '{alias}' already exists", 'error')
        return redirect(url_for('tenant_new'))
    finally:
        conn.close()


@app.route('/api/tenant/<int:tenant_id>', methods=['POST'])
def tenant_update(tenant_id):
    alias = request.form.get('alias', '').strip()
    display_name = request.form.get('display_name', '').strip()
    auth_url = request.form.get('auth_url', '').strip()
    token_url = request.form.get('token_url', '').strip()
    client_id = request.form.get('client_id', '').strip()
    client_secret = request.form.get('client_secret', '').strip()
    groups = request.form.get('groups', '').strip()

    if not all([alias, display_name, auth_url, token_url, client_id, client_secret]):
        flash('All fields except Groups are required', 'error')
        return redirect(url_for('tenant_edit', tenant_id=tenant_id))

    if not ALIAS_PATTERN.match(alias):
        flash('Alias must include an IdP type suffix after a hyphen (e.g. acme-corp-azure)', 'error')
        return redirect(url_for('tenant_edit', tenant_id=tenant_id))

    tenant_slug = derive_slug(alias)

    conn = get_db()
    try:
        conn.execute('''
            UPDATE tenants SET alias=?, tenant_slug=?, display_name=?, auth_url=?, token_url=?,
                client_id=?, client_secret=?, groups=?, updated_at=?
            WHERE id=?
        ''', (alias, tenant_slug, display_name, auth_url, token_url, client_id, client_secret,
              groups, datetime.utcnow(), tenant_id))
        conn.commit()
        flash('Tenant updated', 'success')
        return redirect(url_for('tenant_detail', tenant_id=tenant_id))
    except sqlite3.IntegrityError:
        flash(f"A tenant with alias '{alias}' already exists", 'error')
        return redirect(url_for('tenant_edit', tenant_id=tenant_id))
    finally:
        conn.close()


@app.route('/api/tenant/<int:tenant_id>/activate', methods=['POST'])
def tenant_activate(tenant_id):
    try:
        activate_tenant(tenant_id)
        flash('Tenant activated successfully — IdP and groups created in Keycloak', 'success')
    except Exception as e:
        flash(f'Activation failed: {e}', 'error')
    return redirect(url_for('tenant_detail', tenant_id=tenant_id))


@app.route('/api/tenant/<int:tenant_id>/delete', methods=['POST'])
def tenant_delete(tenant_id):
    try:
        delete_tenant(tenant_id)
        flash('Tenant deleted successfully', 'success')
    except Exception as e:
        flash(f'Delete failed: {e}', 'error')
        return redirect(url_for('tenant_detail', tenant_id=tenant_id))
    return redirect(url_for('index'))


@app.route('/health')
def health():
    return jsonify({'status': 'ok'}), 200


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080, debug=True)
