# Overview

Recurly is a wave2 fan-out declarative-HTTP migration. It reads Recurly accounts, subscriptions,
invoices, transactions, and plans through the Recurly v3 REST API
(`GET https://v3.recurly.com/...`). This bundle targets capability parity with
`internal/connectors/recurly` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide a Recurly private API key via the `api_key` secret. It is sent as the HTTP Basic username
with an empty password (`Authorization: Basic base64(api_key:)`), matching legacy's
`connsdk.Basic(key, "")` (`recurly.go:122`) exactly via the engine's `basic` auth mode. A static
`Accept: application/vnd.recurly.v2021-02-25` header is declared on `base.headers`, matching
legacy's `Requester.Accept` field (`recurly.go:122`) — Recurly's v3 API requires this versioned
Accept header on every request. `base_url` defaults to `https://v3.recurly.com` and may be
overridden for tests/proxies.

## Streams notes

Five streams, all primary-keyed on `id` (Recurly's real wire-format IDs are opaque strings, e.g.
`e28zov4fw0v2`, not integers). Each hits a flat Recurly v3 list endpoint (`/accounts`,
`/subscriptions`, `/invoices`, `/transactions`, `/plans`) whose records live under the top-level
`data` array, matching legacy's `connsdk.Harvest(..., "data", ...)` call (`recurly.go:92`).
Pagination follows RFC 5988 `Link: <url>; rel="next"` headers (`pagination.type: link_header`,
`page_size: 200` matching legacy's `recurlyDefaultPageSize`), matching legacy's own
`connsdk.LinkHeaderPaginator` (`recurly.go:91`) exactly — both sides follow the API's own
`Link` header until it stops appearing, with no separate short-page/total-count stop rule on
either side.

Every stream projects Recurly's real field names verbatim via bare `computed_fields` references
(so the engine's typed extraction preserves each field's native wire type): `accounts` maps
`id`/`code`/`email`/`state`/`created_at`/`updated_at`; `plans` maps `id`/`code`/`name`/`state`/
`updated_at`; `transactions` maps `id`/`account_id`/`status`/`amount`/`created_at`;
`subscriptions` and `invoices` additionally reach into a nested `account`/`plan` object
(`{{ record.account.id }}`, `{{ record.plan.id }}` — the engine's `record.<dotted.path>` reference
walks nested `map[string]any` values) to derive `account_id`/`plan_id`, matching legacy's
`nestedID()` helper's primary case (an embedded `AccountMini`/`PlanMini` object's own `id` field,
confirmed always present on Recurly v3's real wire shape). `total` (invoices) and `amount`
(transactions) are Recurly v3 JSON numbers (`float64` in Recurly's own Go client), preserved as
`"number"` schema types via the same bare-reference typed extraction, matching legacy's raw
`item["total"]`/`item["amount"]` copy exactly.

Legacy's own field-name choice (`state` for accounts/subscriptions/invoices/plans, `status` for
transactions) matches Recurly v3's actual canonical field names for each resource (confirmed
against Recurly's own v3 Go client: `Account`/`Subscription`/`Invoice`/`Plan` structs all declare
`state`, `Transaction` declares `status`, not `state`) — this bundle projects the same primary field
per stream, not a fallback (see Known limits for the two-field-fallback gap this narrows).

None of the five streams expose a server-side incremental filter parameter in legacy (`Read` never
sends a date-filter query param — the paginator only ever sends `limit`), so this bundle declares no
`incremental` block for any stream, matching legacy exactly. `x-cursor-field` is still declared per
schema as informational catalog metadata only (`updated_at` for accounts/subscriptions/plans,
`created_at` for transactions, matching legacy's own `Catalog` `CursorFields` declaration
verbatim for those four streams) — per `docs/migration/conventions.md` §2, `incremental_append` sync
modes are gated on the presence of an `incremental` block, not on `x-cursor-field` alone.

## Write actions & risks

None. Legacy's own `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write`
is `false` and this bundle ships no `writes.json`.

## Known limits

- **`invoices`' declared cursor field (`updated_at`) does not exist on the emitted record — fixed,
  not carried forward.** Legacy's own `Catalog` declares `CursorFields: []string{"updated_at"}` for
  the `invoices` stream (`recurly.go:146`), but `invoiceRecord` (`recurly.go:158-160`) never emits an
  `updated_at` field (only `id`/`account_id`/`state`/`total`/`created_at`) — a pre-existing legacy
  inconsistency (a declared cursor field the stream's own record shape never carries). The engine's
  loader hard-requires `x-cursor-field` to name a property that actually exists in that same schema
  (`cursor_field_missing`/`cursor_fields_exist`, `docs/migration/conventions.md` §2), so this bundle
  declares `created_at` (the field `invoiceRecord` actually emits) as `invoices`' `x-cursor-field`
  instead of byte-copying legacy's inconsistent declaration. This is a corrected catalog-metadata
  defect, not an emitted-record DATA change (invoices' emitted fields are unchanged from legacy).
- **`id`/`code` and `state`/`status` two-field fallbacks are not modeled.** Legacy's `nestedID()`
  helper falls back from an embedded object's `id` to its `code` (`recurly.go:168-173`), and several
  `mapRecord` functions fall back between `state`/`status` or `amount`/`total`
  (`recurly.go:152-166`). The engine's `computed_fields` dialect has no coalesce/fallback-between-
  two-paths primitive (only a single dotted-path reference or a filter chain over ONE reference);
  this bundle therefore projects only the primary field per Recurly v3's real, documented API shape
  (confirmed against Recurly's own v3 Go client field tags — see Streams notes), narrowing legacy's
  defensive-only fallback behavior (a hypothetical missing-primary-field shape not observed in
  Recurly's current API), never its accepted-input behavior for the real API's actual wire shape.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`recurlyDefaultPageSize`/`recurlyMaxPageSize`, `recurly.go:221-231`). The engine's `link_header`
  paginator's `PageSize` (used only as the initial-request `limit` value; the paginator itself
  follows the `Link` header thereafter, never recomputing `limit`) is a bundle-declared constant
  (`streams.json`'s per-stream `query: {"limit": "200"}` plus `base.pagination.page_size: 200`), with
  no per-request config-driven override mechanism. This bundle therefore fixes Recurly's own default
  (`limit=200`) and does not declare `page_size`/`max_pages` in `spec.json` at all (a
  declared-but-unwireable config key is worse than an absent one, per the bitly/searxng/pagerduty F6
  precedent).
- **Every stream fixture ships a single page rather than the usual 2-page requirement (§4).** This is
  the same harness limitation the `gitlab`/`greenhouse`/`freshdesk` bundles document for their own
  `link_header` streams: a fixture file has no field to declare a `Link:` response header, so a
  second page can never be expressed in a static fixture for `link_header` pagination.
  `pagination_terminates` still passes (a single-page fixture with no `Link` header terminates after
  exactly one request, the correct, honest outcome). No live parity test exists for Recurly's 2-page
  Link-header advance in this wave; a future wave could add one (bitly/calendly's `next_url` pattern)
  if this needs live proof.
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a fixed
  2-record set with synthetic values regardless of stream (`recurly.go:100-111`); this is a test-only
  affordance, not part of the live record shape. The engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
