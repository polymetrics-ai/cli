# Overview

DefiLlama is a public, unauthenticated DeFi analytics API split across four hosts (`api.llama.fi`
for TVL/volumes/fees/perps overviews, `stablecoins.llama.fi` for stablecoin data,
`yields.llama.fi` for yield pools). This bundle reads protocols, chains, stablecoins,
stablecoin-chain totals, DEX trading-volume overviews, options-dex volume overviews, perps
open-interest overviews, fees/revenue overviews, yield pools, and global historical chain TVL —
the full practically-syncable surface of DefiLlama's documented free-tier API (Pass B,
2026-07-04; cross-checked against DefiLlama's own published OpenAPI spec,
`https://github.com/DefiLlama/api-docs`'s `defillama-openapi-free.json`). Read-only, no
credentials required. This bundle migrates `internal/connectors/defillama` (the hand-written
connector); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

None. DefiLlama's public API needs no credentials; `spec.json` declares no `x-secret` property at
all (matching legacy, which never builds an `Auth` value for its `connsdk.Requester`).

## Streams notes

- `protocols` (`GET /protocols`, records at the response body root) and `chains` (`GET
  /v2/chains`, records at the response body root) are client-side paginated with
  `pagination.type: offset_limit` (`limit`/`offset` query params, `page_size: 1000`) — DefiLlama's
  real endpoints return the full list in one response with no server-side pagination contract, but
  legacy pages the response client-side anyway to keep payloads bounded, and this bundle
  reproduces that same client-side limit/offset request shape (a short/empty page from the API
  stops the loop, matching legacy's `harvest`'s "short page ends the list" rule exactly).
- `stablecoins` (`GET /stablecoins?includePrices=true`, records at `peggedAssets`) lives on a
  SECOND host, `stablecoins.llama.fi`, distinct from every other stream's `api.llama.fi` host.
  `streams.json` declares this stream's `path` as a static absolute URL
  (`https://stablecoins.llama.fi/stablecoins`) rather than a `{{ config.base_url }}`-relative
  path — the engine's HTTP layer (`connsdk.Requester.resolveURL`) recognizes an `http(s)://`-
  prefixed stream path as already-absolute and dispatches it as-is, bypassing `base.url`/`BaseURL`
  entirely, which is exactly the mechanism this dual-host API needs. See Known limits for the
  conformance-testing consequence of this.
- `dexs` (`GET /overview/dexs`, records at `protocols`), `fees` (`GET /overview/fees`, records at
  `protocols`), `options` (`GET /overview/options`, records at `protocols`), and `open_interest`
  (`GET /overview/open-interest`, records at `protocols`) all send the fixed query params
  `excludeTotalDataChart=true` and `excludeTotalDataChartBreakdown=true` (matching legacy's static
  per-endpoint `query` map for `dexs`/`fees`, extended identically to the two new Pass B overview
  streams since all four share the exact same response envelope shape) and are read as a single,
  unpaginated request — none of the four are ever paged.
- `pools` (`GET https://yields.llama.fi/pools`, records at `data`) is a THIRD host,
  `yields.llama.fi`, distinct from `api.llama.fi`/`stablecoins.llama.fi`. Same absolute-path
  mechanism as `stablecoins` below; unpaginated (DefiLlama returns the full ~15,000-pool list in
  one response) and full-refresh only (`x-primary-key: pool`, the API's own opaque pool GUID).
- `stablecoin_chains` (`GET https://stablecoins.llama.fi/stablecoinchains`, records at the response
  body root) is a per-chain current-circulating-stablecoin-total snapshot, keyed by chain `name`
  (verified unique across the live response). Same second-host absolute-path mechanism as
  `stablecoins`.
- `historical_chain_tvl` (`GET /v2/historicalChainTvl`, records at the response body root,
  `api.llama.fi`) is a global (all-chains-summed) daily TVL time series keyed by `date` (Unix
  seconds, verified unique across the live response) — DefiLlama's own historical-TVL-excluding-
  liquid-staking-and-double-counting endpoint. Unpaginated: the full history (~3200 daily points
  as of 2026) returns in one response.
- No stream declares an `incremental` block: DefiLlama's public API is a full-refresh-only
  analytics snapshot with no updated-since filter, matching legacy exactly (no cursor fields were
  ever declared for any DefiLlama stream) — this includes the newly added `historical_chain_tvl`,
  whose own `date` field could in principle support a lower-bound filter, but the endpoint accepts
  no query parameter to apply one server-side and the full series is cheap enough (~3200 rows) that
  `client_filtered` was not deemed worth the added complexity for this pass; full-refresh only.

## Write actions & risks

None. DefiLlama is a read-only public analytics API (`capabilities.write: false`); legacy's
`Write` unconditionally returns `ErrUnsupportedOperation`.

## Known limits

- **`stablecoins`/`pools`/`stablecoin_chains`'s dual/triple-host absolute paths break
  fixture-replay conformance for those streams only** (`streams.json`'s per-stream
  `conformance.skip_dynamic` marker): the engine substitutes the conformance replay server's origin
  only for `base.url`-relative stream paths; an `http(s)://`-prefixed `stream.path` is dispatched
  exactly as declared, so a dynamic (fixture-replay) read against these streams would dial the real
  `stablecoins.llama.fi`/`yields.llama.fi` instead of the test double. Each affected stream's
  `fixtures/streams/<name>/page_1.json` still exists and matches DefiLlama's real documented
  envelope shape (proving the schema/record-mapping is correct), but the dynamic
  engine-vs-replay-server checks (`read_fixture_nonempty:<name>`, and each stream's exclusion from
  `pagination_terminates`/`records_match_schema`/`cursor_advances`' candidate-stream selection) are
  skipped, not silently faked. The other 7 streams (all on `api.llama.fi`, matching `base.url`) are
  fully dynamically conformance-tested.
- `protocols`/`chains` pagination's declared `page_size: 1000` matches legacy's real default
  (`defillamaDefaultPageSize = 1000`, max `5000`): `spec.json` cannot expose a config-driven
  `page_size` at all (see next bullet), so this bundle fixes one static value rather than legacy's
  config-overridable one, but the fixed value is legacy's own default, not an arbitrarily smaller
  one. `fixtures/streams/{protocols,chains}/page_1.json` request `limit=1000` accordingly and
  return their entire small fixture record set on a single short page (`protocols`'s fixture,
  previously split across two `page_size: 5` pages, is now the single `page_1.json` file with all
  6 records).
- `page_size`/`max_pages`/`stablecoins_base_url` config keys legacy exposed are not declared in
  `spec.json`: the engine's `offset_limit` paginator (and `MaxPages`) read their values only from
  `streams.json`'s statically-declared `pagination` block, with no mechanism to source either from
  `RuntimeConfig.Config` at read time — the same limitation documented for searxng's `page_size`/
  `max_pages` (`docs/migration/conventions.md`'s Tier-1 read-only variant section). A `spec.json`
  property no template ever consumes is dead config (F6, REVIEW.md), so none of the three are
  declared; `base_url` (main host only) remains config-overridable via `{{ config.base_url }}`.
- **Per-identifier detail/history endpoints are out of scope, not silently dropped** — every
  remaining documented endpoint (per-protocol/per-pool/per-chain/per-coin historical time series,
  single-coin current/historical prices, block-lookup) requires either a caller-supplied token
  identifier this connector's `spec.json` has no config surface for (coin/price endpoints — there
  is no enumerable catalog of valid `chain:address` identifiers to fan out from), or would need an
  impractically large/expensive fan-out (e.g. `/protocol/{protocol}` returns ~33MB of embedded daily
  history for a single one of 7778 protocols; `/chart/{pool}` would mean 15,000+ requests to cover
  every `pools` stream entry). See `api_surface.json`'s per-endpoint `excluded` reasons for the
  full accounting; the previously-documented `bridges.llama.fi` API is no longer present in
  DefiLlama's current published OpenAPI surface at all and is marked `deprecated` rather than
  `out_of_scope`.
