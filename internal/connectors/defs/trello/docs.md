# Overview

Trello is an L-size fresh declarative-HTTP migration of `internal/connectors/trello` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Trello **boards** at full parity, plus
**lists**/**checklists** at documented-narrower parity, through the Trello REST API
(`https://api.trello.com/1`). Legacy's **cards** and **actions** streams are blocked — see Known
limits. Read-only: legacy's `Write` returns `connectors.ErrUnsupportedOperation` and there is no
allow-listed reverse-ETL write surface for Trello mutations.

## Auth setup

Provide a Trello API key via the `key` secret and a Trello API token via the `token` secret; both
are required (`spec.json`'s `required`). Trello authenticates every request via TWO query
parameters sent together (`?key=...&token=...`), matching legacy's `keyTokenAuth`
(`trello.go:362-370`). `base.auth` declares `api_key_query` for `key` (the only query-param auth
mode the dialect's `auth` block supports, and it carries exactly one param); `token` is instead
wired as a per-stream/check `query` entry (`"token": "{{ secrets.token }}"`) on every declarative
request path that goes through the engine's normal per-stream query resolution (`check`, `boards`,
and the per-board sub-requests `fan_out` drives for `lists`/`checklists`). Neither `key` nor
`token` is ever logged (`x-secret: true` on both).

`base_url` defaults to `https://api.trello.com/1` and may be overridden for tests/proxies.

## Streams notes

**`boards`** (full parity): `GET /members/me/boards`, records at the bare top-level array (`records.path:
""`), primary key `["id"]`. Every emitted field (`id`, `name`, `desc`, `closed`, `idOrganization`,
`url`, `shortUrl`, `dateLastActivity`) matches legacy's `trelloBoardRecord` mapper
(`streams.go:138-149`) field-for-field, no rename needed since the raw API field names are used
verbatim by legacy's own mapper. No pagination — Trello's `/members/me/boards` returns every
accessible board in one response (`pagination.type: none`), matching legacy's own single
unpaginated call in the discovery path (`trello.go:175-191`).

**`lists`** and **`checklists`** (board_ids-scoped fan_out; auto-discovery narrowed — see Known
limits): both are board-scoped sub-resources, unpaginated per board (`GET
/boards/{id}/lists`/`/boards/{id}/checklists`), matching legacy's `readBoardResource`
non-paginated branch (`trello.go:236-252`). Each declares `fan_out.ids_from.config_key: board_ids`
(the config's comma-separated board ID list, `into.path_var: board_id` referenced in the stream's
own `path` as `{{ fanout.id }}`, `stamp_field: idBoard` writing the current board id onto every
emitted record after projection). `lists` emits `id`/`name`/`closed`/`idBoard`/`pos`/`subscribed`
(`trelloListRecord`, `streams.go:151-160`); `checklists` emits
`id`/`name`/`idBoard`/`idCard`/`pos` (`trelloChecklistRecord`, `streams.go:178-186`) — both
field-for-field matches of their legacy mappers.

No incremental cursor is modeled for any stream. Legacy's catalog declares `dateLastActivity` as a
`CursorFields` hint only on `cards` (a blocked stream, see below); `boards`/`lists`/`checklists`
have no `CursorFields` in legacy's own catalog either, so this is exact parity, not a narrowing.

## Write actions & risks

None. Legacy `trello` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`,
`trello.go:105-107`); `metadata.json` declares `capabilities.write: false` and this bundle ships no
`writes.json`.

## Known limits

- **`cards` and `actions` are BLOCKED (`ENGINE_GAP`), not implemented in this bundle.** Both are
  board-scoped, id-cursor-paginated sub-resources (Trello's `before`-param, last-record-id
  convention: `readBoardResource`'s paginated branch, `trello.go:255-289`) that legacy also
  auto-discovers boards for by default (no `board_ids` configured). Modeling this at Tier-1 would
  need `fan_out.ids_from.request` (auto-discover boards via `GET /members/me/boards`) on a stream
  that ALSO needs `pagination.type: cursor` + `last_record_field: id` for its own per-board
  sub-fetch — but `fan_out.ids_from.request`'s id-listing request reuses the SAME stream-level
  `PaginationSpec` as the per-board sub-fetch (`bundle.go`'s `FanOutIDsRequest` doc comment: "it
  reuses the child stream's [own effective pagination spec]"; `read.go`'s `fanOutIDsFromRequest`
  and `readOneSequence` both read `stream.Pagination`/`b.HTTP.Pagination` identically). Two
  independent problems follow from that reuse: (1) `fanOutIDsFromRequest` sends only
  `page.Query` (the paginator's own query), never `stream.Query` — so Trello's second required
  auth query param (`token`, wired via `stream.Query` since `auth` only carries one query param)
  would be MISSING from the id-listing request, which would 401; and (2) even with auth fixed,
  `lastRecordCursor` (the `cursor`+`last_record_field` paginator, `paginate.go:206-241`) has no
  short-page stop and no loop guard (unlike `tokenPathCursor`/`nextURL`/`linkHeaderPaginator`) — a
  board-listing endpoint that ignores the `before` param entirely (Trello's `/members/me/boards`
  does not support it) would return the same non-empty board list forever, so the id-listing call
  would never terminate. Both are genuine Tier-1 dialect gaps, not workaroundable without Go (a
  `StreamHook`, the sanctioned Tier-2 escape hatch this migration was scoped not to use — see
  `docs/migration/conventions.md`'s explicit "sub-resource fan-out (issue → comments per issue)"
  `StreamHook` trigger, listing exactly this shape). Filed as `ENGINE_GAP` blockers (see this
  migration's reported `blockers[]`); legacy `internal/connectors/trello` remains the only
  implementation of `cards`/`actions` until the engine's `fan_out` dialect supports a
  differently-paginated id-listing request, or a `StreamHook` is authorized for this connector.
- **`lists`/`checklists` require `board_ids` to be explicitly configured — the auto-discovery
  fallback legacy provides by default (no `board_ids` set: discover every accessible board, then
  fetch that board's lists/checklists, `trello.go:216-228`'s `boardIDs`) is NOT modeled.** This is
  the SAME underlying engine limitation as the cards/actions blocker: `fan_out.ids_from.request`
  auto-discovery cannot be used for ANY stream needing the `token` query-param secret on its
  id-listing sub-request (see above) — but unlike cards/actions, `lists`/`checklists` need no
  pagination on their own per-board fetch, so `fan_out.ids_from.config_key: board_ids` (no HTTP
  request for id resolution at all, so the auth gap never applies) is a correct, fully-working
  Tier-1 substitute PROVIDED `board_ids` is set. An unconfigured `board_ids` now yields an empty
  `lists`/`checklists` read (no error — `resolveFanOutIDs`'s `config_key` case returns a nil id
  list for an empty/unset config value) rather than legacy's automatic full-board-set discovery.
  This changes emitted data for that specific legacy-accepted input (unset `board_ids`), so it is
  reported as a deviation, not silently absorbed; there is no security/tenancy concern either way
  (Trello's `key`+`token` auth is member-scoped, not board-scoped — `board_ids` was always a
  convenience filter over boards the same credentials already access, never an access-control
  boundary), and `boards` itself is completely unaffected (still auto-discovers every accessible
  board with no configuration required).
- **The `boards` stream does not honor `board_ids` for its OWN output.** Legacy's `boards` stream
  itself also branches on `board_ids` (`readBoards`/`eachBoardObject`, `trello.go:162-212`): when
  set, it fetches only those specific boards via per-id `GET /boards/{id}`; when unset, it lists
  every accessible board. This bundle's `boards` stream always calls `GET /members/me/boards`
  (the discovery path), regardless of `board_ids`. Modeling the `board_ids`-set branch would need
  the exact same "config_key-with-a-request-mode-fallback" conditional `fan_out` cannot express
  (fan_out is declared as exactly one of `config_key`/`request`, never a runtime choice between
  them) — the common/default case (no `board_ids`) is preserved exactly; only the narrower
  explicit-scoping case changes shape (a strict, always-larger superset of legacy's expected
  boards for that input, never missing data, and no security boundary crossed for the same reason
  given above).
- `dateLastActivity`/`pos`/similar Trello wire values are copied through in their real JSON types
  (string timestamps as `["string","null"]`, `pos` as `["number","null"]`) — no schema widening was
  needed since plain schema projection reproduces legacy's inline `connectors.Record{...}`
  construction field-for-field for every implemented stream.
