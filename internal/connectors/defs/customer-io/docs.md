# Overview

Customer.io is a Tier-1 declarative-HTTP source bundle (`internal/connectors/defs/customer-io/`):
Bearer auth, plain JSON list endpoints under the Customer.io App API, and cursor pagination
where each page's response carries a `next` token echoed back as the `start` query parameter (an
empty/null `next` ends the loop). This port is a pure `streams.json`/`spec.json`/`schemas/*.json`
+ `writes.json` bundle with zero Go — the legacy package (`internal/connectors/customer-io/`) is a
thin connsdk composition with no auth/stream hooks, so it maps directly onto the engine's `cursor` +
`token_path` pagination dialect (the identical shape airtable's `records` stream uses for its
`offset`/`offset` pair).

**Pass B full-surface expansion** (2026-07-04): reviewed the full 159-operation Journeys App API
OpenAPI spec (`https://docs.customer.io/files/journeys-app.json`, the real machine-readable
reference behind `docs.customer.io/api/app/`'s rendered docs page). Added 12 read streams
(`activities`, `messages`, `exports`, `transactional`, `object_types`, `reporting_webhooks`,
`sender_identities`, `snippets`, `subscription_channels`, `subscription_topics`, `workspaces`,
`collections` — beyond the original 4 legacy-parity streams) and 10 write actions; `capabilities.write`
is now `true`. See `api_surface.json` for the complete endpoint-by-endpoint disposition (every one
of the 159 real operations is `covered_by` or `excluded` with a real category+reason) and Known
limits below for what remains out of reach and why.

## Auth setup

Provide `app_api_key` (an App API Key from Customer.io's UI) as a secret; it is sent as
`Authorization: Bearer <app_api_key>` via `streams.json`'s `base.auth`, never logged.

`base_url` is **required** in this bundle, unlike legacy where it was derived from an optional
`region` config value (`US` -> `https://api.customer.io/v1`, `EU` -> `https://api-eu.customer.io/v1`,
with an explicit `base_url` override always taking priority). See Known limits below for why the
derivation itself could not be ported.

## Streams notes

All 4 streams (`campaigns`, `newsletters`, `segments`, `broadcasts`) share the identical shape:
`GET` against the resource path, `records.path` set to the resource's own top-level JSON key (e.g.
`{"campaigns": [...]}`), a `limit` query param sourced from `config.page_size` (default `100`,
matching legacy's `customerIODefaultPage`/`customerIOMaxPage` clamp — the engine dialect does not
enforce a 1-100 range on a config value the way legacy's `customerIOPageSize` did; see Known limits),
and `incremental.cursor_field: updated` with `client_filtered: true` (the Customer.io App API has no
server-side `updated`-since filter parameter, matching legacy's `harvest`, which fetches every page
unconditionally and relies on the caller's own downstream dedup/append semantics — `client_filtered`
is the sanctioned dialect for exactly this "API can't filter server-side" shape, per
`docs/migration/conventions.md` §3).

Every object exposes a numeric `id` and Unix-seconds `created`/`updated` timestamps (segments omit
`created`, matching legacy's `segmentFields`), so every schema declares `x-primary-key: ["id"]` and
`x-cursor-field: "updated"`.

The 12 Pass B streams fall into two pagination shapes:

- **`activities`** shares the base's `cursor`/`start`/`next` pagination exactly like the original 4
  streams (`incremental.cursor_field: timestamp`, `client_filtered: true` — the endpoint's own
  `start`/`limit` params page through results but there is no server-side "since" filter).
- **The other 11** (`messages`, `exports`, `transactional`, `object_types`, `reporting_webhooks`,
  `sender_identities`, `snippets`, `subscription_channels`, `subscription_topics`, `workspaces`,
  `collections`) are genuinely unpaginated per the OpenAPI spec — none of their GET operations
  documents a `page`/`limit`/`start` parameter at all — so each declares a stream-level
  `"pagination": {"type": "none"}` override against the base's `cursor` spec, and each ships a
  single-page fixture (no 2-page requirement per conventions.md §4, which only mandates a 2-page
  fixture when pagination is actually declared for that stream).

`snippets`' primary key is `name` (not `id` — the resource has no numeric identifier; the API's own
docs state the name must be unique), matching the write actions' identical `name`-keyed shape.

## Write actions & risks

Ten write actions, all flat-JSON-body mutations against documented endpoints (`capabilities.write`
is now `true`):

- `create_snippet` / `update_snippet` (`POST`/`PUT /snippets`) — create or overwrite a reusable
  content snippet; `update_snippet` is a bare `PUT /snippets` (the API upserts by the `name` field
  in the body, no path-parameterized identifier).
- `delete_snippet` (`DELETE /snippets/{snippet_name}`) — permanently removes a snippet; irreversible,
  breaks any message/newsletter still referencing it.
- `create_reporting_webhook` / `update_reporting_webhook` / `delete_reporting_webhook` — manage a
  workspace reporting webhook subscription (event delivery to an external URL).
- `create_manual_segment` / `delete_manual_segment` — `create_manual_segment`'s record shape is
  `{"segment": {"name": ..., "description": ...}}`, matching the API's own nested-object body
  exactly (an ordinary nested JSON object, not the "wrap in a named bulk ARRAY" shape that blocks
  Customerly's `create_user`/`create_lead` — a nested object is just normal JSON body construction,
  no special dialect primitive needed).
- `send_email` (`POST /send/email`) — sends a live transactional email; the OpenAPI body is an
  `allOf` of three sub-schemas that merge into one flat object (no `oneOf` discriminator), so it
  maps directly onto a single flat `record_schema`.
- `trigger_broadcast` (`POST /campaigns/{broadcast_id}/triggers`) — triggers an API-triggered
  broadcast to its "default audience" (the simplest of 3 documented audience-targeting variants,
  a `oneOf`; the other two — custom recipient lists / per-recipient data overrides — are excluded,
  see Known limits).

Every action's `risk` field flags it as an external mutation requiring approval; the two deletes and
`send_email`/`trigger_broadcast` explicitly call out irreversibility.

## Known limits

- **`base_url` cannot be derived from a `region` config value.** Legacy's `customerIOBaseURL`
  switches on an optional `region` config key (`US`/`EU`/unset-defaults-to-US) to choose between two
  hardcoded base URLs, only falling back to a directly-configured `base_url` override. The engine's
  `spec.json` `"default"` mechanism materializes a single **fixed literal** default value for an
  absent key (see `docs/migration/conventions.md` §3, "`spec.json` `default` values ARE now
  materialized") — it has no mechanism to derive one config value's default from ANOTHER config
  value's value (the same gap `sentry`'s hostname-derived URL and `chargebee`'s site-derived URL
  hit). Per the sanctioned resolution for this exact shape (conventions §3), `base_url` is declared
  **required** here instead of re-deriving the branch in Go (which would need a 3rd Tier-2 hook
  interface or a Tier-3 escalation neither justified by this connector's otherwise-uniform HTTP
  shape). This is a documented, accepted narrowing of the config surface, never a change to any
  emitted record's data: an operator who previously left `region` unset (or set it to `US`/`EU`) now
  supplies the resolved `https://api.customer.io/v1` or `https://api-eu.customer.io/v1` value
  directly as `base_url`.
- **`page_size`'s 1-100 range is not enforced.** Legacy's `customerIOPageSize` rejects a `page_size`
  outside `[1, 100]` with a config-validation error before the first request. The engine dialect has
  no range-validation primitive for a plain templated query parameter — an out-of-range `page_size`
  here is sent to the Customer.io API as-is and would surface as a live API error rather than a local
  config-validation error. This never changes emitted record DATA for any `page_size` legacy itself
  would have accepted (1-100); it only moves where an out-of-range value is rejected, from local
  config validation to the live API's own response.
- **`max_pages` is not configurable.** Legacy accepts a `max_pages` config value (default unlimited)
  as a client-side page-count cap. The engine's `PaginationSpec.MaxPages` is a static bundle-declared
  integer, not a per-request templated value (mirroring stripe's identical, already-accepted
  `max_pages`/`page_size` dead-config resolution recorded in the parity-deviation ledger, §5 item 3)
  — there is no config-driven override mechanism for it at all. This bundle declares no `max_pages`
  spec property (a declared-but-unwireable key is worse than an absent one, per `conventions.md` F6)
  and leaves pagination unbounded (matching legacy's default `max_pages=0`/unlimited behavior, and the
  paginator's own short-page/empty-token stop signal still terminates every real sync).
- **`oneOf`-discriminated write bodies are not implemented.** `POST /v1/newsletters` (6-way channel
  discriminator) and the "Custom recipients"/per-recipient-data-override variants of
  `POST /v1/campaigns/{broadcast_id}/triggers` all require the caller to pick one of several
  mutually-exclusive body shapes at request time. A `writes.json` action declares exactly one flat
  `record_schema` per action name — there is no discriminated-union primitive, so each of these
  would need either N separate action names (one per variant, awkward and not how the API itself
  models the choice) or a Tier-2 `WriteHook` (not justified here: the shapes differ only in which
  JSON fields are present, no auth/compound-request/binary-payload trigger applies). Only
  `trigger_broadcast`'s "Default audience" variant (no discriminator needed — it's the shape with NO
  extra required fields beyond `broadcast_id`) is implemented; see `api_surface.json`'s `excluded`
  entries for the rest.
- **Per-object detail/metrics/sub-resource endpoints keyed by a single already-covered list's id are
  excluded as `duplicate_of`**, not implemented via `fan_out`: `fan_out.ids_from` needs either a
  `config_key` (a caller-supplied comma-separated id list) or a `request` (a preliminary GET that
  itself returns the id list) — for every excluded sub-resource here, the "parent" ids are already
  the covered list stream's own `id` field, and there is no separate declarative mechanism to feed a
  stream's OWN already-read primary keys back into a second fan_out stream's `ids_from` without a
  config-supplied id list (which would require the operator to enumerate ids up front, defeating the
  point of a catalog sync). This is the same shape as every `*_id`-keyed detail GET this bundle
  excludes (broadcast/campaign/collection/customer/newsletter/segment/sender_identity/transactional
  sub-resources) — see `api_surface.json` for the full list.
