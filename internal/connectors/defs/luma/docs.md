# Overview

lu.ma (Luma) is a read-only declarative-HTTP connector that reads events, event guests, and event
hosts through the lu.ma public REST API (`https://api.lu.ma/public/v1`). This bundle migrates
`internal/connectors/luma` (the hand-written legacy connector, which stays registered and unchanged
until wave6's registry flip) to a Tier-1 defs bundle at capability parity.

## Auth setup

Provide a lu.ma API key via the `api_key` secret; it is sent as the `x-luma-api-key` header
(`auth.mode: api_key_header`) and is never logged. There is no fallback unauthenticated mode — lu.ma
requires the header on every request, matching legacy's hard `Check`/`requester` failure when the
secret is unset.

## Streams notes

All 3 streams share lu.ma's `{entries:[{<entryKey>:{...}}], has_more, next_cursor}` wrapper
envelope and the same cursor pagination shape (`pagination.type: cursor`, `cursor_param:
pagination_cursor`, `token_path: next_cursor`, `stop_path: has_more` — legacy's own stop rule is
`hasMore != "true" || next == ""`, and `stop_path` reproduces the `has_more`-gates-continuation half
of that check while the paginator's built-in same-token loop guard and empty-token stop reproduce
the rest). The wrapper's `entries[]` items are NOT the flat record — the actual object lives one
level deeper (`entries[].event`, `entries[].guest`, `entries[].host`), and the engine's
`RecordsSpec.Path` only selects a body path, not a per-item unwrap, so schema projection alone
cannot flatten this shape. Every field for every stream is therefore populated by a
`computed_fields` bare-reference into the nested raw item (e.g. `"api_id": "{{ record.event.api_id
}}"`), which resolves against the RAW pre-projection record — exactly reproducing legacy's
`unwrapEntry(entry, endpoint.entryKey)` + `mapRecord(item)` two-step in the declarative dialect,
with each schema property's declared type preserved via the engine's bare-single-reference typed
extraction.

- `events` (`GET /calendar/list-events`) — no config requirement.
- `event_guests` (`GET /event/get-guests`) and `event_hosts` (`GET /event/get-hosts`) both require
  the `event_api_id` config value, sent as a query param — matches legacy's hard `event_api_id`
  config requirement for these two sub-streams (an absent value is a hard interpolation error, same
  as legacy's explicit `fmt.Errorf("luma stream %q requires config event_api_id", stream)`).

lu.ma's public API supports only full-refresh syncs; no stream declares an `incremental` block,
matching legacy (`lumaStreams()` declares no `CursorFields` anywhere).

## Write actions & risks

None. lu.ma's public API has no safe reverse-ETL surface; `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy exactly.

## Known limits

- Legacy's config-driven `max_pages` override (a caller-supplied cap on page count, defaulting to
  unlimited) has no declarative equivalent: the engine's `PaginationSpec.MaxPages` is a fixed value
  baked into `streams.json`'s `base.pagination` block, not a runtime-config-driven override
  mechanism (see `docs/migration/conventions.md` §3's rate_limit/dead-config note for the same
  fixed-vs-configurable distinction applied elsewhere). This bundle therefore does not declare a
  `max_pages` spec property at all (a declared-but-unwireable key is worse than an absent one, per
  the F6 dead-config rule) and leaves pagination unbounded (`max_pages` omitted from
  `base.pagination`, i.e. unbounded) — matching legacy's own default (`lumaDefaultMaxPages = 0`,
  unlimited) for the common case, but the override itself is out of scope.
- Only the 3 legacy-parity read streams are implemented; the broader lu.ma public API (event
  creation, coupons, ticket types, `calendar/list-people`) is out of scope until Pass B — see
  `api_surface.json`.
