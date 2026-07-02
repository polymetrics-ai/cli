# Overview

Appfigures is a wave2 fan-out declarative-HTTP migration. It reads Appfigures app-store reviews
through the read-only Appfigures v2 REST API (`GET https://api.appfigures.com/v2/reviews`). This
bundle targets capability parity with `internal/connectors/appfigures` (the hand-written connector
it migrates) for the `reviews` stream only; the legacy package stays registered and unchanged
until wave6's registry flip, and remains authoritative for the `products`/`sales`/`ratings`/
`categories` streams this bundle does not cover (see Known limits).

## Auth setup

Provide an Appfigures Personal Access Token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`appfigures.go:299`). `base_url` defaults to `https://api.appfigures.com/v2`
and may be overridden for tests/proxies.

## Streams notes

`reviews` is a `GET /reviews` list endpoint whose records live at the `reviews` body key, primary
key `["id"]`. Pagination follows Appfigures' page-number convention (`pagination.type:
page_number`, `page_param: page`, `size_param: count`, `page_size: 100`) — legacy's `readPaged`
sends `count=<page_size>&page=<n>` and stops on `this_page >= pages` or a short page; the engine's
`page_number` paginator stops purely on a short page (`recordCount < page_size`), which is
behaviorally identical for every real Appfigures response (the API always returns exactly
`count` records except on the final page) and is the same mapping already used for other
page-number-paginated bundles in this repo (e.g. `appcues`). Optional per-request filters
(`search_store` -> `store`, `group_by` -> `group_by`, `start_date` -> `start`, `end_date` -> `end`)
are wired via the opt-in optional-query dialect (`omit_when_absent: true`), matching legacy's
`appfiguresQuery` exactly: each is sent only when its config value is set, never as an empty
string.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`products`, `sales`, `ratings`, and `categories` are not modeled as streams in this bundle
  (ENGINE_GAP).** All four legacy streams read a JSON object keyed by an arbitrary id (e.g.
  `products/mine` returns `{"111":{...},"222":{...}}`), and legacy's `flattenKeyedObject`
  (`streams.go:224`) turns each value into its own record. The engine's declarative record
  extraction (`connsdk.RecordsAt`, `internal/connectors/connsdk/extract.go:33`) recognizes exactly
  two body shapes at a `records.path`: a JSON array (flattened element-by-element) or a JSON object
  (returned as a SINGLE record, the whole object). There is no dotted-path wildcard or "flatten a
  map-of-objects by key" primitive in the dialect, so a keyed-object endpoint can only ever be
  read as one record covering the whole set — silently wrong (a monolithic pseudo-record standing
  in for N distinct products/report rows), not a defensible approximation. This is a genuine
  ENGINE_GAP, not a Tier-2-fixable shape (no single hook interface flattens the record stream
  itself without also reimplementing the whole read loop, which is a StreamHook — forbidden this
  wave per the fan-out task's hard rules). Legacy stays authoritative for these four streams until
  the engine gains a keyed-object flatten primitive.
- `page_size`/`max_pages` config overrides legacy exposes (`appfiguresPageSize`/
  `appfiguresMaxPages`, clamped 1-500 / `all`/`unlimited`) are not runtime-configurable here: the
  engine's `page_number` paginator's `PageSize` is a static int set once in `streams.json`, not
  template-resolvable, and `PaginationSpec` has no `MaxPages` field read by this paginator type
  (the `MaxPages` cap is a distinct top-level field enforced by the read loop, but no config knob
  wires it). `spec.json` intentionally does not declare `page_size`/`max_pages` (a declared-but-
  unwireable key is worse than an absent one, per conventions.md F6).
