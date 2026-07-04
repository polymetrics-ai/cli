# Overview

Box Data Extract reads and writes Box folder-scoped file data through the Box REST API
(`https://api.box.com/2.0` by default, OpenAPI 3.0 spec at
`https://raw.githubusercontent.com/box/box-openapi/main/openapi.json`) using the OAuth2
client-credentials grant. This bundle originally migrated `internal/connectors/box-data-extract`
(the hand-written legacy connector, read-only); this Pass B revision expands to per-file detail
metadata and a file rename/description write, scoped to the connector-configured folder
(`box_folder_id`) — see `api_surface.json` for the full endpoint-by-endpoint reasoning. The legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id`/`client_secret` (both `x-secret`) for a Box Server Authentication with Client
Credentials Grant app. The engine's `oauth2_client_credentials` auth mode exchanges them at
`token_url` (`https://api.box.com/oauth2/token` by default) for a short-lived bearer token, scoped
by two Box-specific token-request form params sent via `auth[].extra_params` (`box_subject_type`,
`box_subject_id`) — matching legacy's `requester`/`box_subject_type`/`box_subject_id` construction
exactly. `box_subject_type` defaults to `enterprise` (the application service account); set it to
`user` to scope the token to a specific user instead. `box_subject_id` is the enterprise id or user
id being scoped to.

## Streams notes

- `files` reads `/folders/{{ config.box_folder_id }}/items` (`box_folder_id` defaults to `"0"`,
  the root folder), records at `entries` (Box's `{entries:[...], offset, limit, total_count}`
  envelope), primary key `["id"]`. Pagination is `offset_limit` (`limit_param: limit`,
  `offset_param: offset`, `page_size: 100`), stopping on a short/empty page.
- `file_details` is a new Pass B stream: a `fan_out` detail-GET (`GET /files/{{ fanout.id }}`) over
  every id the `files` stream's own listing request returns (`fan_out.ids_from.request`, reusing
  the identical `/folders/{{ config.box_folder_id }}/items` listing, `records_path: "entries"`,
  `id_field: "id"`), stamping the source id onto each emitted record's `file_id` field. Schema
  fields (`sha1`, `size`, `path_collection`, `created_by`/`modified_by`/`owned_by`,
  `content_created_at`/`content_modified_at`, `item_status`, etc.) are derived directly from Box's
  published OpenAPI `File` response schema — richer than the bare `id`/`type`/`name` the `files`
  list endpoint itself returns, which is the entire point of a detail-GET fan-out.

`check` issues a single bounded `GET /folders/{{ config.box_folder_id }}/items?limit=1`, mirroring
legacy's `Check` implementation exactly.

## Write actions & risks

1 write action (`capabilities.write` is now `true`): `update_file` renames a file and/or updates
its description (`PUT /files/{{ record.id }}`, `path_fields: ["id"]`, body `{name?, description?}`)
— requires operator approval.

## Known limits

- **Conformance dynamic checks are skipped** (`metadata.json`'s `conformance.skip_dynamic`, bundle
  level, unchanged from pre-expansion): `oauth2_client_credentials` auth's `token_url` is a
  separate declared `config.token_url` property that conformance's replay-server rewiring cannot
  reach, so every auth-resolving dynamic check (including `write_request_shape:update_file`) would
  otherwise fail identically and uninformatively. Static checks still run and pass. The
  read/pagination/schema-projection/write-request shapes for `file_details`/`update_file` are
  proven by structural review against Box's published OpenAPI spec rather than a
  `paritytest/box-data-extract` package, matching the bundle-level precedent this connector already
  established.
- **`file_details`'s fan-out assumes every entry `files` lists is a `file`-type item, not a
  `folder`.** Box's `/folders/{folder_id}/items` endpoint returns BOTH file and (sub)folder
  entries in the same `entries` array with no declarative way to filter the fan-out id-listing
  request by `type` (the `fan_out.ids_from.request` dialect has no filter field). `GET
  /files/{{ fanout.id }}` against a folder-type id 404s, and a fan-out sub-sequence error is
  fail-fast (per-id, not per-record) — so `file_details` only reads cleanly when the configured
  `box_folder_id` folder's immediate children are all files, no subfolders. This is a genuine,
  documented scope narrowing (not silently wrong): a caller pointing `box_folder_id` at a
  folder containing subfolders will see `file_details` fail outright rather than silently skip or
  mis-map the subfolder entries. Candidate `ENGINE_GAP` if a future increment needs
  `fan_out.ids_from.request` to filter by a field value before extracting ids.
- **`file_text` is NOT migrated (documented scope narrowing, not an `ENGINE_GAP`), unchanged from
  pre-expansion.** Legacy's `Read` for `file_text` unconditionally returns an error outside fixture
  mode; there is no live Box endpoint being called for this stream at all in legacy. Box's real
  extracted-text mechanism (`GET /files/{file_id}` with an `x-rep-hints: [extracted_text]` header,
  an async representation-generation job-polling flow) is a materially different, considerably more
  complex shape a plain declarative GET cannot express — see `api_surface.json`'s excluded entry;
  this would be a genuine `ENGINE_GAP`/Tier-2 `StreamHook` candidate if pursued, not a stream to
  silently approximate.
- Only the connector-configured folder's immediate children are reachable; this connector's
  `spec.json` has no config surface for enumerating arbitrary Box folders/files beyond
  `box_folder_id` — see `internal/connectors/defs/box` for the sibling bundle covering
  account/enterprise-level users/groups/collections/webhooks/etc.
- `box_folder_id` must be a Box numeric folder id; not re-implemented as a `spec.json` pattern
  constraint (an invalid `box_folder_id` surfaces as a Box API error instead).
- Documented parity deviation (unchanged from pre-expansion): legacy only sets the
  `box_subject_id` token-request form param when non-empty; the engine's `auth[].extra_params`
  dialect hard-errors on an unresolved config key rather than supporting `stream.Query`'s
  `omit_when_absent` tolerance, so `box_subject_id` relies on `spec.json`'s
  `default`-materialization mechanism with no explicit default value (`""` when unset) instead. See
  `docs/migration/conventions.md`'s parity-deviation ledger.
- `files` is full-refresh only (no cursor field declared, matching legacy).
