# Overview

Customerly was quarantined in wave1 for an `ENGINE_GAP`: both streams (`users`, `leads`) genuinely
paginate from page 0 (legacy's `for page := 0; ...` sends `page=0` as the true first request), and
the engine's `page_number` paginator unconditionally coerced a zero `StartPage` to 1
(`connsdk.PageNumberPaginator.Start()`). This gap was closed by the S4 engine mini-wave's
`PaginationSpec.StartPage *int` (`"start_page": 0` is now distinguishable from an omitted key) —
this bundle is the unblock build using that dialect addition. It reads Customerly users and leads
through the Customerly v1 REST API. This bundle migrates `internal/connectors/customerly` (the
hand-written connector it replaces at capability parity); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Customerly API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Both streams (`users` -> `GET /users/list`, `leads` -> `GET /leads/list`) share the identical
shape: records live at `data.users`/`data.leads` respectively, and every request sends
`sort=last_update&sort_direction=desc` as static query params (matching legacy's `harvest`
query construction exactly). Pagination is genuinely 0-indexed
(`pagination.type: page_number`, `page_param: page`, `size_param: per_page`, `start_page: 0`,
`page_size: 50` matching legacy's `customerlyDefaultPageSize`) — the first request sends `page=0`,
matching legacy's `for page := 0; ...` loop exactly; a page returning fewer than `per_page` records
stops the read (legacy's `len(records) < pageSize` short-page stop).

Neither stream declares an `incremental` block: legacy's `harvest` sends no server-side filter
parameter derived from a cursor or `start_date` config value at all (only the static
`sort`/`sort_direction`/`page`/`per_page` params above) — per conventions.md §8 rule 2, an
`incremental` block is only declared when legacy actually sends a server-side filter, which it does
not here. `x-cursor-field: last_update` is still declared on both schemas (matching legacy's
published `CursorFields`) for catalog/sync-mode-derivation parity even though no request-time
filtering happens.

## Write actions & risks

None. Customerly's v1 API is exposed read-only here (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`customerlyMaxPages`) independent of the short-page stop signal. The engine's `page_number`
  paginator has no `MaxPages`-equivalent config-driven knob (`PaginationSpec.MaxPages` is a fixed
  bundle-authored value, not wired to a runtime config key); pagination here is bounded only by the
  short-page stop signal, which is Customerly's own real termination behavior and is exercised on
  every read regardless of `max_pages`. `max_pages` is not declared in `spec.json` (a declared but
  unwireable key is worse than an absent one, per conventions.md F6).
- Full Customerly API surface (conversations, attributes, events, project settings, and any
  write/mutation endpoints) is out of scope; see `api_surface.json`'s `excluded` entries. Only the 2
  legacy-parity read streams are implemented.
