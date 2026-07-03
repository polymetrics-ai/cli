# Overview

Whisky Hunter is a wave2 fan-out declarative-HTTP migration. It reads public Whisky Hunter auction
and distillery data (`GET https://whiskyhunter.net/api/...`). This bundle is engine-vs-legacy
parity-tested against `internal/connectors/whisky-hunter` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

No credentials are required: Whisky Hunter's API is fully public. `base_url` defaults to
`https://whiskyhunter.net` and may be overridden for tests/proxies, matching legacy's
`defaultBaseURL`/`baseURL` validation (scheme+host required, trailing slash trimmed).

## Streams notes

Both streams (`auctions`, `distilleries`) are single-page `GET` list endpoints whose response body
is a bare top-level JSON array (`records.path: "."`, matching legacy's `recordsPath: "."` /
`connsdk.RecordsAt(resp.Body, ".")`) — there is no pagination envelope at all, so no `pagination`
block is declared for either stream, matching legacy's single `r.Do(...)` call with no paginator.

`auctions` maps `id` (integer), `dt` (auction date/timestamp string), and `winning_bid` (number),
identical to legacy's `streamEndpoints["auctions"].fields`. `distilleries` maps `id` (integer),
`name`, and `country`, identical to legacy's `streamEndpoints["distilleries"].fields`. Legacy
declares no `CursorFields` for either stream (`streams()` never sets one), so neither stream
declares `x-cursor-field` or an `incremental` block here — both are full-refresh-only, matching
legacy exactly.

Both streams declare `projection: "passthrough"` (§8 rule 1): legacy's `Read` decodes each stream's
response body with `connsdk.RecordsAt(resp.Body, endpoint.recordsPath)` and emits every element
verbatim — `for _, item := range records { ... emit(connectors.Record(item)) }` (`whisky_hunter.go`)
never field-builds a `connectors.Record{...}` from named keys. Schema-mode projection would silently
drop any raw wire field beyond the 3 documented per stream, which schema-mode projection alone
cannot detect; `passthrough` reproduces legacy's verbatim emission exactly.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **Legacy's fixture-mode-only `stream` marker field is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance)
  stamps an extra `stream` field onto every fixture-mode record (`whisky_hunter.go:134`). This is
  not part of the live record shape; this bundle's schemas and fixtures target the live path only.
  The engine's own conformance/fixture-replay harness provides the credential-free test affordance
  this bundle needs, so no fixture-mode equivalent is needed here.
- No pagination or incremental sync is modeled for either stream, matching legacy exactly — Whisky
  Hunter's public API returns each resource as a single flat array with no cursor field.
