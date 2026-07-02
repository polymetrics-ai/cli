# Overview

Intercom is a wave2 fan-out declarative-HTTP migration. It reads Intercom contacts, companies,
conversations, admins, and tags through the Intercom REST API (`https://api.intercom.io`). This
bundle is engine-vs-legacy parity-tested against `internal/connectors/intercom` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide an Intercom access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)`. `base_url` defaults to `https://api.intercom.io` and may be overridden
for tests/proxies. An optional `api_version` config value is sent as the `Intercom-Version` header;
when unset, the header is omitted entirely (not sent empty) — matching legacy's `if version :=
...; version != ""` conditional header construction, and the engine's own conditional-header
omission rule for an optional (not-`required[]`) `config.*` key.

## Streams notes

All 5 streams (`contacts`, `companies`, `conversations`, `admins`, `tags`) are `GET` list
endpoints with records at the `data` key. Pagination follows Intercom's own
`pages.next.starting_after` cursor convention (`pagination.type: cursor`, `token_path:
"pages.next.starting_after"`, `cursor_param: starting_after"`) — no `stop_path` is declared because
legacy's own stop condition is purely "the next cursor token is empty" (`harvest`'s `next ==
""` check), which is exactly the `token_path`-only cursor variant's default behavior with no
additional falsy-body-value gate needed. `admins` and `tags` return a single page with no `pages`
object at all; `StringAt`'s absent-path behavior resolves to `""`, so the paginator stops after one
request exactly like legacy's harvest loop does for those two endpoints. Every list request sends
`per_page=50` (matches legacy's default `page_size`) via each stream's static `query: {"per_page":
"50"}` (admins/tags omit it, matching legacy sending no `per_page` on those two calls either).
`contacts`/`companies`/`conversations` carry `x-cursor-field: updated_at` in their schemas purely
as descriptive metadata (matching legacy's `CursorFields` stream-catalog field) — no
`incremental` block is declared for any stream because legacy never actually sends a
server-side incremental filter parameter; `Read` is always a full harvest of every available
record, exactly as legacy's own `Read` does (its `InitialState`/cursor plumbing is present but
unused by any filter).

## Write actions & risks

None. Intercom is exposed as a read-only source here; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable in the same way legacy allowed.** Legacy exposes
  `page_size` (default 50, max 150) as a config-driven per-request override
  (`intercomPageSize`/`intercomMaxPageSize`). The engine's `cursor` paginator has no
  config-driven page-size knob analogous to `page_number`/`offset_limit`'s `PageSize` field (only
  the static per-stream `query` value controls what is actually sent), so this bundle sends
  Intercom's own default (`per_page=50`) as a fixed literal, matching stripe's `limit=100`
  static-query precedent. `spec.json` still declares `page_size` (default `"50"`) for
  documentation/informational parity with legacy's config surface, but no template in
  `streams.json` consumes it.
- **`max_pages` is not modeled.** Legacy's `intercomMaxPages` config-driven hard page-count cap has
  no equivalent wiring in this bundle; pagination is bounded only by the empty-next-cursor stop
  signal, matching Intercom's own real termination behavior.
