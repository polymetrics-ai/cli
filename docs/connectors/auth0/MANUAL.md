# pm connectors inspect auth0

```text
NAME
  pm connectors inspect auth0 - Auth0 connector manual

SYNOPSIS
  pm connectors inspect auth0
  pm connectors inspect auth0 --json
  pm credentials add <name> --connector auth0 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Auth0 users, clients, connections, roles, organizations, role assignments, and organization memberships, and creates/updates users, clients, roles, and organizations, through the Auth0 Management API v2.

ICON
  asset: icons/auth0.svg
  source: official
  review_status: official_verified
  review_url: https://auth0.com/docs/api/management/v2

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  audience
  base_url
  mode
  access_token (secret)
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  users:
    primary key: user_id
    cursor: updated_at
    fields: blocked(), created_at(), email(), email_verified(), family_name(), given_name(), last_login(), logins_count(), name(), nickname(), picture(), updated_at(), user_id(), username()
  clients:
    primary key: client_id
    fields: app_type(), client_id(), description(), global(), is_first_party(), name(), oidc_conformant()
  connections:
    primary key: id
    fields: display_name(), id(), is_domain_connection(), name(), strategy()
  roles:
    primary key: id
    fields: description(), id(), name()
  organizations:
    primary key: id
    fields: display_name(), id(), name()
  role_users:
    primary key: role_id, user_id
    fields: email(), name(), picture(), role_id(), user_id()
  organization_members:
    primary key: organization_id, user_id
    fields: email(), name(), organization_id(), picture(), user_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /api/v2/users
    risk: external mutation; creates a new Auth0 user account (and, when password is set, a live credential); approval required
  update_user:
    endpoint: PATCH /api/v2/users/{{ record.user_id }}
    required fields: user_id
    risk: external mutation; updates an existing Auth0 user's profile/credential/blocked state; approval required
  create_client:
    endpoint: POST /api/v2/clients
    risk: external mutation; registers a new Auth0 application (client), which can obtain its own OAuth2 credentials; approval required
  update_client:
    endpoint: PATCH /api/v2/clients/{{ record.client_id }}
    required fields: client_id
    risk: external mutation; updates an existing Auth0 application's configuration; approval required
  create_role:
    endpoint: POST /api/v2/roles
    risk: external mutation; creates a new RBAC role (no permissions attached by default); approval required
  update_role:
    endpoint: PATCH /api/v2/roles/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing RBAC role's name/description; approval required
  create_organization:
    endpoint: POST /api/v2/organizations
    risk: external mutation; creates a new Auth0 organization (multi-tenant scoping unit); approval required
  update_organization:
    endpoint: PATCH /api/v2/organizations/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing Auth0 organization's name/display_name; approval required

SECURITY
  read risk: external Auth0 Management API read of user, client, and tenant configuration data, fanned out to per-role and per-organization membership lists
  write risk: creates/updates Auth0 users (including credentials), applications (clients), RBAC roles, and organizations; approval required for every action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect auth0

  # Inspect as structured JSON
  pm connectors inspect auth0 --json

AGENT WORKFLOW
  - Run pm connectors inspect auth0 before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
