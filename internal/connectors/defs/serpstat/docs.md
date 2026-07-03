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

Neither stream is incremental: legacy's `Read` never consults `req.State`/a cursor at all (grep
confirms no state read anywhere in `serpstat.go`) — every read is a full re-fetch of the configured
page range, matching this bundle's schemas declaring no `x-cursor-field` and `streams.json`
declaring no `incremental` block on either stream.

### Declarative path (`streams.json`) vs. the live StreamHook path

`streams.json` still declares complete stream/schema metadata for both streams (identity, PK,
field types) — this is what backs the catalog/manifest surface regardless of which path a read
actually takes. Because `hooks/serpstat/hooks.go`'s `StreamHook.ReadStream` recognizes and handles
both stream names unconditionally (`handled=true`), the declarative fallback in `streams.json` is
**never exercised by production traffic** — `engine.Read` only falls through to it when the
`StreamHook` returns `handled=false` (an unrecognized stream name) or no hooks are registered at
all. Both streams carry an explicit `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(conventions.md SS4/SS6): `internal/connectors/conformance/dynamic.go` honors this by Skipping every
dynamic fixture-replay check for these streams, since a declarative GET-shaped replay can never
faithfully exercise a JSON-RPC POST + in-body-pagination `StreamHook`. The authoritative substitute
these markers name is `internal/connectors/paritytest/serpstat/parity_test.go` (drives the real,
hook-dispatched connector via `engine.HooksFor("serpstat")`) and
`hooks/serpstat/hooks_test.go` — both assert serpstat's real JSON-RPC wire format (request body
shape, in-body pagination, `result.data` record extraction) byte-for-byte against legacy.
`fixtures/streams/<stream>/page_1.json` is retained purely as documentation of the real record
shape each stream emits (and to satisfy `fixtures_present`'s static "first stream ships a fixture"
requirement), not as a load-bearing replay contract.

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
  — both streams carry a `conformance.skip_dynamic` marker naming
  `paritytest/serpstat`/`hooks/serpstat/hooks_test.go` as the authoritative substitute.
- **No incremental filtering, matching legacy exactly.** Neither stream declares
  `x-cursor-field`/`incremental` — legacy never filters or advances reads by any cursor; every read
  is a full page-range re-fetch.
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
