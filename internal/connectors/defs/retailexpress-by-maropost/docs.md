# Overview

Retail Express by Maropost is a wave2 fan-out declarative-HTTP migration. It reads Retail Express
products, customers, orders, stock levels, and stores through the Maropost API
(`GET https://<account>.retailexpress.com.au/api/v2/...`). This bundle targets capability parity
with `internal/connectors/retailexpress-by-maropost` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

This bundle accepts either of two credential shapes, matching legacy's own first-match-wins
precedence exactly: an `access_token` secret (sent as a Bearer token) takes priority when
configured; `api_key` (sent as the `X-API-Key` header via `api_key_header` auth mode) is used only
when `access_token` is not configured. The candidate order in `streams.json`'s `base.auth`
(`bearer` first, `api_key_header` second, each gated by its own `when: {{ secrets.* }}` presence
check) reproduces legacy's `authenticator` function: `if token := secret(cfg, "access_token"); token
!= "" { return Bearer(token) }; if key := secret(cfg, "api_key"); key != "" { return
APIKeyHeader(...) }`. Neither secret is ever logged.

## Streams notes

All five streams (`products`, `customers`, `orders`, `stock_levels`, `stores`) return records at
`data`. Pagination is `page_number` (`page`/`limit`, static `page_size: 100` matching legacy's
`defaultPageSize`) with the identical short-page stop rule legacy's own
`connsdk.PageNumberPaginator` implements — an exact parity match, not an approximation.

Legacy applies four optional config-driven filters (`updated_after`, `created_after`, `store_id`,
`status`) uniformly to every stream's request. This bundle reproduces that exact behavior via the
opt-in optional-query dialect (`query.<param>.omit_when_absent: true`) on `base.query`.

Every stream stamps a static-literal `stream` marker field (`"products"`/`"customers"`/`"orders"`/
`"stock_levels"`/`"stores"`) via `computed_fields`, matching legacy's own `out["stream"] = stream`
line in `mapRecord`. `products`/`customers`/`orders`/`stock_levels` publish `updated_at` as
`x-cursor-field` (matching legacy's own `CursorFields` declarations), but Retail Express's list
endpoints expose no server-side incremental filter parameter and legacy's own `harvest` never
applies one — every read is a full paginated sweep, matching legacy's true read behavior (no
`request_param`/`start_config_key`/`client_filtered` declared). `stores` has no cursor field,
matching legacy.

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's `account`-alias base URL derivation is not modeled; `base_url` is required instead.**
  Legacy accepts either a full `base_url` or a bare `account` config key, deriving
  `https://<account>.retailexpress.com.au/api/v2` from the latter
  (`"https://" + url.PathEscape(account) + ".retailexpress.com.au/api/v2"`). The engine's
  `spec.json` `"default"` materialization mechanism only supports a FIXED literal default, not one
  derived from another config value's own value (per `docs/migration/conventions.md` §3's
  spec-default-materialization note) — so per that section's guidance, this bundle requires
  `base_url` directly and drops the `account`-alias derivation rather than inventing ad hoc Go for
  it. An operator previously using the bare `account` key must supply the fully-formed
  `https://<account>.retailexpress.com.au/api/v2` URL as `base_url` instead; the resolved effective
  base URL behavior once configured is identical.
- **Legacy's `id` fallback chain is narrowed to the raw `id` field only.** Legacy synthesizes a
  missing `id` from `first(out, "id", "uuid", "sku", "code", "order_number")` when the raw record
  has no `id` field at all. The `computed_fields` dialect has no OR-fallback primitive across
  multiple record paths, so this bundle relies on schema projection copying `id` directly when
  present. Documented scope narrowing (identical narrowing class as this wave's reply-io bundle).
- **Legacy's multi-candidate records-path search is narrowed to a single fixed path (`data`).**
  Legacy's `recordsAt` helper falls through `data`/`items`/`records`/`results`/root array if the
  declared path yields no records; this bundle declares only `data` (legacy's own per-endpoint
  `streamEndpoint.recordsPath` value for every stream), matching legacy's actual configured
  behavior exactly since every endpoint entry in legacy's `endpoints` map sets `recordsPath: "data"`
  — the fallback chain is defensive dead code for this connector's specific endpoint
  configuration, not an active behavior difference.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `page_number` paginator has no config-driven page-size or
  request-count-cap knob (mirrors this wave's aha/referralhero/rentcast/reply-io precedent);
  `page_size`/`max_pages` are therefore not declared in `spec.json`, and this bundle sends Retail
  Express's own default (`limit=100`) as a static pagination-block value.
- Full Retail Express API surface (purchase orders, suppliers, gift cards, loyalty, webhooks) is
  out of scope for this wave; see `api_surface.json`'s `excluded` entries.
- `docs_url` (`retailexpress.atlassian.net/wiki/...`) returned HTTP 401 (authentication-gated
  Confluence space) when fetched during this migration; legacy's own Go source
  (`internal/connectors/retailexpress-by-maropost/retailexpress_by_maropost.go`) was used as ground
  truth for every endpoint path, auth mode, and record shape instead, per this migration's
  legacy-over-docs precedence rule.
