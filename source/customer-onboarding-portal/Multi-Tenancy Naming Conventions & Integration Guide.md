

# **External Identity Naming Conventions**

**Version 2.2 | March 2026**

---

# **1\. Purpose and Scope**

This document defines the canonical naming, identifier, and lookup conventions for:

* External corporate tenant identities  
* Human users authenticating via external Identity Providers (IdPs)  
* External robot identities (system users and client identities)

The document was developed to test the hypothesis that MaaS will operate in a federated IDP configuration for an AI Cloud Service Provider scenario.

This document **does not cover**:

* Internal platform users  
* Kubernetes ServiceAccount identities (in-cluster)  
* Non-federated Keycloak human users

  ---

# **2\. Identity Types**

## **2.1 Human Federated User**

A human user authenticated via a tenant-specific external IdP (e.g. Azure AD).

Characteristics:

* Identity originates from external IdP  
* Provisioned in Keycloak via federation (first login)  
* Bound to exactly one tenant  
* Represented in OpenShift via OIDC token

  ---

## **2.2 External Robot (Client Identity)**

### **Two Types of Robot Identity**

**Keycloak Client Robot**

* Used by external systems (MaaS clients, CI/CD pipelines, backup agents, monitoring tools) that run outside the cluster.  
* Authenticates to Keycloak to obtain a token, then uses that token to call OpenShift.

**Kubernetes ServiceAccount Robot**

* Used by workloads running inside the cluster (operators, sidecars, custom controllers).  
* Authenticates using OpenShift ServiceAccount tokens.  
* **This identity type is not covered in this document.**  
  ---

### **Characteristics**

* Identity originates from either:  
  * external IdP (system/service account), or  
  * platform-managed Keycloak client  
* Authenticates using client credentials or stronger mechanisms (e.g. private key JWT)  
* Not interactive (non-human)  
* Always associated with a single tenant  
* Authorised via platform-defined groups

  ---

# **3\. Canonical Identifiers**

## **3.1 Human User Identifiers**

| Identifier | Description | Source | Mutability |
| ----- | :---- | ----- | ----- |
| **Keycloak UUID** | Primary identity anchor | Keycloak | Immutable |
| **Keycloak Username** | Human-readable identifier | IdP | Mutable |
| **Email** | Contact and display identifier | IdP | Mutable |
| **Tenant Slug** | Tenant ownership identifier | Platform | Immutable |
| **OpenShift Username** | Projected identity used in OpenShift | Derived | Immutable |
| **Federated Identity Link** | IdP alias \+ IdP subject | Keycloak | Immutable |

---

## **3.2 Human Identity Rules**

### **Identity Anchor**

The **Keycloak UUID** is the authoritative identity for:

* RBAC ownership  
* audit logging  
* cross-system correlation

  ---

  ### **Keycloak Username**

* MUST be human-readable  
* SHOULD be email-derived  
* MUST be unique within the realm

Example:

fred.bloggs@acme.com

---

### **Tenant Slug**

* Globally unique identifier for a tenant  
* Assigned during onboarding  
* Stored as a **user attribute**  
* MUST NOT change after creation

Example:

acme-corp

---

### **OpenShift Username (Token Projection)**

The OpenShift username is generated dynamically in the token:

| oidc:\<tenant-slug\>:\<keycloak-uuid\> |
| :---- |

Example:

| oidc:acme-corp:f81d4fae-7dec-11d0-a765-00a0c91e6bf6 |
| :---- |

Properties:

* Deterministic  
* Immutable  
* Globally unique

How this works in practice is: When Keycloak issues the token for OpenShift, a mapper constructs `preferred_username` as `oidc:<tenant-slug>:<keycloak-uuid>`. Refer to Appendix for details.

---

### **Federated Identity Link**

Each user MUST have exactly one federated identity: (IdP alias, IdP subject)

Example:

| acme-corp-azure, 12345678-1234-1234-1234-123456789abc |
| :---- |

---

## **3.3 Robot Identifiers**

| Identifier | Description | Source | Mutability |
| :---- | :---- | :---- | :---- |
| **Robot ID** | Canonical robot identifier | Derived | Immutable |
| **Tenant Slug** | Owning tenant | Platform | Immutable |
| **Origin Type** | IdP or Keycloak client | Platform | Immutable |
| **Display Name** | Human-readable label | Platform or IdP | Mutable |
| **OpenShift Username** | Projected identity | Derived | Immutable |

---

## **3.4 Robot Identity Rules**

### **Robot ID Format**

All robots MUST follow:

robot:\<tenant-slug\>:\<name\>

Examples:

| robot:acme-corp:ci-pipelinerobot:acme-corp:backup-agent |
| :---- |

Properties:

* Globally unique  
* Encodes tenant ownership  
* Stable identifier for audit and RBAC

  ---

### **Origin Types**

#### **Keycloak Client Robot**

* Robot ID derived from client ID  
* Managed in Keycloak

#### **External IdP Robot**

* Robot ID derived from IdP identity  
* Must be mapped into platform format during token generation

  ---

### **OpenShift Username (Robots)**

| robot:\<tenant-slug\>:\<robot-id\> |
| :---- |

Example:

| robot:acme-corp:websearch-agent |
| :---- |

---

# **4\. Tenant Naming Rules**

## **4.1 Tenant Slug Format**

* Lowercase  
* Hyphen-separated  
* DNS-safe

Pattern:

| \[a-z0-9\]+(-\[a-z0-9\]+)\* |
| :---- |

Examples:

* acme-corp  
* widgets-inc  
* finbank-au  
  ---

  ## **4.2 Constraints**

* MUST be globally unique  
* MUST NOT be reused  
* MUST NOT be changed after assignment

  ---

  ## **4.3 Storage**

Tenant slug MUST be stored in:

* Keycloak user attributes  
* Keycloak client attributes (if applicable)  
* Platform tenant registry

  ---

# **5\. Group Naming Conventions**

All groups are **tenant-scoped**.

## **5.1 Tenant Access Groups**

A **tenant access group** is a Keycloak group that represents a predefined level of access across all resources owned by a tenant.

It follows the naming format:

tenant:\<tenant-slug\>:\<access-level\>

A tenant access group is used to grant **broad, tenant-wide permissions** by binding the group to Kubernetes Roles or ClusterRoles via RoleBindings or ClusterRoleBindings.

These groups:

* represent access levels such as `viewer`, `developer`, or `admin`  
* are platform-defined and created during tenant onboarding  
* **are the primary mechanism for expressing RBAC in OpenShift**  
* must not be derived directly from external identity provider group names  
* When a user logs into OpenShift these groups are included in the JWT.

An **access level** is a platform-defined label that represents a set of permissions, such as `viewer`, `developer`, or `admin`.

Access levels:

* are encoded in group names  
* do not directly correspond to Kubernetes Roles  
* are mapped to Kubernetes Roles or ClusterRoles through RBAC bindings  
* must be consistently defined across all tenants

Examples:

| tenant:acme-corp:viewertenant:acme-corp:developertenant:acme-corp:premium-tiertenant:acme-corp:admin |
| :---- |

---

## **5.2 Namespace Access Groups**

A **namespace access group** is a Keycloak group that represents a predefined level of access within a specific namespace belonging to a tenant.

It follows the naming format:

tenant:\<tenant-slug\>:namespace:\<namespace\>:\<access-level\>

A namespace access group is used to grant **fine-grained, namespace-scoped permissions** by binding the group to Roles or ClusterRoles via RoleBindings within that namespace.

These groups:

* allow different access levels across environments (e.g. dev vs prod)  
* support separation of duties and least privilege  
* enable multi-team isolation within a single tenant  
* are platform-defined and created during tenant onboarding

Examples:

| tenant:acme-corp:namespace:acme-dev:developerTenant:acme-corp:namespace:acme-prod:viewer |
| :---- |

## **5.3 Platform Groups**

### **Platform Group**

A **platform group** is any Keycloak group that follows the platform naming conventions and is used as the authoritative input to OpenShift RBAC.

| Platform Group├── Tenant Access Group└── Namespace Access Group |
| :---- |

Platform groups:

* are the only groups trusted for authorisation decisions  
* are created and managed by the platform during onboarding  
* are assigned to users via Keycloak identity provider mappers  
* are emitted in tokens and consumed by OpenShift for RBAC evaluation

External identity provider groups must be mapped into platform groups and are never used directly.

### **Group Mapping**

**Group mapping** is the process of translating external identity provider claims (such as Azure AD groups or roles) into platform groups.

This mapping:

* is configured in Keycloak using identity provider mappers  
* assigns users to pre-existing platform groups  
* does not create groups dynamically  
* ensures that only platform-defined group names are used for RBAC

## **5.3 Rules**

* Groups MUST be platform-defined during customer onboarding (or mapper will fail)  
* Groups MUST be the only RBAC authority  
* External IdP roles/groups MUST be mapped into platform groups  
* No direct trust of external group names


# **6\. Token Claim Conventions**

## **6.1 Human Token**

| { "sub": "\<keycloak-uuid\>", "preferred\_username": "oidc:\<tenant-slug\>:\<uuid\>", "email": "\<email\>", "name": "\<display-name\>", "groups": \[   "tenant:\<tenant-slug\>:\<role\>" \]} |
| :---- |

## **6.2 Robot Token**

| { "sub": "robot:\<tenant-slug\>:\<name\>", "preferred\_username": "robot:\<tenant-slug\>:\<name\>", "groups": \[   "tenant:\<tenant-slug\>:\<role\>" \]} |
| :---- |

## **6.3 Rules**

* `preferred_username` MUST be deterministic  
* `groups` MUST only contain platform groups  
* Tenant slug MUST always be present in identity projection

# **7\. Identity Derivation Rules**

This section defines the **rules that govern how a user’s identity is constructed and maintained inside the platform** so they are secure, stable, and deterministic platform identity that can be safely used for multi-tenant RBAC in OpenShift.

## **7.1 Tenant Context**

Tenant slug is derived from:

* IdP configuration (authoritative)

It MUST NOT be derived from:

* email domain  
* group names  
* user input

  ---

  ## **7.2 First Login (Humans)**

* Create Keycloak user  
* Assign UUID  
* Store tenantSlug  
* Import identity attributes  
* Assign baseline group  
* Apply mapped roles

  ---

  ## **7.3 Subsequent Login**

* Refresh profile attributes  
* Re-evaluate role mappings  
  Do NOT modify:  
  * UUID  
  * tenantSlug

  ---

  # **8\. Lookup and Support Workflows**

  ## **8.1 Human Lookup**

  email → Keycloak user → UUID → OpenShift username → groups

  ---

  ## **8.2 Tenant Lookup**

  oidc:\<tenant\>:\<uuid\> → tenantSlug → Keycloak user

  ---

  ## **8.3 Robot Lookup**

  robot:\<tenant\>:\<name\> → tenant → groups → permissions

  ---

  # **9\. Constraints and Non-Goals**

  ## **9.1 Supported**

* Single-tenant human identities  
* External IdP federation  
* Platform-defined RBAC groups  
* External robots via Keycloak or IdP identities

  ---

  ## **9.2 Not Supported (v1)**

* Multi-tenant human identities  
* Direct RBAC from IdP groups  
* In-cluster ServiceAccounts (covered elsewhere)  
* Dynamic tenant reassignment

  ---

# **10\. Examples**

## **10.1 Human**

Tenant: acme-corp  
Email: fred.bloggs@acme.com  
UUID: f81d4fae-...

Result:

| oidc:acme-corp:f81d4fae-... |
| :---- |

| tenant:acme-corp:developer |
| :---- |

---

## **10.2 Robot (External)**

robot:acme-corp:ci-pipeline

Result:

robot:acme-corp:ci-pipeline

tenant:acme-corp:admin

---

# **11\. Summary**

This model establishes:

* UUID as the identity anchor for humans  
* Robot ID as the identity anchor for system users  
* Tenant slug as the isolation boundary  
* Groups as the sole RBAC mechanism  
* Token projection as the OpenShift identity interface

It separates:

* identity origin (IdP vs platform)  
* identity representation (Keycloak)  
* identity projection (OpenShift)

## **Appendix A \- Keycloak System Configuration Requirements for External Identity Naming**

This appendix defines the Keycloak configuration needed to support the external identity naming conventions, especially the projected OpenShift username format `oidc:<tenant-slug>:<keycloak-uuid>`, tenant-scoped groups, and first-login provisioning for federated corporate users. It follows the existing design direction that the platform uses a single shared realm, one IdP per corporate tenant, and claim-based tenant attribution enforced through Keycloak and OpenShift RBAC.

### **A.1 Realm requirements**

The platform uses one shared Keycloak realm, and tenant isolation is enforced through IdP scoping, tenant attribution, and RBAC rather than separate realms. Automatic account linking must be disabled, and Trust Email must be disabled on all external IdPs so that upstream email values are informational only and are never used as a stable trust anchor or routing decision.

The realm must use explicit configuration for token lifetimes, session controls, brute force protection, authentication flows, IdPs, clients, and client scopes. The implementation requirements document is explicit that these are all deliberate decision points rather than defaults.

### **A.2 Per-tenant Identity Provider requirements**

Each corporate tenant must have exactly one Identity Provider entry in the `platform` realm. The IdP alias must follow the naming pattern `{tenant-slug}-{idp-type}`, such as `acme-corp-azure`, and this alias must remain stable after users have authenticated through it because Keycloak uses it as part of the stable federated identity linkage.

Each tenant IdP must be configured with a platform-defined custom first-login broker flow. That flow must create the local Keycloak user, set the tenant attribute from IdP configuration, and assign the user to the appropriate tenant group without relying on any tenant identifier asserted by the upstream IdP. The tenant attribute is set through a hardcoded mapper at the IdP level, and its value is the tenant slug configured by the platform team during IdP registration.

The upstream `sub` claim must be preserved as the federated identity link for audit and reference, but it must never be used as the OpenShift username or Kubernetes identity. The stable identity anchor remains the Keycloak UUID of the local user record.

### **A.3 Required user attributes and identity semantics**

For human federated users, Keycloak must maintain a local user record with an internally generated UUID. That UUID is the master identifier across Keycloak, OpenShift, support tooling, audit trails, and ownership metadata. The naming conventions document also establishes that the OpenShift username is a projection of tenant slug plus Keycloak UUID, while email remains a mutable human-readable search field only.

The local user record must therefore contain, at minimum, a stable UUID, a human-readable username, email and profile attributes imported from the IdP, and a `tenant_id` or `tenantSlug` attribute whose value is the tenant slug. That tenant attribute must be immutable after first creation. Your flow analysis correctly identifies that subsequent logins may refresh profile fields and group mappings, but must not alter UUID or tenant binding.

### **A.4 OpenShift client requirements**

The OpenShift OIDC client in Keycloak must issue a token whose claims are compatible with OpenShift RBAC processing. The implementation requirements document states that every Kubernetes-facing token must carry the same claim model consistently, and your analysis identifies the missing implementation detail: a client-side mapper must generate `preferred_username` in the form `oidc:${tenantSlug}:${uuid}`.

The OpenShift client configuration must therefore include a custom username mapper that constructs `preferred_username` from the stored tenant slug and the Keycloak local user ID. It must also include a group membership mapper that emits platform-defined groups in the `groups` claim, plus standard mappers for email and display name where required. Your analysis also notes the expected client shape: OpenID Connect client, confidential access type, and redirect URIs matching the OpenShift console.

The critical rule is that the OpenShift username is not the Keycloak internal username field. It is generated at token issuance time. This preserves human-readable administration in Keycloak while ensuring a deterministic OpenShift-visible identity. That separation is implicit in the original documents and made explicit by your analysis.

### **A.5 Group and authorisation requirements**

Tenant groups must follow the platform naming model `tenant:{tenant-slug}:{role}`, with optional namespace-scoped forms such as `tenant:{tenant-slug}:namespace:{namespace}:{role}`. These groups are the platform’s RBAC authority and must be the values emitted in the token `groups` claim. OpenShift evaluates those groups against RoleBindings and ClusterRoleBindings.

Each tenant IdP must therefore support at least two assignment mechanisms. First, a baseline default group may be assigned during first login, such as `tenant:acme-corp:viewer`. Second, optional upstream group mappings may translate external IdP groups into platform groups, with those mappings re-evaluated on subsequent logins. Your analysis recommends a hybrid model, and that fits the original design best because it preserves platform control of the RBAC taxonomy while allowing customer-specific federation.

### **A.6 External robot requirements**

The original requirements document defines external robots primarily as Keycloak client identities using client credentials or signed JWT, with credential storage and rotation handled through the platform secrets manager and robot registry.

If the revised naming model is extended to allow some external robot identities to originate in an upstream IdP, the same principles still apply: tenant association must come from platform-controlled configuration rather than upstream self-asserted tenant context, the resulting token must emit only platform-defined groups, and the robot must still map to a single tenant. The authoritative records for robot ownership, tenant binding, and credential lifecycle remain the robot registry and tenant registry.

### **A.7 Minimum required Keycloak objects**

At minimum, the Keycloak implementation must include one shared realm, one OIDC client for OpenShift, one IdP per corporate tenant, a custom browser or broker flow for first login, tenant-scoped groups, and the mapper set required to stamp tenant context, emit groups, and build the OpenShift username projection. This is consistent with the implementation requirements summary, which explicitly lists realm settings, authentication flows, IdPs, and mappers as deliberate configuration points.

A concise implementation checklist is:

* one shared `platform` realm  
* one IdP per tenant with stable alias  
* hardcoded `tenant_id` or `tenantSlug` mapper on each IdP  
* custom first-login broker flow  
* OpenShift client with custom `preferred_username` mapper  
* group membership mapper to `groups`  
* platform-defined tenant groups created ahead of time  
* tenant registry as the authoritative source of tenant slug, IdP alias, namespaces, and group paths  
* robot registry as the authoritative source for robot identities and ownership.

# **Appendix B \- Example First Login Flow (with Tenant Registry Details)**

This example shows how a first federated login creates the OpenShift username `oidc:<tenant-slug>:<keycloak-uuid>` for a corporate user, and how the **tenant registry acts as the authoritative control plane source of truth** throughout the flow.

Ideally the tenant registry is integrated with the end-to-end solution and part of the automation chain. For day-0 this could be as simple as a spreadsheet that is maintained outside the system by the Service Administrator.

## **B.1 Precondition \- Tenant Onboarding (Tenant Registry CREATE)**

Before any user can log in, the tenant must be onboarded through the platform control plane.

### **Step B.1.1 \- Create tenant record (CREATE)**

The platform creates a tenant record in the **tenant registry**:

* `tenantSlug = acme-corp`  
* `tennantUUID`  
* `displayName = Acme Corporation`  
* `idpAlias = acme-corp-azure`  
* `idpType = azure-ad, oidc, saml`  
* `authorisationEndpoint`  
* `tokenEndpoint`  
* `clientID`  
* `clientSecretRef = location of client secret (do not persist here)`  
* `signingKeyRef = location of signing key secret (do not persist here)`  
* `expectedScopes`  
* `claimMappingProfile`  
* `groupMappingProfile`  
* `jitProvisioningEnabled = boolean`  
* `namespaces = [acme-corp-dev, acme-corp-prod]`  
* `groups = [tenant:acme-corp:viewer, tenant:acme-corp:developer, tenant:acme-corp:admin]`  
* `idpProvisioned = boolean`  
* `namespacesProvisioned = boolean`  
* `groupsProvisioned = boolean`  
* `operationalState = active/suspended/retired`  
* `createdAt = timestamp`  
* `updatedAt - timestamp`

This is the **authoritative source of tenant metadata** and is used to help onboard a new tenant IDP and hold the source of truth for essential operational data such as the tenantSlug.

### **Step B.1.2 \- Provision dependent systems (READ from tenant registry)**

Provisioning automation (or manual lookup for day 0\) reads the tenant registry and configures:

**Keycloak**

* Create IdP: `acme-corp-azure`  
* Configure hardcoded mapper:  
  * `tenantSlug = acme-corp`  
* Configure group mappings (optional)

**OpenShift**

* Create namespaces: `acme-corp-*`  
* Create RoleBindings referencing:  
  * `tenant:acme-corp:*`

**Platform groups**

* Create Keycloak groups:  
  * `tenant:acme-corp:viewer`  
  * `tenant:acme-corp:developer`  
  * `tenant:acme-corp:admin`

### **Step B.1.3 \- Persist configuration state (UPDATE)**

Provisioning status is written back to the tenant registry:

* `idpProvisioned = true`  
* `namespacesProvisioned = true`  
* `groupsProvisioned = true`

The tenant is now **active**.

## **B.2 Step 1 \- User starts login in OpenShift**

Fred visits the OpenShift console and selects login.

* OpenShift redirects to Keycloak  
* Keycloak determines tenant via:  
  * home realm discovery, or  
  * user selection of IdP

Fred is routed to: acme-corp-azure

## **B.3 Step 2 \- External IdP authenticates Fred**

Azure AD authenticates Fred and returns:

* `sub` (Azure object ID)  
* `email`  
* `name`  
* optional `groups`

These claims are **untrusted for tenancy** and are treated as input only.

## **B.4 Step 3 \- Keycloak checks for federated identity (READ)**

Keycloak queries its local store: (IdP alias, upstream sub)

Result:

* Not found → first login

Keycloak starts the **first-login broker flow**.

## **B.5 Step 4 \- Keycloak creates the local user (CREATE)**

Keycloak creates a new user:

* generates UUID: f81d4fae-7dec-11d0-a765-00a0c91e6bf6

* imports:  
  * email  
  * display name

This UUID becomes the **global identity anchor**.

## **B.6 Step 5 \- Resolve tenant context (READ from tenant registry → WRITE to Keycloak)**

Keycloak determines tenant context:

* uses IdP alias: `acme-corp-azure`  
* resolves tenant via configuration derived from tenant registry

Keycloak writes user attribute:

tenantSlug \= acme-corp

This value is:

* platform-controlled  
* immutable after creation

## **B.7 Step 6 \- Store federated identity link (CREATE)**

Keycloak stores: (acme-corp-azure, upstream sub)

This enables:

* future login correlation  
* audit traceability

## **B.8 Step 7 \- Assign platform groups (READ \+ COMPUTE)**

Keycloak assigns groups:

### **Baseline assignment**

From tenant registry definition: tenant:acme-corp:viewer

### **Optional mapped roles**

If configured:

Azure AD group → platform group mapping

Example:

Azure "Engineering" → tenant:acme-corp:developer

Only **platform-defined groups** are applied.

## **B.9 Step 8 \- Token generation (DERIVED via mapper)**

Keycloak issues a token for OpenShift.

Mapper constructs:

| preferred\_username \=  "oidc:" \+ tenantSlug \+ ":" \+ user.id |
| :---- |

Result:

| oidc:acme-corp:f81d4fae-7dec-11d0-a765-00a0c91e6bf6 |
| :---- |

Token includes:

| sub \= \<uuid\>groups \= \[tenant:acme-corp:\*\] |
| :---- |

## **B.10 Step 9 \- OpenShift processes identity (CREATE or READ)**

OpenShift receives token.

### **If user does not exist:**

* CREATE User:

|  oidc:acme-corp:\<uuid\> |
| :---- |

**If user exists:**

* READ existing user  
* RBAC evaluation:  
  * matches `groups` to RoleBindings

## **B.11 Step 10 \- Audit and support linkage (READ)**

Support systems can resolve identity via:

email → Keycloak → UUID → OpenShift username → tenant registry

Tenant registry provides:

* tenant metadata  
* namespaces  
* contact details

## **B.12 Result**

Fred is now consistently represented:

| System | Identity |
| ----- | ----- |
| Keycloak | UUID |
| OpenShift | `oidc:acme-corp:<uuid>` |
| Support | email \+ tenantSlug |

The tenant registry remains the **anchor for tenant context**, while Keycloak anchors identity.

## **B.13 Subsequent logins (READ \+ UPDATE)**

On future logins:

### **Keycloak**

* READ federated identity link  
* reuse UUID  
* do not modify:

  * UUID  
  * tenantSlug

### **Attribute refresh (UPDATE)**

* email  
* display name

### **Group evaluation (RECOMPUTE)**

* re-apply mapping rules

### **Token regeneration (DERIVED)**

* same `preferred_username`

## **Summary of Tenant Registry CRUD in Flow**

| Stage | Operation | Description |
| ----- | ----- | ----- |
| Onboarding | **CREATE** | Tenant record created |
| Provisioning | **READ** | Used to configure Keycloak/OpenShift |
| Provisioning | **UPDATE** | Status updated |
| Login | **READ** | Tenant context resolved |
| Runtime | **READ** | Used by support and automation |

## **Key Architectural Insight**

The tenant registry is:

* the **source of truth for tenancy**  
* used at **provisioning time and login time**  
* never overridden by upstream IdP data

While:

* Keycloak is the **source of truth for identity**  
* OpenShift is the **enforcement layer**

## **Appendix 3: Field Data Dictionary**

| Field name | Source of its value |
| ----- | ----- |
| **Organisation name** | **Customer** \- supplied during tenant onboarding. The slug is then derived from this name by platform automation or the platform team. |
| **Tenant slug** | **Service provider / platform** \- derived from the organisation name at onboarding time, recorded in the tenant registry, and then configured into Keycloak. It is stable and must not change after activation. |
| **Tenant record** | **Service provider / platform** \- created during onboarding in the platform registry. |
| **Keycloak IdP alias** | **Service provider / platform** \- constructed from `{tenant-slug}-{idp-type}` when the tenant IdP is registered. |
| **IdP type** | **Other \- selected from the tenant’s chosen provider** such as Azure, Google, Okta, SAML, or OIDC, then encoded by the platform into the IdP alias. |
| **`tenant_id` / `tenantSlug` user attribute in Keycloak** | **Mapper** \- set by a **Hardcoded Attribute mapper** on the tenant IdP. The configured value is the tenant slug that the platform entered at IdP registration time. It is not taken from upstream claims. |
| **Upstream IdP `sub`** | **Customer IdP** \- asserted by the external identity provider after authentication. Keycloak stores it as the federated identity link. |
| **Federated identity link `(IdP alias, upstream sub)`** | **Other \- created by Keycloak** by combining the tenant IdP alias with the upstream `sub` received from the customer IdP. |
| **Keycloak UUID (`user.id`)** | **Service provider / Keycloak** \- generated by Keycloak when the local federated user is first created. |
| **Keycloak username** | **Other \- imported by Keycloak from IdP profile data** according to the configured broker/import behaviour. In your current model this should remain human-readable, typically email-derived. |
| **Email** | **Customer IdP** \- asserted by the external IdP and imported into Keycloak as a mutable profile field. It is informational, not a trust anchor. |
| **Display name / full name** | **Customer IdP** \- asserted by the external IdP and imported into Keycloak. |
| **Baseline tenant group** | **Service provider / Keycloak** \- assigned by platform-defined first-login flow or default group logic for that tenant. |
| **Mapped role groups** | **Mapper** \- produced by IdP group-to-platform-group mapping rules configured by the platform for that tenant. The upstream group names come from the customer IdP, but the emitted platform groups are controlled by the platform. |
| **`groups` token claim** | **Mapper** \- generated by the Keycloak group membership mapper from the user’s platform group memberships. |
| **`sub` token claim for OpenShift** | **Service provider / Keycloak** \- set to Keycloak’s internal UUID for the local user record, not the upstream IdP `sub`. |
| **`preferred_username` token claim** | **Mapper** \- generated by the OpenShift client username mapper as `oidc:<tenant-slug>:<keycloak-uuid>`, using stored `tenantSlug` plus `user.id`. |
| **OpenShift username** | **Other \- created by OpenShift from the token claim**. OpenShift consumes `preferred_username` and creates the cluster user identity from that projected value. |
| **Namespace name** | **Service provider / platform** \- derived from `{tenant-slug}-{environment}` when namespaces are provisioned. |
| **RoleBinding subject group names** | **Service provider / platform** \- created from platform naming conventions and matched against token `groups`. |
| **Robot client ID** | **Service provider / platform** in the original requirements, using `robot-{tenant-slug}-{purpose}`. If you move to customer-defined robot identities in the external IdP, this would need a separate federation model because the current requirements assume Keycloak client robots.  |

## **Appendix 4: Tenant Registry**

## **Tenant Registry**

The **tenant registry** is the platform’s authoritative system of record for tenant identity and tenant configuration. It is owned by the service provider as part of the platform control plane and is not a native feature of Keycloak or OpenShift. Its purpose is to hold the canonical metadata for every tenant so that all other systems \- especially Keycloak, namespace provisioning, RBAC automation, support tooling, and audit workflows \- derive tenant-specific configuration from one trusted source.

For each tenant, the registry stores at minimum:

* tenant slug  
* tenant display name  
* Keycloak IdP alias  
* namespace names  
* tenant tier  
* tenant contact details  
* Keycloak group paths or equivalent platform group definitions.

The tenant registry is created and maintained by platform onboarding and lifecycle automation. During corporate tenant onboarding, the platform creates the tenant record in the registry first, then uses that record to configure the tenant’s IdP in Keycloak, provision the initial namespaces, and create the tenant’s Keycloak group hierarchy. In this way, the tenant registry becomes the root of truth for tenant onboarding, suspension, offboarding, and IdP migration workflows.

The registry also provides the stable reference point needed for naming conventions. Tenant slugs are derived at onboarding time, must remain stable after activation, and are used consistently across IdP aliases, group names, namespace names, audit lookups, and projected OpenShift usernames. The naming documents assume that support and automation can resolve a tenant slug back to full tenant details through the tenant registry.

Operationally, the tenant registry is consumed by:

* Keycloak provisioning automation, to create or update tenant IdP definitions and mapper configuration  
* namespace provisioning automation, to create the correct namespaces and labels  
* RBAC automation, to create and bind the correct tenant groups  
* support tooling, to resolve slugs, IdP aliases, namespaces, and contacts during incident handling  
* lifecycle automation, for suspension, reactivation, migration, and offboarding.

## **Appendix 5: System Architecture**

![][image1]

## **Appendix A: To-do List**

### **Robot identity**

- [ ] Define the external robot authentication flows end to end.

- [ ] Specify which robot types are supported in v1.

- [ ] Define how tenant context is assigned to robots.

- [ ] Define robot token claim rules, especially `sub`, `preferred_username`, and `groups`.

- [ ] Define whether robots may use tenant access groups, namespace access groups, or both.

- [ ] Define robot onboarding, rotation, suspension, and revocation behaviour.

### **Token contract**

- [ ] Add a normative token contract section for human users.

- [ ] Add a normative token contract section for robot identities.

- [ ] Specify required claims, optional claims, formats, and stability guarantees.

- [ ] Decide whether `tenantSlug` is emitted as its own claim or only encoded in `preferred_username`.

### **Group mapping**

- [ ] Define the supported Keycloak mapping patterns from external IdP claims to platform groups.

- [ ] Define how baseline group assignment works.

- [ ] Define conflict handling where multiple mappings apply.

- [ ] Define behaviour when no mapping exists.

- [ ] Define whether mapping is static only or allows pattern-based rules.

### **Failure and rejection behaviour**

- [ ] Define what happens when tenant resolution fails.

- [ ] Define what happens when required Keycloak groups do not exist.

- [ ] Define what happens when the tenant is suspended or retired.

- [ ] Define what happens when the IdP configuration is invalid or incomplete.

- [ ] Define whether failures result in login denial, login without access, or administrative error state.

- [ ] Define required audit events for all failure cases.

### **Tenant registry runtime model**

- [ ] Decide whether the tenant registry is provisioning-time only or also runtime authoritative.

- [ ] Define the relationship between tenant registry state and realised Keycloak configuration.

- [ ] Define drift detection and reconciliation rules between the registry and Keycloak.

- [ ] Define which fields are authoritative in the registry versus in Keycloak.

### **Deprovisioning and lifecycle**

- [ ] Add tenant suspension workflow.

- [ ] Add tenant retirement and offboarding workflow.

- [ ] Add user deprovisioning workflow for removal from external IdP.

- [ ] Define timing of access revocation.

- [ ] Define expected token/session invalidation behaviour after suspension or removal.

- [ ] Define robot deprovisioning workflow.

### **IdP alias rules**

- [ ] Define uniqueness requirements for `idpAlias`.

- [ ] Define allowed character set and length limits.

- [ ] Define the exact relationship between `tenantSlug`, IdP type, and alias.

- [ ] Define whether alias is immutable after activation.

### **Namespace ownership and enforcement**

- [ ] State explicitly whether a namespace may belong to only one tenant.

- [ ] Define how cross-tenant namespace bindings are prevented.

- [ ] Define the enforcement mechanism for namespace ownership.

- [ ] Define whether shared namespaces are unsupported in v1.

### **RBAC contract**

- [ ] Add standard access-level to RBAC mappings, for example `viewer`, `developer`, `admin`.

- [ ] State whether these mappings are globally fixed or tenant-customisable.

- [ ] Add example RoleBinding and ClusterRoleBinding patterns.

- [ ] Define when tenant access groups are used versus namespace access groups.

### **Audit and correlation**

- [ ] Define the canonical identifiers that must appear in audit logs.

- [ ] Define how events are correlated across tenant registry, Keycloak, and OpenShift.

- [ ] Define whether `tenantUUID` must be logged alongside `tenantSlug`.

- [ ] Define audit expectations for first login, subsequent login, mapping changes, suspension, and deprovisioning.

### **Human versus robot policy differences**

- [ ] Define which access levels are allowed for humans versus robots.

- [ ] Define whether robots may hold admin-equivalent groups.

- [ ] Define least-privilege expectations for robot identities.

- [ ] Define whether namespace-scoped robot access is preferred over tenant-wide access.

### **Policy, quota, and billing linkage**

- [ ] Define how identity attributes feed into rate limiting and quota policy.

- [ ] Define which identifiers are used for metering and billing correlation.

- [ ] Define whether policy engines should use `groups`, `preferred_username`, `tenantSlug`, or `tenantUUID`.

- [ ] Define how tenant suspension affects policy enforcement.

### **Editorial clean-up**

- [ ] Normalise `role` versus `access-level` terminology throughout the document.

- [ ] Correct field name inconsistencies such as `tenant_id` versus `tenantSlug`.

- [ ] Correct typos and formatting issues, including `tennantUUID`.

- [ ] Ensure all examples use one consistent naming pattern.

[image1]: <data:image/png;base64,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CN30uPVWHZUZool22nEbZZZlCV0AgIQgdBGx0O6683bKLMsSugAACUHopsvcS6bMEE2z8247KrMsS+gCACQEoYuIhXbvPTopsyxL6AIAJAShmz5HDX5UmSGa5L577aLMsiyhCwCQEIRu+rz80jOUGaJJHrDvbsosyxK6AAAJQeimU4fP6qLBHnTAHsosyxK6AAAJQegiYqHtetDeyizLEroAAAlB6KbXG6+5QJkhmuDhXfdVZlmW0AUASAhxAcb0Wl5WpswQk/agvTZTZllXvvYCAABAPMrlAUDCDJcHAAAAALWhkTwASJh35QEAAABAbVkjDwAS5G15AAAAAFAX+BwgmMKb8gAAAACgrnwvDwASYKA8AAAAACgEc+QBQIl5RR4AAAAAFIp6OVvKQ4AS8ZI8AAAAACg0fG4XkuA5eQAAAABQLJbJA4Ai8ow8AAAAACg2vMObMC7YwunyawsAAADJ09gheBNDriVILYQuAABAChDR+2PODvIOKDxyLUFqIXQBAAAsoDxnfacyiIXdo7uhJviVJFZ9wzN537JlyyL75dvIc+HHH3+s7K8tNb3tPffc465du1Ye15gZM2bIIyMoKyvzVwldAAAAixERBDVEFNKGRYC/LS8FM2fOVGY6vvvuu2BdHC+fb8cdd/SW9957b+R8EydO9LbvvvvuYCb47W9/G9kO32batGnujz/+6N2nmH/99dfBMeH7zbcuGDJkSDB/6qmngrnYXr16dd7QbdWqVbD+4Ycfekv5vNXdp7yeb/uRRx7Ju2/WrFne+nvvveePCV0AAIAM8D95AHr8cArjb8tLHxFX8iwffuiGj/3yyy+97eeffz6YffbZZ95Svj/5PuTtfLOzzjrLW+60007e8phjjvGW4eNWrFgR2fbfcfZnX3zxRWRbECd05ccS3h41alRke+7cud5Sdxt5uddee3nLLbbYwlvecccd3lKwIbIJXQAAgAyxTh6AiiikDYsAf1tehsk3k8kXusOHD/e2t91222Am061bN28p34e8nW82efLkyHa+0M23nW8W3o4Tuj7idv/973+rPZ88++qrryLb8vKwww7zlhdccIG3DPPqq6+KBaELAACQMUQoQDX4seSvhkaR2bp167zP58r7169fnzfgBH7oNm7c2Fsecsgh3lK+L3kZJ3Q7dOjgLbfeeutgJpDP5Yfu9ttv7y3H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>