# Overview

Qonto is a wave2 fan-out declarative-HTTP migration. It reads Qonto bank transactions,
memberships, and accounts through the Qonto REST API
(`GET https://thirdparty.qonto.com/v2/<resource>`). This bundle targets capability parity with
`internal/connectors/qonto` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Qonto API key via the `api_key` secret; it is typically formatted
`<organization-slug>:<secret-key>` and is sent **verbatim** as the `Authorization` header (no
`Bearer` prefix), matching legacy's `connsdk.APIKeyHeader("Authorization", token, "")` construction
exactly. Never logged. `base_url` defaults to `https://thirdparty.qonto.com/v2`.

`iban` (the bank account IBAN) is required by the `transactions` stream and by `Check` itself
(legacy's `Check` always calls its `iban(cfg)` validator before issuing any request), but is not
required by `memberships`/`accounts`. It is declared as an optional `spec.json` property (not in
top-level `required`) since it is per-stream, not global — matching legacy's own per-stream
`requiresIBAN` flag.

## Streams notes

All three streams share the identical Qonto page-number-via-body envelope: `GET /<resource>`
returns `{"<resource>":[...],"meta":{"current_page","next_page"}}`, records live at the top-level
resource-named key (`transactions`/`memberships`/`accounts` — matching legacy's per-stream
`recordsPath`). Pagination is declared as `cursor` (`cursor_param: page`,
`token_path: meta.next_page`): the next page's `page` query value is read from the response body's
`meta.next_page` (an integer page number, not an opaque token — Qonto's own "next_page" convention),
and pagination stops when that value is null/absent, matching legacy's own
`strings.TrimSpace(nextPage) == ""` stop rule exactly (no `stop_path` is declared since legacy has no
separate boolean stop signal beyond `next_page` itself).

Every request sends `per_page=100` via each stream's static `query` (matches legacy's
`defaultPageSize`), not via `pagination.size_param`/`page_size` (the `cursor`+`token_path` paginator
constructor never reads those fields).

`start_date`, when configured, is sent as the `start_date` query parameter on every `transactions`
request only (matching legacy's `if start := ...; start != "" { query.Set("start_date", start) }` —
scoped inside the loop that only runs for the requested stream's path/spec, so it is a per-request,
per-stream passthrough, not a cross-stream global). This is declared via the opt-in optional-query
object dialect, omitting the parameter entirely when `start_date` is unset.

`transactions` renames the raw API's `transaction_id` field to the schema's `id` via a bare
`computed_fields` reference (`{{ record.transaction_id }}`), matching legacy's
`first(item, "transaction_id", "id", "slug")` id-resolution for the transactions resource shape
(Qonto's transactions objects carry `transaction_id`, not a bare `id`). `memberships`/`accounts`
objects carry a bare `id` directly, which schema projection copies with no rename needed — this
also matches legacy's own `mapRecord`, which applies the identical coalesce logic uniformly across
all three streams but only ever needs the `transaction_id`-to-`slug` fallback chain for the
transactions resource shape in practice.

Only `transactions` publishes `settled_at` as `x-cursor-field` (matching legacy's own
`CursorFields: []string{"settled_at"}` — `memberships`/`accounts` declare no cursor field in
legacy either). No stream declares an `incremental.request_param`: `start_date` is a static config
passthrough (see above), never a computed/resolved incremental lower bound.

**Legacy's declared field catalog is the schema-parity source** (`streams()`'s
`Fields: []connectors.Field{id, amount, side, settled_at, updated_at}`, applied identically to all
three streams even though `memberships`/`accounts` raw objects carry no meaningful `amount`/`side`/
`settled_at` values in practice) — every schema in this bundle mirrors that exact 5-field catalog
per stream, matching legacy's actual declared parity surface rather than each resource's full raw
wire shape.

## Write actions & risks

None. Legacy `qonto` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`. Qonto's
write surface (transfers, payment initiation) is money-movement and was never modeled by legacy;
`api_surface.json` excludes it `destructive_admin`.

## Known limits

- **`label`/`email`/`created_at` fields are not modeled.** Legacy's `mapRecord` additionally sets
  `label: first(item, "label", "name")`, `email` (unused by any stream in practice), and
  `created_at`, but these are NOT part of legacy's own declared stream field catalog
  (`streams()`'s `Fields` — see Streams notes above) — they exist only inside the record map,
  never advertised as a schema/catalog field. This bundle's schemas mirror the declared catalog
  exactly, consistent with the "schema is a projection of legacy's parity surface" convention.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `cursor` paginator has no config-driven page-size or request-count-cap
  knob; neither is declared in `spec.json`, and this bundle sends Qonto's own default
  (`per_page=100`) as a static per-stream query value.
- **Legacy's `raw` passthrough field is not modeled.** Legacy's `mapRecord` stashes the entire raw
  item under a `raw` key on every emitted record; this bundle's schema projection keeps only the
  declared parity fields (`id`, `amount`, `side`, `settled_at`, `updated_at`).
