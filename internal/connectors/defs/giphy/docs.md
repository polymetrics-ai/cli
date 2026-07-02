# Overview

Giphy is a read-only declarative-HTTP connector migrated from `internal/connectors/giphy` (legacy
wave2 fan-out). It reads GIFs, stickers, and clips from the Giphy search and trending REST
endpoints. This bundle is capability-parity with the legacy hand-written connector; the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Giphy API key via the `api_key` secret; it is sent as the `api_key` query parameter on
every request (`auth: [{"mode": "api_key_query", "param": "api_key", ...}]`) and is never logged.
`base_url` defaults to `https://api.giphy.com/v1` and may be overridden for tests or proxies.

## Streams notes

4 streams: `gif_search` (`GET /gifs/search`), `sticker_search` (`GET /stickers/search`),
`clip_search` (`GET /clips/search`), and `trending_gifs` (`GET /gifs/trending`). All 4 share the
same record shape (Giphy's media object) and primary key `["id"]`; records live at `data` in every
response. Pagination is `offset_limit` (`offset`/`limit` query params, default page size 25) —
Giphy's real stop signal combines a short page AND a `pagination.total_count` bound, but the
engine's `offset_limit` paginator's short-page stop rule (fewer than `page_size` records returned)
is the exact SAME primary check legacy's own `harvest` loop applies first (legacy checks the short
page before consulting `total_count`), so no behavior is lost porting to the declarative paginator.

The 3 search streams each require a non-empty search query: `gif_search` reads `query_for_gif`,
`sticker_search` reads `query_for_stickers`, `clip_search` reads `query_for_clips` — matching
legacy's per-stream `queryConfigKey`. `trending_gifs` requires no query. An optional `rating`
config value (content rating filter: y/g/pg/pg-13/r) is sent as a `rating` query param on every
stream's request when set (`omit_when_absent: true`), omitted entirely otherwise — matching
legacy's conditional `if rating := ...; rating != "" { base.Set("rating", rating) }`.

Legacy additionally falls back to a generic `query` config key when a stream-specific query key
(`query_for_gif`/`query_for_stickers`/`query_for_clips`) is unset. The engine's templating dialect
resolves exactly one config key per field with no fallback-chain primitive, so this bundle declares
only the stream-specific keys (legacy's primary/first-checked key per stream) in `spec.json`; the
generic `query` fallback alias is dropped. See Known limits.

## Write actions & risks

None. The Giphy API is a read-only search/trending source with no sensible reverse-ETL target in
legacy (`capabilities.write: false`, matching exactly); there is no `writes.json`.

## Known limits

- Only the 4 legacy-parity read streams are implemented; other Giphy endpoints (get-by-id, random,
  translate, categories, stickers/clips trending) are out of scope for this migration wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- The legacy generic `query` config alias (fallback when a stream-specific query key is unset) is
  dropped; only the stream-specific keys (`query_for_gif`/`query_for_stickers`/`query_for_clips`)
  are declared. ACCEPTABLE per the parity-deviation meta-rule: never changes behavior for any
  caller using the stream-specific key, only removes an alternate generic-key fallback.
- `gif_search`/`sticker_search`/`clip_search` have no `required` enforcement on their query config
  key at the `spec.json` level (legacy raises a runtime error only at Read time if the resolved
  query is empty, not at config-validation time) — the engine's `Interpolate` similarly hard-errors
  at read time when the referenced `config.query_for_*` key is entirely absent from the caller's
  RuntimeConfig, matching legacy's runtime (not upfront) enforcement point.
