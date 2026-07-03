# Overview

ClickUp-api reads ClickUp workspaces (teams), spaces, folders, lists, and tasks through the
ClickUp v2 REST API (`https://api.clickup.com/api/v2`). This bundle is a wave2 fan-out migration
of `internal/connectors/clickup-api` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip. ClickUp is read-only here — legacy's
own package doc: "ClickUp's writes (creating tasks, etc.) are not a natural reverse-ETL target";
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide a ClickUp personal API token via the `api_token` secret; it is sent RAW in the
`Authorization` header with no `Bearer ` prefix (`mode: api_key_header`, `header: Authorization`,
`prefix: ""`), matching legacy's `connsdk.APIKeyHeader("Authorization", token, "")` exactly (a
ClickUp personal token convention, not OAuth Bearer). `base_url` defaults to
`https://api.clickup.com/api/v2` and may be overridden for tests/proxies.

## Streams notes

Streams are declared `tasks`/`teams`/`spaces`/`folders`/`lists` in this bundle's `streams.json`
(reordered from legacy's own catalog order `teams, spaces, folders, lists, tasks` purely so the
one genuinely-paginated stream, `tasks`, is exercised by conformance's `pagination_terminates`
dynamic check, which always picks the bundle's first declared stream — catalog position carries
no legacy-parity meaning in this dialect, unlike stream NAMES, which are unchanged).

- **`teams`** (`GET /team`): unparameterised, matching legacy's default stream.
- **`spaces`** (`GET /team/{{ config.team_id }}/space`): requires config `team_id` (an absent
  `team_id` hard-errors on both sides — legacy: `"clickup-api stream spaces requires config
  team_id"`; engine: an unresolved `config.team_id` path-template key — same failure
  classification, different literal text, per conventions.md §5's config-validation-parity
  precedent).
- **`folders`**/**`lists`** (`GET /space/{{ config.space_id }}/folder` and `.../list`): both
  require config `space_id` (same absent-key parity pattern as `spaces`/`team_id`). Both stamp a
  `space_id` computed field from the raw API's nested `{{ record.space.id }}` reference (ClickUp
  nests a partial space object on folder/list records), matching legacy's own
  `nestedID(item["space"])` helper.
- **`tasks`** (`GET /team/{{ config.team_id }}/task`): requires config `team_id`. Computed fields
  extract every nested-object id ClickUp returns (`status` from `{{ record.status.status }}`,
  `creator_id` from `{{ record.creator.id }}`, `list_id`/`folder_id`/`space_id` from their
  respective nested `.id` references) — bare single-reference `computed_fields` entries copy the
  RAW typed value (so `creator_id` stays a JSON integer, matching ClickUp's real wire shape and
  legacy's own `nestedID` passthrough), matching legacy's `clickupTaskRecord`/`statusName`/
  `nestedID` exactly.

Pagination (`tasks` only; every other stream is a single unpaginated read, matching legacy's
`fetchOnce` for those 4 streams): `pagination.type: page_number`, `page_param: page`, **no
`size_param`** (legacy never sends a page-size query parameter — ClickUp's task list endpoint
returns a fixed ~100 items per page server-side, confirmed by ClickUp's own docs: "Responses are
limited to 100 tasks per page"). `streams.json` pins `page_size: 100`, matching that real
server-side page size, as the client-side short-page-stop threshold (no query param is ever sent
for it — `size_param` is absent — so this value only ever decides when the paginator stops, never
what is requested). The required first-stream 2-page fixture proof (conventions.md §4) follows
chargify's/clarif-ai's identical precedent: page 1 returns exactly 100 tasks (a full page, so the
paginator advances) and page 2 returns 1 (a short page, so it terminates).

Legacy's real stop signal is `last_page == "true" || len(records) == 0`; the engine's
`page_number` paginator has no `stop_path`-equivalent field (that is exclusive to the `cursor`
pagination type in this dialect), so this bundle relies purely on the short-page-stop heuristic —
DATA-identical to legacy in the overwhelmingly common case, diverging only in emitted REQUEST
COUNT (never emitted RECORD data) for a final page landing on an exact multiple of the declared
threshold.

**`start_page` is declared `0`, matching ClickUp's real 0-indexed first page.** Legacy's own loop
is `for page := 0; page < clickupMaxPages; page++` — ClickUp's task list endpoint is genuinely
0-indexed (`page=0` is the real first page). This was originally blocked by an `ENGINE_GAP`
(`PaginationSpec.StartPage` was a plain Go `int`, and `engine/paginate.go`'s `newPaginator`
unconditionally coerced an explicit `"start_page": 0` to `1`, since the zero value could not be
distinguished from "never set" — the identical gap class documented in `algolia`'s and `datadog`'s
`docs.md`). That gap is now closed (S4 engine mini-wave item 1): `PaginationSpec.StartPage` is a
`*int`, so an explicit `start_page: 0` is honored verbatim rather than coerced, and this bundle now
reads `tasks` starting from ClickUp's real first page — full read parity for every stream in this
bundle (`teams`, `spaces`, `folders`, `lists`, `tasks`).

`tasks`'s incremental cursor is `date_updated` with `client_filtered: true` (ClickUp's real task
list endpoint exposes no server-side updated-since filter parameter — legacy's own `Read`/
`harvestPaged` never sends one), matching legacy's full-scan-then-emit-everything behavior exactly;
the engine drops already-seen records client-side by comparing `date_updated` against the
persisted lower bound.

`archived`/`include_closed_tasks` config values are forwarded verbatim as the literal query value
when present (`omit_when_absent`/`default` dialect), matching legacy's own `fetchOnce`
(`archived=true` sent only when configured true, otherwise omitted entirely — ClickUp's own docs
confirm `archived` defaults server-side to `false`, so omission and an explicit `archived=false` are
DATA-equivalent) and `harvestPaged` (`archived` ALWAYS sent, `true` or `false`; `include_closed`
sent only when true). See Known limits for the narrowed accepted-value vocabulary this requires.

## Write actions & risks

None. ClickUp-api is a read-only source connector; `capabilities.write` is `false` and this bundle
ships no `writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`include_archived`/`include_closed_tasks` accept only the literal strings `"true"`/`"false"`
  (or absent), narrower than legacy's `boolConfig` helper (which also accepts `"1"`/`"yes"` as
  truthy synonyms).** The engine's query-templating dialect has no boolean-normalization filter —
  a `{{ config.include_archived }}` reference forwards whatever string the operator set verbatim
  as the query value, with no equivalent of legacy's `strings.EqualFold`-based truthy-alias
  matching. Declaring the accepted vocabulary as `"true"`/`"false"` only (documented here rather
  than silently accepting-and-mismapping `"1"`/`"yes"`) keeps every ACCEPTED input's emitted
  records identical to legacy; an operator who previously relied on `"1"`/`"yes"` must switch to
  the literal `"true"` string. This is a documented config-surface narrowing, not a data-shape
  regression for any `"true"`/`"false"`/absent input.
- **`tasks` pagination has no `last_page`-boolean stop signal wired.** The engine's `page_number`
  paginator type supports only a short-page stop threshold (no `stop_path`, which is exclusive to
  the `cursor` pagination type in this dialect) — see Streams notes above for why this is
  data-identical to legacy in every case that matters (an over-threshold "extra" request against an
  already-exhausted endpoint returns zero records and stops immediately either way).
- Full ClickUp API surface (task/comment creation, list-scoped task retrieval) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity read streams are implemented.
