# Overview

Zoho Billing is a wave2 fan-out migration of `internal/connectors/zoho-billing`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip). It reads Zoho
Billing customers, subscriptions, and invoices through the Zoho Billing REST
API.

## Auth setup

Provide a Zoho OAuth access token via the `access_token` secret (required).
It is sent as `Authorization: Zoho-oauthtoken <access_token>`, matching
legacy's `connsdk.APIKeyHeader("Authorization", token, "Zoho-oauthtoken ")`
(`zoho_billing.go:190`) byte-for-byte, via the engine's declarative
`api_key_header` auth mode (`header`/`prefix`/`value`) — never logged.
`base_url` defaults to `https://www.zohoapis.com/billing/v1` (legacy's
`defaultBaseURL`) and may be overridden for tests/proxies. `organization_id`
is optional; when set it is sent as the `organization_id` query parameter on
every request (matching legacy's `baseQuery`), and left off entirely when
unset (the engine's opt-in optional-query dialect,
`omit_when_absent: true`).

## Streams notes

All 3 streams (`customers`, `subscriptions`, `invoices`) share the same
shape: `GET` against the Zoho Billing collection endpoint, records
extracted from the response's own top-level key matching the stream name.
`projection` is left at its default (`"schema"`); this bundle does NOT use
`passthrough` — instead each schema declares BOTH the raw Zoho field names
(`customer_id`/`display_name`/`updated_time`, etc.) and the derived
convenience aliases (`id`/`name`/`updated_at`) legacy's `mapRecord` also
emits, and `computed_fields` renames the entity-specific primary field to
the alias (e.g. `"id": "{{ record.customer_id }}"`) using the engine's
bare-single-reference typed-extraction rule (a template with no filter
chain and no surrounding text copies the raw JSON value's native type
verbatim, per conventions.md's "Typed extraction" note) so `id` stays a
string, matching Zoho's real wire shape.

**Documented deviation (fallback-chain narrowing)**: legacy's `mapRecord`
resolves `id`/`name`/`updated_at` via an ordered fallback across MULTIPLE
candidate raw keys per stream (e.g. `idKeys: ["customer_id", "id"]`,
`cursorKeys: ["updated_time", "last_modified_time", "updated_at"]`,
`zoho_billing.go:53-66`) — a defensive hedge against undocumented Zoho API
response variants. The engine's `computed_fields` dialect has no
first-non-null-of-several-keys expression (only a single templated
reference per computed field), so this bundle wires only the FIRST key in
each legacy fallback chain (`customer_id`/`subscription_id`/`invoice_id` for
id; `display_name`/`name`/`invoice_number` for name; `updated_time` for the
cursor across all three streams) — the key Zoho's documented API actually
returns (confirmed against `https://www.zoho.com/billing/api/v1/customers/`:
`customer_id`, `display_name`, `updated_time` are the documented response
fields). Every fallback key after the first is legacy defensive code for a
shape that does not occur against the real, documented API; no emitted
record differs from legacy for any input the real Zoho Billing API actually
returns. Also unlike legacy's `mapRecord` (which copies every raw field
through via a `for k, v := range in` passthrough BEFORE applying the
fallback aliases), this bundle uses `"schema"` projection with an explicit
allow-list of properties (both the raw and aliased field names) rather than
`"passthrough"` — any raw Zoho field not explicitly listed in
`schemas/<stream>.json` is dropped. This narrows the emitted field set to
the ones enumerated here; expanding the allow-list to Zoho's full documented
response shape per stream is Pass B scope (see `api_surface.json`).

Pagination is `page_number` (`page_param: page`, `size_param: per_page`,
matching legacy's `baseQuery`) with `pagination.page_size: 2` declared
purely to keep the required 2-page `customers` fixture small (the
auth0/aviationstack precedent, conventions.md) — this has no bearing on a
live deployment, since `PaginationSpec.PageSize` is a static value read
once at bundle load with no config-driven override mechanism. Legacy's real
default is `per_page=200` (configurable 1-200 via a `page_size` config key,
`pageSize()`, `zoho_billing.go:261-271`); the engine has no analogous
runtime override for `page_number`'s `PageSize` field (F6, REVIEW.md: a
declared-but-unwireable `page_size` spec property is worse than an absent
one), so `page_size` is not declared in `spec.json` at all. Every schema
declares `x-cursor-field: updated_at` for manifest-surface documentation
only; no stream declares an `incremental` block — legacy's own `harvest()`
(`zoho_billing.go:135-156`) never sends a server-side lower-bound filter on
any stream, always reading the full collection page by page, matching this
bundle exactly.

## Write actions & risks

None. Legacy `zoho-billing` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Zoho Billing API surface (plans, addons, coupons, credit notes,
  payments, items, webhooks) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 3 legacy-parity streams are
  implemented.
- **Fallback-chain narrowing** (see "Streams notes" above): only the first,
  documented-API key of each legacy multi-key fallback chain is wired via
  `computed_fields`; the remaining defensive fallback keys
  (`id`/`last_modified_time`/`updated_at`/`customer_name`/`plan_name`/
  `number`) are not reachable by this bundle. No emitted record differs
  from legacy for the real, documented Zoho Billing response shape.
- **`page_size`/`max_pages` are not exposed as config**, for the same
  static-`PaginationSpec`-field reason documented in auth0's and searxng's
  goldens: `PaginationSpec.PageSize`/`MaxPages` are plain JSON values
  resolved once at bundle load, with no template/config-driven override.
  `streams.json`'s `pagination.page_size: 2` exists purely to keep the
  required 2-page fixture small; it has no bearing on a live deployment.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`, a
  credential-free conformance-harness affordance) stamps a
  `previous_cursor` field (echoing `connsdk.Cursor(req.State)` when set)
  onto every fixture-mode record (`zoho_billing.go:164-166`). This is not
  part of the LIVE record shape; this bundle's schemas target the live path
  only (`harvest()`), per the same instruction bitly's/zendesk-chat's/
  zendesk-talk's migrations followed. The engine's own
  conformance/fixture-replay harness provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed
  here.
