# Overview

SmartReach is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full
documented SmartReach v3 API surface (SmartReach's own OpenAPI 3.0 spec, embedded in its
ReadMe-hosted reference docs — see `api_surface.json`'s `docs` field). It reads SmartReach teams,
campaigns, prospects, email settings, do-not-contact records, users, and accounts, and writes
prospect/account/campaign/DNC/task mutations, through the SmartReach v3 API
(`https://api.smartreach.io/api/v3/...`). This bundle migrates `internal/connectors/smartreach`
(the hand-written connector); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a SmartReach API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-API-KEY", key, "")`
(`smartreach.go:143`). `base_url` defaults to `https://api.smartreach.io/api/v3` and may be
overridden for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

Seven streams hit their own `GET <resource>` endpoint: the 5 legacy-parity streams
(`campaigns`, `prospects`, `teams`, `email_settings`, `do_not_contact` — `campaigns`, `prospects`,
`teams`, `email-settings`, `do-not-contact`, matching legacy's `streamEndpoints` map exactly),
plus 2 Pass-B-added streams confirmed against SmartReach's own OpenAPI spec: `users` (`GET
/users`, records at `users`) and `accounts` (`GET /search/accounts`, records at `accounts` — a
CRM-style company/account object distinct from `teams`). Records live at
`campaigns`/`prospects`/`teams`/`email_settings`/`dnc`/`users`/`accounts` respectively. None of
the streams paginate (a single request per read) — `pagination.type: none` is declared.
`team_id` is applied as an optional query filter to every stream EXCEPT `teams` (matching
legacy's `if stream != "teams"` guard — `teams` itself is not team-scoped); the live API's own
OpenAPI spec documents `team_id` as REQUIRED on every one of these endpoints, but this bundle
preserves legacy's read-side optional modeling unchanged (not a Pass B read-behavior change; see
Known limits) — `older_than`/`newer_than` are optional passthrough query filters applied to
every stream, omitted entirely when unset, matching legacy's `copyConfig`. Every stream declares
`projection: "passthrough"`: legacy's `readRecords` emits each decoded record verbatim
(`emit(connectors.Record(rec))`, `smartreach.go:115`, no field-building or `mapRecord` step), so
this bundle emits every raw field the API returns rather than narrowing to the `id`/`name`/
`created_at` triple `streams()`'s catalog happens to declare — schema-mode projection on a
verbatim-emitting legacy would silently drop real API fields (`conventions.md` §8 rule 1). The
`id`/`name`/`created_at` properties in `schemas/*.json` remain the documented, guaranteed-present
fields; they are a floor, not a ceiling, on what a record contains. `users`/`accounts` schemas
were authored directly from the real OpenAPI `User`/`ProspectAccount` component schemas (not a
legacy catalog guess, since legacy never modeled these streams at all).

## Write actions & risks

Legacy's own connector is read-only, but Pass B full-surface expansion adds every
dialect-expressible mutation SmartReach's own OpenAPI spec documents (`api_surface.json`), all
JSON bodies, each requiring approval. Every write's `path` embeds SmartReach's REQUIRED
`team_id` query parameter directly as a literal `?team_id={{ config.team_id }}` path suffix
(the write dialect has no separate query-param mechanism — see Known limits) — `config.team_id`
must be set for any write to succeed, even though it remains optional for reads:

- **`add_or_update_prospects`** (`POST prospects`, `upsert`): creates or updates a prospect,
  deduped server-side by SmartReach's own `unique_identifier_columns` rule (email by default).
- **`add_prospects_to_campaign`** / **`unassign_prospects_from_campaign`** (`POST` / `PUT
  campaigns/{{ record.campaign_id }}/prospects`, `custom`): enroll/unenroll prospects in an
  outbound campaign. Risk: enrollment is the highest-impact prospect-level write — it triggers
  real scheduled outreach messages.
- **`update_prospect_campaign_status`** (`PUT prospects/prospect_status_change`, `update`):
  changes a prospect's per-campaign engagement status (`replied`/`pause`/`unpause`/
  `resume_later`).
- **`update_campaign_status`** (`PUT campaigns/{{ record.campaign_id }}/status`, `update`):
  starts/schedules/stops an entire campaign. Risk: the highest-impact campaign-level write —
  directly gates whether any outreach is sent at all.
- **`remove_from_do_not_contact`** (`DELETE do-not-contact`, `delete`): removes suppression-list
  entries by id, re-enabling outreach to those emails/domains.
- **`add_emails_to_do_not_contact`** / **`add_domains_to_do_not_contact`** (`POST
  do-not-contact/email` / `POST do-not-contact/domain`, `create`): blacklist emails or entire
  domains (with an optional per-domain exclusion list) from all future outreach.
- **`create_or_update_account`** (`POST accounts`, `upsert`): creates or updates a CRM-style
  account (company) record, distinct from the `accounts` read stream's own `search/accounts`
  endpoint (both hit the same underlying object, one paginated list, one single-object upsert).
- **`update_task_status`** (`PUT tasks/{{ record.task_id }}/status`, `update`): changes a sales
  task's status. The real API's request body is a `oneOf`/discriminator union of 4 shapes
  (`Due`/`Snoozed`/`Done`/`Skipped`, keyed by `status_type`) that this dialect's draft-07 subset
  cannot express as a true union (no `oneOf` support, same limitation stripe's
  `minProperties`-approximation deviation documents) — `record_schema` instead declares
  `status_type` plus every variant's own optional field (`due_at`/`snoozed_till`) as a flat
  optional set, and `body_fields` allow-lists exactly `status_type`/`due_at`/`snoozed_till` so an
  unrelated record field never leaks into the body. This is strictly more permissive than the real
  discriminated union (e.g. it does not reject `due_at` set alongside `status_type: "done"|
  "skipped"`), never stricter, matching the stripe `minProperties` precedent (`conventions.md` §5).

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
- **Write actions embed `team_id` as a literal `?team_id={{ config.team_id }}` path suffix, not a
  declared query parameter.** `WriteAction` (`engine/bundle.go`) has no `query` field at all —
  only `path`/`path_fields`/`body_type`/`body_fields` — so a write's required query parameter must
  be expressed as a literal-text query string appended directly to `path`, exactly like
  commercetools's `?version={{ record.version }}` golden pattern. `InterpolatePath` percent-encodes
  per `/`-delimited segment but leaves literal `?`/`=` characters outside any `{{ }}` marker alone,
  so this composes correctly as long as the resolved `config.team_id` value itself contains no `/`
  (a query-string value, never a path segment, so this is the correct treatment). A write attempted
  with no `team_id` configured hard-errors at interpolation (an absent required `config.*` key is
  always a hard error outside `when`/opt-in query-object tolerance, per `conventions.md` §3) —
  correct, since the live API genuinely rejects these calls without `team_id`.
- **Write request/response body shapes are sourced directly from SmartReach's own published
  OpenAPI 3.0 specification** (embedded in the ReadMe-hosted reference docs at
  `https://help.smartreach.io/reference/getprospects`, extracted 2026-07-03), not from legacy Go
  (which has no write path at all — `capabilities.write` was `false` prior to this Pass B
  expansion). This is a stronger source of truth than the SmartEngage/Smartwaiver writes in this
  same bundle group, which were inferred from prose documentation/integration references only.
