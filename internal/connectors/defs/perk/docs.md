# Overview

Perk is a wave2 fan-out declarative-HTTP migration, expanded in Pass B beyond legacy parity. It
reads Perk/TravelPerk trips, invoices, invoice lines, invoice profiles, and per-trip custom fields
through read-only REST endpoints (`GET https://api.travelperk.com/...`). The 2 original streams
(`trips`, `invoices`) are engine-vs-legacy parity-tested against `internal/connectors/perk` (the
hand-written connector this bundle migrates); the legacy package stays registered and unchanged
until wave6's registry flip. The 3 Pass B streams (`invoice_lines`, `invoice_profiles`,
`trip_custom_fields`) have no legacy counterpart — legacy never modeled them — so they carry no
parity constraint and are authored directly from TravelPerk's own published OpenAPI 3.1
documentation (fetched directly during this review from `https://developers.perk.com/reference/
<slug>.md`, one machine-readable fragment per endpoint; the prior docs_url manual-intervention note
is resolved — see Known limits).

## Auth setup

Provide a Perk/TravelPerk API key via the `api_key` secret; it is sent as an `Authorization: ApiKey
<api_key>` header (`auth.mode: api_key_header`, `header: Authorization`, `prefix: "ApiKey "`) and is
never logged, matching legacy's `connsdk.APIKeyHeader("Authorization", key, "ApiKey ")`
(`perk.go:128`). Every request also carries a static `Api-Version: 1` header, matching legacy's
`DefaultHeaders`. `base_url` defaults to `https://api.travelperk.com` and may be overridden for
tests/proxies. This is confirmed against TravelPerk's own OpenAPI security scheme
(`"type": "apiKey", "in": "header", "name": "Authorization"`) and every published code sample
(`curl -H 'Authorization: ApiKey 123456.abcdef...' -H 'Api-Version: 1' ...`).

## Streams notes

`trips` and `invoices` are `GET` list endpoints returning records at a top-level key matching the
stream name (`trips`/`invoices`). Pagination is offset+limit (`pagination.type: offset_limit`,
`limit_param: limit`, `offset_param: offset`, `page_size: 50`), stopping on a short page —
identical to legacy's `connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset",
PageSize: size}`; confirmed against the real OpenAPI spec's `offset`/`limit` query parameters on
both `/trips` and `/invoices`. `trips`' incremental cursor field is `modified`, sent as
`modified_gte`; `invoices`' incremental cursor field is `issuing_date`, sent as
`issuing_date_gte` — both computed from the sync's persisted cursor or, on a fresh sync, the
RFC3339 `start_date` config value, matching legacy's `firstNonEmpty(req.State["cursor"],
req.Config.Config["start_date"])` exactly (`param_format: rfc3339`, i.e. forwarded verbatim, same
as legacy's `base.Set(endpoint.startParam, start)` with no reformatting). Note the real API's own
`modified_gte`/`issuing_date_gte` OpenAPI parameter schemas are `format: date` (bare `YYYY-MM-DD`),
narrower than legacy's verbatim-RFC3339 forwarding; this bundle keeps legacy's exact (looser)
forwarding behavior for these 2 streams to preserve parity — not tightened in this pass.

`invoices`' primary key is `serial_number` (not `id`) — legacy's own `streams()` declares
`PrimaryKey: []string{"serial_number"}` for this stream; this bundle's schema matches that exactly.

Both parity streams declare `"projection": "passthrough"` (post-wave2 review §8 rule 1): legacy's
`Read` emits `emit(connectors.Record(rec))` for both `trips` and `invoices` — a verbatim type-cast
of the raw harvested record, with no `mapRecord`-style field-building — so schema-mode projection
would silently drop any raw field this bundle's schema omits. Each schema remains a documentation
surface only; it does not gate what is emitted.

**Pass B additions** (no legacy counterpart; authored from TravelPerk's own OpenAPI 3.1 docs):

- `invoice_lines` (`GET /invoices/lines`, records at `invoice_lines`): TravelPerk's own
  `InvoiceLineExtended` model — per-line-item detail underlying every invoice, richer than the
  `invoices` stream's own summary. Same offset+limit pagination shape as every other stream, and
  the identical `issuing_date_gte`/incremental filter param the real API documents for this
  endpoint (confirmed from its own OpenAPI parameter list, alongside `invoices`' matching param).
  `"projection": "passthrough"` since this is a new stream authored straight from the raw API
  response shape (no legacy record-shaping function exists to diverge from).
- `invoice_profiles` (`GET /profiles`, records at `profiles`): the account's configured invoice
  profiles (billing legal entities). No incremental filter is documented for this endpoint (its
  OpenAPI spec exposes only `offset`/`limit`/`exclude_personal`), so no `incremental` block is
  declared — full-refresh only, matching the real API's own capability.
- `trip_custom_fields` (`GET /trips/{trip_id}/custom-fields`, single-object response, `records:
  {"path": ".", "single_object": true}`): TravelPerk's own docs state this returns "all custom
  field values associated with a trip" — a per-trip sub-resource with no list endpoint of its own,
  so this stream uses `fan_out` (`ids_from.request` against `trips`' own `GET /trips` endpoint,
  `into.path_var: trip_id`, `stamp_field: trip_id`) to drive one `/trips/{id}/custom-fields`
  sub-sequence per trip, stamping the source trip's id onto every emitted record (overwriting the
  raw response's own numeric `trip_id` field with the fan-out id's string form — documented on the
  schema property itself). `"projection": "passthrough"` since the sub-resource's `custom_fields`
  array shape is bundle-defined per account and not meaningfully constrainable to a fixed schema.

## Write actions & risks

None. Legacy `perk.Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write`
is `false` and this bundle ships no `writes.json`. TravelPerk's own docs DO document 2 real,
dialect-expressible write endpoints on this same host (`POST /webhooks`, `PATCH /webhooks/{id}`),
but both are deliberately excluded — see Known limits.

## Known limits

- **`docs_url` previously pointed at the API host, not a docs page; this is now resolved.**
  TravelPerk's developer docs hub (`https://developers.travelperk.com/docs`, migrated under the
  `developers.perk.com` rebrand) is directly reachable and publishes a machine-readable OpenAPI 3.1
  fragment per endpoint at `https://developers.perk.com/reference/<slug>.md` (confirmed by direct
  fetch during this Pass B review — every endpoint this bundle covers or excludes cites its own
  fragment's exact request/response schema, not a guess). The bundle's `metadata.json.docs_url`
  field itself is left unchanged (still the API host, matching this connector's historical
  convention) since the schema does not require it to be the docs hub specifically, but the real
  authoritative source used for this review is the docs hub above, not legacy code guesswork.
- **The documented TravelPerk/Perk surface spans THREE separate hosts, only one of which this
  bundle targets.** `api.travelperk.com` (this bundle's `base_url`, spec `travelperk-api-v1` v1.5)
  covers trips/invoices/invoice-lines/profiles/webhooks. Cost Centers (`GET/POST/PATCH
  /cost_centers`, `PUT /cost_centers/{id}/users`, `PATCH /cost_centers/bulk_update`) lives on a
  genuinely distinct host, `https://api.perk.com` (spec "Cost Centers API" v1.6). SCIM user
  provisioning (`POST/PUT/PATCH /Users`) lives on yet another distinct host,
  `https://app.perk.com/api/v2/scim` (or the `app.travelperk.com` equivalent), with its own SCIM
  2.0 auth/versioning model entirely. This bundle's `spec.json` declares exactly one `base_url` and
  one `api_key` credential, matching legacy's single-host design; covering either of the other two
  hosts would require a second base_url/credential pair this bundle does not have and legacy never
  needed — out of scope for this pass, recorded per-endpoint in `api_surface.json` as
  `out_of_scope` with this exact cross-host reasoning.
- **Webhook creation/update (`POST /webhooks`, `PATCH /webhooks/{id}`) are real, dialect-expressible
  endpoints on this bundle's own host, but are deliberately not implemented.** Both register/repoint
  an outbound event-delivery URL of the caller's choosing; consistent with this dialect's general
  caution around webhook-URL mutations elsewhere in this migration (persona, bitly), these are
  deferred pending their own dedicated security review rather than wired in this pass.
- **`trip_custom_fields`' fan-out sub-sequence sends unused `limit`/`offset` query params on the
  per-trip request.** This stream declares no stream-level `pagination` override, so it (and its
  `fan_out.ids_from.request` id-listing request) both inherit the bundle's base `offset_limit`
  pagination spec — the dialect's own documented behavior (conventions.md's fan_out section: the
  id-listing request "reuses the surrounding stream's" effective pagination spec). TravelPerk's
  `/trips/{trip_id}/custom-fields` endpoint is a single-object response with no documented
  pagination parameters of its own, so the extra `limit=50&offset=0` sent alongside every per-trip
  request is inert (ordinary REST tolerance for unrecognized query params) but not itself
  meaningful; `connsdk.OffsetPaginator` still terminates correctly after one page since a
  single-object response always yields exactly 1 record (`1 < 50`, the short-page stop condition).
  No emitted-record behavior is affected; this is a wire-level curiosity, not a data-loss risk.
- Full TravelPerk/Perk API surface on `api.travelperk.com` is now covered: every GET list endpoint
  is a stream (8 of 10 endpoints on this host); the remaining 2 (`GET /trips/{id}`, `GET
  /invoices/{serial_number}`) are excluded `duplicate_of` an already-covered list stream's per-item
  shape, and the PDF download is `binary_payload` — see `api_surface.json`.
- `fixtures/streams/**` and `fixtures/check.json` use synthetic values only; real Perk/TravelPerk
  responses were not available (no live credential), so fixture shapes are derived from
  TravelPerk's own published OpenAPI response examples (confirmed real wire shape, not guessed) for
  every new stream, and from legacy's own `readFixture` record shape / `perk_test.go`'s live-server
  fixture for the 2 parity streams.
