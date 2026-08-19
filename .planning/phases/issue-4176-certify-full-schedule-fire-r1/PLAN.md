# Issue #4176 — Full certification must execute flow and scheduled firing

## Task Delivery Header

- Issue: Refs #4176 — fix(certification): make full runs cover flow and installed schedule firing.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: committed branch, pushed, with a pull request open against `integration/4015-mvp-flat-r1`; its API-reported base must match exactly.
- Working branch: `fm/cli-certify-full-omits-flow-schedule-r1`
- Task: make `pm certify --full` a superset of the ordinary certification stage set; execute the installed schedule through `pm schedule fire`, observe its flow result and restore the redirected backend byte-for-byte; explicitly distinguish the hermetic backend execution from an unobserved external scheduler daemon.
- Verification: red-first focused certification tests and the real CLI construction-path test; targeted package test/vet/build; derived checks and GSD workflow verification; post-open GitHub API base read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Full certification contains all ordinary stage names and adds full-only coverage | live | Two runner reports are compared after real Runner construction; each ordinary stage must be present in the full report and a full-only sweep stage must also be present. |
| An installed schedule executes its stored flow | live | The certification run invokes `pm schedule fire` after `install`; the returned `ScheduleFire` flow status and durable fire receipt are asserted before removal. |
| Backend state is restored | live | The redirected crontab fixture is byte-for-byte equal to its pre-install snapshot after remove. |
| A refused/unavailable schedule fire cannot pass | live | A test sabotages the install assertion and asserts that `schedule_fire` refuses before its own CLI invocation, names the refusal, and makes the report fail before backend removal is recorded as success. |
| A real scheduler daemon fired the installed backend entry | fake | The test uses the existing redirected crontab backend to protect the operator's scheduler; no daemon/credentialed scheduler is authorized in this worktree. The report will call this boundary `not_live`, never `pass`. |

## Scope and ownership

- Target connector: `sample` for hermetic certification mechanics. It exercises the shared `internal/connectors/certify` runner through the real in-process `pm` CLI construction path.
- Production paths: `internal/connectors/certify/{stages_source.go,stages_glue.go,report.go}` and directly related tests only.
- No credentialed provider run, no external scheduler activation, no connector definition change, no new dependency, and no change to #4166's validation-only scope.

## GSD lifecycle and skills

- Resolved and executed inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` via `scripts/gsd`; `go run ./cmd/agentcontractgen check` passed before planning.
- Inline/manual fallback: the canonical contract permits one delivery worker only and this runtime has no compatible isolated Pi workers. This records the fallback; it does not waive the lifecycle.
- Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-documentation`.
- CLI help/manual/website parity: not applicable. `--full`, `schedule fire`, report output, and their documents already exist; this change only corrects certification execution and its report semantics. Runtime help, docs, website, and generated artifacts will be checked for unintended surface drift.

## TDD slices

1. **Full stage-set superset (happy path).** Add a real Runner report comparison that fails on the baseline because `flow_roundtrip` and `schedule_roundtrip` are absent from Full mode. Green makes ordinary tail glue stages execute in both modes and preserves full-only sweep work.
2. **Installed fire observation (happy path).** Add an assertion for an actual `schedule fire` invocation, successful terminal flow status/receipt, and byte-identical cleanup. Green inserts the fire after install and before remove.
3. **Silent-pass refusal (bad path).** Add a named sabotage/refusal test that makes install invalid and asserts `schedule_fire`, `schedule_roundtrip`, and `Report.Passed` are false. It must prove zero fire invocations before the refusal and must not classify cleanup as a pass.
4. **Hermetic scheduler boundary (edge case).** Add a test for an empty pre-existing crontab snapshot; after fired schedule removal it remains byte-identical. The report’s external-daemon boundary must be `not_live` with a reason, not `pass` plus a non-live excuse.

## Verification matrix

| Area | Command | Required result |
| --- | --- | --- |
| Red / green certification | `go test -timeout 20m ./internal/connectors/certify -run 'Test(FullCertificationStageSetIsStrictSuperset|GlueStages.*Schedule)' -count=1 -v` | Baseline red names absent full glue stages; green proves stage set, fire flow receipt, refusal, and cleanup. |
| Affected package | `go test -timeout 20m ./internal/connectors/certify -count=1` | Pass. |
| CLI construction path | `go test -timeout 20m ./internal/cli -run 'TestCertifyCLISingleConnectorPassExitsZero' -count=1 -v` | The shipped CLI reaches the certification runner. |
| Static / build | `gofmt -w` changed Go files; `go vet ./internal/connectors/certify ./internal/cli`; `go build ./cmd/pm`; `git diff --check` | Pass. |
| Derived checks | individual relevant `make` gates including `agent-contract-check` and certification/surface checks | Pass. |
| Workflow / review | `scripts/verify-gsd-workflow`; inline verify-work and code review records | Pass with red/green evidence and no silent-skip finding. |

## Commit checkpoints

1. Planning artifacts and red-test checkpoint.
2. Green implementation and focused proof.
3. Verification/review artifacts and any in-scope fixes.
