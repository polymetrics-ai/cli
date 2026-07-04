# Overview

Avni is a read-only source connector. It reads Avni subjects, encounters, program enrolments,
program encounters, group subjects, locations, and entity approval statuses through Avni's REST
API (`https://app.avniproject.org` by default, overridable for self-hosted instances) using HTTP
Basic authentication. This bundle migrates `internal/connectors/avni` (the hand-written legacy
connector); the legacy package stays registered and unchanged until wave6's registry flip. Legacy
is pure `connsdk`-HTTP with no signature auth, no custom stream handling, and no writes, so it maps
to a Tier-1 declarative bundle with zero Go.

**Pass B full-surface expansion**: this bundle was reviewed against Avni's published
`external-api.yaml` OpenAPI spec (`https://raw.githubusercontent.com/avniproject/avni-server/master/
avni-server-api/src/main/resources/api/external-api.yaml`, linked from
`https://avni.readme.io/docs/api-guide`) and now covers every list/paged GET endpoint that spec
documents: `subjects`, `encounters` (legacy-parity, wave0/wave1), plus newly added
`program_enrolments`, `program_encounters`, `group_subjects`, `locations`, and
`approval_statuses`. See `api_surface.json` for the full endpoint-by-endpoint disposition.

## Auth setup

Provide `username` (plain config) and `password` (`x-secret`) for HTTP Basic authentication —
`Authorization: Basic base64(username:password)` — matching legacy's
`connsdk.Basic(cfg.Config["username"], secret(cfg, "password"))` exactly.

## Streams notes

All 7 streams (`subjects`, `encounters`, `program_enrolments`, `program_encounters`,
`group_subjects`, `locations`, `approval_statuses`) share the same shape: `GET` against an Avni
list endpoint returning the `{items:[...], next_page}` envelope (`records.path: "items"`).
Pagination is `cursor`/`token_path` (`cursor_param: page`, `token_path: next_page`): the response
body's `next_page` field carries the literal next page number as a string, and pagination stops
when `next_page` is empty — identical to legacy's `readPaged` loop (`page = next; if next == ""
{ return nil }`) and unchanged from the legacy-parity `subjects`/`encounters` streams. Every
request sends `page_size` (default `100`, matching legacy's `defaultPageSize`) and an optional
`start_date`/`lastModifiedDateTime` query parameter, sent only when the `start_date` config value
is set (`omit_when_absent`) — legacy only calls `query.Set("start_date", start)` when
`strings.TrimSpace(cfg.Config["start_date"]) != ""`. `start_date` is a passthrough config value,
not a computed incremental lower bound (legacy never reads it back from a persisted sync cursor),
so no `incremental` block is declared for any stream — each stream's real-world updated-time field
stays a schema-only cursor candidate (`x-cursor-field`), matching legacy's own
`CursorFields: []string{"updated_at"}` catalog declaration with no server-side incremental filter
wired to it (§8 rule 2).

- `program_enrolments` reads `/api/programEnrolments` (Avni's program-enrolment resource — a
  subject's enrolment into a specific program, e.g. "Pregnancy" or "ANC program"). Primary key
  `["id"]`.
- `program_encounters` reads `/api/programEncounters` (program-scoped encounters, children of a
  program enrolment — e.g. a monthly visit within a pregnancy program). Primary key `["id"]`.
- `group_subjects` reads `/api/groupSubjects` (household/group membership: which member subjects
  belong to which group subject). Primary key `["id"]`.
- `locations` reads `/api/locations` (Avni's address/catchment hierarchy — villages, blocks,
  districts). Primary key `["id"]`.
- `approval_statuses` reads `/api/approvalStatuses` (entity-approval-workflow status per
  Subject/ProgramEnrolment/ProgramEncounter/Encounter/ChecklistItem entity); this endpoint's real
  query parameter is `lastModifiedDateTime` (not `start_date`), matching the OpenAPI spec exactly —
  the same `start_date` config value is reused and re-templated onto that parameter name for this
  one stream. Primary key `["entity_id", "entity_type"]` (a composite key; an approval status is
  scoped per entity-type, not globally unique per entity id alone, per the OpenAPI's own
  `EntityApprovalStatusBody` shape).

`check` issues a single bounded `GET /api/subjects?page_size=1`, mirroring legacy's `Check`
implementation exactly (a 1-record probe confirms auth and connectivity without mutating
anything).

**Field-shape note (pre-existing, inherited, not introduced by this expansion)**: Avni's published
OpenAPI spec (`external-api.yaml`) documents a different, richer envelope for these resources than
what this bundle (and the legacy Go connector it was migrated from) actually reads —
`{content:[...], totalElements, totalPages, pageSize}` with PascalCase-with-spaces field names
(`"ID"`, `"External ID"`, `"Registration date"`, etc.), rather than the `{items:[...], next_page}`
lowercase-snake envelope this bundle uses. The legacy connector was hand-written years before this
migration against real Avni server responses (evidenced by its own hardcoded `items`/`next_page`
decode calls, `connsdk.RecordsAt(resp.Body, "items")`/`connsdk.StringAt(resp.Body, "next_page")`),
and this bundle's job is to reproduce that ALREADY-SHIPPED, ALREADY-PARITY-TESTED behavior exactly,
not to retarget every stream at the public OpenAPI's documented (but not legacy-exercised) response
shape — doing so would be a silent, untested accepted-input-behavior change forbidden by the §5
meta-rule. The newly added streams (`program_enrolments`/`program_encounters`/`group_subjects`/
`locations`/`approval_statuses`) follow the SAME `items`/`next_page` envelope convention as the
existing `subjects`/`encounters` streams for internal consistency, since no legacy Go
implementation exists for them to port from and the existing streams are the only proven
ground-truth this repo has for how this specific Avni deployment shape actually responds. An
operator connecting a self-hosted Avni instance that instead ships the OpenAPI's documented
`content`/`totalElements` envelope for these paths would see empty results rather than an error —
flagged here as a documented known limit, not silently masked.

## Write actions & risks

None. Avni is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation`.

## Known limits

- All 7 documented list/paged GET endpoints in Avni's published `external-api.yaml` are now
  implemented as streams; see `api_surface.json` for the full disposition of every excluded
  endpoint (single-entity detail lookups, every create/update/patch/delete mutation, async
  bulk-migration/task/user-admin endpoints).
- **New-stream field shape follows the legacy `items`/`next_page` envelope, not the OpenAPI's
  documented `content`/`totalElements` envelope** — see the Streams notes field-shape callout
  above. This is a documented, deliberate consistency choice (matching the only proven ground-truth
  shape this repo has), not an oversight.
- `approval_statuses`'s composite primary key (`entity_id`+`entity_type`) means two different
  entity types sharing a numeric/UUID id collide only if the API itself would also return them as
  logically distinct records — this mirrors the OpenAPI's own `EntityApprovalStatusBody` shape,
  which has no single globally-unique id field.
- Documented parity deviation: legacy accepts a runtime `max_pages` config override
  (`intConfig(req.Config, "max_pages", defaultMaxPages)`); the engine's `pagination.max_pages` field
  is a static integer with no template support (no `PaginationSpec` field is ever resolved via
  `{{ }}` interpolation — see `conventions.md` §3's pagination table), so a per-request runtime
  override cannot be expressed declaratively. This bundle instead declares a fixed
  `pagination.max_pages: 100`, matching legacy's own `defaultMaxPages` constant exactly — the
  request-count cap every existing caller actually observes (nothing in the repo's config surface
  overrides `max_pages` from its default today). A caller that previously set a non-default
  `max_pages` would see a behavior change (this bundle always caps at 100); this is judged
  ACCEPTABLE as a documented scope narrowing rather than an `ENGINE_GAP`, since `max_pages` is a
  defensive request-count ceiling, not data-shaping logic, and every default-configured caller
  (the common case) is unaffected.
- `start_date` is sent verbatim as a static query parameter on every page of every read (matching
  legacy); it is not a computed, cursor-state-driven incremental lower bound.
