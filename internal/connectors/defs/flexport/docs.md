# Overview

Flexport is a wave2 fan-out declarative-HTTP migration. It reads Flexport companies, locations,
products, invoices, and shipments through the Flexport REST API
(`GET https://api.flexport.com/...`). This bundle migrates `internal/connectors/flexport` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Flexport API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(secret)`
(`flexport.go:256`). `base_url` defaults to `https://api.flexport.com` and may be overridden for
tests/proxies.

## Streams notes

All 5 streams share the identical shape: `GET` against the Flexport list endpoint, records nested
two levels deep at `data.data` (Flexport's own envelope: `{"data":{"data":[...],"next":"<url>"}}`),
primary key `["id"]`. Pagination follows Flexport's absolute-next-URL convention
(`pagination.type: next_url`, `next_url_path: "data.next"`), matching legacy's `harvest` loop
exactly (it fetches `data.next` verbatim as the next request's full URL, clearing the base query
once an absolute next URL is followed). The `per` page-size query param is declared via the
optional-query dialect (`{"template": "{{ config.page_size }}", "default": "100"}`, conventions.md
§3) so a configured `page_size` overrides Flexport's default of 100 exactly like legacy's
`flexportPageSize`, and the same default (100) applies when unset.

Every stream also declares `x-cursor-field: updated_at` on its schema, matching legacy's own
catalog declaration (`CursorFields: []string{"updated_at"}`) — **but legacy's `Read`/`harvest`
never actually sends an incremental request parameter or filters by this field client-side; it is
a full-refresh-only read path** (confirmed: `flexport.go`'s `harvest` has no reference to
`req.State`/cursor anywhere in its live request-building logic). This bundle therefore declares NO
`incremental` block on any stream — adding one would be new, behavior-changing filtering legacy
never had (the meta-rule in conventions.md §5 forbids inventing behavior legacy doesn't exercise).
`x-cursor-field` is retained purely as the same catalog-level candidate-cursor annotation legacy
publishes, without implying an incremental read mode the engine would derive from a nonexistent
`incremental` block (conventions.md §2's sync-mode-derivation rule: `incremental_append` requires
an `incremental` block to apply at all, so this bundle correctly surfaces as full-refresh-only,
matching legacy's real behavior).

## Write actions & risks

None. Flexport is read-only in legacy (`capabilities.write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); this bundle ships no `writes.json`.

## Known limits

- **Fixture requests do not assert the `per` query value.** `fixtures/streams/**/page_1.json` and
  `fixtures/check.json` declare only `method`+`path` (no `query` block) for the replay match.
  Conformance's synthetic runtime config (`internal/connectors/conformance`'s
  `runtimeConfigForEngine`) populates EVERY declared, non-secret `spec.json` property with a fixed
  synthetic placeholder string, so the optional-query dialect's `{{ config.page_size }}` reference
  resolves to that placeholder (not its `default: "100"`) during conformance replay — the `default`
  only ever fires for a genuinely ABSENT config key, which conformance's synthetic-config generator
  never produces for a declared property. A fixture asserting the literal `per=100` would therefore
  never match a conformance-driven request and always 404; omitting the query assertion (matching
  calendly's identical `count`-templated-query precedent) is the correct fixture-authoring pattern
  for any templated, non-secret query param, not a coverage gap — the literal `per=100` value is
  still exercised for real by any operator who leaves `page_size` genuinely unset in production.
- **`fixtures/streams/**` ship one page per stream, per conventions.md §4's sanctioned `next_url`
  exception.** A `next_url` stream's next-page URL is the replay server's own runtime address,
  unknown until the harness picks a port — a static fixture file cannot embed the correct absolute
  URL for a second page. Every stream in this bundle uses `next_url` pagination (unlike bitly's
  pilot, which had three non-paginated streams available), so `pagination_terminates` exercises the
  bundle's first stream (`companies`) against its single-page fixture — a true, if necessarily thin,
  termination check (one page in, one request out, read returns cleanly). Real 2-page `next_url`
  correctness for Flexport's shape is already proven by bitly's/calendly's live `paritytest/<name>`
  precedent for the identical paginator; a dedicated flexport parity suite was not authored in this
  fan-out pass (no parity-suite requirement for non-golden wave2 connectors per
  `docs/migration/conventions.md` §7).
- **`max_pages` (legacy's request-count cap override) is not modeled.** The engine's `next_url`
  paginator has no `MaxPages`-wired config knob analogous to legacy's `flexportMaxPages`; pagination
  is bounded only by the next-page-URL-absent stop signal (Flexport's own real termination
  behavior), matching every other next_url bundle's documented gap (see bitly's `docs.md`).
- **Legacy's fixture-mode-only `previous_cursor` field is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps `record["previous_cursor"] =
  req.State["cursor"]` when a prior cursor happens to be set — this is a credential-free
  conformance-harness affordance, not part of the live wire shape, and is out of scope per the same
  reasoning as bitly's fixture-mode-only fields.
