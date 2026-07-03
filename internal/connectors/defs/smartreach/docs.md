# Overview

SmartReach is a wave2 fan-out declarative-HTTP migration. It reads SmartReach teams, campaigns,
prospects, email settings, and do-not-contact records through the SmartReach v3 API (`GET
https://api.smartreach.io/api/v3/...`). This bundle migrates `internal/connectors/smartreach`
(the hand-written connector); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a SmartReach API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-API-KEY", key, "")`
(`smartreach.go:143`). `base_url` defaults to `https://api.smartreach.io/api/v3` and may be
overridden for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

All five streams (`campaigns`, `prospects`, `teams`, `email_settings`, `do_not_contact`) hit
their own `GET <resource>` endpoint (`campaigns`, `prospects`, `teams`, `email-settings`,
`do-not-contact`), matching legacy's `streamEndpoints` map exactly. Records live at
`campaigns`/`prospects`/`teams`/`email_settings`/`dnc` respectively, matching legacy's
`recordsPath` per stream. None of the streams paginate in legacy (a single `r.Do` call per read,
no loop) — `pagination.type: none` is declared, one request per read. `team_id` is applied as an
optional query filter to every stream EXCEPT `teams` (matching legacy's `if stream != "teams"`
guard — `teams` itself is not team-scoped); `older_than`/`newer_than` are optional passthrough
query filters applied to every stream, omitted entirely when unset, matching legacy's
`copyConfig`. Every stream declares `projection: "passthrough"`: legacy's `readRecords` emits
each decoded record verbatim (`emit(connectors.Record(rec))`, `smartreach.go:115`, no
field-building or `mapRecord` step), so this bundle emits every raw field the API returns rather
than narrowing to the `id`/`name`/`created_at` triple `streams()`'s catalog happens to declare —
schema-mode projection on a verbatim-emitting legacy would silently drop real API fields
(`conventions.md` §8 rule 1). The `id`/`name`/`created_at` properties in `schemas/*.json` remain
the documented, guaranteed-present fields; they are a floor, not a ceiling, on what a record
contains.

## Write actions & risks

None. SmartReach's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **The `teamid` config-key alias is not modeled; only `team_id` is accepted.** Legacy accepts
  EITHER `teamid` OR `team_id` as the config key naming the team filter, trying `teamid` first
  and falling back to `team_id` (`firstConfig(cfg, "teamid", "team_id")`, `smartreach.go:149-151,
  180-187`) — both aliases map to the same `team_id` query parameter. The engine's
  `stream.Query` dialect (`docs/migration/conventions.md` §3) resolves each output query key from
  exactly one template reference; there is no declarative "first non-empty of two config keys"
  primitive (unlike `base.auth`'s first-match-wins candidate list, which exists only for
  authentication, not query params). This bundle therefore declares only `team_id` as the
  accepted config key and does not declare `teamid` at all (a declared-but-unwireable config key
  is worse than an absent one, per the searxng/smaily precedent). An operator who previously
  configured only `teamid` must switch to `team_id`; this is a documented config-surface
  narrowing (one accepted input alias dropped), not a data-shape change — the emitted records for
  a given team are byte-identical either way once `team_id` resolves to the same value.
