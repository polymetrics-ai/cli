# Plan — Zoom Tasks documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3939](https://github.com/polymetrics-ai/cli/issues/3939).
- Scope: Zoom's published **Tasks** artifact only: all seventeen documented actions, the one
  necessary redirect-safe multipart foundation, typed command contracts, fixtures, generated Zoom
  docs/site output, endpoint reconciliation, and this phase evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. This provider-category phase
  is not registered by the official runtime and the parent contract forbids role spawning, so this
  is the documented inline manual-GSD fallback with explicit discussion, plan, RED, GREEN,
  verification, and review evidence.

## Inline discuss-phase record

Tasks is Zoom's own published provider category. Assignee, collaborator, comment, file import,
and task-item resources all appear in the same `tasks.md` artifact, so this slice does not invent a
delivery category. All 17 paths are independently implementable except that the documented file
upload requires an engine redirect capability absent from the existing deliberately redirect-refusing
direct-write transport; the task is required to build that narrow capability here rather than defer
the operation.

## Live artifact audit — completed before RED

The source was fetched afresh rather than trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/tasks.md` |
| Retrieval | `2026-08-08T15:22:01Z` |
| HTTP / bytes | `200` / `68,605` |
| SHA-256 | `53920b1c473e4d8ccdad03475d6d14f13b6b0b54ce036127dd644e51850f09be` |
| Artifact | OpenAPI `3.1.1`, normal API server `https://api.zoom.us/v2` |
| Ledger delta | `0` — all 17 `provider_module=tasks` rows match the live method, path, title, and source URL |

The live artifact contains exactly these seventeen operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/tasks/items/{taskId}/assignees` | Get assignees of a task |
| POST | `/tasks/items/{taskId}/assignees` | Add assignees to a task |
| DELETE | `/tasks/items/{taskId}/assignees/{userId}` | Remove Assignee from task |
| GET | `/tasks/items/{taskId}/collaborators` | Get collaborators of a task |
| POST | `/tasks/items/{taskId}/collaborators` | Add collaborators to a task |
| DELETE | `/tasks/items/{taskId}/collaborators/{userId}` | Remove collaborator from task |
| GET | `/tasks/items/{taskId}/comments` | Get a task's comments |
| POST | `/tasks/items/{taskId}/comments` | Add a comment to a task |
| DELETE | `/tasks/items/{taskId}/comments/{commentId}` | Delete a task's comment |
| POST | `/tasks/files` | Upload a file in tasks |
| POST | `/tasks/imports` | Submit a task import job |
| GET | `/tasks/imports/{importId}` | Get import job status |
| GET | `/tasks/items` | List tasks |
| POST | `/tasks/items` | Create a new task |
| GET | `/tasks/items/{taskId}` | Get task details |
| DELETE | `/tasks/items/{taskId}` | Delete a task |
| PATCH | `/tasks/items/{taskId}` | Update task fields |

No provider request-parameter section declares `page`, `per_page`, `limit`, a cursor, date range,
or any other query input. Response-only paging and filter-looking fields will not become CLI flags.
The four DELETEs and task PATCH return `204 No Content`, so each must be a status-only action.

## Redirect-safe multipart foundation

`POST /tasks/files` is an ordinary documented operation, not an exclusion: it accepts a `.json`
file up to 10 MB at `https://fileapi.zoom.us/v2`, tells callers to follow HTTP 30x redirects, and
requires the Authorization header to survive a redirect to a different hostname. The existing
typed multipart `rest_write` foundation from #3761 intentionally refuses every redirect, including
307/308 body replay, so it cannot honestly implement this endpoint as-is.

The foundation will be its own red/green commits before Tasks authoring. It remains narrow:

- Add a closed declaration-owned redirect contract only for a typed multipart `rest_write` with a
  fixed literal HTTPS operation base URL and declared bearer authentication.
- Bind a finite redirect-hop cap and a provider-owned hostname-suffix allowlist into the preview
  definition. Zoom's declaration permits only `zoom.us` hosts; no command accepts a URL, hostname,
  header, or redirect policy value.
- Manually follow only an admitted HTTPS redirect, rebuild the already snapshot-bound multipart
  body for that one declared redirect hop, and reapply only the declared bearer authentication.
  Ordinary direct writes retain their no-retry/no-redirect behavior.
- Preserve root confinement, file media/type and 10 MB caps, approved SHA-256 binding,
  single-use approval, bounded response capture, and redacted error/result policy. A redirect to a
  non-allowlisted host, scheme downgrade, body/method change, or excess hop count fails before a
  second send.

This reusable engine extension unblocks provider-declared multipart writes that require a bounded,
same-provider cross-host redirect while retaining credentials. It is not a generic redirect or HTTP
write facility; its scope and future connector applicability will be stated in the parent handoff
and eventual PR body.

## Locked decisions

- Implement every audited operation: six bounded `rest_read` / `direct_read` commands and eleven
  approval-gated `rest_write` / `direct_write` commands. There are no duplicate exclusions and no
  `unsafe_or_disallowed` rows.
- CLI paths use the provider resources: `tasks assignees`, `tasks collaborators`, `tasks comments`,
  `tasks files`, `tasks imports`, and `tasks items`.
- JSON mutations use named root-object flags tied to one closed provider schema. The file upload
  accepts only a project-root-contained `.json` source path under the declared 10 MB multipart
  contract; it is not a raw body or arbitrary upload facility.
- Every mutation retains plan → no-network preview → single-use approval → execute. The four
  DELETE actions additionally require typed destructive confirmation, and all documented 204
  responses assert status only.
- Normal reads/writes use the Zoom `/v2` bearer transport. Assignee, collaborator, task, comment,
  import, file ID, project, title, description, URL, email, display-name, avatar, user ID, and
  generic token values are redacted in previews, errors, and output. Fixtures are synthetic.
- Whole-file derived metadata comes only from `surface-sync` and reconciliation. Docs/site are
  generated repository-wide, unrelated generated connector docs are restored, and only Zoom output
  is retained.

## TDD execution

1. **Plan checkpoint** — commit this source audit, target accounting, required foundation, and
   inline manual-GSD fallback before test or production changes.
2. **RED checkpoint** — add only the Tasks command-surface test plus foundation tests. Capture the
   current 67 → 84 executable, 1,775 → 1,758 Zoom-local, 38 → 44 direct-read, and 24 → 35
   direct-write failures; prove every path is unknown through real preflight. The multipart
   redirect test must show that a declared safe redirect is currently rejected before any
   production change.
3. **GREEN foundation** — add the redirect-safe typed multipart capability in its own commits and
   prove allowed redirect, bearer retention, bounded hop/method/body behavior, unapproved-host
   refusal, downgrade refusal, and unchanged ordinary no-redirect behavior.
4. **GREEN connector** — declare typed Tasks operations/commands/fixtures, reconcile only Tasks
   rows, generate derived metadata/docs/site, and test every command through the live runner.
5. **Verify/review** — run fixture lifecycle tests, surface/ledger checks, fresh binary base/group/
   every-command help, generated-output scope checks, and inline review.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 67 | 84 |
| Zoom-local implementable rows | 1,775 | 1,758 |
| Direct reads | 38 | 44 |
| Direct writes | 24 | 35 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN real-commandrunner preflight and fixture lifecycle tests for all 17 actions.
- Foundation loopback tests for the permitted Zoom-suffix HTTPS multipart redirect plus all
  rejection paths; fixture tests prove exact initial/final request method/body/auth and 10 MB cap.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=tasks`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom tasks`, and every exact command
  `--help`.
- Scoped CI-equivalent vet/lint/docs/website/CLI/contract/surface/boundary/release gates from
  `AGENTS.md`; the full repository suite remains CI-owned.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/tasks.md`
