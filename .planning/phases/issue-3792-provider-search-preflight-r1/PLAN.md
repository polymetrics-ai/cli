# PLAN — issue #3792 provider-search runtime preflight

Issue: #3792 (parent #3769)
Branch: `fm/cli-found-provider-operations-r1`

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- `scripts/gsd prompt discuss-phase 3792`: executed inline; fixed decisions are in `CONTEXT.md`
  and `DISCUSSION-LOG.md`.
- `scripts/gsd prompt plan-phase 3792 --tdd`: executed inline in this plan.
- `scripts/gsd prompt execute-phase 3792`, `verify-work 3792`, and `code-review 3792`: execute
  inline after the applicable slices and record their results in this phase directory.
- Inline/manual fallback: compatible isolated GSD roles are unavailable and the canonical
  single-worker/no-spawn contract forbids role spawning. This fallback does not change TDD,
  verification, review, or human gates.

## Required skills loaded

- `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`,
  `golang-context`, and `golang-cli`.

## Scope and ownership

Only these production seams may change:

- `internal/connectors/commandrunner/runner.go`: `validateOperationDirectReadCommand` and its
  immediately necessary consumer-owned no-network preflight interface.
- `internal/connectors/engine/direct_read.go`: operation-direct-read admission/preflight, factored
  so runtime dispatch and preflight use the same executable checks.
- Focused tests in the corresponding package test files.

The current defect is that `validateOperationDirectReadCommand`
(`commandrunner/runner.go:518-538`) checks only a generic reader, one relative GET/POST endpoint,
and a supported policy. It does not load or inspect the referenced operation; the executor later
rejects unsupported kinds (`engine/direct_read.go:32-67`).

## Blocking result — 2026-08-06

This issue cannot be delivered independently in the current repository state.
A strict, no-network candidate was assessed and then completely reverted (no
production change is retained): its focused tests passed, but the real
`TestEveryImplementedCommandPassesRuntimePreflight` sweep reported
`178 of 1239 commands marked "implemented" fail runtime Preflight`.
That is an acceptance failure, not a tolerable migration warning.

The candidate cannot be narrowed to bundle connectors or to `provider_search`
alone: that would leave other operation-backed implemented commands on the
generic-reader preflight, contradicting the issue's required shared eligibility
contract and real sweep. Requiring the local contract from every reader instead
blocks native executable readers that do not expose the required operation
metadata. Details and next dependencies are recorded in `VERIFICATION.md`.

## Slice A — RED: prove generic-reader preflight is insufficient

1. Add a commandrunner regression using a fake that implements the proposed no-network preflight
   method and returns a deterministic unsupported/mismatched-operation error.
2. Assert `Preflight` rejects the implemented operation-backed command before `OperationDirectRead`
   is called. On baseline this test must fail because the current validator never invokes the
   method.
3. Add engine table cases against the existing provider-search fixture for unsupported kind,
   mismatched method, mismatched path, mismatched policy, and missing/non-positive response cap.
   No HTTP server call is required for these admission cases.

## Slice B — GREEN: bind the preflight to the loaded executable operation

1. Introduce a small commandrunner-local interface for the no-network operation preflight. Do not
   modify the shared `connectors` interfaces/types.
2. Implement the method on `*engine.Connector` in `engine/direct_read.go`, avoiding
   `engine/connector.go` (the #3740 collision path).
3. Factor existing operation-direct-read admission checks into one engine helper used by both the
   executor and preflight. The helper accepts only `rest_read` or landed `provider_search`, requires
   the exact declared REST method/path, a connector-relative endpoint, matching API-surface row,
   and a positive cap. The preflight also requires exact command operation, method, path, and output
   policy agreement.
4. Keep body-schema validation, the existing deadline, clamp, and no-network-on-invalid-input
   behavior unchanged. Do not add a second provider-search executor.

## Slice C — REFACTOR and integration

1. Confirm a valid existing provider-search test fixture reaches the same admission helper and a
   valid commandrunner fake receives the exact operation/body parameters.
2. Run the actual `TestEveryImplementedCommandPassesRuntimePreflight` sweep. It must now exercise
   the engine's no-network method rather than a validator duplicate.
3. Inspect the changed paths to prove no #3788/#3797, #3771, #3852, or #3740-owned files changed.

## Verification plan

- RED/green focused: `go test ./internal/connectors/commandrunner -run 'OperationDirectRead.*Preflight|Preflight.*OperationDirectRead' -count=1` and `go test ./internal/connectors/engine -run 'ProviderSearch|OperationDirectRead' -count=1`.
- Package regression: `go test ./internal/connectors/commandrunner` and `go test ./internal/connectors/engine`.
- Runtime sweep: `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`.
- Adjacent CLI package only: `go test ./internal/cli -run Connector -count=1`.
- Formatting/static/build: `gofmt -w internal/connectors/commandrunner/runner.go internal/connectors/commandrunner/runner_test.go internal/connectors/engine/direct_read.go internal/connectors/engine/direct_read_test.go`; `go vet ./internal/connectors/commandrunner`; `go vet ./internal/connectors/engine`; `go build ./cmd/pm`.
- Repository gates individually: `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` as permitted by the no-live/no-credential task scope.
- Do not run aggregate `go test ./...` or `make verify` locally; CI owns the timeout-prone 550+ connector suite.
- CLI help/manual/website parity: N/A. No command declaration, flag, help text, output format,
  generated surface, docs/manual, website page, or bare namespace behavior changes. Confirm through
  changed-path inspection.

## Commit checkpoints

1. Plan/context/TDD artifact checkpoint.
2. RED test checkpoint retaining observed failure.
3. GREEN runtime-admission/checkpoint with focused and package tests.
4. Verification/review artifact checkpoint.

## Safety boundaries

- No production provider declaration, citation, `connectorgen` rule, schema, bundle, capability,
  query interface, redaction policy, or redaction implementation changes.
- No credentials or live provider calls; admission tests are pure and execution tests use existing
  `httptest` fixtures only.
- Do not modify any function owned by #3771/#3775; do not run no-mistakes until firstmate directs
  the post-commit validation stage.
