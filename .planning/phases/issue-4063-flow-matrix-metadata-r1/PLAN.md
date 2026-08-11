---
issue: 4063
parent_issue: 3897
phase: issue-4063-flow-matrix-metadata-r1
type: tdd
wave: 1
depends_on: []
files_modified:
  - .planning/phases/issue-4063-flow-matrix-metadata-r1/
  - internal/connectors/certifications/flow-matrix.json
autonomous: true
requirements: []
---

# #4063 TDD Plan: Refresh flow-authoring discovery metadata

## Manual-GSD execution record

Resolved commands:

- scripts/gsd prompt discuss-phase issue-4063-flow-matrix-metadata-r1 --auto
- scripts/gsd prompt plan-phase issue-4063-flow-matrix-metadata-r1 --tdd
- scripts/gsd prompt execute-phase issue-4063-flow-matrix-metadata-r1
- scripts/gsd prompt verify-work issue-4063-flow-matrix-metadata-r1
- scripts/gsd prompt code-review issue-4063-flow-matrix-metadata-r1

phase-op and plan-phase report phase_found=false because ROADMAP.md is an
archive entry, not an active numbered issue roadmap. This is the documented
manual-GSD fallback. The one active worker executes each lifecycle stage
inline; no GSD role is spawned.

Required skills loaded: github-issue-first-delivery, golang-how-to,
golang-cli, golang-testing, golang-error-handling, golang-security,
golang-safety, golang-lint, gsd-discuss-phase, gsd-plan-phase,
gsd-execute-phase, gsd-verify-work, and gsd-code-review.

## TDD feature

**Feature:** the generated flow matrix accurately names the existing
flow_authoring runtime source coordinate.

**Why TDD applies:** the canonical generator check is a deterministic,
observable drift contract. It supplies RED and GREEN evidence without
inventing a new behavior test or modifying production Go code.

### Task 1 — RED: observe exact-head drift

<read_first>
- internal/cli/flow_cli.go
- internal/connectors/certifications/flow-matrix.json
- cmd/connectorgen/certificationmatrix.go
- cmd/connectorgen/certificationflow.go
- .planning/phases/issue-3897-flow-connection-scope-r1/TDD-LEDGER.md
- .planning/phases/issue-3897-flow-connection-scope-r1/RUN-STATE.json
</read_first>

<action>
At exact head 002ddf674a447bf0872486aa979efdaa078f602c, run go run
./cmd/connectorgen certification-matrix --check. It must exit 1 because the
flow matrix retains internal/cli/flow_cli.go:20 while the annotated func
runFlow begins at line 21. Record the exit and non-secret error text without
modifying any generated file.
</action>

<acceptance_criteria>
- The command exits 1 before generator execution.
- internal/cli/flow_cli.go shows pmcert:workflow flow_authoring at line 20
  and func runFlow at line 21.
- flow-matrix.json shows flow_authoring.discovery_source at line 20.
</acceptance_criteria>

### Task 2 — GREEN: canonical regeneration and semantic boundary

<read_first>
- cmd/connectorgen/certificationmatrix.go
- cmd/connectorgen/certificationflow.go
- internal/cli/flow_cli.go
- internal/connectors/certifications/flow-matrix.json
</read_first>

<action>
Run go run ./cmd/connectorgen certification-matrix once. Inspect the complete
semantic diff against the exact starting head. Accept only
workflow_kinds[id=flow_authoring].discovery_source changing from
internal/cli/flow_cli.go:20 to internal/cli/flow_cli.go:21 in
internal/connectors/certifications/flow-matrix.json. Do not hand-edit JSON or
accept any other generated artifact, capability fact, status, evidence, or
connector-definition change.
</action>

<acceptance_criteria>
- git diff --name-only against exact head lists flow-matrix.json plus only
  this issue's planning evidence.
- JSON semantic comparison reports one changed leaf, discovery_source.
- The changed leaf has old value :20 and new value :21.
</acceptance_criteria>

### Task 3 — verification and delivery gates

<read_first>
- .planning/phases/issue-4063-flow-matrix-metadata-r1/TDD-LEDGER.md
- .planning/phases/issue-4063-flow-matrix-metadata-r1/VERIFICATION.md
- internal/connectors/certifications/flow-matrix.json
</read_first>

<action>
Run every required focused test, generator check, vet, lint, GSD workflow
guard, diff check, Shepherd/no-mistakes local-only path, inline verify-work,
and inline code review. Commit planning, then generated correction and
closeout evidence as coherent checkpoints. Push only
feat/3897-flow-connection-scope-nm; verify #4060 remains a draft targeting
feat/3988-github-certification and wait for current CI including Snyk.
</action>

<acceptance_criteria>
- Every command listed in VERIFICATION.md is recorded PASS, expected RED, or
  a bounded non-blocking external status.
- No command enables no-mistakes push, PR, or CI stages.
- #4060 exact head/base/draft evidence is recorded after push.
</acceptance_criteria>

## Commit checkpoints

1. docs(issue-4063): plan flow matrix metadata correction
2. fix(certification): refresh flow authoring discovery source
3. docs(issue-4063): record verification and review closeout

## Scope fence

This plan does not change a CLI command, flag, output, help topic, docs, or
website surface. CLI help/manual/website parity is explicitly not applicable.
