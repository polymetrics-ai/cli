# Overview

Braze is a declarative-HTTP migration (unblocked from quarantine now that the engine's
`page_number` paginator supports an explicit 0-indexed `start_page`). It reads Braze campaigns,
Canvases, and segments through the Braze REST API list-export endpoints. This bundle targets
capability parity with `internal/connectors/braze` (the hand-written connector it migrates) for
the `campaigns`/`canvases`/`segments` streams; the legacy package stays registered and unchanged
until wave6's registry flip, and remains authoritative for the `events` stream this bundle does
not cover (see Known limits).

## Auth setup

Provide a Braze REST API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`braze.go:252`). Braze has no single global host — each customer is
provisioned on a regional REST endpoint (e.g. `https://rest.iad-01.braze.com`) — so `base_url` is
**required**, matching legacy's `brazeBaseURL`'s hard requirement (`braze.go:275-277`, no
built-in default).

## Streams notes

All 3 streams (`campaigns`, `canvases`, `segments`) share the same shape: `GET` against the Braze
list-export endpoint, records at the endpoint's own top-level array field
(`campaigns`/`canvases`/`segments`), primary key `["id"]`. `campaigns` and `canvases` additionally
declare `last_edited` as the incremental cursor field (legacy's `CursorFields`); `segments` has
none, matching legacy exactly (`streams.go`'s `brazeStreams`). Braze paginates its list-export
endpoints with a genuinely 0-based `?page=` query parameter and never sends a page-size query
param at all (legacy's `harvest` loop starts `for page := 0; ...` and only ever sets `page`, never
a size param) — expressed here as `pagination.type: page_number`, `page_param: page`,
`size_param: ""` (never send a size param), `start_page: 0`. The engine's `page_number` paginator
stops on a short page (fewer than the declared `page_size: 100` records), identical to legacy's
own `len(records) < pageSize` stop rule.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy exactly
(the Braze REST API surface this connector targets has no safe, generic reverse-ETL action).

## Known limits

- **The `events` stream is not modeled by this bundle (ENGINE_GAP, different from the pagination
  gap this connector was originally quarantined for).** Braze's `/events/list` endpoint returns an
  array of bare event-NAME STRINGS (not objects) — legacy's `decodeRecords` (`braze.go:172-213`)
  special-cases this shape by wrapping every string element into `{"name": <string>}` before
  mapping. The engine's declarative record extraction (`connsdk.RecordsAt`,
  `internal/connectors/connsdk/extract.go:33-56`) only ever yields a `Record` for an array element
  that decodes as a JSON object (`map[string]any`); a bare string element is silently skipped, not
  wrapped — there is no declarative primitive to turn a scalar array element into a one-field
  record. This is a genuine ENGINE_GAP (an omission would silently under-report Braze's events, not
  a defensible parity approximation), not a Tier-1/2-fixable shape (no `computed_fields` or hook
  interface runs before/instead of `RecordsAt`'s own array-element filtering). Legacy stays
  authoritative for the `events` stream until the engine gains an array-of-scalars wrapping
  primitive; see `api_surface.json`'s `excluded` entry for `/events/list`.
- Full Braze API surface (message send endpoints, user data export/import, catalogs, campaign/
  Canvas analytics data series) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `page_size`/`max_pages` config overrides legacy exposes (`brazePageSize`/`brazeMaxPages`, clamped
  1-100 / `all`/`unlimited`) are not runtime-configurable here: the engine's `page_number`
  paginator's `PageSize` is a static int set once in `streams.json`, not template-resolvable, and
  `PaginationSpec` has no config-driven `MaxPages` override wired to any spec key. `spec.json`
  intentionally does not declare `page_size`/`max_pages` (a declared-but-unwireable key is worse
  than an absent one, per conventions.md F6).
