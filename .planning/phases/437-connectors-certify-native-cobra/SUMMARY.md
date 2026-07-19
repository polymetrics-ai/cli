# Phase 437 Summary

Status: seventh bounded correction reopened; verification false until new RED/GREEN/full gates pass; stacked PR #466 remains open with human/parent review fallback pending.

## Identity

- Session: `issue-437-sol-high-review-correction-20260719`
- Profile: Sol/high
- Branch: `refactor/437-connectors-certify-native-cobra`
- Original exact start/base: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` / `feat/cli-architecture-v2`
- Fourth-correction exact start: `1e27b14012f65ffa24c01ed855d0405c24401eee`
- Parent #397, umbrella #407, draft parent PR #438

## Delivered before this correction

`connectors` is a native Cobra subtree with `list`, `catalog`, `inspect`, `man`, `docs`, hidden positional help, and nested `certify`. Single, batch, and sweep certification use invocation-local runtime seams while preserving in-process execution, report rendering, telemetry, cancellation, bounded workers, events, and exit mapping. Canonical help, generated CLI docs, golden transcripts, and website data describe the bounded certification surface.

Two prior review corrections were completed, verified, committed, and pushed. They made unsupported controls fail closed, restored execution ordering and help behavior, made batch write-disable controls dominate credential-file writes, gated writes on sandbox, rejected unsupported credential-file limits, constrained skip values, and corrected stale docs. Their recorded full gates passed at their respective heads.

## Third accepted correction

All findings in `/tmp/pm-397-rereview2-437.log` are accepted:

- certify subtrees must reject every unknown flag before credential loading, runners, sweep, or effects, including write-like typos;
- sweep age must be strictly positive and reasonably bounded;
- ordinary completed prior reports must be reusable with `--resume`, without fabricated future timestamps, while incomplete reports rerun;
- credential-file `exec` must remain prohibited and reject before effects; generic external execution code and claims must be removed;
- usage exit docs, release-stage token examples (`ga`), flag/docs audit, and terminal planning state must be accurate.

The plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state were reopened before RED tests or production changes. RED reproduced every finding. GREEN rejects unknown certify flags and unsafe ages before effects, rejects credential-file exec with no external execution path, reuses valid completed reports on ordinary resume runs while rerunning incomplete reports, and corrects canonical/generated/golden/website docs.

Focused, repeated, race, resume/sweep/no-effect, and flag/docs audit tests pass. Runtime help parity, invalid-action/typo exits, docs generation, golden transcripts, website hash-stable regeneration, and credential-free local sample certification pass. Full CLI passed in `446.382s`; full certify passed in `350.637s`; gofmt, clean diff, vet, full tests, and build pass. `make verify` exited 0 in `14m58.384s`, and explicit connectorgen validation checked 547 bundles with zero findings.

## GSD / skills / execution decision

`scripts/gsd doctor` and `scripts/gsd list` passed. `scripts/gsd prompt programming-loop ...` is absent from the adapter registry, so the manual universal runtime loop is the recorded fallback. Execution decision is `local_critical_path`: one bounded correction in the existing isolated issue worktree, no subagent tool, and no credentials, external commands, services, dependencies, PR, or review.

Loaded skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-context`, `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-spf13-cobra`, and `golang-lint`.

## Safety

Implementation and verification used fixture/temp/in-process paths and the repository's existing ordered local smoke only. No credential values, credentialed connector checks, external credential commands, live services, external writes or sweeps, new dependencies, generic tools, destructive/admin actions, production changes, PR, or review were used.

## Fourth bounded review-correction cycle

Start: `1e27b14012f65ffa24c01ed855d0405c24401eee`, clean and equal to the local/remote active branch. Launcher: `openai-codex/gpt-5.6-sol`, thinking `high`. Inputs: independent correctness and security exact-head reviews named in PLAN.md. Every overlapping item is accepted and consolidated into F1–F10: preview/approval gating; secret-safe rendering/config/report modes; invocation-local crontab isolation; durable provenance-bound ledger/sweep; context/cancellation with bounded post-mutation cleanup; strict bounded credential files; strict boolean/parallel/age controls; prerequisite DAG; resume identity/fingerprint; and temp-only tests.

GSD doctor/list and plan prompt passed; programming-loop remained absent, so the manual universal loop applied. Execution stayed `local_critical_path` because the user prohibited subagents and constrained work to this isolated issue branch. Planning `07d0b5a4`, RED `43acd262`, GREEN `2c0a550c`, and lint/resource-fix `b06816ad` are pushed.

All accepted groups are corrected: every reverse mutation now requires a successful nonleaky preview; secret detection is opaque and report-safe; credential config and files fail closed; reports/progress/ledgers are private and atomic; crontab selection is invocation-local; durable ledgers are provenance-bound and consumed from fresh temporary sweep workspaces; context reaches nested CLI calls and cancellation permits only bounded cleanup after a successful mutation; controls, workers, ages, and prerequisites are bounded/gated; resume requires exact schema/manifest/effective-options identity; and tests cannot pollute the source tree. Final-component symlink races, ledger tag authority, file descriptors, and sweep source-preparation failures were also closed during refactor.

Verification passed: focused ×10 and race matrices; standalone CLI/certify checkpoints and schedule tests; runtime help/bare/flag-help parity, invalid exit 2, JSON manual, credential-free sample; temporary CLI docs generation, goldens, website drift, docs validation; connectorgen 547/0; gofmt/diff/vet; explicit final-code `go test ./...` in real `456.93s` (CLI `452.912s`, certify `346.633s`); build; and final `make verify` exit 0 in real `464.41s` (CLI `439.981s`, certify `330.355s`, lint 0). The first verify attempt exposed unchecked fixture writes at lint; `b06816ad` corrected them and the entire gate was rerun successfully.

Safety remained fixture/fake/in-process/temp-only plus the repository's existing local smoke. No credential values, credentialed/live checks, external credential commands, system crontab, external writes/sweeps, services, connector-def changes, dependencies, generic write tools, reverse execution outside fake/temp/local smoke, PR, integration, parent mutation, or main merge occurred.

## Fifth bounded review-correction cycle — reopened

Status: complete; full `make verify` passed and `verificationPassed=true`.

Identity: `issue-437-fifth-review-correction-20260720`; exact clean/matched start `05d9c6658f52e542b6a74e87e29bdcad7275ea9d`; launcher `openai-codex/gpt-5.6-sol`, thinking `high`; no subagents. The correctness/security rereviews were consolidated into seven accepted findings: post-verification cleanup ledger mutation; tag-safe sweep authority with numeric GitHub resources denied; bounded secret-safe ledger loading; pre-effect sweep credential-constraint validation; surfaced current/history report persistence with leak-dominant precedence; pre-logger/telemetry certify syntax validation; and structurally complete outcome-recomputed resume evidence.

Recovery-budget exception: unresolved P1 destructive authority and retryability findings justify this additional narrowly bounded correction cycle. GSD doctor/list passed, the required programming-loop is still absent, the applicable audit-fix dry-run prompt was generated, and the manual universal loop remains the fallback. Execution is `local_critical_path` because the user prohibited subagents and parent/PR/integration work.

Planning `8acf62a9`, RED `e2559f64`/`3d69b7a4`, and GREEN `e9ce945e` are pushed. Focused RED captured every accepted finding before production: premature cleaned state, numeric-resource effects, unbounded/reflected ledgers, ignored sweep constraints, silent persistence failure, invalid-flag telemetry creation, and trusted minimal/edited resume artifacts. GREEN defers ledger cleaning until exact verification, verifies sweep absence and denies numeric authority, bounds/sanitizes ledgers, validates sweep constraints before telemetry/workspace, surfaces persistence failure with leak dominance, prevalidates complete certify syntax before effects, and accepts only structurally complete outcome/leak-consistent resume evidence.

Verification passed for the prior fifth cycle: focused/repeated/race; full CLI `443.427s` and certify `327.840s`; schedule/safety; runtime help/bare/invalid/JSON; docs/goldens `16.600s`; website drift-free; connectorgen 547/0; gofmt/diff/vet; explicit `go test ./...` real `7m34.316s`; build; and final `make verify` exit 0 real `7m52.496s` (CLI `449.572s`, certify `332.793s`, lint 0). Safety remained fake/temp/effect-recorder and existing ordered local smoke only, with no credentials, live services/connectors, system crontab, external credential commands/writes/sweeps, connector defs, dependencies, generic write tools, PR, parent, integration, or main mutation.

## Continuation — parent reconcile and stacked PR

Session `issue-437-continuation-20260719T211738Z`; exact start `86eea0f966814e6848e5a52143eea15dd46ff801`; latest parent `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa`. `scripts/gsd doctor` passed, `scripts/gsd prompt plan-phase 437 --skip-research` generated the official prompt, and `programming-loop` remains unavailable (`unknown GSD command`), so manual universal-loop fallback applies.

Plan/TDD/verification were refreshed before parent merge or production edits. Diff audit is clean: 42 files are limited to #437 phase artifacts, native connectors/certify CLI code/tests/goldens, certify safety code/tests, the `cmd/pm` in-process harness seam, and required docs/website parity; no dependency or connector-def delta. Certify safety corrections are directly applicable to the native Cobra certify surface because declared flags, credential files, reports, ledgers, sweep, resume, and cleanup authority must fail closed before effects. Latest parent `a5474bcb` merged cleanly into `dc4aed23`; #462 design docs/skills/traces retained and not in the stacked PR diff. Verification passed: focused CLI `119.151s`, certify `7.344s`; runtime help byte-equal (`8391` bytes); docs/golden `10.347s`; website docs data regenerated with no drift; credential-free `./pm connectors certify sample --root <tmp> --json` exit 0/pass/stderr 0; gofmt/diff/vet/full tests/build; `make verify` exit 0 with local Makefile smoke, lint `0 issues`, and connectorgen green; explicit connectorgen validation 547/0. Delivery complete for worker scope: branch pushed and non-draft stacked PR opened at https://github.com/polymetrics-ai/cli/pull/466 targeting `feat/cli-architecture-v2`. CI queued on PR creation. Claude is disabled and Copilot quota is exhausted, so human/parent fallback review coverage remains pending. No new behavior edit was needed; existing RED evidence remains valid.


## Sixth bounded exact-head review-correction cycle — reopened

Status: planning checkpoint in progress; verification reopened and false until full gates pass.

Identity `issue-437-sixth-review-correction-20260719T220843Z`; exact clean/matched start `8e7e2533c75451114c4d6ae38f89b7fd1ede6c34` on branch `refactor/437-connectors-certify-native-cobra`, PR #466. Accepted findings require another bounded recovery-budget exception: unresolved High approval-replay resource authority and leak-dominance failures can leave live resources untracked or mask exit 3; Medium stale-leak consistency can make reports internally contradictory.

Planned fixes: move approval idempotency before final cleanup or otherwise make it cleanup-covered; preserve leak-dominant exit 3 and safe batch evidence when runner/progress errors coexist with leaks; clear stale top-level leak claims when exact cleanup verification proves absence while keeping the cleanup stage failure honest.

GSD: `scripts/gsd doctor` passed; plan-phase prompt generated; `programming-loop` remains absent, so manual universal-loop fallback applies. Execution decision is `local_critical_path`; this worker has no subagent tool. Parent orchestrator records the invocation as spawned.

Safety: tests/fixes limited to local fakes, temp roots, sample/outbox fixture path, and existing CLI/certify code. No credentials, live connector checks, services, external writes/sweeps, dependencies, generic write tools, bot review requests, PR merge, parent mutation, or main merge.


### Sixth-cycle RED evidence captured

- `go test ./internal/connectors/certify -run 'TestRereviewApprovalReplaySuccessIsCoveredByFinalCleanup|TestRereviewCleanupFailureThenAbsenceProofClearsStaleLeak|TestRunBatchRunnerErrorWithLeakedReportKeepsExit3' -count=1` failed as intended in `25.435s`: approval replay success left final outbox action `"create"`; cleanup absence proof left stale top-level `Report.Leaks`; runner report+error with a leaked report produced batch exit `2` instead of `3`.
- `go test ./internal/cli -run 'TestRereviewBatchProgressErrorWithLeaksEmitsReportAndExit3' -count=1` failed as intended in `0.572s`: progress persistence error with leaked batch returned exit `1` and emitted an `Error` envelope instead of the safe `ConnectorCertificationBatch` evidence with leak-dominant exit `3`.


### Sixth-cycle GREEN focused evidence

Implementation reorders approval idempotency before final cleanup, skips replay checks unless `write_create` actually consumed the plan, removes stale leak records when exact cleanup verification proves absence, preserves the cleanup stage failure as a non-leaked write-action failure, and lets leaked runner reports dominate ancillary runner/progress errors.

Focused GREEN commands:

- `go test ./internal/connectors/certify -run 'TestRereviewApprovalReplaySuccessIsCoveredByFinalCleanup|TestRereviewCleanupFailureThenAbsenceProofClearsStaleLeak|TestRunBatchRunnerErrorWithLeakedReportKeepsExit3' -count=1` => pass (`21.288s`).
- `go test ./internal/cli -run 'TestRereviewBatchProgressErrorWithLeaksEmitsReportAndExit3' -count=1` => pass (`0.565s`).
- Repeated: same certify focus `-count=3` => pass (`62.605s`); same CLI focus `-count=3` => pass (`0.567s`).
- Race: same certify focus with `-race` => pass (`229.386s`); same CLI focus with `-race` => pass (`1.666s`).
- Coherent-slice gates: `go vet ./...` => pass; `go build ./cmd/pm` => pass.

A prior repeated attempt timed out before the approval-replay test preserved invocation-local crontab options in its wrapper; the wrapper was corrected and the focused, repeated, and race gates above reran successfully.


### Sixth-cycle terminal verification evidence

Implementation head before terminal artifacts: `791b1a1d` (`52cd1e05` GREEN plus `791b1a1d` lint/test close fix). All accepted findings are corrected.

Verification commands/results:

- Focused affected packages: `go test ./internal/connectors/certify -count=1` => pass (`351.189s`); `go test ./internal/cli -count=1` => pass (`445.489s`).
- Formatting/full gates: `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; `go test ./...` => pass (CLI `445.908s`, certify `350.792s`); `go build ./cmd/pm` => pass.
- Runtime help parity: `./pm help connectors`, bare `./pm connectors`, and `./pm connectors --help` byte-identical (`8391` bytes).
- Docs/golden/website parity: `go test ./internal/cli -run 'TestConnectorsManual|TestNativeConnectors|TestNativeCertify' -count=1` => pass (`11.149s`); `cd website && node scripts/gen-docs-data.mjs` wrote 11 pages; tracked docs/website/golden diff clean. No canonical help/manual/website text change was required by this fix.
- Fixture-only sample smoke: `./pm connectors certify sample --root <temp> --json` => exit `0`, kind `ConnectorCertification`, connector `sample`, passed `true`, stderr bytes `0`.
- Connector validation: `go run ./cmd/connectorgen validate internal/connectors/defs` => `547 connector(s) checked, 0 findings`.
- `make verify`: first attempt failed only at certify lint (`errcheck` on `f.Close` in the new replay fixture helper); `791b1a1d` fixed it. Complete rerun passed: CLI `453.202s`, certify `357.136s`, docs validate, ordered local smoke `smoke ok`, lint `0 issues`, connectorgen `547 connector(s) checked, 0 findings`.

Safety: no credentials, live certification, services, system crontab, external writes/sweeps, dependencies, connector defs, generic write tools, bot review request, parent/main merge, or quality-gate reduction. The only reverse ETL execution was the repository's local fixture/temp smoke path (plan → preview → approval → execute) inside `make verify` and the explicit sample/outbox certification smoke.

## Seventh bounded correction — reopened

Status: planning checkpoint in progress; `verificationPassed=false` until seventh full gates pass.

Identity: `issue-437-seventh-bounded-correction-20260720`; exact clean/matched start `6e9e7d9422050a609306d8900d6a06c8bb1fc223` on branch `refactor/437-connectors-certify-native-cobra`, PR #466. Local branch, remote branch, and PR head were confirmed equal. Sixth-cycle stale pending fields are corrected: PR #466 body already records head `6e9e7d9422050a609306d8900d6a06c8bb1fc223`; seventh terminal evidence is not claimed yet.

Accepted findings: `validResumeEvidence` rejects the no-leak cleanup-failure/absence-proof shape introduced by the sixth cleanup fix, causing `--resume` to rerun and repeat effects; certify prevalidation lets bare value-required flags become `true`, so single/batch/sweep paths can reach runner, credential loading, telemetry, or wrong validation category before usage rejection. Recovery-budget exception: unresolved effect-before-usage and resume-replay risks justify this additional bounded cycle.

GSD: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 437 --skip-research` generated; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable, so the manual universal-loop fallback applies. Execution decision: `local_critical_path`; parent records this worker as spawned.

Plan: commit/push planning first; add RED tests for the exact resume absence-proof case and bare required-value no-effect/exit-2 cases; implement minimal fixes in `internal/connectors/certify/batch.go` and `internal/cli/certify_cli.go`; run focused/repeated/race, full affected packages, runtime help/no-effect/sample smoke, docs/website if needed, gofmt/diff/vet/full tests/build/`make verify`, and connectorgen; update PR #466 body with final head. No bot review request or merge.
