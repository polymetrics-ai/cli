# Overview

Huntr is a wave2 fan-out declarative-HTTP migration. It reads Huntr organization members,
candidates, activities, notes, and actions through the Huntr organization REST API
(`GET https://api.huntr.co/org/...`). This bundle is a capability-parity port of the hand-written
connector at `internal/connectors/huntr` (`huntr.go`/`streams.go`), which stays registered and
unchanged until wave6's registry flip. The connector's `docs_url` in the wave2 migration manifest
is `"manual intervention needed"` (no reachable public API reference on file); legacy source is
the authoritative ground truth for this bundle per conventions.md's "legacy is ground truth over
any doc" rule, and every stream/field/pagination shape below is derived directly from
`huntr.go`/`streams.go`, not from external documentation.

## Auth setup

Provide a Huntr organization API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`huntr.go:234`). `base_url` defaults to `https://api.huntr.co/org` and
may be overridden for tests/proxies (legacy's own `huntrBaseURL` validates scheme+host the same
way; the engine's base-URL resolution has no equivalent runtime validation, but every
parity/conformance fixture only ever points at an httptest server, so this is not exercised
differently on either side).

## Streams notes

All five streams (`members`, `candidates`, `activities`, `notes`, `actions`) are simple list
endpoints; records live at the `data` key on every response (`connsdk.RecordsAt(resp.Body,
"data")`, `huntr.go:151`). `notes` reads from `/notes/members` (legacy's own routing table,
`streams.go:22`), not `/notes` — this bundle's `path` matches that exactly.

Pagination is Huntr's own next-cursor convention: the `next` query parameter carries the token
read from the response body's own `next` field (`pagination.type: cursor`, `cursor_param: next`,
`token_path: next`), matching legacy's `harvest` loop (`huntr.go:136-175`): request with `limit`
(and `next=<token>` once a token exists), read `data[]`, then read the body's `next` field for the
next token. Legacy stops on an empty, `null`, or the literal string `"inf"` cursor, or when the
token repeats the immediately-prior one (`huntr.go:169`). The engine's `tokenPathCursor` paginator
stops on an empty/absent token (a JSON `null` at `next` reads back as `""` via `connsdk.StringAt`,
so the `null`-stop case is covered identically) and loop-guards against ANY token repeating across
the whole page history (`p.seen` — every previously-seen token, not just the immediately-prior
one), a strictly stronger guard than legacy's single-previous-token check, never a weaker one.

`limit=100` is declared as a static per-stream `query` value (`streams.json`'s `"query":
{"limit": "100"}`), matching legacy's `huntrDefaultPageSize` (`huntr.go:30`) sent via
`query.Set("limit", strconv.Itoa(pageSize))` on every page (`huntr.go:143`) — the engine's
declarative `stream.Query` is merged into every page request identically (`engine/read.go`'s
`mergeQuery`), so `limit=100` is re-sent on every page on both sides.

## Write actions & risks

None. Huntr's organization API supports only `full_refresh` reads (legacy's own package doc: "no
safe reverse-ETL writes... read-only"); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The literal `"inf"` cursor-stop value is not separately modeled.** Legacy treats the literal
  string `"inf"` as an additional stop signal alongside empty/null (`huntr.go:169`); no Huntr
  response has ever been observed to actually emit `"inf"` as a real page-token value (it reads as
  a defensive extra case in legacy, not a confirmed real wire behavior). The engine's `stop_path`
  mechanism only recognizes a designated body path's `"true"`/falsy value, not an arbitrary
  sentinel token string, so `"inf"` cannot be wired as a stop condition without inventing a new
  engine primitive. If a real Huntr response ever returns the literal token `"inf"`, this bundle's
  pagination would follow one more page than legacy before its loop-guard (`p.seen`) catches a
  repeat, or paginate until the API's own empty-page/short-page reality intervenes — a defensive
  gap, not an observed data-parity break with real legacy inputs, and no fixture ever exercises
  this path since it targets undocumented, never-seen-in-the-wild behavior. Documented here rather
  than silently dropped per §5's meta-rule.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing `req.State["cursor"]`
  when set) onto every fixture-mode record (`huntr.go:209-213`). This bundle's schemas and
  fixtures target the live `harvest` record shape only; the engine's own conformance/fixture-replay
  harness supplies the credential-free test affordance legacy's fixture mode existed for.
- **`page_size`/`max_pages` are not runtime-configurable beyond `page_size`'s spec default.**
  Legacy exposes `page_size` (1-100) and `max_pages` (integer, `all`, or `unlimited`) as
  config-driven overrides (`huntrPageSize`/`huntrMaxPages`, `huntr.go:267-295`). This bundle
  declares `page_size` in `spec.json` with a `default: 100` for documentation purposes, but the
  `cursor`/`token_path` paginator has no config-driven page-size or max-pages knob wired into
  `streams.json`'s pagination block (unlike `page_number`/`offset_limit`, which read
  `PaginationSpec.PageSize` directly) — `limit=100` is therefore a static per-stream query literal,
  not a `{{ config.page_size }}` template, matching the bitly `size=50` precedent
  (`docs/migration/conventions.md`). `max_pages` is likewise not modeled; pagination is bounded
  only by the empty/short-token stop signal, matching Huntr's own real termination behavior.
