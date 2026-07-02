# Overview

Easypromos is a wave2 fan-out declarative-HTTP migration. It reads Easypromos promotions,
organizing brands, stages, users, participations, and prizes through the Easypromos REST API
(`GET https://api.easypromosapp.com/v2/...`). This bundle migrates
`internal/connectors/easypromos` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. Easypromos is read-only: legacy exposes no reverse-ETL
writes, so `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide an Easypromos API bearer token via the `bearer_token` secret; it is sent as
`Authorization: Bearer <bearer_token>` and never logged, matching legacy's
`connsdk.Bearer(secret)` (`easypromos.go`'s `requester`). `base_url` defaults to
`https://api.easypromosapp.com/v2` and may be overridden for tests/proxies.

## Streams notes

`promotions` and `organizing_brands` are account-level lists (`GET /promotions`,
`GET /organizing_brands`); the remaining 4 streams (`stages`, `users`, `participations`, `prizes`)
are scoped to a single promotion via the `promotion_id` config value, substituted into their
`/{resource}/{promotion_id}` path (`{{ config.promotion_id }}`, urlencoded by
`InterpolatePath`'s per-segment default) — matching legacy's `perPromotion` routing exactly
(`easypromos.go`'s `Read`: `path = endpoint.resource + "/" + url.PathEscape(promotionID)`). An
absent `promotion_id` hard-errors identically on both sides (legacy: `"easypromos stream %q
requires config promotion_id"`; engine: an unresolved `config.promotion_id` path-template key —
same failure classification, different literal text, per conventions.md §5's precedent for
config-validation parity).

All 6 streams share Easypromos's cursor pagination convention: list responses carry
`{"items":[...], "paging":{"next_cursor":"..."|null}}`; the next page is requested with
`next_cursor=<token>` until `paging.next_cursor` is null (`pagination.type: cursor`,
`cursor_param: next_cursor`, `token_path: paging.next_cursor`), matching legacy's `harvest`
function exactly. Records live at the `items` key on every stream.

`promotions` and `prizes` require `computed_fields` renames for their nested reference objects
(`organizing_brand.id`/`.name` -> `organizing_brand_id`/`organizing_brand_name`;
`prize_type.id`/`.name` -> `prize_type_id`/`prize_type_name`), matching legacy's inline
`item["organizing_brand"].(map[string]any)` / `item["prize_type"].(map[string]any)` flattening in
`easypromosPromotionRecord`/`easypromosPrizeRecord`. The other 4 streams' raw field names already
match this bundle's snake_case schema property names exactly (Easypromos's own wire shape is
snake_case, unlike e.g. e-conomic's camelCase), so no rename is needed for them beyond the
identity pass-through schema projection performs.

## Write actions & risks

None. Easypromos is a read-only source in legacy (`easypromos.go`'s package doc: "the upstream
source supports only full_refresh syncs and exposes no safe reverse-ETL writes");
`capabilities.write` is `false` and no `writes.json` is shipped.

## Known limits

- **No incremental cursor.** Legacy exposes no incremental cursor field for any Easypromos stream
  (`easypromosStreams()` declares no `CursorFields` anywhere: "The API only supports full_refresh,
  so no cursor fields are published") — every stream is full-refresh only in both legacy and this
  bundle; no `incremental` block is declared on any stream, matching legacy exactly.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps deterministic placeholder records across
  every field this connector might ever emit (a superset shared by all 6 streams' mappers), which
  is not the live wire shape any single stream actually returns. This bundle's schemas and
  fixtures target the LIVE record shape only (`easypromos.go`'s `harvest`/`mapRecord` functions),
  per the bitly-pilot precedent (`docs/migration/conventions.md`'s worked example): the engine's
  own fixture-replay conformance harness supersedes the need for an in-connector fixture-mode
  branch.
