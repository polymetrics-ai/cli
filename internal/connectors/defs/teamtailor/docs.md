# Overview

Teamtailor is a declarative-HTTP migration of `internal/connectors/teamtailor` (the hand-written
legacy connector this bundle migrates; the legacy package stays registered and unchanged until
wave6's registry flip). It reads Teamtailor jobs, candidates, job applications, departments,
locations, roles, stages, teams, users, and regions, and writes approved recruiting mutations
through the Teamtailor JSON:API REST API (`https://api.teamtailor.com/v1/<resource>`).

This is a Pass B full-surface expansion: the wave2 migration covered only the single legacy-parity
`jobs` stream; every other stream and every write action here is new coverage researched against
Teamtailor's published Postman collection API reference (`api_surface.json`).

## Auth setup

Legacy accepts either of two credential shapes with a defined precedence: a plain config value
`api` (read first) OR a secret `api_key` (fallback when `api` is unset). Both are marked
`x-secret: true` here and are wired as TWO auth candidates in `streams.json` `base.auth`, in the
SAME precedence order legacy uses — `api` first, `api_key` second. Either credential is sent as
`Authorization: Token token=<key>` (`api_key_header` mode, `prefix: "Token token="`). An optional
`x_api_version` config value is sent as the `X-Api-Version` header; when unset, the header is
omitted entirely. Never logged.

`base_url` defaults to `https://api.teamtailor.com/v1` and may be overridden for tests/proxies.

## Streams notes

Every stream shares Teamtailor's JSON:API envelope: records live at `data`, and every attribute is
nested under each record's own `attributes` object using HYPHENATED keys (`first-name`,
`created-at`, etc.) rather than the schema's snake_case output names — plain schema projection
alone would silently drop every such field, so each stream's `computed_fields` rehydrates them from
the raw pre-projection record. Pagination follows a 1-based page-number convention
(`pagination.type: page_number`, `page_param: page[number]`, `size_param: page[size]`,
`page_size: 100`) shared by every stream via `base.pagination`.

- `jobs` (legacy-parity): `GET /jobs`, primary key `["id"]`. `title`/`created_at` computed-field-
  renamed from `attributes.title`/`attributes.created-at`.
- `candidates`: `GET /candidates`, primary key `["id"]`, `x-cursor-field: created_at`.
  `first_name`/`last_name`/`email`/`created_at` computed-field-renamed from
  `attributes.first-name`/`attributes.last-name`/`attributes.email`/`attributes.created-at`.
- `job_applications`: `GET /job-applications`, primary key `["id"]`, `x-cursor-field: created_at`.
  `created_at` renamed from `attributes.created-at`; `candidate_id`/`job_id` are derived from the
  JSON:API `relationships.candidate.data.id`/`relationships.job.data.id` linkage objects (NOT plain
  attributes) via the same `computed_fields` dotted-path walk.
- `departments`, `locations`, `roles`, `stages`, `teams`, `regions`: each `GET /<resource>`,
  primary key `["id"]`, `name`(/`city`/`country` for locations) computed-field-renamed from
  `attributes.name` etc. No incremental cursor (Teamtailor's own `attributes` for these resources
  carry no `created-at`/`updated-at` field consistently enough to model one).
- `users`: `GET /users`, primary key `["id"]`, `x-cursor-field: created_at`. `name`/`email`/
  `created_at` computed-field-renamed from `attributes.name`/`attributes.email`/
  `attributes.created-at`.

## Write actions & risks

Every Teamtailor write body is the JSON:API envelope `{"data": {"type", "attributes",
"relationships"}}` — modeled by declaring the RECORD ITSELF as that envelope shape (a `record`
whose one top-level key is `"data"`). The engine's `body_type: json` write path sends record fields
verbatim as the top-level JSON body (`buildJSONBody`), so the body produced is byte-for-byte the
JSON:API shape the live API requires — no engine change was needed (the same nested-object-passes-
through pattern bitly's `create_qr_code.destination` and Teamwork's wrapper-object writes both
already prove). `type` fields use a single-value `enum` rather than draft-07's `const` keyword
(the engine's schema compiler does not implement `const`; a one-element `enum` is the equivalent
constraint this dialect DOES support).

- `create_job` (create, `POST /jobs`): creates a new job posting; low-risk, no approval required.
- `create_candidate` (create, `POST /candidates`): creates a new candidate; stores personal data
  (name/email) about a real individual, subject to data-protection obligations.
- `update_candidate` (update, `PATCH /candidates/{id}`, path resolved via
  `{{ record.data.id }}` — no `path_fields` declared, since excluding the whole `data` wrapper
  from the body would strip the entire JSON:API payload the PATCH body legitimately needs to repeat
  its own `id`/`type` inside): mutates an existing candidate's personal data.
- `create_job_application` (create, `POST /job-applications`): links a candidate to a job as a new
  application via `relationships.candidate`/`relationships.job`; moves the candidate into that
  job's active pipeline and may trigger applicant notifications.
- `update_job_application` (update, `PATCH /job-applications/{id}`): mutates an existing
  application (e.g. moves it to a different stage); may trigger applicant-facing notifications.
- `create_department` (create, `POST /departments`): low-risk, no approval required.
- `create_location` (create, `POST /locations`): low-risk, no approval required.
- `create_team` (create, `POST /teams`): low-risk, no approval required.
- `create_todo` (create, `POST /todos`): creates a to-do reminder, optionally assigned to a user
  against a candidate via `relationships`; low-risk, no approval required.
- `create_note` (create, `POST /notes`): creates an internal recruiter note on a candidate via
  `relationships.candidate`; stores potentially sensitive commentary about a real individual.

`metadata.json` now declares `capabilities.write: true`.

## Known limits

- User-account provisioning/administration, permission-role management, and custom-field-schema
  administration are excluded as `requires_elevated_scope` — these require Teamtailor
  administrator-tier credentials beyond ordinary recruiting read/write access.
- Destructive deletes beyond the covered writes are excluded as `destructive_admin`.
- Binary-upload endpoints (resumes/documents) are excluded (`binary_payload`) — the engine's
  `body_type` dialect has no multipart shape.
- Video-question answers/forms, custom fields, divisions, interviews, job offers, movements
  (stage-history audit trail), NPS surveys, onboardings, partner-integration results, referrals,
  reject-reason vocabulary, headcount requisitions, and interview scorecards are excluded as
  `out_of_scope` — each is a separate niche/feature-specific object domain distinct from core
  jobs/candidates/applications data; Pass B breadth-vs-cost triage.
- `todos`/`notes` list streams are NOT modeled (only their `create_*` writes are covered) — see
  `api_surface.json`'s reasoning: each is a per-user/per-candidate feed rather than primary
  recruiting-pipeline data, while creating a reminder or note is still a common, worthwhile write.
- All fixtures (`fixtures/streams/**`, `fixtures/writes/**`, `fixtures/check.json`) represent
  Teamtailor's real JSON:API wire shape, including hyphenated attribute keys before their
  `computed_fields` renames, and the `{"data": {...}}` envelope write request bodies described
  above.
- Dual-auth precedence (`api` over `api_key`) is exercised at authoring time per the ordering rule
  above; a both-present parity assertion is left to a future parity-suite extension (no
  `paritytest/teamtailor` package exists in this wave).
