# Overview

DefiLlama is a public, unauthenticated DeFi analytics API split across a small number of hosts
(`api.llama.fi` for most endpoints, `stablecoins.llama.fi` for the stablecoins overview). This
bundle reads protocols, chains, stablecoins, DEX trading-volume overviews, and fees/revenue
overviews. Read-only, no credentials required. This bundle migrates
`internal/connectors/defillama` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

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
- `dexs` (`GET /overview/dexs`, records at `protocols`) and `fees` (`GET /overview/fees`, records
  at `protocols`) both send the fixed query params `excludeTotalDataChart=true` and
  `excludeTotalDataChartBreakdown=true` (matching legacy's static per-endpoint `query` map) and are
  read as a single, unpaginated request — legacy never pages these two endpoints either.
- No stream declares an `incremental` block: DefiLlama's public API is a full-refresh-only
  analytics snapshot with no updated-since filter, matching legacy exactly (no cursor fields were
  ever declared for any DefiLlama stream).

## Write actions & risks

None. DefiLlama is a read-only public analytics API (`capabilities.write: false`); legacy's
`Write` unconditionally returns `ErrUnsupportedOperation`.

## Known limits

- **`stablecoins`'s dual-host absolute path breaks fixture-replay conformance for that stream
  only** (`streams.json`'s per-stream `conformance.skip_dynamic` marker): the engine substitutes
  the conformance replay server's origin only for `base.url`-relative stream paths; an
  `http(s)://`-prefixed `stream.path` is dispatched exactly as declared, so a dynamic
  (fixture-replay) read against this stream would dial the real `stablecoins.llama.fi` instead of
  the test double. `fixtures/streams/stablecoins/page_1.json` still exists and matches
  DefiLlama's real documented `peggedAssets` envelope shape (proving the schema/record-mapping is
  correct), but the dynamic engine-vs-replay-server checks (`read_fixture_nonempty:stablecoins`,
  and this stream's exclusion from `pagination_terminates`/`records_match_schema`/
  `cursor_advances`' candidate-stream selection) are skipped, not silently faked. The other 4
  streams (all on `api.llama.fi`, matching `base.url`) are fully dynamically conformance-tested.
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
- Full DefiLlama surface (per-protocol/per-chain historical TVL charts, yields/pools, bridges) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
