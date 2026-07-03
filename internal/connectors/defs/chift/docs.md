# Overview

Chift is an L-size fresh migration. This bundle reads Chift consumers, connections, and syncs
through the Chift REST API, migrating `internal/connectors/chift` (the legacy hand-written
connector, which stays registered and unchanged until wave6's registry flip) at capability parity.
The API is read-only; this bundle exposes no write actions.

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

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub
(`connectors.ErrUnsupportedOperation`).

## Known limits

- `max_pages` runtime override is not exposed (see Streams notes above) — every read is unbounded
  (walks every page to exhaustion via the short-page stop rule), matching legacy's own default
  (`max_pages` unset/`0`/`all`/`unlimited`) unconditionally. A caller who previously bounded a sync
  to N pages now always reads to exhaustion instead; this never changes any single emitted
  record's DATA, only how many requests a sync issues — parity-deviation ledger candidate,
  ACCEPTABLE under the meta-rule (no accepted-input record-data change).
- The session-token exchange (`POST /token`) is Tier-2 hook-covered, not declarative
  (`AuthSpec.mode: custom`); `conformance`'s dynamic (fixture-replay) checks exercise the hook for
  real against the replay harness (the hook's token-exchange POST runs against the same
  `httptest.Server` the declarative reads do) — no `skip_dynamic` marker is needed or declared.
- Full Chift API surface (per-connector accounting/commerce/pos sub-APIs, per-consumer detail,
  consumer creation) is out of scope until Pass B; see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries. Legacy itself never implemented these.
