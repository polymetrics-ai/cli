---
name: pm-auth0
description: Auth0 connector knowledge and safe action guide.
---

# pm-auth0

## Purpose

Reads Auth0 users, clients, connections, roles, organizations, role assignments, and organization memberships, and creates/updates users, clients, roles, and organizations, through the Auth0 Management API v2.

## Icon

- id: auth0
- asset: icons/auth0.svg
- source: official
- review_status: official_verified
- review_url: https://auth0.com/docs/api/management/v2

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- audience
- base_url (required)
- mode
- access_token (secret)
- client_id (secret)
- client_secret (secret)

## ETL Streams

- users:
  - primary key: user_id
  - cursor: updated_at
  - fields: blocked(boolean), created_at(string), email(string), email_verified(boolean), family_name(string), given_name(string), last_login(string), logins_count(integer), name(string), nickname(string), picture(string), updated_at(string), user_id(string), username(string)
- clients:
  - primary key: client_id
  - fields: app_type(string), client_id(string), description(string), global(boolean), is_first_party(boolean), name(string), oidc_conformant(boolean)
- connections:
  - primary key: id
  - fields: display_name(string), id(string), is_domain_connection(boolean), name(string), strategy(string)
- roles:
  - primary key: id
  - fields: description(string), id(string), name(string)
- organizations:
  - primary key: id
  - fields: display_name(string), id(string), name(string)
- role_users:
  - primary key: role_id, user_id
  - fields: email(string), name(string), picture(string), role_id(string), user_id(string)
- organization_members:
  - primary key: organization_id, user_id
  - fields: email(string), name(string), organization_id(string), picture(string), user_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_user:
  - endpoint: POST /api/v2/users
  - required fields: connection
  - risk: external mutation; creates a new Auth0 user account (and, when password is set, a live credential); approval required
- update_user:
  - endpoint: PATCH /api/v2/users/{{ record.user_id }}
  - required fields: user_id
  - risk: external mutation; updates an existing Auth0 user's profile/credential/blocked state; approval required
- create_client:
  - endpoint: POST /api/v2/clients
  - required fields: name
  - risk: external mutation; registers a new Auth0 application (client), which can obtain its own OAuth2 credentials; approval required
- update_client:
  - endpoint: PATCH /api/v2/clients/{{ record.client_id }}
  - required fields: client_id
  - risk: external mutation; updates an existing Auth0 application's configuration; approval required
- create_role:
  - endpoint: POST /api/v2/roles
  - required fields: name
  - risk: external mutation; creates a new RBAC role (no permissions attached by default); approval required
- update_role:
  - endpoint: PATCH /api/v2/roles/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing RBAC role's name/description; approval required
- create_organization:
  - endpoint: POST /api/v2/organizations
  - required fields: name
  - risk: external mutation; creates a new Auth0 organization (multi-tenant scoping unit); approval required
- update_organization:
  - endpoint: PATCH /api/v2/organizations/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing Auth0 organization's name/display_name; approval required

## Security

- read risk: external Auth0 Management API read of user, client, and tenant configuration data, fanned out to per-role and per-organization membership lists
- write risk: creates/updates Auth0 users (including credentials), applications (clients), RBAC roles, and organizations; approval required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Auth0's declared typed write actions.
- Usage: pm auth0 <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Reverse ETL writes
- Other Commands
  - create client apply - POST /api/v2/clients (create_client) [intent=reverse_etl availability=implemented write=create_client]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; registers a new Auth0 application (client), which can obtain its own OAuth2 credentials; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
  - create organization apply - POST /api/v2/organizations (create_organization) [intent=reverse_etl availability=implemented write=create_organization]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; creates a new Auth0 organization (multi-tenant scoping unit); approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
  - create role apply - POST /api/v2/roles (create_role) [intent=reverse_etl availability=implemented write=create_role]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; creates a new RBAC role (no permissions attached by default); approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
  - create user apply - POST /api/v2/users (create_user) [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; creates a new Auth0 user account (and, when password is set, a live credential); approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --connection (required)
  - update client apply - Typed action update_client [intent=reverse_etl availability=partial write=update_client]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v2/clients/{client_id} disagrees with covered api_surface path /api/v2/clients/{id}.; risk: external mutation; updates an existing Auth0 application's configuration; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v2/clients/{client_id} disagrees with covered api_surface path /api/v2/clients/{id}.; flags: --client-id (required)
  - update organization apply - PATCH /api/v2/organizations/{id} (update_organization) [intent=reverse_etl availability=implemented write=update_organization]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; updates an existing Auth0 organization's name/display_name; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
  - update role apply - PATCH /api/v2/roles/{id} (update_role) [intent=reverse_etl availability=implemented write=update_role]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation; updates an existing RBAC role's name/description; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
  - update user apply - Typed action update_user [intent=reverse_etl availability=partial write=update_user]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v2/users/{user_id} disagrees with covered api_surface path /api/v2/users/{id}.; risk: external mutation; updates an existing Auth0 user's profile/credential/blocked state; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v2/users/{user_id} disagrees with covered api_surface path /api/v2/users/{id}.; flags: --user-id (required)

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

## Commands

### Inspect as a manual

```bash
pm connectors inspect auth0
```

### Inspect as structured JSON

```bash
pm connectors inspect auth0 --json
```

## Agent Rules

- Run pm connectors inspect auth0 before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
