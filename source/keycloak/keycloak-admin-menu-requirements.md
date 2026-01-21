# Menu-Based Keycloak Admin Script - Implementation Plan

---

## Script Architecture

```
keycloak-admin-menu.sh
├── Setup & Authentication Module
├── Utility Functions Module
├── Main Menu
│   ├── 1. Manage Tenants → Tenant CRUD Submenu
│   ├── 2. Manage Tenant Groups → Group CRUD Submenu
│   ├── 3. Manage Tenant Users → User CRUD & Membership Submenu
│   └── 0. Exit
└── Cleanup & Exit Handler
```

---

## Main Menu Structure

```
╔════════════════════════════════════════════════╗
║   Keycloak Admin - MaaS Tenants Realm         ║
║   Realm: maas-tenants                         ║
║   Token expires in: 250 seconds               ║
╚════════════════════════════════════════════════╝

  1. Manage Tenants
  2. Manage Tenant Groups  
  3. Manage Tenant Users
  
  0. Exit

Select option:
```

---

## Menu 1: Manage Tenants

### Menu Display
```
╔════════════════════════════════════════════════╗
║   Manage Tenants                              ║
╚════════════════════════════════════════════════╝

  1. Create Tenant
  2. List All Tenants
  3. View Tenant Details
  4. Update Tenant Attributes
  5. Delete Tenant
  
  0. Back to Main Menu

Select option:
```

### 1.1 Create Tenant
```bash
Prompts:
- Enter tenant name (e.g., acme-inc-2):

Actions:
1. Validate tenant name (alphanumeric, hyphens, lowercase)
2. Check if tenant already exists
3. Create parent group /{tenant-name}
4. Create default subgroup /{tenant-name}/{tenant-name}-dedicated
5. Set default attributes:
   - description: "Tenant: {name}"
   - created_at: {timestamp}

Display:
✓ Tenant 'acme-inc-2' created
  - Group ID: {id}
  - Path: /acme-inc-2
  - Default group: /acme-inc-2/acme-inc-2-dedicated

Press Enter to continue...
```

### 1.2 List All Tenants
```bash
Actions:
1. GET /groups (search for all groups)
2. Filter only top-level groups (no "/" in name except prefix)
3. For each, get subgroups count and members count

Display:
┌─────────────────┬──────────────────────────┬───────────┬──────────┐
│ Tenant Name     │ Group ID                 │ Subgroups │ Members  │
├─────────────────┼──────────────────────────┼───────────┼──────────┤
│ acme-inc-2      │ 8f3a7d9e-1234-5678...    │ 2         │ 5        │
│ contoso-org     │ 9a2b8c1d-2345-6789...    │ 1         │ 3        │
└─────────────────┴──────────────────────────┴───────────┴──────────┘

Press Enter to continue...
```

### 1.3 View Tenant Details
```bash
Prompts:
- Enter tenant name:

Actions:
1. GET /groups?search={tenant-name}
2. GET /groups/{id}/children (list subgroups)
3. GET /groups/{id}/members (list members)
4. Display all attributes

Display:
Tenant: acme-inc-2
─────────────────────────────────────────────
Group ID:     8f3a7d9e-1234-5678-90ab-cdef12345678
Path:         /acme-inc-2
Created:      2026-01-21 10:30:00
Description:  Acme Inc Tenant

Subgroups:
  • /acme-inc-2/acme-inc-2-dedicated (5 members)
  • /acme-inc-2/acme-inc-2-shared (2 members)

Total Members: 5
  • acme-user-1@acme-inc-2.com
  • acme-user-2@acme-inc-2.com
  ...

Press Enter to continue...
```

### 1.4 Update Tenant Attributes
```bash
Prompts:
- Enter tenant name:
- Enter description (current: {current}):
- Add custom attribute? (y/n):
  - Key:
  - Value:

Actions:
1. GET group
2. PUT /groups/{id} with updated attributes

Display:
✓ Tenant 'acme-inc-2' updated successfully

Press Enter to continue...
```

### 1.5 Delete Tenant
```bash
Prompts:
- Enter tenant name:
- WARNING: This will delete ALL subgroups and remove ALL user memberships
- Type tenant name to confirm:

Actions:
1. GET group and all children
2. For each member, remove group membership
3. Delete all subgroups (bottom-up)
4. Delete parent group

Display:
✓ Deleted 2 subgroups
✓ Removed 5 user memberships
✓ Tenant 'acme-inc-2' deleted

Press Enter to continue...
```

---

## Menu 2: Manage Tenant Groups

### Menu Display
```
╔════════════════════════════════════════════════╗
║   Manage Tenant Groups                        ║
╚════════════════════════════════════════════════╝

  1. Create Tenant Group
  2. List Tenant Groups
  3. View Group Details
  4. Update Group Attributes
  5. Delete Tenant Group
  
  0. Back to Main Menu

Select option:
```

### 2.1 Create Tenant Group
```bash
Prompts:
- Enter tenant name (parent):
- Enter group name (e.g., dedicated, shared, enterprise):
  (Will be created as: /{tenant}/{tenant}-{group})

Actions:
1. Verify parent tenant exists
2. Construct full group name: {tenant}-{group}
3. POST /groups/{parent-id}/children

Display:
✓ Group created: /acme-inc-2/acme-inc-2-shared
  - Group ID: {id}
  - Members: 0

Press Enter to continue...
```

### 2.2 List Tenant Groups
```bash
Prompts:
- Enter tenant name (or leave empty for all):

Actions:
1. If tenant specified: GET /groups/{tenant-id}/children
2. If empty: GET all groups, display hierarchy

Display:
Tenant: acme-inc-2
┌──────────────────────────────┬──────────────────────────┬──────────┐
│ Group Name                   │ Group ID                 │ Members  │
├──────────────────────────────┼──────────────────────────┼──────────┤
│ acme-inc-2-dedicated         │ 9a2b8c1d-2345-6789...    │ 5        │
│ acme-inc-2-shared            │ 1b3c9d2e-3456-7890...    │ 2        │
│ acme-inc-2-enterprise        │ 2c4d0e3f-4567-8901...    │ 1        │
└──────────────────────────────┴──────────────────────────┴──────────┘

Press Enter to continue...
```

### 2.3 View Group Details
```bash
Prompts:
- Enter full group path (e.g., /acme-inc-2/acme-inc-2-dedicated):

Actions:
1. Search for group
2. GET /groups/{id}
3. GET /groups/{id}/members

Display:
Group: /acme-inc-2/acme-inc-2-dedicated
─────────────────────────────────────────────
Group ID:     9a2b8c1d-2345-6789-01bc-def123456789
Parent:       /acme-inc-2
Tier Level:   dedicated
Created:      2026-01-21 10:35:00

Members (5):
  • acme-user-1@acme-inc-2.com (Acme User1)
  • acme-user-2@acme-inc-2.com (Acme User2)
  ...

Attributes:
  • tier: dedicated
  • resources: gpu-enabled

Press Enter to continue...
```

### 2.4 Update Group Attributes
```bash
Prompts:
- Enter group path:
- Update attributes:
  - Key:
  - Value:
  - Add another? (y/n)

Actions:
1. GET group
2. PUT /groups/{id} with attributes

Display:
✓ Group '/acme-inc-2/acme-inc-2-dedicated' updated

Press Enter to continue...
```

### 2.5 Delete Tenant Group
```bash
Prompts:
- Enter group path:
- WARNING: This will remove ALL user memberships from this group
- Type group name to confirm:

Actions:
1. GET group members
2. For each member, DELETE membership
3. DELETE /groups/{id}

Display:
✓ Removed 5 user memberships
✓ Group '/acme-inc-2/acme-inc-2-shared' deleted

Press Enter to continue...
```

---

## Menu 3: Manage Tenant Users

### Menu Display
```
╔════════════════════════════════════════════════╗
║   Manage Tenant Users                         ║
╚════════════════════════════════════════════════╝

  1. Create User
  2. List Users
  3. View User Details
  4. Update User
  5. Delete User
  6. Add User to Group
  7. Remove User from Group
  
  0. Back to Main Menu

Select option:
```

### 3.1 Create User
```bash
Prompts:
- Enter email (will be username):
- Enter first name:
- Enter last name:
- Enter password:
- Email verified? (y/n):
- Enabled? (y/n, default: y):

Actions:
1. POST /users
2. PUT /users/{id}/reset-password

Display:
✓ User created: acme-user-1@acme-inc-2.com
  - User ID: {id}
  - Status: Enabled
  - Email Verified: Yes
  - Groups: (none)

Add to groups now? (y/n):
  [If yes, jump to Add User to Group]

Press Enter to continue...
```

### 3.2 List Users
```bash
Prompts:
- Filter by:
  1. All users
  2. Users in specific tenant
  3. Users in specific group
  4. Search by email/name
  
- Enter choice:
- [Based on choice] Enter search criteria:

Actions:
1. GET /users (with search params)
   OR
2. GET /groups/{id}/members (if filtering by group)

Display:
Users (15 found):
┌────────────────────────────────┬──────────────────────┬─────────┬────────┐
│ Email                          │ Name                 │ Groups  │ Status │
├────────────────────────────────┼──────────────────────┼─────────┼────────┤
│ acme-user-1@acme-inc-2.com     │ Acme User1          │ 2       │ ✓      │
│ acme-user-2@acme-inc-2.com     │ Acme User2          │ 1       │ ✓      │
└────────────────────────────────┴──────────────────────┴─────────┴────────┘

Press Enter to continue...
```

### 3.3 View User Details
```bash
Prompts:
- Enter user email:

Actions:
1. GET /users?email={email}
2. GET /users/{id}/groups

Display:
User: acme-user-1@acme-inc-2.com
─────────────────────────────────────────────
User ID:       dc3af004-e88c-4efb-b02d-136016ec2ec5
Username:      acme-user-1@acme-inc-2.com
First Name:    Acme
Last Name:     User1
Email:         acme-user-1@acme-inc-2.com
Email Verified: ✓
Enabled:       ✓
Created:       2026-01-21 11:00:00

Groups (2):
  • /acme-inc-2/acme-inc-2-dedicated
  • /serverless

Attributes:
  • phone: +1-555-0100
  • department: Engineering

Press Enter to continue...
```

### 3.4 Update User
```bash
Prompts:
- Enter user email:

Submenu:
  1. Update name
  2. Update email
  3. Update password
  4. Update attributes
  5. Enable/Disable user
  0. Back

Actions:
Based on submenu choice, PUT /users/{id}

Display:
✓ User 'acme-user-1@acme-inc-2.com' updated successfully

Press Enter to continue...
```

### 3.5 Delete User
```bash
Prompts:
- Enter user email:
- WARNING: This will permanently delete the user
- Type email to confirm:

Actions:
1. GET user
2. Display user details
3. DELETE /users/{id}

Display:
✓ User 'acme-user-1@acme-inc-2.com' deleted
  (Automatically removed from 2 groups)

Press Enter to continue...
```

### 3.6 Add User to Group
```bash
Prompts:
- Enter user email:
- Current groups:
  [Display current group memberships]
  
Select group to add:
  1. Search for group by path
  2. Select from tenant groups
  
Actions:
1. If option 1: Enter full path
2. If option 2:
   - Enter tenant name
   - Display tenant groups
   - Select from list

3. PUT /users/{id}/groups/{group-id}

Display:
✓ Added 'acme-user-1@acme-inc-2.com' to '/acme-inc-2/acme-inc-2-dedicated'

Current groups (3):
  • /acme-inc-2/acme-inc-2-dedicated
  • /serverless
  • /premium-users

Add to another group? (y/n):

Press Enter to continue...
```

### 3.7 Remove User from Group
```bash
Prompts:
- Enter user email:
- Current groups:
  1. /acme-inc-2/acme-inc-2-dedicated
  2. /serverless
  3. /premium-users
  
- Select group number to remove (or 0 to cancel):

Actions:
1. DELETE /users/{user-id}/groups/{group-id}

Display:
✓ Removed 'acme-user-1@acme-inc-2.com' from '/serverless'

Current groups (2):
  • /acme-inc-2/acme-inc-2-dedicated
  • /premium-users

Remove from another group? (y/n):

Press Enter to continue...
```

---

## Common Functions Module

```bash
# Authentication
get_token()                    # Get or refresh access token
check_token_expiry()          # Warn if token expires soon

# API Wrappers
api_get()                     # GET request wrapper
api_post()                    # POST request wrapper
api_put()                     # PUT request wrapper
api_delete()                  # DELETE request wrapper
check_response()              # Validate HTTP response

# Search/Lookup
find_group_by_path()          # Search group by path
find_group_by_id()            # Get group by ID
find_user_by_email()          # Search user by email
find_user_by_id()             # Get user by ID

# Display
draw_menu()                   # Draw menu with title
draw_table()                  # Draw ASCII table
show_error()                  # Display error message
show_success()                # Display success message
pause()                       # Pause with message

# Validation
validate_email()              # Validate email format
validate_tenant_name()        # Validate tenant naming
validate_group_name()         # Validate group naming
confirm_action()              # Confirm destructive action

# Utilities
extract_id_from_location()    # Parse Location header
format_date()                 # Format timestamp
truncate_id()                 # Show short ID
```

---

## Configuration & Setup

```bash
# Configuration
KEYCLOAK_URL="https://keycloak.apps.ethan-sno-kk.sandbox3469.opentlc.com"
REALM="maas-tenants"
CLIENT_ID="realm-admin-cli"
CLIENT_SECRET="${KEYCLOAK_ADMIN_SECRET}"  # From environment

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Settings
DEBUG=false                    # Enable debug output
AUTO_PAUSE=true               # Auto pause after operations
TOKEN_CACHE_FILE="/tmp/kc-token-cache"
LOG_FILE="/tmp/keycloak-admin.log"
```

---

## Error Handling

```bash
- HTTP error codes mapped to user-friendly messages
- Token expiry warnings (< 60 seconds)
- Network error detection
- Invalid input validation
- Confirmation for destructive operations
- Option to retry failed operations
```

---

## Sample Session Flow

```bash
User executes: ./keycloak-admin-menu.sh

1. Script authenticates → gets token
2. Shows main menu
3. User selects: 1 (Manage Tenants)
4. Shows tenant submenu
5. User selects: 1 (Create Tenant)
6. User enters: acme-inc-2
7. Script creates tenant and default group
8. Returns to tenant submenu
9. User selects: 0 (Back)
10. Returns to main menu
11. User selects: 3 (Manage Users)
12. ... continues workflow
```


