# Overview

Box reads and writes through the Box Platform REST API (`https://api.box.com/2.0`, OpenAPI 3.0
spec at `https://raw.githubusercontent.com/box/box-openapi/main/openapi.json`) using the OAuth2
client-credentials grant (Server Authentication). This bundle originally migrated
`internal/connectors/box` (the hand-written legacy connector, read-only); this Pass B revision
expands to the account/enterprise-level catalog surface reachable from this connector's actual
config shape (see `api_surface.json` for every one of the 294 endpoints, each `covered_by` a
stream/write or `excluded` with a reason). The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide `client_id`/`client_secret` (both `x-secret`) for a Box Server Authentication with
Client Credentials Grant app. The engine's `oauth2_client_credentials` auth mode exchanges them
at `token_url` (`https://api.box.com/oauth2/token` by default) for a short-lived bearer token,
scoped by two Box-specific token-request form params sent via `auth[].extra_params`
(`box_subject_type`, `box_subject_id`) — matching legacy's `authenticator`/`boxSubject` exactly.
`box_subject_type` defaults to `enterprise` (the application service account); set it to `user` to
scope the token to a specific user instead. `box_subject_id` is the enterprise id or user id being
scoped to.

## Streams notes

12 read streams total.

- `users`, `groups`, `collections`, `folder_items` are the 4 original legacy-parity streams:
  `GET` against a Box list endpoint returning the `{entries:[...], offset, limit, total_count}`
  envelope (`records.path: "entries"`), primary key `["id"]`, pagination `offset_limit`
  (`limit_param: limit`, `offset_param: offset`, `page_size: 100`), stopping on a short/empty page.
  `folder_items` reads `/folders/{{ config.folder_id }}/items` (`folder_id` defaults to `"0"`, the
  root folder).
- `webhooks`, `retention_policies`, `legal_hold_policies`, `storage_policies`, `sign_requests`,
  `metadata_templates` are new Pass B streams using Box's **marker-based** cursor pagination
  (`type: "cursor"`, `cursor_param: "marker"`, `token_path: "next_marker"`) — a distinct pagination
  shape from the 4 offset/limit streams above, confirmed against the published OpenAPI response
  envelopes (`Webhooks`/`RetentionPolicies`/etc. all declare `next_marker`/`prev_marker` fields).
  `metadata_templates` reads the enterprise-scoped `/metadata_templates/enterprise` endpoint and
  renames 3 camelCase raw fields (`templateKey`/`displayName`/`copyInstanceOnItemCopy`) via
  `computed_fields`.
- `terms_of_services` uses `pagination: {"type": "none"}` — Box's real API has no pagination
  parameters on this endpoint at all (confirmed against the OpenAPI spec: no `limit`/`marker`/
  `offset` parameters declared), returning every terms-of-service record in one response.
- `pending_collaborations` reads `/collaborations` with a REQUIRED `status=pending` query param
  (Box's API only supports listing the `pending` subset account-wide; non-pending collaborations
  have no account-wide list endpoint at all, only per-file/per-folder/per-group scoped lists) —
  declared as a static `query: {"status": "pending"}` entry.

`users`, `groups`, `folder_items`, `webhooks`, `retention_policies`, `legal_hold_policies`,
`terms_of_services`, `pending_collaborations` expose `modified_at`/`created_at` as a schema cursor
candidate, but Box's list endpoints have no server-side incremental filter parameter, so no
`incremental` block is declared on any stream (per §8 rule 2) — full refresh only.

`check` issues a single bounded `GET /users?limit=1`, mirroring legacy's `Check` implementation
exactly.

## Write actions & risks

9 write actions (`capabilities.write` is now `true`):

- `create_group`/`update_group`/`delete_group` — manage a Box enterprise group
  (`POST /groups`, `PUT /groups/{id}`, `DELETE /groups/{id}`; delete is idempotent, 404 counts as
  written).
- `create_webhook`/`update_webhook`/`delete_webhook` — manage a webhook subscription that POSTs
  event payloads to an external address (`POST /webhooks`, `PUT /webhooks/{id}`,
  `DELETE /webhooks/{id}`).
- `create_collaboration`/`update_collaboration`/`delete_collaboration` — grant, change, or revoke
  a user/group's access to a Box file or folder (`POST /collaborations`,
  `PUT /collaborations/{id}`, `DELETE /collaborations/{id}`).

Every write action requires operator approval (`metadata.json`'s `risk.approval`); the 3 delete
actions additionally set `confirm: "destructive"`.

## Known limits

- **Conformance dynamic checks are skipped** (`metadata.json`'s `conformance.skip_dynamic`, bundle
  level, unchanged from pre-expansion): `oauth2_client_credentials` auth's `token_url` is a
  separate declared `config.token_url` property that conformance's replay-server rewiring cannot
  reach, so every auth-resolving dynamic check (including `write_request_shape` for the 9 new write
  actions) would otherwise fail identically and uninformatively. Static checks (spec/schema
  validity, `interpolations_resolve`, docs/fixtures presence, secret redaction, `surface_complete`)
  still run and pass. The read/pagination/schema-projection/write-request shapes for every new
  stream and write are proven by structural review against Box's published OpenAPI spec (fetched
  live from `github.com/box/box-openapi` during this revision) rather than a `paritytest/box`
  package, matching the bundle-level precedent this connector already established.
- **Scope is account/enterprise-level catalog resources plus the one connector-configured
  `folder_id`, not a full file/folder-tree crawl.** This connector's `spec.json` has no config
  surface for enumerating an arbitrary set of file/folder ids (only a single `folder_id`), so the
  ~86 file/folder/web-link-item-scoped endpoints (content upload/download, per-file/folder
  metadata/comments/tasks/versions/shared-links/watermarks — anything requiring a specific
  `file_id`/`folder_id` beyond the one configured `folder_id`, or a `web_link_id`) are out of scope
  — see `api_surface.json`'s `scope` field and per-endpoint `excluded` entries. This mirrors
  `github`'s own precedent of narrowing `api_surface.json` to the endpoints a real config shape can
  reach rather than enumerating literally everything the API publishes.
- **Several per-parent-scoped list endpoints are deferred as `fan_out` candidates, not modeled this
  pass**: `/groups/{group_id}/memberships`, `/users/{user_id}/memberships`,
  `/users/{user_id}/email_aliases`, `/retention_policies/{id}/assignments` would each require
  iterating every id from an already-covered stream (groups/users/retention_policies respectively)
  using the engine's `fan_out` dialect — judged lower-value than the 8 new streams already covered
  this pass; candidates for a future increment.
- Higher-risk account-admin writes (`create`/`update`/`delete` user, retention/legal-hold policy
  create/update/delete, session termination, ownership transfer) are deliberately excluded — see
  `api_surface.json`'s `destructive_admin` category entries — as either irreversible
  enterprise-identity actions or carrying legal/compliance consequences beyond a generic
  reverse-ETL write's risk profile.
- Documented parity deviation (unchanged from pre-expansion): legacy only sets the
  `box_subject_id` token-request form param when non-empty; the engine's `auth[].extra_params`
  dialect hard-errors on an unresolved config key rather than supporting `stream.Query`'s
  `omit_when_absent` tolerance, so `box_subject_id` relies on `spec.json`'s
  `default`-materialization mechanism with no explicit default value (`""` when unset) instead. See
  `docs/migration/conventions.md`'s parity-deviation ledger.
- `folder_id` must be a Box numeric folder id; not re-implemented as a `spec.json` pattern
  constraint (an invalid `folder_id` surfaces as a Box API error instead).
- `users`/`groups`/`folder_items`/`webhooks`/`retention_policies`/`legal_hold_policies`/
  `terms_of_services`/`pending_collaborations` are full-refresh only (no server-side incremental
  filter parameter on any of these Box endpoints), even though each declares a schema cursor
  candidate.
