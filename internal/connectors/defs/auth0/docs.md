# Overview

Reads Auth0 users, clients, connections, roles, organizations, role assignments, and organization
memberships, and creates/updates users, clients, roles, and organizations, through the Auth0
Management API v2.

Readable streams: `users`, `clients`, `connections`, `roles`, `organizations`, `role_users`,
`organization_members`.

Write actions: `create_user`, `update_user`, `create_client`, `update_client`, `create_role`,
`update_role`, `create_organization`, `update_organization`.

Service API documentation: https://auth0.com/docs/api/management/v2.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Auth0 Management API access token, used directly as a
  Bearer token when set; never logged. Takes priority over the client_id/client_secret M2M mode
  below (dual-auth ordering, conventions.md 3).
- `audience` (optional, string); Audience form parameter sent on the M2M OAuth2 client-credentials
  token request, scoping the issued token to the Management API (Auth0 requires this explicitly;
  there is no engine mechanism to derive it from base_url - see docs.md's Known limits). Required
  when using client_id/client_secret auth; typically <base_url>/api/v2/.
- `base_url` (required, string); format `uri`; Auth0 tenant domain, e.g.
  https://your-tenant.us.auth0.com.
- `client_id` (optional, secret, string); Auth0 M2M application client id, used with client_secret
  to perform an OAuth2 client-credentials exchange against <base_url>/oauth/token when access_token
  is not set. Never logged.
- `client_secret` (optional, secret, string); Auth0 M2M application client secret; paired with
  client_id for the OAuth2 client-credentials exchange. Never logged.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`, `client_id`, `client_secret`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- OAuth 2.0 client credentials authentication with extra token parameters `audience` using
  `config.base_url`, `secrets.client_id`, `secrets.client_secret`, `config.audience` when `{{
  secrets.client_id }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2/clients`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 0; page size 50.

- `users`: GET `/api/v2/users` - records path `users`; query `include_totals`=`true`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 0; page size 50.
- `clients`: GET `/api/v2/clients` - records path `clients`; query `include_totals`=`true`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 0; page size
  50.
- `connections`: GET `/api/v2/connections` - records path `connections`; query
  `include_totals`=`true`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 0; page size 50.
- `roles`: GET `/api/v2/roles` - records path `roles`; query `include_totals`=`true`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 0; page size 50.
- `organizations`: GET `/api/v2/organizations` - records path `organizations`; query
  `include_totals`=`true`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 0; page size 50.
- `role_users`: GET `/api/v2/roles/{{ fanout.id }}/users` - records path `users`; query
  `include_totals`=`true`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 0; page size 50; fan-out; ids from request `/api/v2/roles`; id-list records path
  `roles`; id field `id`; id inserted into the request path; stamps `role_id`.
- `organization_members`: GET `/api/v2/organizations/{{ fanout.id }}/members` - records path
  `members`; query `include_totals`=`true`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 0; page size 50; fan-out; ids from request
  `/api/v2/organizations`; id-list records path `organizations`; id field `id`; id inserted into the
  request path; stamps `organization_id`.

## Write actions & risks

Overall write risk: creates/updates Auth0 users (including credentials), applications (clients),
RBAC roles, and organizations; approval required for every action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/api/v2/users` - kind `create`; body type `json`; required record fields
  `connection`; accepted fields `blocked`, `connection`, `email`, `email_verified`, `family_name`,
  `given_name`, `name`, `password`, `phone_number`, `phone_verified`, `username`; risk: external
  mutation; creates a new Auth0 user account (and, when password is set, a live credential);
  approval required.
- `update_user`: PATCH `/api/v2/users/{{ record.user_id }}` - kind `update`; body type `json`; path
  fields `user_id`; required record fields `user_id`; accepted fields `blocked`, `email`,
  `email_verified`, `family_name`, `given_name`, `name`, `password`, `phone_number`, `user_id`,
  `username`; risk: external mutation; updates an existing Auth0 user's profile/credential/blocked
  state; approval required.
- `create_client`: POST `/api/v2/clients` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `app_type`, `description`, `is_first_party`, `name`, `oidc_conformant`;
  risk: external mutation; registers a new Auth0 application (client), which can obtain its own
  OAuth2 credentials; approval required.
- `update_client`: PATCH `/api/v2/clients/{{ record.client_id }}` - kind `update`; body type `json`;
  path fields `client_id`; required record fields `client_id`; accepted fields `app_type`,
  `client_id`, `description`, `is_first_party`, `name`, `oidc_conformant`; risk: external mutation;
  updates an existing Auth0 application's configuration; approval required.
- `create_role`: POST `/api/v2/roles` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `name`; risk: external mutation; creates a new RBAC role
  (no permissions attached by default); approval required.
- `update_role`: PATCH `/api/v2/roles/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  external mutation; updates an existing RBAC role's name/description; approval required.
- `create_organization`: POST `/api/v2/organizations` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `display_name`, `name`; risk: external mutation; creates a
  new Auth0 organization (multi-tenant scoping unit); approval required.
- `update_organization`: PATCH `/api/v2/organizations/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `display_name`, `id`,
  `name`; risk: external mutation; updates an existing Auth0 organization's name/display_name;
  approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 7 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=3, destructive_admin=12, duplicate_of=6, non_data_endpoint=13, out_of_scope=25,
  requires_elevated_scope=1.
