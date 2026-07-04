# Overview

Sharetribe reads and writes Sharetribe listings, users, transactions, events, marketplace, files,
and file-attachment data through the Sharetribe Integration API
(`https://flex-api.sharetribe.com/v1/integration_api/<resource>/<action>`). Read behavior for
`listings`/`users`/`transactions`/`events` is capability-parity migrated from
`internal/connectors/sharetribe` (the hand-written connector it migrates; the legacy package stays
registered and unchanged until wave6's registry flip and was read-only). This Pass B expansion adds
3 new read streams (`marketplace`, `files`, `file_attachments`) and 15 write actions covering the
listing/user/transaction/availability/stock mutation surface, going beyond strict read-only parity
per the Pass B full-surface-expansion charter (see Write actions & risks).

## Auth setup

Provide a Sharetribe Integration API OAuth2 access token via the `oauth_access_token` secret; it is
sent as a Bearer token (`Authorization: Bearer <oauth_access_token>`), matching legacy's
`connsdk.Bearer(token)` (`sharetribe.go:102`) exactly, and is never logged. `base_url` defaults to
`https://flex-api.sharetribe.com/v1` (legacy's `defaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

All 4 streams (`listings`, `users`, `transactions`, `events`) declare `"projection": "passthrough"`:
legacy's `Read` (`sharetribe.go:88-90`) emits `emit(connectors.Record(rec))` verbatim from
`connsdk.Harvest`'s decoded record with no `mapRecord`-style field-building anywhere in the read
path, so schema-mode projection (which would silently drop any raw field not enumerated in
`schemas/*.json`) would be a parity regression versus legacy's every-field-survives behavior.
`schemas/*.json` stay a documentation surface of the well-known fields; passthrough mode means the
schema does not gate which fields actually reach the caller.

All 4 streams share the same shape: `GET
/integration_api/<resource>/query`, records at the response body's `data` array, and `page_number`
pagination (`page`/`per_page` query params, 1-based start page) — an exact port of legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` (`sharetribe.go:87`). A page returning fewer records than `page_size` signals the last
page. Every stream shares an identical field set (`id`, `type`, `attributes`, `updated_at`),
matching legacy's own single shared `fields` slice (`sharetribe.go:115`) — the raw records may
carry additional attributes beyond this set, all of which now survive under passthrough.

No stream is incremental; legacy's `streams()` declares no `CursorFields` for any of the 4 streams
and never filters by time (matching this bundle's omission of any `incremental` block).

Pass B adds 3 new read streams, sharing the same `page_number` pagination/passthrough-projection
shape as the 4 legacy streams (except `marketplace`, a singleton with `pagination: {"type":
"none"}`):

- `marketplace` (`GET /integration_api/marketplace/show`) — a singleton resource (the marketplace's
  own name/description); no query params, no pagination, always exactly one record.
- `files` (`GET /integration_api/files/query`) — every documented query parameter (`ids`,
  `messageId`, `ownerId`, `createdAtStart`, `createdAtEnd`) is optional, so this genuinely supports
  an unfiltered list-all-files read, unlike `messages/query`/`availability_exceptions/query`/
  `stock_adjustments/query` below.
- `file_attachments` (`GET /integration_api/file_attachments/query`) — same reasoning: every
  query parameter (`ids`, `fileIds`, `messageId`) is optional.

## Write actions & risks

Pass B adds 15 write actions covering the Integration API's create/command surface (all `POST`,
even the ones that read as REST "updates" — Sharetribe's Integration API uses action-suffixed
paths like `/update`/`/close`/`/approve` rather than HTTP PATCH/PUT for mutations):

- **Listings**: `create_listing`, `update_listing` (cannot change listing state — Sharetribe
  documents state changes as separate commands), `close_listing`, `open_listing`,
  `approve_listing` (approves a `pendingApproval` listing to `published`; review before enabling
  in a caller with untrusted input if the marketplace relies on manual moderation).
- **Users**: `approve_user` (activates a pending account — higher scrutiny, grants marketplace
  access), `update_user_profile`, `update_user_permissions` (changes `postListings`/
  `initiateTransactions`/`read` access-control flags — higher scrutiny), `verify_user_email`.
- **Transactions**: `transition_transaction` (drives the marketplace's transaction process —
  accept/decline/mark-paid/etc; only operator-actor transitions are permitted by the API itself;
  can trigger real payment capture/payout depending on the process definition — review before
  enabling in a caller with untrusted input), `transition_transaction_speculative` (simulates a
  transition without changing state — safe to call freely, useful for previewing a price
  breakdown), `update_transaction_metadata`.
- **Availability & stock**: `create_availability_exception`, `delete_availability_exception`
  (`kind: delete`, `delete.idempotent: true` — the API's own delete command has no documented
  "already deleted" distinguished status, so idempotency is asserted at the dialect level rather
  than via a `missing_ok_status` list), `set_listing_stock` (`stock/compare_and_set` — a
  compare-and-swap that only applies if the listing's current stock matches the submitted
  `oldTotal`), `create_stock_adjustment` (stock adjustments are immutable once created).

`publicData`/`privateData`/`protectedData`/`metadata` object fields on `update_listing`/
`update_user_profile`/`update_transaction_metadata` are merged with the existing object on the
**top level only** (not deep-merged) by the API itself — a caller wanting to clear a nested key
must still submit the full desired value for that top-level key, matching Sharetribe's own
documented merge semantics exactly (this bundle does not attempt to reimplement or validate that
merge behavior client-side).

**Deliberately NOT implemented**: `POST /integration_api/images/upload` (excluded in
`api_surface.json` as `binary_payload` — a multipart file upload, not JSON). Image/profile-image
attachment by previously-uploaded id already flows through `create_listing`/`update_listing`/
`update_user_profile`'s `images`/`profileImageId` fields.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable per the engine dialect.** Legacy exposes
  both as config-driven overrides (`sharetribe.go:79-86`, `positiveInt`/`parseMaxPages`, `page_size`
  clamped 1-1000, `max_pages` defaulting to 1). The `page_number` paginator's `page_size` is a fixed
  value baked into `streams.json`'s `base.pagination` block, and there is no per-request `max_pages`
  override mechanism at all (conventions.md §3). Neither key is declared in `spec.json` (a
  declared-but-unwireable key is worse than an absent one — searxng precedent).
- `page_size` is baked at `100`, matching legacy's own default (`sharetribe.go:19`,
  `defaultPageSize`); fixtures use a full 100-record page 1 for `listings`, matching a valid first
  page that legacy would read before the default `max_pages: 1` cap stops pagination.
- Legacy's own default `max_pages` is `1` (`sharetribe.go:19`, `defaultMaxPages`); this bundle bakes
  in `max_pages: 1`, matching the legacy default. The legacy runtime config override remains
  intentionally absent because the current pagination dialect has no per-request `max_pages`
  template.
- **`messages/query`, `availability_exceptions/query`, and `stock_adjustments/query` are NOT
  modeled as streams** (`api_surface.json`'s `excluded: {category: out_of_scope}` entries): each
  requires a mandatory parent-resource filter (`transactionId`/`ids` for messages; `listingId` for
  the other two) with no unfiltered list-everything shape — there is no marketplace-wide "all
  messages"/"all availability exceptions"/"all stock adjustments" endpoint to sync top-level,
  unlike `listings`/`users`/`transactions`/`events`/`files`/`file_attachments`. A future capability
  expansion could model one of these as a `fan_out` stream keyed off the `transactions`/`listings`
  stream's own ids (conventions.md §3's `fan_out` dialect) if per-parent enumeration becomes
  valuable.
- **`stock_reservations/show` is NOT modeled** (excluded `out_of_scope`): it is a single-id GET
  with no corresponding list/query endpoint at all — stock reservations are reachable only via the
  `stockReservation` relationship of a stock adjustment, so there is no way to enumerate reservation
  ids to iterate this endpoint over.
