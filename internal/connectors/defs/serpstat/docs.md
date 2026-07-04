# Overview

Serpstat exposes its SEO data through a single JSON-RPC-over-HTTP endpoint (`POST /v4`) whose
request body carries `{"id": <page>, "method": "<procedure>", "params": {...}}` and whose
pagination state (the current page number) lives INSIDE that request body, not in the query
string. This is a Tier-2 `StreamHook` migration (quarantine.json's original `ENGINE_GAP` finding):
`internal/connectors/engine/bundle.go`'s `StreamSpec.Body` field exists but
`internal/connectors/engine/read.go`'s declarative read path never sends a body (`read.go`'s
`readOneSequence` always issues `rt.Requester.Do(ctx, method, reqPath, query, nil)` — the body
argument is hard-coded `nil`), so a POST-body JSON-RPC read with in-body pagination state cannot be
expressed in `streams.json` alone. `internal/connectors/hooks/serpstat/hooks.go` implements
`StreamHook`, porting `internal/connectors/serpstat/serpstat.go`'s JSON-RPC body construction,
in-body page-number pagination, and `result.data` record extraction verbatim. This bundle is
engine-vs-legacy parity-tested against `internal/connectors/serpstat` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Serpstat authenticates via a single `api_key` token sent as the **`token` query-string parameter**
on every request (never a header, never in the JSON-RPC body) — matching legacy's
`requester()`/`Read()` (`serpstat.go:94,104`: `query := url.Values{"token": []string{token}}`,
appended to every JSON-RPC POST). This bundle wires the identical shape **declaratively, with no
AuthHook needed at all**: `streams.json`'s `base.auth` declares a single `api_key_query` candidate
(`{"mode":"api_key_query","param":"token","value":"{{ secrets.api_key }}"}`). Auth is intentionally
NOT part of `hooks/serpstat/hooks.go` — the engine's `connsdk.Requester` (built by
`engine.newRuntime` before any hook runs) already appends the resolved `token` query param to every
request the hook issues via `rt.Requester.Do`, so the hook only needs to add the JSON-RPC body; it
never touches auth itself. This is why serpstat needed only a `StreamHook`, not the `AuthHook`
interface, despite being filed as a Tier-2 connector in `quarantine.json`.

## Streams notes

Legacy defines 2 streams, each a distinct JSON-RPC `method` value POSTed to the SAME `/v4` endpoint
(`serpstat.go`'s `streamEndpoints` map): `domain_keywords`
(`SerpstatDomainProcedure.getKeywords`) and `domain_competitors`
(`SerpstatDomainProcedure.getCompetitors`). Both streams share the IDENTICAL page-number pagination
shape (`serpstat.go:94-120` `Read`): for `page := 1; pages == 0 || page <= pages; page++`, POST a
body of `{"id": page, "method": "<procedure>", "params": {"domain", "se", "page", "size"}}`,
extract records at `result.data`, and stop the instant a page returns fewer than `page_size`
records (a short-page stop, identical semantics to the engine's own `page_number` paginator's
stop rule — just carried in the request BODY instead of the query string, which is exactly the
shape the declarative `page_number` paginator cannot express).

`hooks/serpstat/hooks.go`'s `ReadStream` ports this loop verbatim, including:

- `pages_to_fetch` (`config.pages_to_fetch`, legacy's `parsePages`): `0` means unbounded (fetch
  until a short page), a positive integer caps the page count at that value — mirrored exactly,
  including legacy's non-negative-integer validation error.
- `domain`/`region_id` config (legacy's `domain`/`region_id`, defaulting to `serpstat.com`/`g_us`
  when unset) are threaded into every page's JSON-RPC `params`.
- `page_size` (`config.page_size`, legacy's `positiveInt(..., 1, 1000, ...)`): both the `size`
  JSON-RPC param and the short-page stop threshold.

Neither `domain_keywords` nor `domain_competitors` is incremental: legacy's `Read` never consults
`req.State`/a cursor at all (grep confirms no state read anywhere in `serpstat.go`) — every read is
a full re-fetch of the configured page range, matching this bundle's schemas declaring no
`x-cursor-field` and `streams.json` declaring no `incremental` block on either stream.

**`domain_urls`** (`SerpstatDomainProcedure.getDomainUrls`) is a Pass B full-surface-expansion
addition, NOT a legacy-ported stream: it returns the list of URLs within the analyzed domain and
each URL's ranking-keyword count. It was chosen as the ONE additional Serpstat JSON-RPC method to
add because it shares the EXACT SAME `{domain, se, page, size}` params shape and in-body
page-number pagination that `hooks/serpstat/hooks.go`'s existing loop already implements —
extending `jsonRPCMethod`'s stream-name-to-procedure map by one entry and adding one schema was the
entire change needed; every other candidate Serpstat method surveyed (`getDomainsInfo`,
`getKeywordsInfo`, `getTopUrls`, etc. — see `api_surface.json`'s per-endpoint exclusion reasons)
takes a materially different params shape (a caller-supplied array, a keyword-driven query, a
different resource model entirely) that would require a SECOND request-builder shape inside the
hook, judged out of proportion to a Pass B pass (see Known limits). `domain_urls` also declares no
`incremental`/`x-cursor-field`, matching the other two streams' full-refresh-only shape (Serpstat's
real API has no updated-since filter for any of these 3 JSON-RPC methods).

### Declarative path (`streams.json`) vs. the live StreamHook path

`streams.json` still declares complete stream/schema metadata for all 3 streams (identity, PK,
field types) — this is what backs the catalog/manifest surface regardless of which path a read
actually takes. Because `hooks/serpstat/hooks.go`'s `StreamHook.ReadStream` recognizes and handles
all 3 stream names unconditionally (`handled=true`), the declarative fallback in `streams.json` is
**never exercised by production traffic** — `engine.Read` only falls through to it when the
`StreamHook` returns `handled=false` (an unrecognized stream name) or no hooks are registered at
all. All 3 streams carry an explicit `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(conventions.md SS4/SS6): `internal/connectors/conformance/dynamic.go` honors this by Skipping every
dynamic fixture-replay check for these streams, since a declarative GET-shaped replay can never
faithfully exercise a JSON-RPC POST + in-body-pagination `StreamHook`. For `domain_keywords`/
`domain_competitors`, the authoritative substitute the marker names is
`internal/connectors/paritytest/serpstat/parity_test.go` (drives the real, hook-dispatched
connector via `engine.HooksFor("serpstat")` against a live `httptest.Server`, asserting byte-for-
byte parity against legacy) and `hooks/serpstat/hooks_test.go`. For `domain_urls` (no legacy
equivalent to compare against — see Known limits), the marker instead names
`hooks/serpstat/hooks_test.go`'s `TestReadStream_DomainUrlsUsesGetDomainUrlsProcedure` as the sole
authoritative proof of its JSON-RPC request shape and record extraction.
`fixtures/streams/<stream>/page_1.json` is retained purely as documentation of the real record
shape each stream emits (and to satisfy `fixtures_present`'s static "first stream ships a fixture"
requirement), not as a load-bearing replay contract, for all 3 streams.

## Write actions & risks

None. Serpstat is a read-only source connector (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- **`StreamSpec.Body` is unwired (ENGINE_GAP, documented, non-blocking; same gap monday's bundle
  already ledgers).** The engine's declarative read path never sends a request body, so a
  POST-body JSON-RPC read with in-body pagination state cannot be expressed in `streams.json`
  alone. `hooks/serpstat/hooks.go`'s `StreamHook` implements the real JSON-RPC POST + in-body
  pagination entirely within the sanctioned Tier-2 hook seam, reusing `rt.Requester` (the engine's
  already-built HTTP client/auth/base-URL plumbing, including the `token` query param) exactly as
  the declarative path itself would.
- **The declarative `streams.json` path is never live-dispatched** (see "Declarative path" above)
  — all 3 streams carry a `conformance.skip_dynamic` marker naming their respective authoritative
  substitute (`paritytest/serpstat` for the 2 legacy streams, `hooks/serpstat/hooks_test.go` alone
  for `domain_urls`).
- **`domain_urls` has no legacy `internal/connectors/serpstat` equivalent, so it has no
  `paritytest/serpstat` coverage.** It is genuinely new Pass B capability, not a ported behavior;
  its correctness is proven only by `hooks/serpstat/hooks_test.go`'s
  `TestReadStream_DomainUrlsUsesGetDomainUrlsProcedure` (JSON-RPC method/params/pagination shape)
  and by `connectorgen validate`'s static checks (schema validity, PK, docs, fixtures presence).
- **Only ONE additional Serpstat JSON-RPC method (`getDomainUrls`) was added this pass**, not the
  full documented Serpstat surface. `hooks/serpstat/hooks.go`'s `StreamHook` implements exactly one
  request-BUILDER shape (`{domain, se, page, size}` params, in-body page-number pagination,
  `result.data` extraction); every other candidate method surveyed during this review
  (`getDomainsInfo`/`getKeywordsInfo`'s array params, `getTopUrls`'s keyword-driven query,
  `getCategoryTopDomains`'s category-id config, the entire backlinks/rank-tracker/site-audit
  product surfaces' own distinct resource models) needs a materially different params shape or an
  entirely different config surface (category id, project id, audit job id) — adding any of them
  would mean a 2nd (or 3rd, 4th...) request-builder shape inside this hook, which is exactly the
  "needing a 3rd hook interface OR unbounded shape growth" signal that should escalate scope rather
  than silently bloat a single Tier-2 hook past what a single migration pass should cover. See
  `api_surface.json`'s per-endpoint `excluded` reasons for the specific incompatibility of each
  surveyed method.
- **No incremental filtering, matching legacy exactly, and extended identically to `domain_urls`.**
  No stream declares `x-cursor-field`/`incremental` — legacy never filters or advances reads by any
  cursor for its 2 streams, and Serpstat's real `getDomainUrls` response has no updated-since
  filter either; every read is a full page-range re-fetch.
- **`updated_at` on `domain_keywords` is a legacy fixture-mode-only artifact, not a real API
  field.** Legacy's live (non-fixture) read path emits the raw JSON-RPC record verbatim with no
  `updated_at` key — only `mode: fixture`'s `readFixture` stamps a static `fixtureUpdatedAt`
  literal. It is declared nullable in this bundle's schema purely for catalog-field parity with
  legacy's published `streams()` Fields list, not because Serpstat's real API sends it.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this bundle.** Legacy's
  `readFixture`/`fixtureMode` (`serpstat.go:151-169`) emit synthetic records without any network
  call when `config.mode == "fixture"` — a legacy-only testing convenience. Parity is asserted
  against legacy's LIVE (httptest-driven) read path only, matching the wave1-pilot convention
  (monday's docs.md carries the identical note).
- **`region_id` defaults to `g_us`, matching legacy's `region` fallback exactly** (`serpstat.go:90-93`).
