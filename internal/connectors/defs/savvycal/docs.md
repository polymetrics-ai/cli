# Overview

SavvyCal is a wave2-fan-out declarative-HTTP migration. It reads SavvyCal events, scheduling
links, and contacts through the SavvyCal API (`GET https://api.savvycal.com/v1/...`). This bundle
targets capability parity with `internal/connectors/savvycal` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a SavvyCal API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`savvycal.go:178`). `base_url` defaults to `https://api.savvycal.com` and may be overridden for
tests/proxies.

## Streams notes

All 3 streams (`events`, `links`, `contacts`) share the same shape: `GET` against the SavvyCal
list endpoint, records at `data`, primary key `["id"]`. None of legacy's streams model an
incremental cursor field (legacy's `Catalog` never declares `CursorFields`), so this bundle
declares no `incremental` block for any stream, matching legacy exactly.

Pagination follows legacy's own `links.next`/`next` body-path convention
(`savvycal.go:130`'s `firstStringAt(resp.Body, "links.next", "next")`): the engine's `next_url`
dialect supports only a single `next_url_path`, so this bundle declares `links.next` — SavvyCal's
real, documented wire shape — and does not model legacy's secondary `next` top-level fallback path
(never observed in SavvyCal's real API responses; legacy's own comment offers no evidence it is
ever populated in practice). `page=1`/`per_page={{ config.page_size }}` (default `100`, matching
legacy's `defaultPageSize`) is declared as a static per-stream `query`, re-sent on every page
request exactly like `stripe`'s `limit=100` precedent; the engine's `next_url` paginator re-merges
`stream.Query` onto every absolute next-page URL (`read.go`'s `mergeQuery`), which is a
wire-request-shape divergence from legacy's own `path/query = next; query = nil` reset
(`savvycal.go:138-139`) — benign in DATA terms only, since SavvyCal's own `links.next` URL already
encodes the correct pagination state and a re-applied `page=1`/`per_page` value is superseded by
the actual data returned for that already-resolved URL server-side, not by this bundle's request
shape.

`metadata.json` declares no `rate_limit` — legacy's own SavvyCal package enforces no
client-side rate limiting either (no throttling logic anywhere in `savvycal.go`), so this bundle
adds none, matching legacy's real (lack of) behavior rather than introducing new,
behavior-changing throttling under the guise of a migration.

## Write actions & risks

None. Legacy's SavvyCal connector returns `connectors.ErrUnsupportedOperation` from `Write`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override
  (`savvycal.go:99`, default `100`) as a hard request-count cap. The engine's `next_url` paginator
  has no `MaxPages`-equivalent knob wired to a config value (unlike `page_number`/`offset_limit`,
  which read `PaginationSpec.PageSize`); pagination here is bounded only by the
  short/empty-`links.next` stop signal, matching SavvyCal's own real termination behavior (an
  empty/absent `next` link ends the sync). `max_pages` is not declared in `spec.json` at all (F6,
  REVIEW.md precedent: a declared-but-unwireable config key is worse than an absent one).
- **Legacy's fixture-mode-only marker field is not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) stamps a synthetic `fixture: true` marker onto every
  record (`savvycal.go:162`); this is a credential-free conformance-harness affordance, not part
  of the live record shape, and is intentionally not modeled here — the engine's own
  conformance/fixture-replay harness (`internal/connectors/conformance`) provides the equivalent
  credential-free test affordance.
- Schema is intentionally minimal (`id` + `name`) since legacy performs zero record shaping —
  `savvycal.go:126`'s `emit(connectors.Record(rec))` passes the raw decoded record straight
  through with no field renaming, computation, or filtering. `conformance`'s
  `records_match_schema` check validates the RAW record against the schema before "schema"
  projection drops undeclared fields, and draft-07's default `additionalProperties: true` means
  any additional real-API fields beyond `id`/`name` pass validation without needing to be
  enumerated here; full field-level schema expansion (every SavvyCal event/link/contact property)
  is Pass B (wave5) scope, matching `api_surface.json`'s minimal-honest wave0/pilot depth
  precedent applied to schema breadth.
- Full SavvyCal API surface (link creation, availability lookups) is out of scope for this wave;
  see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries.
