# Overview

Opsgenie is a wave2 fan-out declarative-HTTP migration. It reads Opsgenie alerts, incidents, users,
teams, and services through the Opsgenie REST API (`GET https://api.opsgenie.com/v2/...`). This
bundle migrates `internal/connectors/opsgenie` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Opsgenie API integration token via the `api_token` secret; it is sent as
`Authorization: GenieKey <api_token>` (`api_key_header` mode, `header: "Authorization"`,
`prefix: "GenieKey "`), matching legacy's `connsdk.APIKeyHeader("Authorization", secret,
"GenieKey ")` (`opsgenie.go:225`) exactly. Never logged. `base_url` defaults to
`https://api.opsgenie.com/v2` and may be overridden for tests/proxies.

## Streams notes

All 5 streams (`alerts`, `incidents`, `users`, `teams`, `services`) share an identical shape:
records live under the top-level `data` array, and pagination follows Opsgenie's own
`paging.next` **absolute** URL convention (`pagination.type: next_url`, `next_url_path:
"paging.next"`), matching legacy's `harvest` loop exactly (`opsgenie.go:131-169`): the next page is
requested verbatim at the URL Opsgenie returns, which already carries its own advancing
`offset`/`limit` query parameters.

Each stream declares a static `limit=100` per-stream query value (Opsgenie's own default page
size, matching legacy's `opsgenieDefaultPageSize`) sent on the FIRST request only in effect: the
engine's `next_url` paginator returns an empty `page.Query` on every subsequent page
(`engine/paginate.go`'s `nextURL.Next`), so `stream.Query`'s `limit=100` is re-merged on top of
every page's own query — but since Opsgening's own `paging.next` URL always carries its own
`limit` value matching the one this bundle requested, the re-apply is idempotent (identical to
Bitly's documented `size=50` precedent, `docs/migration/conventions.md`). Legacy also sends an
explicit `offset=0` on the first request; this bundle does not declare a static `offset` (an
`offset` value is NOT constant across pages, unlike `limit` — declaring it statically would
incorrectly re-apply `offset=0` on every subsequent page, overwriting the advancing value the
`paging.next` URL itself carries). Opsgenie's API defaults an absent `offset` to `0`, so omitting it
on the first request is behaviorally identical to legacy's explicit `offset=0` and does not change
which records are returned or their order.

Raw API field names differing from this bundle's snake_case schema fields (`tinyId`, `createdAt`,
`updatedAt`, `lastOccurredAt`, `ownerTeam`, `impactedServices`, `fullName`, `timeZone`, `teamId`)
are renamed via `computed_fields`, each a bare single `{{ record.<path> }}` reference so the
engine's typed extraction preserves each field's real JSON type (objects/arrays/booleans survive
untouched — `role`, `details`, `tags`, `responders`, `blocked`, `verified`), matching legacy's raw
passthrough exactly (`streams.go`'s `opsgenieAlertRecord`/etc.).

None of the 5 streams expose a genuine server-side incremental filter parameter in legacy (legacy
declares `CursorFields: []string{"created_at"}` on `alerts`/`incidents` for catalog metadata only —
`InitialState` seeds an empty cursor but `harvest` never reads or filters by it); this bundle
mirrors that: `x-cursor-field: created_at` is declared on `alerts`/`incidents`' schemas for the same
catalog-metadata purpose, but no `incremental` block is declared for any stream, matching legacy's
full-refresh-only read behavior exactly.

## Write actions & risks

None. Opsgenie's alert/incident mutation endpoints (create, acknowledge, close, escalate) are not
modeled; `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size`/`max_pages`
  as config-driven overrides (`opsgeniePageSize`/`opsgenieMaxPages`, `opsgenie.go:257-285`). The
  engine's `next_url` paginator has no config-driven page-size/max-pages knob (same bitly-documented
  limitation, `docs/migration/conventions.md`); `limit=100` is sent as a static per-stream query
  literal instead, and neither `page_size` nor `max_pages` is declared in `spec.json` (F6: a
  declared-but-unwireable config key is worse than an absent one).
- **Legacy's fixture-mode-only `previous_cursor` echo field is not modeled.** Legacy's
  `readFixture` path stamps `previous_cursor` onto every fixture-mode record when a prior state
  cursor happens to be set (`opsgenie.go:200-203`). This is not part of the LIVE record shape; this
  bundle's schemas target the live path only, per the bitly-documented precedent. The engine's own
  conformance/fixture-replay harness provides the credential-free test affordance this bundle needs.
- **Fixtures are single-page** for every stream (conventions.md §4's sanctioned `next_url`
  exception): a `next_url` stream's next-page URL is the replay server's own runtime address, which
  a static fixture file cannot embed. Each fixture's `paging` is `{}` (no `next` key), terminating
  pagination on page 1 — `pagination_terminates` is satisfied by the paginator's normal
  empty-next-value stop path, not a second recorded page.
