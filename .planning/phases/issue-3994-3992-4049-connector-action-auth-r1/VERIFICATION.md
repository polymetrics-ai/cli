# Verification checklist

- [x] Corrected R1-R8 have retained RED and observable GREEN evidence; superseded per-fire grant
      evidence is labelled and removed from production.
- [x] Every named refusal asserts its typed error and zero sends/writes/files/sentinels/checkpoints.
- [x] Production call chains are exercised from a fresh `cmd/pm` binary.
- [x] Available GitHub/PostgreSQL live evidence is recorded without credentials; unavailable R2
      destination proof names the source-only base contract and excluded owner issues.
- [x] Focused package and selected race tests pass with `-timeout 20m`.
- [x] Changed Go files are gofmt-clean; focused package builds/tests, selected race coverage, and
      diff checks pass. Broad duplicate compilation was stopped under shared-machine load above 500
      after the same-commit focused proofs had already passed, per the focused-test task constraint.
- [x] CLI help/manual/website parity checks pass.
- [x] Derived artifacts are regenerated in one final pass and drift checks leave only intentional
      committed task changes.
- [x] Inline `verify-work` and `code-review` complete with no unresolved actionable findings.
- [x] PR #4168 is open and the API reports base `integration/4015-mvp-flat-r1`.
- [x] PR #4168 was rebased onto moved base `6b9cfe492`; its failed `certify-timing` job was
      reproduced locally, corrected in the branch-owned certification harness, and rerun green.

## Production-path evidence

- Prepared scheduled action: `cmd/pm` -> `cli.Run` -> `flowRun` ->
  `schedule.BeginFire` -> `resolveManifestJobs` -> `flow.Engine.Run` ->
  `connectorFlowActionRunner.ExecuteStep` -> `App.ExecuteAuthorizedFlowAction` ->
  `PrepareAuthorizedFlowAction` -> `ExecutePreparedFlowAction` -> connector
  `Write`/`Read` -> `FireLease.Complete` or `Park`.
- Shared rate budget: `cmd/pm` -> `cli.Run` -> connector command ->
  `commandrunner` -> `engine.Runtime.RequesterFor` -> `rateLimitResolver` ->
  `RateBudgetRefusalError(shared_coordinator_unavailable)` before transport.

`TestPMBinaryExecutesInstalledApprovedJobFlow` executes the exact installed crontab argv and
observes destination read-back, a terminal schedule receipt, and its `pex_` identity.
`TestPMBinaryRefusesRequiredSharedRateBudgetBeforeSend` observes the stable JSON policy code and
zero HTTP sends from a fresh binary.

The production certification route now also proves the no-carrier contract: `stageScheduleRoundtrip`
invokes `schedule create` without `--authorization`, rejects authorization/approval/credential
reference markers in every schedule envelope and installed crontab, verifies the direct rooted
`flow run` payload, then proves sentinel removal and byte-identical backend restoration.

## PR #4168 verify gap closure

- RED: GitHub run `31886502846` and local
  `TestCertifyCLISingleConnectorPassExitsZero` failed because certification still supplied
  `--authorization auth_0123456789abcdef` and demanded `authorization_reference` output.
- Root cause: the production schedule model had been corrected, but
  `internal/connectors/certify/stages_glue.go` and its scripted driver retained the superseded
  schedule-authorization proof. This is branch-owned and unrelated to #4167.
- GREEN: the exact fresh-binary CLI test passes; the full focused certification package passes;
  `go run ./cmd/certifytiming --max-duration 3m30s` passes both packages in 33.032s test elapsed
  (48.458s wall elapsed).
- Rebase coverage: focused app/flow/schedule/CLI authorization tests, rate-refusal packages, and the
  prepared-execution race test all pass on base `6b9cfe492`.

## Live-evidence boundary

No GitHub certification credential or database-integration runtime variable was present. No secret
was requested or printed. The fresh-binary proofs therefore use the real production composition
with local deterministic providers; issue #4166 owns the separate credentialed live-certification
coverage. The PostgreSQL destination gaps #4125 and #4158 remain explicitly out of scope.

## Drift and parity

The final generation pass ran `connectorgen surface-sync`, CLI docs generation, golden transcript
generation, and website docs generation together. `surface-sync --check` subsequently scanned 552
bundles with zero changes. CLI docs validation, golden transcript comparison, website docs diff,
agent-contract projection, GitHub parity, certification matrix, connector boundary/canon, module
tidy, focused vet/lint, `git diff --check`, and a clean `git status` confirmed no derived drift.
