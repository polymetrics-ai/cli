# Issue 4293 — Source-cited binary-media wildcard repair R1

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base: immutable commit `19467858a5c81bd24081bc261ec6fd76f14f5c0f` on `codex/4293-mapping-media-evidence-r1`.
- Delivery: one mapping-only review-fix commit, pushed for a fresh independent review; never merge from this repair.
- Scope: reject wildcard-bearing parsed MIME types in binary mapping evidence while retaining source-cited concrete provider MIME types.
- Boundaries: no runtime, execution, transport, certification, source-lock, connector-definition, or Foundation Atlas changes.

## Manual GSD fallback and plan

The repository GSD lifecycle is executed inline because this narrow review fix has no compatible isolated role and the project contract forbids role spawning. `scripts/gsd doctor` and the sources for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` passed before the final review. Skills used: `go-engineering`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`. CodeGraph was attempted first; this fresh worktree has no `.codegraph/` index, so normal source inspection is used.

1. **Red:** make request/upload and response/download mapping checks reject `application/*+xml` and `*/pdf`; prove the existing helper wrongly admits them.
2. **Green:** treat a parsed MIME type as concrete only when neither type component contains `*`.
3. **Refactor/verify:** retain standard-library parameter normalization and source-citation authority, then run focused mapping, cohort, engine, vet, contract, and diff checks.

## TDD ledger

### Red:

Before changing production code, the focused integration tests reproduced the
review finding:

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'TestSourceOperationMappingCheck(RejectsUncitedOrContradictoryBinaryMedia|RejectsWildcardBinaryDownloadMedia|AcceptsSourceCitedConcreteBinaryMedia)$' \
  -count=1
```

The source-cited `application/*+xml` assertion incorrectly admitted both an
upload and a download binary lane (`0 finding(s)`). `*/pdf` was already
rejected by the standard parser; it remains in both directions to keep every
wildcard-bearing type pattern explicitly covered.

### Green:

`sourceOperationMappingConcreteMediaType` now rejects a parsed media type
whenever either the top-level or subtype component contains `*`. This is a
general concrete-type check, not a provider allow-list. The focused red command
passes after the one-line repair, including the existing positive cases for
`application/gzip`, parameterized gzip, and `application/x-provider-archive`.

### Verification and review

```text
go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestBatch1SourceOperationMappingCohort|TestSourceOperationMapping)' \
  -count=1
# ok  polymetrics.ai/cmd/connectorgen  39.321s

go test -timeout 20m ./internal/connectors/engine -count=1
# ok  polymetrics.ai/internal/connectors/engine  20.404s

go vet ./cmd/connectorgen ./internal/connectors/engine
# exit 0

go run ./cmd/connectorgen source-operation-mapping-cohort \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check
# connectorgen source-operation-mapping-cohort: 10 connector(s), 4341 source operation(s), 0 finding(s)

go run ./cmd/connectorgen declaration-admission
# connectorgen declaration-admission: 1 connector(s), 1 source operation(s), 0 finding(s)

go run ./cmd/agentcontractgen check
# agentcontractgen: canonical contract and registered projections are current

git diff --check
# exit 0
```

No JSON files changed, so no JSON validity check applies. Final scoped diff
review, push, and a fresh independent review remain pending. This trace does
not declare merge readiness.
