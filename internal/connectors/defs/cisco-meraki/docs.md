# Overview

Cisco Meraki reads Meraki Dashboard API v1 organization data. This bundle is a **partial** wave2
fan-out migration of `internal/connectors/cisco-meraki` (the hand-written connector it migrates);
the legacy package stays registered and unchanged (both for the migrated stream, until wave6's
registry flip, and permanently for the 3 unmigrated streams described below). Only the
`organizations` stream is implemented here — legacy's other 3 streams
(`organization_networks`, `organization_devices`, `organization_admins`) require an
organization-scoped fan-out read pattern this bundle's declarative dialect cannot express; see
"Known limits" and the reported `ENGINE_GAP` blocker. `capabilities.write` is `false` and this
bundle ships no `writes.json` (Meraki is read-only in legacy: "no obvious safe reverse-ETL write
for these streams").

## Auth setup

Provide a Cisco Meraki Dashboard API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(secret)` exactly, and is
never logged. `base_url` defaults to `https://api.meraki.com/api/v1` and may be overridden for
tests/proxies.

## Streams notes

`organizations` (`GET /organizations`) is the only stream implemented: it lists every organization
the API key can access. The response is a top-level JSON array (`records.path: ""`), matching
legacy's `connsdk.RecordsAt(resp.Body, "")` call. Pagination follows Meraki's own RFC 5988
`Link: <url>; rel="next"` header convention (`pagination.type: link_header`) — the byte-accurate
parity choice, since legacy's own `harvest` function IS Link-header following
(`connsdk.LinkHeaderPaginator`). The first request sends `perPage=1000` (matches legacy's
`merakiMaxPageSize`) via a static per-stream `query: {"perPage": "1000"}`; the engine re-applies
this SAME static query onto every subsequent Link-header-supplied absolute URL too (`read.go`'s
`mergeQuery` runs unconditionally, even when `page.URL` is an absolute next-page URL) — a
wire-request-shape divergence from legacy, which explicitly resets to an empty query once it
follows an absolute Link-header URL (`cisco_meraki.go`'s `harvest`: "Link-header next URLs are
absolute and already carry pagination params ... query = url.Values{}"). This is verified benign
in DATA terms only because Meraki's own next-page URL already carries the identical `perPage=1000`
value the engine re-applies (the replace is idempotent) — matching bitly's identical documented
`next_url` divergence precedent (`docs/migration/conventions.md`, bitly's own `docs.md`). No
`incremental` block is declared: Meraki's organizations endpoint exposes no incremental cursor
field, matching legacy (full refresh, stable primary key `id`).

## Write actions & risks

None. Cisco Meraki is a read-only source connector (legacy's own package doc: "no obvious safe
reverse-ETL write for these streams... Capabilities.Write is false"); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`ENGINE_GAP` (reported blocker, not a workaround): `organization_networks`,
  `organization_devices`, and `organization_admins` are NOT implemented in this bundle.** All
  three are legacy `orgScoped` streams: `Read()` first lists every accessible organization (via
  `organizationIDs`, itself a bounded `harvest` over `/organizations`), then issues one additional
  request PER organization id against an org-parameterized sub-path
  (`organizations/{organizationId}/networks|devices|admins`), stamping the originating
  `organizationId` onto every emitted record. This is a genuine sub-resource fan-out read — the
  set of requests to issue for stream N depends on the RESULT of reading a DIFFERENT stream
  (`organizations`) first, at read time. Nothing in this dialect expresses "read stream A, then use
  each of its records to parameterize N reads of stream B": `streams.json` has no
  parent-stream/depends-on field, `stream.path` interpolation can only reference `config.*`/
  `secrets.*` (resolved once, before any request), never another stream's OWN read results, and
  `PaginationSpec` has no per-page-becomes-look-up-a-different-resource concept at all. This is
  precisely the "sub-resource fan-out reads" trigger `docs/migration/conventions.md` §1 lists as a
  legitimate Tier-2 `StreamHook` case (`ReadStream(ctx, stream, req, rt, emit)` — a whole-stream
  override capable of issuing the organizations lookup internally, then fanning out) — Tier-1 JSON
  genuinely cannot express it, and per this wave's hard rules, authoring a Tier-2 hook package is
  out of scope (a follow-up wave with hooks authority handles this). Reported as a typed
  `ENGINE_GAP` blocker (see the migration result's `blockers[]`) rather than approximated (e.g.
  hardcoding a single organization id, or omitting the `organizationId` stamp) — either
  approximation would silently diverge from legacy's actual multi-organization fan-out behavior for
  any account with more than one accessible organization, a real emitted-record-DATA change the
  conventions.md §5 meta-rule forbids. Legacy `internal/connectors/cisco-meraki` remains
  authoritative for these 3 streams; `api_surface.json` documents each endpoint's exclusion with
  this same reasoning.
- **The fixture-replay harness cannot exercise `link_header`'s real 2-page continuation.**
  `fixtures/streams/**` (`conformance/replay.go`'s `fixtureResponse`) has no field for declaring
  HTTP response headers — only `status` and `body` — so a fixture page can never carry the
  `Link: <url>; rel="next"` header the real Meraki API sends; `pagination_terminates` can only
  observe the paginator's natural single-page stop (no Link header present = no next page).
  Identical structural limitation to `buildkite`'s/`gitlab`'s own documented `link_header`
  bundles in this repo — not a cisco-meraki-specific shortcut. The `organizations` fixture here is
  a single, representative page; the engine's own `link_header` pagination codepath
  (`internal/connectors/engine/paginate.go`'s `linkHeaderPaginator`) is exercised by the shared
  engine's own test suite, not by this bundle's fixtures.
