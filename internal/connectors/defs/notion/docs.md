# Overview

Notion is a Tier-2 hooks migration of the quarantined `notion` legacy connector
(`docs/migration/quarantine.json`'s `notion` entry, `blocker_type: ENGINE_GAP`). It reads Notion
databases, pages, and users through the Notion REST API, read-only. This bundle is parity-tested
against `internal/connectors/notion` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Authentication is a plain Bearer integration token — this is fully expressible by the engine's
declarative `bearer` auth mode, no `AuthHook` needed. Provide the token via the `token` secret;
`streams.json`'s `base.headers` also sets the required `Notion-Version: 2022-06-28` header on
every request (the Notion API rejects requests without it, matching legacy's
`notionAPIVersion` constant).

## Streams notes

Three streams: `databases` and `pages` (both `POST /search`, distinguished only by an
`object: database`/`object: page` request-body filter) and `users` (`GET /users`). Every list
response is `{results:[...], next_cursor:string|null, has_more:bool}`.

**Pagination — Tier-2 StreamHook, not declarative (matches the quarantine's ENGINE_GAP reason
verbatim)**: `databases`/`pages` page via `POST /search` with the cursor (`start_cursor`) and
object filter injected into the JSON **request body** on every page. The engine's declarative read
path (`engine/read.go`'s `readDeclarative`) always issues `rt.Requester.Do(ctx, method, path,
query, nil)` — the body argument is hardcoded `nil` on every declarative read; `StreamSpec.Body`
(`engine/bundle.go`'s `json:"body,omitempty"` field, commented "POST-body streams") is declared but
never read anywhere in `read.go` — dead/unwired. This is the identical gap
`docs/migration/quarantine.json` documents for `notion`. `hooks/notion/hooks.go` implements
`StreamHook`, porting legacy's `harvest` loop verbatim: for `databases`/`pages`, a POST body
`{"page_size": N, "filter": {"property": "object", "value": "database"|"page"}, "start_cursor":
"<cursor>" (omitted on the first page)}`; for `users`, a plain `GET` with `page_size`/`start_cursor`
as query parameters (this shape alone IS independently declarative-expressible via the engine's
`cursor` pagination type, but is routed through the same hook as `databases`/`pages` for one
consistent dispatch path across the bundle — seed value: keeping all three streams' pagination
logic in one reviewable place rather than splitting `users` onto a declarative path while
`databases`/`pages` stay hook-driven). Every page's stop condition is `has_more != true` (compared
as the literal string `"true"` after JSON decode) OR an empty `next_cursor`, matching legacy
exactly.

Every stream in this bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason":
"..."}` marker (`docs/migration/conventions.md` §4/§6): `internal/connectors/conformance/dynamic.go`
honors this by Skipping every dynamic fixture-replay check for these streams, since the StreamHook
(always `handled=true`) is what every real `Read()` call actually dispatches through, and a
declarative-only fixture replay cannot exercise a POST-body cursor at all. The authoritative
substitute this marker names is `paritytest/notion`'s dedicated 2-page tests
(`TestParityNotion_DatabasesSearchBodyCursorPagination`,
`TestParityNotion_UsersQueryCursorPagination`) and `hooks/notion/hooks_test.go`. `streams.json`'s
own `base.pagination` stays declared `{"type": "none"}` (a single, honest request) since it is
never dynamically exercised now.

`databases`/`pages` declare `incremental.cursor_field: last_edited_time` (matching legacy's
published `CursorFields`); `users` declares no `incremental` block (legacy publishes no
`CursorFields` for it either — Notion users have no edit timestamp).

## Write actions & risks

None. Notion is read-only in legacy (`Write` returns `connectors.ErrUnsupportedOperation`,
`notion.go:336-338`); `capabilities.write` is `false` and no `writes.json` is declared.

## Known limits

- Full Notion API surface (block children, comments, page/database mutation, search-by-title) is
  out of scope for this migration; see `api_surface.json`'s `excluded` entries. Only the 3
  legacy-parity read streams are implemented.
- Pagination is a Tier-2 `StreamHook` (`hooks/notion/hooks.go`), not declarative — see "Streams
  notes" above. Candidate future engine feature: wiring `StreamSpec.Body` into the declarative read
  path (a POST-body-with-templated-cursor primitive) would let `databases`/`pages` drop the hook;
  not implemented in this phase per the `ENGINE_GAP` recurrence rule (`conventions.md` §6) — this is
  the same gap several other quarantined connectors (chift, monday, canny, plaid, pocket) hit.
- `streams.json`'s declared `base.pagination: {"type": "none"}` is not the real production
  pagination behavior; every stream's `conformance.skip_dynamic` marker documents why (see
  "Streams notes").
- `max_pages` is consumed only by `hooks/notion/hooks.go`'s `StreamHook` (mirroring legacy's
  `notionMaxPages`), not by a declarative `PaginationSpec.MaxPages` field, since pagination here is
  entirely hook-driven. Declared in `spec.json` so it is not dead config.
- `spec.json`'s `token` secret intentionally has a single canonical name (unlike legacy, which
  accepts `credentials.access_token`/`access_token`/`token` as aliases for the same value) — the
  multi-alias fallback was a legacy secrets-plumbing convenience with no bearing on the emitted
  record data; an operator migrating a legacy config supplies the same token value under this
  bundle's one canonical `token` secret key.
