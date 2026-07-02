# Overview

Ubidots reads devices, variables, dashboards, and events from the Ubidots Industrial API (`GET
{base_url}/api/v2.0/<resource>/`). This bundle migrates the hand-written
`internal/connectors/ubidots` legacy package to a declarative Tier-1 defs bundle at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Requires one secret: `token` (Ubidots API token), sent as the `X-Auth-Token` header on every
request via `streams.json` `base.auth`'s `api_key_header` mode — matching legacy's
`connsdk.APIKeyHeader("X-Auth-Token", token, "")` exactly (no prefix). `base_url` defaults to
`https://industrial.api.ubidots.com` (legacy's `defaultBaseURL`), overridable for tests or proxies.

## Streams notes

Four streams sharing the identical `page_number` pagination shape (legacy's own
`page_size`/`page` query convention) and the identical record mapping (`id`, `label`, `name`,
`created_at`): `devices` (`GET api/v2.0/devices/`), `variables` (`GET api/v2.0/variables/`),
`dashboards` (`GET api/v2.0/dashboards/`), `events` (`GET api/v2.0/events/`) — all read records from
the paginated envelope's `results` array, matching Ubidots' real Django REST Framework-style list
response shape (`count`/`next`/`previous`/`results`). Pagination sends
`page_size=<page_size>&page=<n>` and stops on a short page (fewer than `page_size` records),
matching legacy's `harvest` loop exactly; `page_size` defaults to 100 (legacy's `defaultPageSize`)
and is a fixed bundle-authored value (see Known limits).

No stream declares `x-cursor-field`: legacy's own `streams()` catalog declares no `CursorFields` for
any of these four streams either (`ubidots.go:212-220`), so there is nothing to reproduce — this is
not a scope narrowing, it mirrors legacy's own manifest exactly. `created_at` is still projected as
a schema property (legacy's `standardRecord` always maps it via `first(item, "created_at",
"createdAt")`, preferring `created_at` — Ubidots' actual REST API returns `created_at` natively in
snake_case, so plain schema projection already reproduces the common case field-for-field; the
`createdAt` camelCase fallback exists in legacy purely as defensive handling for a shape variant this
bundle does not additionally model, since Ubidots' documented API never emits it — see Known limits).

## Write actions & risks

None. This is a read-only connector; `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`page_size`/`max_pages` are not exposed as runtime config.** Legacy accepts `config["page_size"]`
  (1-1000, default 100) and `config["max_pages"]` (default 1; `"all"`/`"unlimited"` for unbounded)
  as caller-overridable values. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are plain
  JSON integers fixed at bundle-authoring time in `streams.json`'s `base.pagination` block — there is
  no templated/config-driven override mechanism for either field. Declaring `page_size`/`max_pages`
  as `spec.json` properties that no template in the bundle ever consumes would be dead config (F6,
  REVIEW.md; see also searxng's identical precedent), so neither is declared. This bundle bakes in
  legacy's own DEFAULT values instead: `page_size: 100`, `max_pages: 1` — reproducing the exact
  behavior a caller who never overrides either config key already gets from legacy.
- **`createdAt` (camelCase) fallback is not modeled.** Legacy's `first(item, "created_at",
  "createdAt")` defensively tolerates a record whose `created_at` key is absent by falling back to a
  camelCase `createdAt` key instead. This bundle's schema projection copies `created_at` by exact key
  match only (Ubidots' documented API always emits snake_case `created_at`, matching legacy's own
  primary/preferred key), so the defensive camelCase fallback path is not reproduced; the
  `computed_fields` dialect has no "try key A, then key B" fallback-chain primitive (only a single
  reference per field), and legacy's own precedence prefers `created_at` first in all real traffic,
  so this narrowing has no observed effect against Ubidots' actual wire shape.
