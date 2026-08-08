# Plan — Zoom Workforce Management documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3938](https://github.com/polymetrics-ai/cli/issues/3938).
- Scope: Zoom's published **Workforce Management** artifact only: all eighteen documented actions,
  the necessary typed CSV multipart validation foundation, command contracts, fixtures, derived
  Zoom docs/site output, endpoint reconciliation, and phase evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. This provider-category phase
  is not registered by the official runtime and the parent contract forbids role spawning, so this
  is the documented inline manual-GSD fallback with explicit discussion, plan, RED, GREEN,
  verification, and review evidence.

## Inline discuss-phase record

Workforce Management is Zoom's own published category and one Markdown artifact. Its resources
are filter groups, forecasts, imports, organizational groups, reports, scheduling groups, and
users; they belong together without inventing a cross-cutting delivery lane. All eighteen actions
are independently implementable through the declarative direct read/write engine, except that the
two published CSV file imports needed a declared CSV validator and staffing needs the source's
numeric `forecast_duration_weeks` range (`1`–`4`). Those narrow foundations ship here in their own
red/green commits rather than classifying uploads as unsafe or deferring them.

## Live artifact audit — completed before RED

The source was fetched afresh rather than trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/workforce-management.md` |
| Retrieval | `2026-08-08T16:39:34Z` |
| HTTP / bytes | `200` / `91,852` |
| SHA-256 | `d9d32ad906fed900608ba486b55c446923d99d06245c87bae6e2602061d3b414` |
| Artifact | OpenAPI `3.1.1`, API server `https://api.zoom.us/v2` |
| Ledger comparison | exactly 18 local `provider_module=workforce-management` rows; method, path, title, and source URL match (delta `0`) |

The live artifact contains exactly these eighteen operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/workforce-management/filter-groups` | List filter groups |
| GET | `/workforce-management/forecasts` | List forecasts |
| GET | `/workforce-management/forecasts/{forecastId}/scheduling-groups/{schedulingGroupId}` | Get a forecast for a scheduling group |
| POST | `/workforce-management/imports/historical-agent-status` | Upload historical agent status |
| DELETE | `/workforce-management/imports/historical-agent-status` | Delete historical agent status |
| POST | `/workforce-management/imports/historical-queue-metrics` | Upload historical queue metrics |
| POST | `/workforce-management/imports/staffing` | Upload forecast staffing |
| GET | `/workforce-management/imports/{importId}/historical-queue-metrics` | Get historical queue metrics import metadata |
| GET | `/workforce-management/organizational-groups` | Get multiple organizational groups |
| POST | `/workforce-management/organizational-groups` | Create an organizational group |
| GET | `/workforce-management/organizational-groups/{organizationalGroupId}` | Get a single organizational group |
| DELETE | `/workforce-management/organizational-groups/{organizationalGroupId}` | Delete organizational group |
| PATCH | `/workforce-management/organizational-groups/{organizationalGroupId}` | Update an organizational group |
| GET | `/workforce-management/reports/adherence/agents` | List agents' adherence data |
| GET | `/workforce-management/reports/schedules/agents` | List all schedule agents |
| GET | `/workforce-management/schedules/agents` | List all schedule agents |
| GET | `/workforce-management/scheduling-groups` | List scheduling groups |
| GET | `/workforce-management/users` | List users |

The artifact has no `Path Parameters` or `Query Parameters` sections. Response pagination fields
such as `page_size` and `next_page_token` are not request parameters; no `page`, `per_page`,
`limit`, `page-size`, or cursor flags will be hand-authored. The two DELETE actions return
`204 No Content` and must remain status-only destructive actions.

## Typed CSV multipart foundation

`POST /workforce-management/imports/historical-queue-metrics` publishes a multipart file that is
"Only CSV", and `POST /workforce-management/imports/staffing` requires a `.csv` source no larger
than 1 MB. Extension checking alone would admit arbitrary bytes renamed `.csv`; MIME sniffing
reports ordinary CSV as text/plain and cannot prove CSV grammar. Before declaring either upload,
add a closed `content_validation: "csv"` option to the existing bounded multipart policy:

- it is declaration-owned, file-part-only, and requires the existing positive file or aggregate
  byte cap;
- it accepts only a complete syntactically valid CSV stream from the already snapshot-bound source;
- it composes with a declared `.csv` extension list and a truthful declared `text/csv` part header;
- it adds no caller-selected validator, header, URL, MIME policy, raw body, or generic file upload;
  and
- malformed CSV or an extension mismatch fails before dispatch. The staffing source cap is the
  published 1 MB; the historical-queue artifact omits a maximum, so its declaration uses a
  documented-in-plan 10 MB local transport safety cap rather than an unbounded upload.

The staffing operation's `forecast_duration_weeks` remains a provider-declared number constrained
by Draft-07 `minimum: 1` and `maximum: 4`; it is not weakened to an unbounded number or narrowed
to an invented integer-only enum. The engine's numeric-bound compiler/validator foundation must
validate finite numeric instances and reject an unsatisfiable declared range before the connector
is considered executable.

## Locked decisions

- Implement all eighteen audited operations: eleven bounded `rest_read` / `direct_read` commands
  and seven approval-gated `rest_write` / `direct_write` commands. There are no duplicate
  exclusions and no `unsafe_or_disallowed` rows.
- CLI paths use source resources under `workforce-management filter-groups`, `forecasts`,
  `imports`, `organizational-groups`, `reports`, `schedules`, `scheduling-groups`, and `users`.
- JSON bodies use named root-object flags tied to closed source schemas. CSV file inputs are
  project-root-contained, snapshot-bound paths with declared extension and grammar validation.
- Every mutation retains plan → no-network preview → single-use approval → execute. Both DELETEs
  require typed destructive confirmation; they are status-only. All provider IDs, names,
  descriptions, emails, schedule/adherence detail, timestamps, file/import values, and generic
  token fields are redacted in previews, errors, and output. Fixtures remain synthetic.
- Whole-file derived metadata comes only from `surface-sync` and reconciliation. Docs/site are
  generated repository-wide, unrelated output is restored, and only Zoom output is retained.

## TDD execution

1. **Plan checkpoint** — commit this source audit, target accounting, CSV foundation decision, and
   inline manual-GSD fallback before test or production changes.
2. **RED checkpoint** — add only the Workforce Management command-surface test and a CSV
   multipart-foundation test. Capture the current `84 → 102` executable, `1,758 → 1,740`
   Zoom-local, `44 → 55` direct-read, and `35 → 42` direct-write failure, and prove every
   provider-native path is unknown through real preflight. Prove `content_validation: "csv"` is
   rejected by the existing closed policy before production change.
3. **GREEN CSV foundation** — add CSV syntax validation in its own commits, test valid upload and
   pre-network malformed/extension rejections, and preserve the JSON policy.
4. **RED/GREEN numeric range foundation** — capture and separately fix the closed schema
   compiler's rejection of the source-required `minimum`/`maximum` numeric bounds. Validate
   source boundary values, fractional in-range values, out-of-range values, compile-time
   contradictory ranges, and nonnumeric Draft-07 applicability.
5. **GREEN connector** — declare all 18 Workforce Management operations and CLI paths, add exact
   fixtures, reconcile only this module, generate derived metadata/docs/site, and run every action
   through the real runner.
6. **Verify/review** — run fixture lifecycle tests, source/ledger checks, fresh binary base/group/
   every-command help, generated-output scope checks, and inline review.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 84 | 102 |
| Zoom-local implementable rows | 1,758 | 1,740 |
| Direct reads | 44 | 55 |
| Direct writes | 35 | 42 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN real-commandrunner preflight and fixture lifecycle tests for all 18 actions.
- CSV foundation tests for a valid snapshot-bound `.csv`, malformed CSV, wrong extension, positive
  maximum requirement, and JSON policy regression.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=workforce-management`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom workforce-management`, and every
  exact command `--help`.
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
- `https://developers.zoom.us/docs/api/workforce-management.md`
