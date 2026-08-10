# Overview

Reads public Docker Hub repositories and image tags for a configured target user or organization via
the Docker Hub registry API, and, when an optional Personal Access Token is configured, manages
personal and organization access tokens, organizations, groups (teams), invites, audit logs,
repositories, and (with a second, separately-configured SCIM token) SCIM-provisioned users.

Readable streams: `repositories`, `tags`, `repository_detail`, `tag_detail`.

Direct reads, status-only existence checks, reverse-ETL writes, and approved direct credential
exchanges cover the 50 remaining documented Docker Hub operations: repositories,
personal/organization access tokens, audit logs, groups (teams), invites, organization
settings/members, SCIM, and authentication. Together with the four stream-backed reads, all 54
documented Docker Hub API operations are modelled and implemented. Implementation is not a claim
that every operation is certified against every Docker account tier; see the phase verification
record for live-account coverage and provider-gated operations.

Service API documentation: https://docs.docker.com/docker-hub/api/latest/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://hub.docker.com/v2`; format `uri`; Docker Hub
  registry API base URL override for tests or self-hosted proxies.
- `auth_url` (optional, string); default `https://hub.docker.com/v2`; format `uri`;
  Docker Hub authentication API base URL. It is intentionally independent of `base_url` so a
  data API proxy cannot receive Docker Hub credential-exchange requests; override it only for a
  trusted HTTPS test or self-hosted authentication proxy. It must not include URL userinfo, query,
  or fragment components.
- `docker_username` (required, string); Docker Hub username used to authenticate `docker_pat`.
  Lowercase alphanumerics, underscores, and hyphens only.
- `namespace` (required, string); Docker Hub user or organization namespace whose repositories and
  tags to read. This is distinct from `docker_username`: a request for another namespace is honored
  rather than silently replaced by the authentication identity. Lowercase alphanumerics,
  underscores, and hyphens only.
- `docker_pat` (optional, secret string); a Docker Hub Personal Access Token for `docker_username`.
  Omitted, the connector stays exactly as before: read-only, unauthenticated, public repositories
  and tags only. When set, it authenticates account-scoped, organization, group/team, invite,
  access-token, audit-log, and repository-write commands. The three direct authentication
  exchanges use their supplied credential fields instead and deliberately carry no inherited
  connector authorization.
- `scim_bearer_token` (optional, secret string); a Docker Hub Enterprise SCIM API bearer token,
  distinct from `docker_pat`. Docker Hub's OpenAPI document declares SCIM (`/v2/scim/2.0/**`)
  under a separate `bearerSCIMAuth` security scheme from the rest of the API's `bearerAuth`; the
  two credentials are never interchangeable. Required only for the `scim-*` commands.
- `page_size` (optional, integer); default `100`; Page size (1-100) for the initial request of each
  paginated stream (Docker Hub's page_size query param); subsequent pages follow the API's own
  absolute next URL verbatim.
- `tier` (optional, string); default `free`; Docker Registry subscription profile for the cited
  image-pull quota. Set `paid` for Pro, Team, Business, or partner/unlimited Registry access. This
  does not manufacture a numeric rate limit for the separate Docker Hub API abuse guard.
- `auth_type` (optional, string); default `unauthenticated`; Docker Registry pull profile for the
  cited quota. Set `authenticated` only when the request path is backed by a Registry bearer token;
  the optional Hub API PAT/session exchange remains a distinct authentication mechanism.
- `registry_client_ip` (optional, string); public IPv4 address or IPv6 /64 used only as the
  non-secret scope of an unauthenticated Registry pull quota. Required when `base_url` targets
  `registry-1.docker.io` with `auth_type=unauthenticated`; it is not a credential and is never sent
  to Docker.
- `repository` (optional, string); Repository name (without the namespace prefix) the
  'tags'/'repository_detail'/'tag_detail' streams are scoped to. Required only when reading one of
  those streams.
- `tag` (optional, string); Tag name the 'tag_detail' stream reads a single tag record for (e.g.
  'latest'). Required only when reading the 'tag_detail' stream.

Default configuration values: `base_url=https://hub.docker.com/v2`,
`auth_url=https://hub.docker.com/v2`, `page_size=100`, `tier=free`,
`auth_type=unauthenticated`.

Authentication behavior:

- No `docker_pat` configured: no inherited connector authentication. The 4 public-read streams and
  the 3 status-only existence checks remain usable; the 3 approved direct authentication exchanges
  (`auth token create`, `auth login create`, and `auth 2fa-login create`) use their supplied
  credential fields with `auth_mode:none` and do not require `docker_pat`.
- `docker_pat` configured: every non-SCIM request authenticates via a `dockerhub` AuthHook
  (`internal/connectors/hooks/dockerhub`) that exchanges `docker_username`/`docker_pat` for a
  short-lived session bearer JWT through Docker Hub's API-root-normalized `POST /v2/users/login`
  endpoint under the independent `auth_url` base,
  caches it until 60s before the JWT's own `exp` claim (a conservative 4-minute cache is used if
  `exp` cannot be parsed), and sends it as `Authorization: Bearer <jwt>` on every subsequent
  request. The long-lived PAT itself is never sent as a static bearer token and never logged.
- `scim_bearer_token` configured: every `/v2/scim/2.0/**` request instead sends
  `Authorization: Bearer <scim_bearer_token>` — the same AuthHook routes by request path
  (`dualAuth`) and never substitutes one credential for the other. A SCIM command with no
  `scim_bearer_token` configured fails closed with a named error, even if a session JWT is already
  cached from a prior non-SCIM command.
- The two credentials are independently sufficient. `scim_bearer_token` alone (no `docker_pat`)
  authenticates every SCIM command, and every non-SCIM command then fails closed naming
  `docker_pat` rather than being sent unauthenticated. SCIM routing derives from the request path
  after the same `base_url`/API-root normalization used for dispatch, so a proxy `base_url` carrying
  its own path prefix still routes SCIM requests to the SCIM credential instead of falling back to
  the account session JWT.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/namespaces/{{ config.namespace }}/repositories`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

Pagination by stream: next_url: `repositories`, `tags`; none: `repository_detail`, `tag_detail`.

- `repositories`: GET `/namespaces/{{ config.namespace }}/repositories` - records path
  `results`; query `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from
  the response body; URL path `next`; next URLs stay on the configured API host.
- `tags`: GET `/namespaces/{{ config.namespace }}/repositories/{{ config.repository }}/tags`
  - records path `results`; query `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `repository_detail`: GET `/namespaces/{{ config.namespace }}/repositories/{{ config.repository }}` - single-object response; records at response root.
- `tag_detail`: GET `/namespaces/{{ config.namespace }}/repositories/{{ config.repository }}/tags/{{ config.tag }}` - single-object response; records at response root.

Direct-read commands (repositories/tokens/groups/invites/orgs/audit-logs/scim listed below) are
each a single bounded HTTP request, not a full paginated crawl — Docker Hub's own list endpoints
return a provider-default first page only per command invocation (a repo-wide, currently open
limitation tracked separately, not specific to this connector).

The 3 `*-check` commands (`repository check`, `repository tags check`, `repository tag check`) are
status-only: Docker Hub's `HEAD` existence-check endpoints never return a response body, so these
commands return `{"status_code": N}` — the HTTP status is the entire signal. Both answers the check
exists to give are results, not errors: `{"status_code": 200}` when the resource exists and
`{"status_code": 404}` when it does not. Every other status is still a failure — 401/403 report an
auth problem rather than a fact about the resource, and 429/5xx are the provider declining to
answer. This is a new runtime capability (`internal/connectors/engine/direct_read.go`'s HEAD
branch), narrowed to status-only HEAD operations; no other connector's commands and no non-HEAD
direct read are affected by it (a GET 404 remains an error).

## Write actions & risks

Requires `docker_pat` (or, for SCIM, `scim_bearer_token`). Reverse ETL mutations require plan,
preview, explicit approval, and execute; destructive actions (DELETE and group/org membership
removal) additionally require typed destructive confirmation.

- Repositories: `repository create`, `repository immutable-tags update`,
  `repository immutable-tags verify` (direct read), `repository group assign`.
- Personal access tokens: `access-tokens list`/`get` (direct read), `access-tokens create` (the
  provider may return the raw token once; runtime discards it by default. Add `--show-token` only
  to the approved execution to print it immediately, deliberately exposing it in terminal or JSON
  output where it cannot be redacted; it is never retained in plan/run history/state), `access-tokens update`, `access-tokens delete`
  (destructive).
- Organization access tokens: `org access-tokens list`/`get` (direct read), `org access-tokens
  create` (the provider may return a raw token once; runtime discards it by default. Add
  `--show-token` only to the approved execution to print it immediately, deliberately exposing it
  in terminal or JSON output where it cannot be redacted; it is never retained in plan/run
  history/state. `resources`, the token's repo/org scope grants, is a nested array of objects with
  no typed scalar leaf and is not flag-mapped — supply it via a reverse-ETL source record), `org access-tokens update`, `org
  access-tokens delete` (destructive).
- Audit logs: `audit-logs list`, `audit-logs actions list` (both direct reads).
- Authentication: `auth token create`, `auth login create`, and `auth 2fa-login create` are
  implemented approved direct writes. Their help and execution paths warn that supplied credential
  values are retained in plaintext local project state for the command plan. Each command accepts
  its credential field normally, or via `--from-env` and `--value-stdin`; prefer the latter two
  because argv can be observed by other local processes and shell history. The operator-visible
  plan sample redacts `secret`, `password`, `login_2fa_token`, and `code` as appropriate, while
  the execution record is retained after the warning so the approved request can run. The provider
  response, including a resulting token, is returned unchanged immediately and must be handled as
  secret material; session-token response bodies are not retained in reverse-run history. The
  internal `dockerhub` AuthHook continues to perform its separate session-login exchange when
  `docker_pat` is configured.
- Groups (teams): `groups list`/`get` (direct read), `groups create`, `groups replace` (PUT, full
  update), `groups update` (PATCH, partial update), `groups delete` (destructive), `groups members
  list` (direct read), `groups members add`, `groups members remove` (destructive).
- Invites: `invites bulk-create`, `invites cancel` (destructive), `invites resend`, `invites list`
  (direct read).
- Organizations & settings: `org settings get` (direct read), `org settings update` (the provider's
  `restricted_images` body is a nested object with three required booleans; this command flattens
  it into three top-level scalar flags one-for-one), `org members list` (direct read), `org members
  export` (binary CSV download), `org members update`, `org members remove` (destructive).
- SCIM (requires `scim_bearer_token`): `scim-resource-types list`/`get`, `scim-schemas list`/`get`,
  `scim-service-provider-config get`, `scim-users list`/`get` (all direct reads), `scim-users
  create`, `scim-users update` (PUT full replace; `enabled` is required by this command even though
  the provider defaults a missing value to `false`/deactivated, to prevent accidental
  deactivation-by-omission). The provider documents SCIM's media type as `application/scim+json`
  (RFC 7644); the declared write sends that media type.

## Known limits

- Batch defaults: read_page_size=100.
- The documented pull quota is selector-scoped to `registry-1.docker.io`, not to the Hub API host
  used by this connector's 54 documented management operations. The bundle declares 100
  unauthenticated pulls per 21,600-second window by public IP/IPv6-/64 and 200 authenticated free
  pulls per window by Docker username; paid Registry profiles deliberately match no fixed budget.
  Docker's separate Hub API abuse limiter publishes no numeric budget, so no synthetic limiter is
  declared for it. A bare 429 still follows the shared requester's provider `Retry-After` backoff.
- API coverage: all 54 documented Docker Hub operations are modelled and implemented (4
  stream-backed, 47 direct-read/status-check/reverse-write, and 3 approved direct credential
  exchanges). This is not certification: provider-plan, provider-permission, and Enterprise-only
  operations remain explicitly recorded in the phase verification evidence.
- Two independent optional credentials gate different command groups: `docker_pat` (most commands)
  and `scim_bearer_token` (SCIM only) — Docker Hub's own OpenAPI document declares these under
  separate security schemes (`bearerAuth` vs `bearerSCIMAuth`) and this bundle never substitutes
  one for the other.
- Direct-read commands are single bounded requests, not full paginated crawls (see Streams notes).
- Other cited artifact endpoints are explicitly classified in `api_surface.json`; no undocumented
  legacy endpoint is exposed by this bundle.
