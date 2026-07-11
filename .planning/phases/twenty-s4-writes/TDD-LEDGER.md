# Twenty S4 writes TDD ledger (#281)

Status: GREEN for PR #304 F1 review fix. Manual GSD fallback because `scripts/gsd prompt programming-loop init --phase twenty-s4-writes --dry-run` is unavailable (`unknown GSD command: programming-loop`).

Loaded skills: `gsd-core`; fallback Go skills `golang-how-to`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-naming`, `golang-code-style`, `golang-documentation`; `caveman` for handoff. Repo-local `.pi/skills/go-implementation/SKILL.md` missing in worktree (recorded fallback; no out-of-scope edit to replace it).

## Red evidence captured before production edits

```text
research preflight:
true
0
true
field manifest preflight:
28
546
api surface GET preflight:
56
api surface covered streams preflight:
56
initial red counts:
writes.json missing (expected before S4)
56
56
0
```

## Review fix F1 red/green ledger

Finding accepted: Twenty batch REST creates expect `request.body` to be `Partial<ObjectRecord>[]` (bare JSON array), not `{ "records": [...] }`. Existing S4 batch actions use `body_fields:["records"]`, so the current engine sends an object wrapper.

Red evidence before production code:

```bash
go test ./internal/connectors/engine -run 'TestWrite.*BodyField|TestWriteRawBodyField' -count=1
```

```text
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/write_test.go:209:3: unknown field BodyField in struct literal of type WriteAction
internal/connectors/engine/write_test.go:255:3: unknown field BodyField in struct literal of type WriteAction
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

Meaning: tests require the new narrow `body_field` dialect before production implementation.

Green evidence:

- `TestWriteRawBodyFieldSendsTopLevelArray` proves `body_field:"records"` sends a top-level JSON array without `{records: ...}` wrapper.
- `TestWriteRawBodyFieldRequiresField` proves a missing raw body field errors before a request can succeed.
- All 28 Twenty `batch_` actions now have `body_field:"records"`, no `body_fields`, and still require outer `records` in `record_schema`.
- Required local gates passed: jq, Python shape check, focused body-field tests, connectorgen validate, twenty conformance, focused packages, `go vet ./...`, `go build ./cmd/pm`, `gofmt -l cmd internal`, `go test ./... -count=1`, `scripts/verify-gsd-workflow 1a86cc1a`.

## Original green evidence summary

- Generated `internal/connectors/defs/twenty/writes.json`: 84 actions = 28 create + 28 update + 28 batch; delete count 0; unique names.
- Updated `internal/connectors/defs/twenty/api_surface.json`: 140 endpoints = 56 GET + 56 POST + 28 PATCH; DELETE count 0; 84 `covered_by.write` rows.
- Batch modeled as object-wrapped `records` array (`body_fields:["records"]`) based on local Twenty research/docs: `/rest/batch/{objects}`, reverse_etl, batch max 60 records/request.
- Immutable/system prune validation: no `createdAt`, `updatedAt`, `deletedAt`, `createdBy`, `updatedBy`, `searchVector`; no `id` in create/batch item bodies; `id` present only as update path field/schema property.
- `connectorgen validate`, twenty conformance, focused packages, `gofmt -l`, `go vet`, `go build`, and `go test ./... -count=1` passed. `make verify` not run because Makefile `smoke` executes `pm reverse run`; S4 safety forbids reverse-ETL execution.

## Ledger

| # | Red / validation-first gate | Green implementation | Refactor / notes | Status |
|---|---|---|---|---|
| 1 | Initial count red: `writes.json` missing; API surface 56/56 GET/0 write rows. | Phase artifacts created; manual GSD fallback recorded. | Planning artifacts only before production JSON. | DONE |
| 2 | Candidate writes absent/fails 84-action target. | Generated 28 create, 28 update, 28 batch actions from S2 schemas and field manifest. | Object-major/action-minor order; no schema edits. | DONE |
| 3 | Writes without API coverage would fail validation. | Appended 84 S4 `covered_by.write` rows to `api_surface.json`. | Existing 56 GET rows preserved as data; only scope text changed. | DONE |
| 4 | Batch body shape must be representable in current write dialect. | Implemented one input-record batch container with `records` array and `body_fields:["records"]`. | No aggregation, raw HTTP, credentials, or live execution. | DONE |
| 5 | Final count/shape assertions fail until complete. | Actions/API counts and shape checks all returned expected true/counts. | Row-level `source_url`/`execution_model` omitted because schema disallows unknown keys. | DONE |
| 6 | Validator may find schema/surface issues. | `go run ./cmd/connectorgen validate internal/connectors/defs --json` returned `findings: []`, `warnings: []`. | No validator/engine/schema changes. | DONE |
| 7 | Conformance may require fixtures. | `go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1` passed without adding write fixtures. | Write fixtures deferred; no fixture edits. | DONE |
| 8 | Local gates. | Focused packages, format, vet, build, full tests, and `scripts/verify-gsd-workflow 1a86cc1a` passed. | `make verify` skipped by reverse-ETL safety gate. | DONE |
| 9 | Review/PR. | Commit created; push/PR pending. | Parent PR #285 remains draft/human-gated. | PENDING |
| 10 | PR #304 F1 red test: focused engine test fails to compile because `WriteAction.BodyField` does not exist. | Added `WriteAction.BodyField`, writes schema `body_field`, JSON raw field payload path, Twenty batch migration, conventions docs. Focused and broad gates passed. | Narrow dialect only; no generic raw HTTP. | DONE |

## Safety ledger

- Reverse ETL execution: NOT RUN.
- Live credentials: NOT USED.
- Destructive/delete actions: NOT ADDED in S4.
- Generic HTTP/raw write tools: NOT EXPOSED; F1 fix remains action-specific declarative `body_field` gated by `record_schema`.
- New dependencies: NONE.
- CLI/help/docs/website parity: deferred to S6 #283/S7 #284 by parent DAG; no S4 edits.
