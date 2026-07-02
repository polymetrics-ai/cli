# Overview

Fastly is a CDN/edge platform. This bundle reads Fastly services, the authenticated current user,
the current customer (account), and points-of-presence datacenters through the Fastly REST API
(`https://api.fastly.com`). It migrates `internal/connectors/fastly` (the hand-written connector),
which stays registered and unchanged until wave6's registry flip. Read-only: Fastly has no obvious
safe reverse-ETL write surface, matching legacy's `Capabilities.Write: false`.

## Auth setup

Fastly authenticates every request with a `Fastly-Key: <token>` header carrying an API token
(legacy's `fastlyAuthHeader`/`connsdk.APIKeyHeader`). This bundle wires the identical shape via
`streams.json` `base.auth`: `{"mode": "api_key_header", "header": "Fastly-Key", "value": "{{
secrets.fastly_api_token }}"}`. The `fastly_api_token` secret is required for every non-fixture
read/check; `base_url` defaults to `https://api.fastly.com` (materialized via `spec.json`'s
`default`, matching legacy's `fastlyDefaultBaseURL` fallback) and may be overridden for a test
server.

## Streams notes

- **services** (`GET /service`) — a top-level JSON array, paginated with `page`/`per_page`
  (`page_number` pagination, 1-based). `page_size: 2` is a small, fixed client-side short-page-stop
  threshold chosen for cheap 2-page fixtures (`docs/migration/conventions.md` §4's 2-page fixture
  requirement); legacy's own default/hard-max was 100 (`fastlyDefaultPageSize`/
  `fastlyMaxPageSize`), which is out of reach here since the Tier-1 `page_number` pagination
  block's `page_size` is a static bundle value, not config-driven (see Known limits). Primary key
  `id`. Legacy's catalog decoratively declares `CursorFields: []string{"updated_at"}` but `Read()`
  never actually filters or advances by it (no incremental logic anywhere in `fastly.go`/
  `streams.go`) — this bundle matches that exact behavior: `updated_at` is declared as
  `x-cursor-field` on the schema for manifest-surface parity only, but no `incremental` block is
  declared on the stream, so the engine performs a full read every time, identical to legacy.
- **current_user** (`GET /current_user`) and **current_customer** (`GET /current_customer`) are
  singleton endpoints returning one JSON object each; `records.path: ""` with `single_object: true`
  wraps the root object into a single record, matching legacy's `readSingle`/`RecordsAt(resp.Body,
  "")` behavior. No pagination.
- **datacenters** (`GET /datacenters`) is a paginated top-level array (`page`/`per_page`,
  identical shape to `services`). Primary key `code` (matching legacy — datacenters have no `id`
  field, they are keyed by their airport-style `code`).

All four streams map their raw JSON fields 1:1 onto the schema's declared properties (field names
already match — no `computed_fields` renames are needed anywhere in this bundle). `version`
(services) is declared as the real wire type (`integer`) since plain schema projection copies the
raw JSON value's native type without any stringification (typed-extraction convention,
`docs/migration/conventions.md` §3).

## Write actions & risks

None. Fastly is read-only for reverse ETL purposes — legacy's own comment: "the Fastly API has no
obvious safe reverse-ETL write surface" — `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **No runtime page-size/max-pages config override.** Legacy accepted `page_size` and `max_pages`
  config keys to override pagination at read time. The Tier-1 declarative dialect's `page_number`
  pagination fields (`page_size`, `max_pages`) are static values baked into `streams.json`'s
  `pagination` block — there is no mechanism to route a `spec.json` config value into them at read
  time (`docs/migration/conventions.md`'s `PaginationSpec` fields are read directly off the loaded
  bundle, not templated). `page_size: 2` (smaller than legacy's own default/max of 100) is used
  purely as a cheap, fixed short-page-stop threshold to keep the conformance fixture small
  (`docs/migration/conventions.md` §4); `max_pages` is left unset (unbounded), matching legacy's own
  default (`0`/"unlimited"). Declaring `page_size`/`max_pages` as `spec.json` properties that no
  template anywhere in this bundle consumes would itself be dead config (F6) — they are
  intentionally not declared.
- **`services`' `updated_at` cursor field is decorative, not functional**, matching legacy exactly
  (see Streams notes above) — this is not a scope narrowing versus legacy, since legacy itself never
  implemented incremental filtering for this stream.
