# Overview

Chift is an L-size fresh migration, expanded in Pass B to the full practical core-platform API
surface. This bundle reads Chift consumers, connections, syncs, integrations, datastores, and
webhook event definitions, and writes consumers (create/update/delete), through the Chift REST API.
It originally migrated `internal/connectors/chift` (the legacy hand-written connector, read-only,
which stays registered and unchanged until wave6's registry flip) at capability parity for the
first 3 streams; Pass B's remaining streams and every write action are new capability beyond legacy
parity. Chift's per-connector unified sub-APIs (accounting, banking, invoicing, ecommerce,
point-of-sale, pms, payment) are reviewed and deliberately deferred — see Known limits.

## Auth setup

Chift does not use a standard form-encoded OAuth2 client-credentials grant: it exchanges
credentials via a POST to `<base_url>/token` with a **JSON body** of
`{"clientId": ..., "clientSecret": ..., "accountId": ...}`, returning
`{"access_token": ..., "expires_in": ...}`, which is then carried as a Bearer token
(`Authorization: Bearer <access_token>`) on every data request. The engine's declarative
`oauth2_client_credentials` auth mode always builds a form-encoded token request with the standard
`grant_type`/`client_id`/`client_secret`/`scope` field names (`connsdk.OAuth2ClientCredentials`) —
it cannot express Chift's JSON body or its non-standard `clientId`/`clientSecret`/`accountId` field
names. This is a legitimate Tier-2 token-exchange-auth trigger (`docs/migration/conventions.md`
§1's Tier-2 table): `internal/connectors/hooks/chift/hooks.go` implements `engine.AuthHook`,
porting legacy `chift/auth.go`'s `sessionTokenAuth` verbatim (JSON POST to `/token`, in-memory
token cache with a 60-second-before-expiry refresh margin, a 3600-second default TTL when
`expires_in` is absent/non-positive). `streams.json`'s `base.auth` declares a single
`{"mode": "custom", "hook": "chift"}` candidate. Provide `client_id`, `client_secret`, and
`account_id` as secrets (all three `x-secret: true` in `spec.json`); none are ever logged, and the
minted access token itself is never persisted outside the hook's in-process cache.

## Streams notes

All 3 streams (`consumers`, `connections`, `syncs`) share the same shape: `GET` against the Chift
list endpoint, which returns a top-level JSON array (`records.path: ""`, matching
`connsdk.RecordsAt`'s empty-path-means-document-root convention — identical to legacy's
`recordsPath: ""`). Pagination is offset/limit (`pagination.type: offset_limit`, `limit_param:
size`, `offset_param: offset`, `page_size: 100` — matches legacy's `chiftDefaultPageSize`),
stopping on a short page (fewer than `page_size` records), exactly matching legacy's
`harvest`/`len(records) < pageSize` stop condition. None of the 3 streams are incremental in
legacy (no cursor-field handling anywhere in `chift/streams.go` or `chift.go`), so no
`streams.json` entry declares an `incremental` block, and no schema declares `x-cursor-field`.
Primary keys: `consumers` uses `consumerid`, `connections` uses `connectionid`, `syncs` uses
`syncid` — matching legacy's declared `PrimaryKey`.

`spec.json` intentionally does NOT declare a `max_pages` runtime-configurable property (unlike
legacy, which accepts a `0`/`all`/`unlimited`/`N` override): the `offset_limit` paginator's
`MaxPages` is a static JSON integer literal (`PaginationSpec.MaxPages`) read from `streams.json`
alone, never from a templated `config.*` value — there is no mechanism in this dialect to wire a
spec property into that field (F6, `conventions.md`: a declared-but-unwireable spec property is
worse than an absent one). See Known limits.

**Pass B new streams**: `integrations` (`GET /integrations`, primary key `integrationid`, no
pagination — Chift's own docs show no page/size params on this endpoint, so `pagination.type:
none` overrides the base offset_limit spec), `datastores` (`GET /datastores`, primary key `id`, also
`pagination.type: none`), and `webhook_definitions` (`GET /webhooks`, the catalog of POSSIBLE
webhook events an account can subscribe to — NOT the configured webhook-instance list, which
`api_surface.json` documents as excluded; composite primary key `["event", "api"]` since the
response has no per-row id field at all, only an event name and an optional owning-api category).
None of these 3 streams have legacy precedent.

## Write actions & risks

Pass B adds write capability (new beyond legacy, which was entirely read-only):
`create_consumer` (`POST /consumers`, requires only `name`), `update_consumer`
(`PATCH /consumers/{consumerid}`, every field optional), and `delete_consumer`
(`DELETE /consumers/{consumerid}`, `missing_ok_status: [404]` — idempotent delete). All three map
directly onto Chift's real, confirmed request/response shape (`docs.chift.eu/api-reference`); the
consumer body is flat JSON with no wrapper envelope. All three carry `"risk": "external mutation;
approval required"` (`delete_consumer` additionally flagged irreversible); `metadata.json`'s
`capabilities.write` is now `true`.

## Known limits

- `max_pages` runtime override is not exposed (see Streams notes above) — every read is unbounded
  (walks every page to exhaustion via the short-page stop rule), matching legacy's own default
  (`max_pages` unset/`0`/`all`/`unlimited`) unconditionally. A caller who previously bounded a sync
  to N pages now always reads to exhaustion instead; this never changes any single emitted
  record's DATA, only how many requests a sync issues — parity-deviation ledger candidate,
  ACCEPTABLE under the meta-rule (no accepted-input record-data change).
- **`metadata.json` carries a bundle-level `skip_dynamic` marker** (sole auth candidate is
  `mode:custom`, and conformance's synthetic replay config can never populate a real `base_url` the
  hook's token-exchange POST could reach — matching gmail's identical documented shape). This Skips
  every auth-dependent dynamic check bundle-wide, including `write_request_shape` for the 3 new Pass
  B write actions — their request-shape correctness is proven by `connectorgen validate`'s
  `write_schemas_valid`/`surface_complete` static checks and this docs.md's worked request/response
  examples, not a live fixture-replay assertion. (A prior draft of this file claimed no marker was
  needed; that was inaccurate given `metadata.json`'s actual declaration and has been corrected
  here.)
- **Chift's per-connector unified sub-APIs (accounting, banking, invoicing, ecommerce,
  point-of-sale, pms, payment) are reviewed and deliberately out of scope for this pass** — see
  `api_surface.json`'s top-level `scope` note and the `/consumers/{consumer_id}/<api>/**` entries.
  Every one of these sub-APIs' list endpoints is scoped `GET /consumers/{consumer_id}/<api>/<resource>`
  (confirmed against Chift's live API reference, e.g. `GET /consumers/{consumer_id}/pos/orders`),
  requiring a fan_out over every consumer id. Unlike chargify's per-product-family `components` gap,
  this IS mechanically expressible with the existing `fan_out` dialect (Chift's consumer object is
  flat, so `fan_out.ids_from.request.id_field: "consumerid"` resolves correctly, with no nested-envelope
  obstacle) — the reason these are deferred is combinatorial scope, not an engine limitation: 7
  sub-APIs x 30-90 endpoints each x every connected consumer is not a single pass's worth of
  reviewable, testable surface. A follow-up pass should size ONE sub-API (accounting is the largest
  and most billing-relevant) as its own dedicated scope rather than attempt all 7 at once.
- Per-consumer detail (`GET /consumers/{id}`), per-sync/connection execution history, and
  webhook-instance (subscribed callback) management are excluded as duplicate-of-stream-data or
  operational/observability data rather than durable business objects — see `api_surface.json`'s
  per-endpoint reasons.
