# Overview

Miro is a wave2 fan-out declarative-HTTP migration. It reads Miro boards, board members, items,
tags, and connectors through the Miro REST API v2 (`GET https://api.miro.com/v2/...`). This bundle
targets capability parity with `internal/connectors/miro` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Miro REST API v2 access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`, `auth.mode: bearer`), matching legacy's
`connsdk.Bearer(secret)` (`miro.go:224`); the secret is never logged. `base_url` defaults to
`https://api.miro.com` and may be overridden for tests/proxies (legacy's own `miroBaseURL`
validates scheme+host the same way; the engine's base-URL resolution has no equivalent runtime
validation, but every fixture only ever points at a replay server, so this is not exercised
differently on either side).

## Streams notes

`boards` is the only non-scoped stream (`GET /v2/boards`); the other four (`board_users`,
`board_items`, `board_tags`, `board_connectors`) are nested under `/v2/boards/{board_id}/...` and
require the `board_id` config value (an absent `board_id` hard-errors on both sides: legacy's own
`"miro stream %q requires config board_id"`; the engine's unresolved `config.board_id` path-template
key — same failure classification, different literal text, per conventions.md §5's precedent).
Every stream's records live at the top-level `data` array (Miro's `{data:[...], total, size, offset,
limit}` envelope) and use Miro's offset/limit pagination (`pagination.type: offset_limit`,
`limit_param: limit`, `offset_param: offset`) — a page shorter than the declared page size stops
pagination, matching legacy's own short-page stop rule (`miro.go:171-174`) exactly.

Board-scoped streams stamp `board_id` onto every emitted record via a `computed_fields`
`{{ config.board_id }}` reference, matching legacy's own `item["board_id"] = boardID` mutation
before `mapRecord` runs (`miro.go:164-166`). `boards` renames Miro's camelCase wire fields
(`viewLink`/`createdAt`/`modifiedAt`) to the schema's snake_case names and reaches into the nested
`owner`/`team` objects for `owner_id`/`team_id` via `computed_fields` (`{{ record.owner.id }}` /
`{{ record.team.id }}`), matching legacy's `miroBoardRecord` exactly. `board_items` renames
`createdAt`/`modifiedAt`; `board_tags` renames `fillColor` to `fill_color`. None of the five streams
expose an incremental cursor field in Miro's REST API v2 (legacy's own `miroStreams()` declares no
`CursorFields` for any stream), so every stream is full-refresh only, matching legacy exactly.

## Write actions & risks

None. Miro's REST API v2 only supports full-refresh reads for the streams modeled here (legacy's
own package doc: "Miro's REST API v2 only supports full-refresh reads, so the connector is
read-only"); `capabilities.write` is `false` and this bundle ships no `writes.json`, matching
legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable, and is fixed small (2) rather than at legacy's default
  (50).** Legacy exposes `page_size` as a config-driven override (`miroPageSize`, `miro.go:257-270`,
  default 50, max 50). The engine's `offset_limit` paginator's `PageSize` is a static bundle-authored
  int (not templated), so there is no way to expose it as a config override; `page_size` is
  therefore not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key
  is worse than an absent one). It is set to `2` (not legacy's default of 50) specifically so the
  mandatory 2-page conformance fixture (`fixtures/streams/boards/{page_1,page_2}.json`) is realistic
  to author and honestly exercises the short-page stop rule (`conformance`'s `pagination_terminates`
  check requires the replay server to serve exactly one request per fixture page — a `page_size` of
  50 against a small hand-authored fixture would never reach a genuine short page), matching
  callrail's/bamboo-hr's identical documented precedent (`docs/migration/conventions.md`). This
  changes the real per-page record count from legacy's 50 to 2 — a REST-shape difference (more,
  smaller requests), never a data-emission difference (every record is still read exactly once,
  across more, smaller pages).
- **`max_pages` is not runtime-configurable.** Legacy exposes `max_pages` as a config-driven hard
  request-count cap override (`miroMaxPages`, `miro.go:272-285`, default 0/unbounded). The engine's
  `PaginationSpec.MaxPages` is a static bundle-authored int (not templated) — there is no
  config-driven knob to wire it to, so it is left unset (unbounded), matching legacy's own default
  behavior (`max_pages=0`/`all`/`unlimited`); a caller that relied on legacy's config-driven override
  to bound page count has no equivalent here, a documented scope narrowing, not a data difference.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) synthesizes
  deterministic records directly in Go rather than exercising `harvest`/`mapRecord` at all
  (`miro.go:183-208`). This bundle's schemas and fixtures target the LIVE `harvest`/`mapRecord` path
  only; the engine's own conformance/fixture-replay harness (`internal/connectors/conformance`)
  provides the credential-free test affordance legacy's fixture mode existed for, so no fixture-mode
  equivalent is needed here.
