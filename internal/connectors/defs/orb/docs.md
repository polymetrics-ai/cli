# Overview

Orb is a read-only declarative migration of `internal/connectors/orb` (legacy Go connector). It
reads Orb customers, subscriptions, plans, and invoices via Orb's REST API. This bundle is
capability-parity with legacy; legacy stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Orb API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. There is no unauthenticated fallback mode
— legacy hard-errors when `api_key` is unset (`orb connector requires secret api_key`), matching
this bundle's `required: ["api_key"]`.

## Streams notes

All 4 streams (`customers`, `subscriptions`, `plans`, `invoices`) share the identical shape: `GET`
against the Orb list endpoint, records at `data`, primary key `["id"]`, incremental cursor field
`created_at`. Pagination follows Orb's `pagination_metadata` envelope
(`pagination.type: cursor` with `token_path: pagination_metadata.next_cursor` and
`stop_path: pagination_metadata.has_more`): the next page's `cursor` query param is
`pagination_metadata.next_cursor`, and pagination stops whenever `pagination_metadata.has_more` is
falsy — matching legacy's `nextCursor` helper (`!env.Pagination.HasMore` returns `""`, stopping the
loop) exactly, including the case where `has_more` is true but `next_cursor` happens to be empty
(the engine's `stop_path`-gated `tokenPathCursor` stops on either signal being falsy, same as
legacy's combined check).

`limit` is sent on every request from the `page_size` config value (default `100`, matching
legacy's `defaultPageSize`) via each stream's `query.limit` object-form entry (`default: "100"`).
Incremental reads send `created_at[gte]` verbatim (RFC3339, `param_format` defaults to `rfc3339`)
computed from the sync's persisted cursor, falling back to the RFC3339 `start_date` config value on
a fresh sync — identical to legacy's `lowerBound` helper (`connsdk.Cursor(req.State)` falling back
to `req.Config.Config["start_date"]`). When neither resolves, no `created_at[gte]` param is sent at
all (matches legacy's `if lower := lowerBound(req); lower != "" { ... }` guard).

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size`/`max_pages` config validation (legacy's numeric-range and `all`/`unlimited` keyword
  parsing) is not reproduced at the bundle-config level; the engine treats `page_size` as an opaque
  string substituted directly into the `limit` query param, and has no runtime-config-driven
  "max_pages" page-count-cap mechanism tied to a spec property. Declared in `spec.json` for parity
  of intent/documentation, but a caller-supplied malformed value is sent to Orb as-is rather than
  rejected client-side the way legacy's `strconv.Atoi` validation would. This never changes emitted
  record DATA for any legacy-valid input; it only narrows client-side input validation, out of
  scope for wave2 fan-out (Pass B).
