# Overview

Plaid is a Tier-2 (StreamHook) migration of `internal/connectors/plaid` (quarantine.json:
`AUTH_COMPLEX`, "Plaid's entire API is POST-only with ALL credentials (client_id/secret) and ALL
pagination/filter state (count/offset/country_codes) carried in the JSON request body, never in
headers or query params"). It reads Plaid institution and category metadata — the only two catalog
endpoints legacy covers that need no end-user `access_token`. This bundle is parity-tested against
`internal/connectors/plaid` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: legacy's `Write` always returns
`ErrUnsupportedOperation`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide two secrets: `client_id` and `secret` (both `x-secret` in `spec.json`, never logged).
Unlike every header/query-based auth mode the engine's declarative dialect supports, Plaid expects
`client_id`/`secret` as two ordinary fields of the SAME JSON body that also carries the
request-specific fields (`count`/`offset`/`country_codes` for `institutions/get`, nothing extra for
`categories/get`) — there is no request this API accepts with credentials anywhere else. The
engine's read path (`engine/read.go`'s `readDeclarative`) always calls `Requester.Do(ctx, method,
path, query, nil)` — the body argument is hard-coded `nil` on every declarative read, and
`StreamSpec.Body` (`bundle.go`), while declared on the struct, is never actually read back out
anywhere in `read.go`. This is a genuine `ENGINE_GAP`: there is no declarative way to send a body at
all on a read, let alone one carrying pagination state — the same gap monday's GraphQL POST body
hit (`docs/migration/conventions.md`'s monday precedent), solved the same way here.

`hooks/plaid/hooks.go` implements `StreamHook` (both streams) and `CheckHook`, porting legacy
`plaid.go`'s `authBody`/`readPage`/`Check` verbatim: it builds the full JSON body itself (starting
from `{"client_id": ..., "secret": ...}` and adding stream-specific fields), POSTs via
`rt.Requester.Do` (reusing the engine's already-resolved base URL/retry/rate-limit plumbing — the
declarative `base.auth` is `mode: none` since there is no header/query credential to declare), and
decodes the response. `client_id`/`secret` values flow only into the outgoing request body; they
are never logged and never appear in an error string.

## Streams notes

Two streams: `institutions` (paginated) and `categories` (single call, no pagination — legacy's
`categories/get` takes no `count`/`offset`/`country_codes` fields at all). `institutions` pages via
Plaid's own in-body `count`/`offset` fields (not query params): each request's body sets
`count: config.page_size` (default 100), `offset` (0, then advancing by `page_size`), and
`country_codes` (from `config.country_codes`, default `["US"]`); a page returning fewer than
`count` records stops pagination, mirroring legacy's `readPage` short-page stop exactly.
`max_pages` (default 3, matching legacy's `defaultMaxPages`) is a hard page-count cap, also ported
verbatim. Records are extracted from each response body's `institutions`/`categories` array;
`country_codes`/`hierarchy` (both raw JSON arrays) are joined into a legacy-matching
comma-separated, sorted string (`joinAny`, ported verbatim from `plaid.go`).

Both streams' declarative `streams.json` entries exist so `metadata.json`/`spec.json`/schemas stay
uniform with every other connector and so `connectorgen validate` can check schema/primary-key
shape — but neither stream is ever actually read declaratively; `StreamHook.ReadStream` returns
`handled=true` unconditionally for both, matching monday's documented shadow-declarative-path
precedent.

## Write actions & risks

None — Plaid is read-only for this bundle's scope. `capabilities.write: false`, no `writes.json`
file, matching legacy's `ErrUnsupportedOperation` (`plaid.go`'s `Write`).

## Known limits

- **The full Plaid API surface (Item, Auth, Transactions, Identity, Balance, Liabilities,
  Investments) is out of scope**, matching legacy exactly: every one of those endpoints requires a
  Link-issued end-user `access_token`, which legacy's own package doc explicitly excludes ("covers
  Plaid catalog endpoints that do not require an end-user access_token"). Modeling the Link
  token-issuance/OAuth-like exchange and per-end-user access-token storage is a materially
  different, larger scope than this migration's mandate (Pass B, if ever prioritized).
- **`ENGINE_GAP` — no declarative body-carrying read path.** `StreamSpec.Body` is declared on
  `engine/bundle.go`'s struct but is dead: `engine/read.go`'s declarative read loop always calls
  `Requester.Do` with a `nil` body, so no JSON-body request (credentials or pagination state) can
  be expressed in `streams.json` alone today. This is why a `StreamHook` is required rather than a
  declarative fixture-replay-provable bundle; see `metadata.json`'s `conformance.skip_dynamic`
  marker. Wiring `StreamSpec.Body` into `read.go` (interpolated the same way `Query`/`Path` are,
  with the same `omit_when_absent`/typed-value support the `computed_fields`/`stream.Query` dialect
  already has) would close this gap generally — filed here as the concrete evidence for a future
  engine mini-wave increment (this migration's scope is limited to the two named connectors, not
  engine changes).
- **`TestConformance/plaid`'s dynamic (fixture-replay) checks are `skip_dynamic`'d** because the
  fixture-replay harness has no way to represent a body-carried credential/pagination-state request
  at all (not merely an auth-resolution limitation like gmail/nexus-datasets's custom-mode gap) —
  see `metadata.json`'s `conformance.reason`. The hook's own unit tests
  (`internal/connectors/hooks/plaid/hooks_test.go`, a real `httptest.Server` asserting the exact
  body shape/pagination/short-page-stop behavior) and the pre-existing legacy test suite
  (`internal/connectors/plaid/plaid_test.go`, unchanged, still passing against the read-only legacy
  package) remain the authoritative correctness bar for this connector.
