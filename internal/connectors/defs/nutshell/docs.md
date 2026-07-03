# Overview

Nutshell was quarantined during wave1 for an `ENGINE_GAP`: Nutshell's REST API is genuinely
0-indexed (legacy's harvest loop `for page := 0; ...` sends `page[page]=0` on the very first
request, including the health `Check` call itself), and the engine's `page_number` pagination
could not express a 0-indexed start (a plain Go `int` `StartPage` field could not distinguish an
explicit `0` from an omitted key, so a declared `"start_page": 0` silently coerced to `1`). This
gap was closed by the S4 engine mini-wave's `PaginationSpec.StartPage *int` (`"start_page": 0` is
now distinguishable at the Go layer and honored verbatim) — this bundle is the unblock build using
that dialect addition. It reads Nutshell CRM accounts, contacts, leads, activities, and users
through the Nutshell REST API. This bundle migrates `internal/connectors/nutshell` (the
hand-written connector it replaces at capability parity); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Nutshell account `username` (config) and API token as the `password` secret; both flow
into HTTP Basic auth (`auth: [{"mode": "basic", ...}]`) exactly as legacy's `connsdk.Basic(username,
secret)`. The token is never logged.

## Streams notes

`accounts`, `contacts`, `leads`, and `activities` share the identical shape: `GET` against the
Nutshell list endpoint, records nested one level under a per-stream envelope key matching the
stream name (`records.path: "<stream>"`, e.g. `{"accounts": [...]}` — legacy's
`connsdk.RecordsAt(resp.Body, endpoint.recordsKey)`). Pagination is genuinely 0-indexed
(`pagination.type: page_number`, `page_param: "page[page]"`, `size_param: "page[limit]"`,
`start_page: 0`, `page_size: 500` matching legacy's `nutshellDefaultPageSize`/`nutshellMaxPageSize`)
— the first request sends `page[page]=0`, matching legacy's loop exactly; a page returning fewer
than `page[limit]` records stops the read (legacy's `len(records) < pageSize` short-page stop).

`users` is legacy's one reference (non-paginated) endpoint (`paginated: false` in
`nutshellStreamEndpoints`): it is read in a single request with no `page[page]`/`page[limit]` query
params at all, matching legacy's `harvest` branch that skips setting those params entirely when
`endpoint.paginated` is false. This bundle expresses that with a stream-level `"pagination":
{"type": "none"}` override, which replaces the base-level `page_number` spec wholesale for that one
stream (conventions.md §3: "Stream-level `pagination` replaces the base-level spec **wholesale**").

None of the 5 streams declares an `incremental` block: legacy's `harvest`/`Read` sends no
server-side filter parameter derived from a cursor or `start_date`-shaped config value at all
(only the page[page]/page[limit] params above) — per conventions.md §8 rule 2, an `incremental`
block is only declared when legacy actually sends a server-side filter, which it does not here.
`x-cursor-field: modifiedTime` is still declared on every schema that legacy publishes a
`CursorFields` entry for (`accounts`/`contacts`/`leads`/`activities`; `users` has none) for
catalog/sync-mode-derivation parity even though no request-time filtering happens.

## Write actions & risks

None. Nutshell is exposed read-only here (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- Full Nutshell API surface (tasks, emails, products, custom fields, and any write/mutation
  endpoints) is out of scope; see `api_surface.json`'s `excluded` entries. Only the 5
  legacy-parity read streams are implemented.
- `page_size`'s valid range (1-500) is enforced by the connection spec's description only, matching
  legacy's `nutshellMaxPageSize` constant; the engine does not itself clamp an out-of-range
  `page_size` value — an operator-supplied value outside 1-500 is passed through verbatim to the
  Nutshell API, whose own server-side behavior for an out-of-range `page[limit]` is unspecified
  here (legacy validated this range in Go before ever making a request; this bundle's declarative
  `page_size` config has no equivalent runtime bounds-check mechanism in the dialect).
