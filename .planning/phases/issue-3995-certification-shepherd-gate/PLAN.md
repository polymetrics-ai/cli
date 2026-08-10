# PLAN — issue #3995 shared connector-certification Shepherd gate

Issue: #3995. Parent: #3988. Parent PR: #4018. Branch:
`feat/3995-certification-shepherd-gate`. Required child PR base:
`feat/3988-github-certification`.

## GSD path

- `scripts/gsd doctor` passed; `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review`.
- Generated prompt sources: `discuss-phase issue-3995-certification-shepherd-gate --auto`,
  `plan-phase issue-3995-certification-shepherd-gate --tdd --skip-research`,
  `execute-phase issue-3995-certification-shepherd-gate --interactive`,
  `verify-work issue-3995-certification-shepherd-gate --auto`, and
  `code-review issue-3995-certification-shepherd-gate --depth=standard`.
- The adapter has no registered numbered phase for #3995. Inline/manual execution is the
  documented fallback; this directory records each phase's decisions, RED/GREEN evidence, gaps,
  verification, and review. No extra GSD role is spawned.

## TDD slices

### Slice 1 — read-only, fail-closed certification evaluation

1. RED: add an `internal/agentcontract` test that loads a current or temporary generated GitHub
   certification input, evaluates `integrate_sub_pr`, and requires `RETRY` with the exact
   `capability/github/capability:check/live_evidence` failure ID. Add a paired all-green fixture
   that requires `PROCEED`; current code has no evaluator and cannot produce either verdict.
2. RED: cover each binding criterion independently; reachability/file/implementation-only input;
   unknown and missing schema; unknown JSON fields; omitted required gate/adaptor fields; malformed
   or unmatched evidence records; and a read-only fixture that detects no created evidence/provider
   action.
3. GREEN: add the canonical gate schema, strict typed artifacts/evidence decoder, stable failure
   coordinates, and pure evaluator under `internal/agentcontract`. Read only the declared paths
   beneath the supplied root. Treat schema/proof incompatibility as fail-closed `HALT`.
4. REFACTOR: separate parsing, validation, identity construction, and evaluation so that verdict
   ordering is deterministic and evidence matching remains auditable.

### Slice 2 — canonical four-harness registration and projections

1. RED: require an OpenCode harness and its two projections; require the shared certification gate
   input/verdict block in all four projections; make check reject a missing or drifted projection.
2. GREEN: extend the canonical contract, renderer, strict projection checker, and generated
   projection list. Run `go run ./cmd/agentcontractgen sync` only after the source is canonical.
3. REFACTOR: retain a single render block for all harnesses so no adapter-local gate schema can
   diverge. Keep generated files generator-owned.

### Slice 3 — workflow enforcement and delivery validation

1. Add state-boundary helpers that require the gate before `integrate_sub_pr`, `accepted`,
   `ready_parent`, or `human_ready`, preserving the exact verdict/failure IDs for Shepherd input.
2. Prove that the current GitHub baseline produces `RETRY` while
   `go run ./cmd/agentcontractgen check` remains green without trying to certify every connector.
3. Run focused tests, projection smoke/drift tests, formatting, static/build checks, individual
   repository verification gates, inline GSD verify/gap handling/review, then `no-mistakes` without
   `--yes`. At most five production/test correction rounds are permitted.

## Verification plan

- `go test -timeout 20m ./internal/agentcontract` and `go test -timeout 20m ./cmd/agentcontractgen`.
- `go run ./cmd/agentcontractgen sync`, generated four-harness smoke tests, and
  `go run ./cmd/agentcontractgen check`.
- `gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, and `git diff --check`.
- Individual local gates: `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`.
- Changed-path audit confirming no `cmd/connectorgen/certification*.go`, connector definition,
  provider, credential, or evidence mutation.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. Red evaluator/projection tests with recorded failing output.
3. Green canonical contract, evaluator, generated projections, and focused tests.
4. Verification/review/no-mistakes fixes and final delivery record.
