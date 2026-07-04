# Overview

Fastly is a CDN/edge platform. This bundle reads Fastly services, the authenticated current user,
the current customer (account), points-of-presence datacenters, per-service configuration details,
and per-customer account users through the Fastly REST API (`https://api.fastly.com`). It migrates
`internal/connectors/fastly` (the hand-written connector), which stays registered and unchanged
until wave6's registry flip. Read-only: Fastly has no obvious safe reverse-ETL write surface,
matching legacy's `Capabilities.Write: false`. **Pass B full-surface expansion** (this revision)
adds `service_details` and `users` — the two additional GET resources practical to model as
declarative fan-out reads over legacy's 4-stream base — and researches the FULL documented Fastly
REST API (~20 resource collections: services, TLS, purging, ACLs/dictionaries, load balancing,
Next-Gen WAF, observability, 20+ third-party logging integrations, billing, IAM) to close out every
remaining endpoint with a specific, closed-vocabulary exclusion reason in `api_surface.json` — see
Known limits and `api_surface.json` for the full accounting.

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
  (`page_number` pagination, 1-based). `pagination.page_size` is `100`, matching legacy's real
  production default/hard-max (`fastlyDefaultPageSize`/`fastlyMaxPageSize`) — this is the actual
  value a live deployment's paginator sends, not a fixture convenience (see Known limits for why
  legacy's runtime override isn't wired). Primary key `id`. Legacy's catalog decoratively declares
  `CursorFields: []string{"updated_at"}` but `Read()` never actually filters or advances by it (no
  incremental logic anywhere in `fastly.go`/`streams.go`) — this bundle matches that exact
  behavior: `updated_at` is declared as `x-cursor-field` on the schema for manifest-surface parity
  only, but no `incremental` block is declared on the stream, so the engine performs a full read
  every time, identical to legacy.
- **current_user** (`GET /current_user`) and **current_customer** (`GET /current_customer`) are
  singleton endpoints returning one JSON object each; `records.path: ""` with `single_object: true`
  wraps the root object into a single record, matching legacy's `readSingle`/`RecordsAt(resp.Body,
  "")` behavior. No pagination.
- **datacenters** (`GET /datacenters`) is a paginated top-level array (`page`/`per_page`,
  identical shape to `services`). Primary key `code` (matching legacy — datacenters have no `id`
  field, they are keyed by their airport-style `code`). This stream declares a stream-level
  `pagination` override (`page_size: 2`) so its required 2-page conformance fixture
  (`fixtures/streams/datacenters/{page_1,page_2}.json`, §4 of `docs/migration/conventions.md`) can
  stay small and readable; since stream-level `pagination` replaces the base spec wholesale, this
  is an intentional, ledgered per-stream deviation from legacy's uniform 100-record page size —
  `datacenters` reads in smaller, more numerous pages than legacy would, `services` is unaffected
  and uses legacy's true 100-record page size end-to-end (matching its fixture's `per_page=100`
  request/response).
- **service_details** (`GET /service/{service_id}/details`, Pass B addition) — a per-service fan-out
  read: the id list comes from a preliminary, fully-paginated `GET /service` request
  (`fan_out.ids_from.request`, `records_path: "."`, `id_field: "id"` — the same top-level array
  `services` itself reads), and each resolved service id is substituted into the details path via
  `into.path_var: "service_id"`. `stamp_field: "service_id"` stamps the fanned-out id onto every
  emitted record. The response is a single nested JSON object per service (active/draft version
  summary, version history) — modeled with `records.path: "."` + `single_object: true`, matching
  the `current_user`/`current_customer` singleton shape. `version`/`versions`/`activated_version`/
  `environments` are typed `["object"|"array", "null"]` and passed through verbatim (no
  `computed_fields` flattening — Fastly's own nested shape varies by service type (`vcl` vs
  `wasm`), so no single flat schema could losslessly represent every service's version metadata).
  No incremental cursor: legacy has no notion of this stream at all, so there is no prior cursor
  field to preserve or omit.
- **users** (`GET /customer/{customer_id}/users`, Pass B addition) — a per-customer fan-out read:
  the id comes from a preliminary `GET /current_customer` request (`fan_out.ids_from.request`,
  `records_path: "."`, `id_field: "id"` — the SAME singleton-object-as-one-record extraction
  `current_customer` itself uses, since `connsdk.RecordsAt` wraps a bare JSON object into a
  one-element record list), substituted into the path via `into.path_var: "customer_id"`. In
  practice this always resolves to exactly the caller's own customer id (Fastly API tokens are
  scoped to one customer), so the fan-out here is a single-id sequence, not a true multi-parent
  fan-out — declared this way instead of a plain `config.customer_id` spec property because the
  bundle should not require the operator to already know their own Fastly customer id when it is
  directly derivable from `/current_customer`. `stamp_field: "customer_id"` stamps the resolved id
  onto every emitted user record.

All streams map their raw JSON fields 1:1 onto the schema's declared properties (field names
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
  bundle, not templated). `services`' `pagination.page_size` is fixed at legacy's own default/max
  (`100`), reproducing legacy's default-configuration behavior exactly; `datacenters` keeps a
  smaller `page_size: 2` as a ledgered per-stream deviation purely to keep its 2-page conformance
  fixture small (see Streams notes) — an operator who had overridden legacy's `page_size` away from
  its default cannot reproduce that override here, but every request this bundle sends by default
  matches legacy's own default cadence. `max_pages` is left unset (unbounded) on every stream,
  matching legacy's own default (`0`/"unlimited"). Declaring `page_size`/`max_pages` as `spec.json`
  properties that no template anywhere in this bundle consumes would itself be dead config (F6) —
  they are intentionally not declared.
- **`services`' `updated_at` cursor field is decorative, not functional**, matching legacy exactly
  (see Streams notes above) — this is not a scope narrowing versus legacy, since legacy itself never
  implemented incremental filtering for this stream.
- **Full documented Fastly surface is intentionally NOT all implemented as streams/writes** — see
  `api_surface.json`'s `scope` field and per-endpoint `excluded.category`/`reason` for the complete
  accounting. Every mutation endpoint (service/version create-update-delete-activate-clone-lock,
  domain/ACL/dictionary/TLS-subscription/token provisioning, cache purging) is `destructive_admin`
  or `requires_elevated_scope`: none is a safe reverse-ETL write target, matching legacy's own
  read-only design (`capabilities.write` stays `false`). Read-only resources outside this
  connector's account/service-inventory scope (Next-Gen WAF, API Security, DDoS Protection,
  Client-Side Protection, Domain Management, Observability alerting, real-time-logging provider
  integrations, load-balancing directors, billing/invoices) are `out_of_scope`/
  `requires_elevated_scope` — each is a genuinely separate Fastly product surface with its own data
  model, not an oversight. Endpoints nesting TWO levels deep under a parent id (e.g.
  `/service/{service_id}/version/{version_id}/...`) cannot be expressed by the engine's
  single-level `fan_out` dialect (`ids_from` resolves exactly one id list per stream) — these are
  `out_of_scope` pending a nested/2-level fan_out engine capability, not silently approximated.
  `/service/{service_id}`, `/service/search`, `/user/{user_id}`, and `/customer/{customer_id}` are
  `duplicate_of`: each returns a strict subset of data already covered by an implemented stream
  (`service_details`, `services`, `users`, `current_customer` respectively).
