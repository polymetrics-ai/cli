# Phase 437 Plan — connectors and certify native Cobra

Current status: ninth bounded CI flake correction complete; `verificationPassed=true` for the full local rerun at test-fix implementation head `828be4de4145d1246347b820d433d52bd1e92002`. PR #466 CI run `29711194607` at exact head `9f004ac5d96d84bd1f8b186496e1f594a183a18b` failed only in `internal/connectors/certify`:
`--- FAIL: TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit (0.25s)` / `batch_test.go:372: elapsed = 252.003235ms, want well under 3x80ms serial time (parallelism not happening)` / `FAIL polymetrics.ai/internal/connectors/certify 638.060s`. Parent base `c91b90cf9671b5caabc0ef4ec24d81897f870458` was reconciled by merge `9678d4dda2fcf331b3199f042804001c06eccf64` with disjoint #397 artifacts retained.

Issue: #437
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/437-connectors-certify-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`
Invocation: `issue-437-pi-sol-high-20260719T095145Z`; profile `Sol`; thinking `high`.
Execution decision: `local_critical_path` — this is the final assigned serialized Phase 9 namespace in an isolated worktree; router changes collide with sibling units, no subagent tool is exposed, and the user requested implementation/commit/push with no PR/review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: pass (69 commands).
- `scripts/gsd prompt plan-phase 437 --skip-research`: generated and executed inline.
- Required `programming-loop` command is absent from the adapter registry (`unknown GSD command`), so the manual fallback is `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` with these six pre-production artifacts and strict RED → GREEN → refactor.

## Required context and skills

Read issue #437 and parents #407/#397; `AGENTS.md`; GSD/manual/issue/parent contracts; connector migration handoff, conventions, v2 design, certification design/contracts; CLI Architecture v2 plan/execution prompt; ADR-0002; CLI help/docs/website parity; current connectors/certify/router/golden/manual/website code.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## Scope

- Replace only the legacy top-level `connectors` wrapper and `runConnectors`/`runCertify` namespace parsers with native Cobra commands.
- Native connector actions: `list`, `catalog`, `inspect`, hidden positional `help`, and compatibility aliases `man`/`docs`; preserve current metadata-only behavior and output.
- Native nested certify actions: single connector, `--all`, and `--sweep`, declaring every currently consumed flag: `credential`, `from-env`, `config`, `stream`, `limit`, `modes`, `skip`, `rate-limit`, `budget`, `record`, `replay`, `live-all-modes`, `allow-production-writes`, `keep-workdir`, `write`, `full`, `credentials-file`, `parallel`, `resume`, `older-than`.
- Preserve legacy repeated-last, bare, assigned, space, operand, ignored trailing, unknown-flag, literal `--`, malformed unknown, action/operand discovery, invocation-global, text/JSON, and error-category behavior where current handlers consume it.
- Preserve `cli.Run` in-process re-entrancy, certify exit 0/1/2/3 mapping, context cancellation, bounded cross-connector concurrency, event sequence, telemetry span names/status, and secret/credential-value exclusion.
- Keep dynamic `pm <connector> <path...>` dispatch and its legacy `parseFlags` path exactly sanctioned and unchanged.
- Certify verification is fixture/replay/local only. No live credential checks or writes.

Excluded: connector defs, connector runtime behavior, new certify semantics/flags, live tests, credential values, external services, new dependencies, dynamic parser removal, other namespaces, Phase 16 dashboard, Phase 19 help-tree churn, PR/review.

## TDD and checkpoints

1. Commit/push these six planning artifacts before test or production edits.
2. RED: add focused tests for native tree shape, complete current flags, connector/certify actions and operands, bare/text/JSON/topic/positional/trailing help, literal separator/malformed unknown/action discovery/globals, exact outputs/errors, exits 0/1/2/3, re-entrancy, cancellation/concurrency/events/telemetry, and planted credential-value absence. Capture failure caused by absent native constructors/runtime seam; commit/push.
3. GREEN: introduce the smallest typed flag structs, native constructors, handler adaptation/runtime seam, and compatibility normalization. Remove only connectors/certify namespace `parseFlags` calls and connectors legacy registration.
4. Refactor: focused ×10, race, router/golden/full CLI/certify; exact-base differential; connector validation; docs/manual/website generation; runtime smoke; gofmt/vet/test/build/make verify.
5. Finalize six artifacts, commit/push, no PR/review.

## Parity and safety

Bare `pm connectors` must render the canonical manual and exit 0; invalid actions remain usage exit 2. Update the canonical connectors manual to document certify commands and 0/1/2/3 exits, regenerate `docs/cli/connectors.md`, and mirror the bounded surface in `website/content/docs/cli-reference.mdx`; generated website data follows existing scripts. Completion registration is unchanged and Phase 15/19 work is not pulled forward.

All certify command tests use sample fixtures, replay/local fakes, `t.TempDir`, injected runner/sweeper seams, synthetic credential variable names, and planted non-secret sentinel values solely to assert absence. Never print, summarize, or persist a credential value. No real env-secret resolution, live connector check, write, sweep against external systems, reverse ETL execution, model/runtime service, dependency, or broad path.

## Completion

Implemented through three test-first slices: absent native constructors, contextual trailing action help, and direct inspect help before private operand capture. Final implementation keeps dynamic connector parsing unchanged, passes exact-start operations 21/21, focused/repeated/race/router/golden/full CLI/certify gates, runtime help/docs/website parity, explicit local sample certify smoke, connector validation 547/0, and final `make verify`. Delivery remains commit/push only with no PR/review.

## Accepted review correction — exact HEAD `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`

Session `issue-437-review-correction-20260719T113319Z` reopens the phase for all five findings in `/tmp/pm-397-review-437.log`. The adapter still lacks `programming-loop`; `scripts/gsd doctor` passed and the manual universal runtime loop remains the recorded fallback. Execution decision is `local_critical_path`: one bounded safety-critical correction in the existing isolated issue worktree, no subagent capability, no services/credentials/PR/review.

Correction slices, in order:

1. **Artifacts/checkpoint:** mark verification false and record the accepted findings, skills, commands, and safety scope; commit and push before tests or production.
2. **RED:** add differential tests proving unsupported safety/mode controls cannot run as no-ops; single certify starts exactly one span before connector validation and option parsing, with connector validation precedence; batch loads the credential file before parsing `--parallel` and preserves byte-exact load/run wrappers; only exact connectors help tokens render manuals; docs distinguish CLI pre-report exits from completed-report exits. Commit/push failing tests before production.
3. **GREEN:** reject and unadvertise controls with no existing typed runner/stage support (`record`, `replay`, `allow-production-writes`, `rate-limit`, `budget`, `live-all-modes`); restore the legacy single and batch execution ordering/wrappers; tighten connectors-only help normalization; correct canonical/generated/website docs. Preserve write opt-in/cleanup gates and all existing redaction behavior.
4. **VERIFY:** focused differential, repeated/race, certify exits/redaction, unsupported replay no-live/runtime-call test, local sample smoke, docs/golden/website generation, full CLI/certify, gofmt/vet/test/build/`make verify`, and connector validation. Use fixture/temp inputs only.
5. **DELIVER:** finalize all six artifacts, coherent commits and pushes only; no PR or automated review per user instruction.

Correction completed at implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`. All six controls without typed runner/stage support are hidden and fail closed before any runner invocation; replay therefore cannot touch credentials or live/write stages. Legacy single span and connector-validation-before-option precedence, batch file-before-parallel precedence and exact load/run wrappers, and exact-only connectors help behavior are restored. Canonical/generated/website docs now separate pre-report CLI exits from completed-report outcomes. Differential 5/5, focused/repeated/race, redaction/replay/exit, local sample, docs/golden/website, full CLI/certify, gofmt/vet/test/build, `make verify`, and connector validation all pass.

Loaded for the correction: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-spf13-cobra`.

## Second accepted safety correction — exact HEAD `0d743e54e06c9e27e550eacce9be7899a9e23d19`

Session `issue-437-second-safety-correction-20260719` reopens the phase for all three priority groups in `/tmp/pm-397-rereview-437.log`; every finding is accepted. `scripts/gsd doctor` and `scripts/gsd prompt plan-phase 437 --skip-research` passed. The required `programming-loop` command remains absent (`unknown GSD command`), so the manual universal runtime loop is the recorded fallback. Execution decision: `local_critical_path` because this is one safety-critical correction in the existing isolated issue worktree, no subagent tool is exposed, and the user prohibited credentials, services, dependencies, PR, and review.

Second-correction slices, in order:

1. **Artifacts/checkpoint:** mark verification false and record accepted P1/P2/P3 findings, required skills, strict safety scope, RED/GREEN boundaries, full flag audit, verification matrix, and commit/push checkpoints. Commit and push before tests or production edits.
2. **RED/checkpoint:** add effect-recorder tests proving batch `--write=false` and `--skip=write` dominate credential entries with `write: true`; configured credential-file sandbox/rate/budget (and every other unsupported accepted control discovered by the audit) fail before runner effects; unsupported or mode-inapplicable flags and skip values fail before effects; every declared certify flag is either used or rejected. Add docs/help claim tests. Run focused tests to capture failure, then commit and push before production edits.
3. **GREEN P1/P2:** preserve safe base behavior while adding batch write-disable overrides; fail closed rather than discard unsupported credential-file constraints; hide/reject unsupported single `--credential`, `--limit`, and `--modes`; allow only implemented `--skip=write`; reject flow/schedule/unknown skip values and all mode-inapplicable controls before runner effects. Keep dynamic dispatch, certify exit mapping, telemetry, and redaction unchanged.
4. **GREEN P3:** remove stale rejected-option examples and claims from certification architecture and the universal-loop PRD; correct the connector-help name claim to match the namespace manual behavior; regenerate CLI docs, golden/help artifacts, and website generated data through repository generators only.
5. **VERIFY:** focused effect/no-op and flag-audit tests; repeated and race runs; local sample fixture smoke; full CLI/certify/docs/website gates; gofmt, vet, full test, build, `make verify`, and `connectorgen validate`. Fixture/temp inputs only; no credentials, live services, or external writes.
6. **DELIVER:** finalize all six artifacts, commit coherent green/verification checkpoints, and push only the active issue branch. No PR or review.

Loaded for this cycle: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-spf13-cobra`.

Second correction completed at implementation head `7b6eaa58`. Batch write-disable controls now dominate credential-file write settings; sandbox gates enabled writes; unsupported rate, budget, and read-limit constraints fail before runner construction. Unsupported single controls are hidden/rejected, skip is allowlisted to write, and every mode rejects inapplicable declared controls before runtime effects. Architecture/PRD/help claims and generated CLI/website data are accurate. Focused, repeated, race, no-op/effect, local sample, full CLI/certify, docs/website, gofmt/vet/test/build, `make verify`, and connectorgen gates pass without credentials, services, dependencies, PR, or review.

## Third accepted safety/correctness correction — exact HEAD `437d13cf`

Session `issue-437-third-safety-correction-20260719`; all findings in `/tmp/pm-397-rereview2-437.log` are accepted. `scripts/gsd doctor` and `scripts/gsd list` passed, but the required `scripts/gsd prompt programming-loop ...` route remains absent (`unknown GSD command`), so the manual universal runtime loop is the recorded fallback. Execution decision: `local_critical_path` because this is one bounded correction in the existing isolated issue worktree, no subagent tool is exposed, and the user prohibited credentials, external commands, services, dependencies, PR, and review.

Third-correction slices, in order:

1. **Artifacts/checkpoint:** set verification/state to honestly not-yet-verified and record accepted findings, skills, strict fixture/temp safety scope, RED/GREEN boundaries, verification matrix, and commit/push checkpoints before tests or production edits.
2. **RED/checkpoint:** add native-Cobra subtree tests proving every unknown flag—including write-like typos—returns usage before credential loading, runner, or sweep effects; add strictly-positive/reasonably-bounded `--older-than` no-effect cases; add credential-file `exec` rejection before runner effects; add a two-run batch effect-recorder proving ordinary completed reports resume while incomplete reports rerun. Update the flag/docs audit assertions. Capture and commit the failures before production edits.
3. **GREEN safety:** remove certify subtree unknown-flag whitelisting while preserving non-certify connector behavior only where required; validate sweep age before `runtime.Sweep`; reject any credential-file `exec` entry before batch effects; remove the generic external execution implementation and claims.
4. **GREEN correctness/docs:** make ordinary `--resume` reuse a valid completed prior report without timestamp fabrication, while incomplete/failed report artifacts rerun; correct CLI usage-exit docs, stage examples to `ga`, resume/sweep safety wording, canonical/generated/website docs, and architecture examples.
5. **VERIFY:** focused and repeated unknown-flag/no-effect/exec/resume/sweep tests; race; full flag/docs audit; docs/golden/website generation and drift; local credential-free sample smoke; full CLI and certify suites; gofmt, vet, full tests, build, `make verify`, and connectorgen validation. Fixture/temp inputs only; no external command invocation, credentials, services, or writes.
6. **DELIVER:** finalize all six artifacts with truthful terminal state, commit coherent GREEN/verification checkpoints, and push only the active issue branch. No PR or review.

Loaded for this cycle: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-context`, `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-spf13-cobra`, and `golang-lint`.

Third correction completed at implementation head `f56bc825`. Certify now rejects every unknown flag before credential loading or runner/sweep effects; sweep age is limited to greater than zero and at most 8760h; credential-file exec is rejected at load, resolver, and batch boundaries with the external execution implementation removed; and ordinary `--resume` reuses valid completed prior reports and their outcomes while incomplete reports rerun. Usage exits, `ga`, resume/sweep/exec wording, generated CLI docs, goldens, and website data are accurate. Focused, repeated, race, no-effect/audit, runtime help, docs generation, credential-free local sample smoke, full CLI/certify, gofmt/vet/test/build, `make verify`, and connectorgen gates pass. No external credential command, live credential, service, dependency, PR, or review was used.

## Fourth bounded review-correction cycle — exact HEAD `1e27b14012f65ffa24c01ed855d0405c24401eee`

Session `issue-437-sol-high-review-correction-20260719`; launcher model `openai-codex/gpt-5.6-sol`, thinking `high`. The isolated worktree, branch, clean tree, exact HEAD, and local/remote branch equality were confirmed before work. Independent exact-head inputs are `/tmp/pm-397-437-correctness-review.out` and `/tmp/pm-397-437-security-review.out`. They are review evidence only; every finding was traced through reachable production paths.

`doctor` and `list` pass with 69 commands. `scripts/gsd prompt plan-phase 437 --skip-research` generated successfully. `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable (`unknown GSD command`), so this cycle records the manual `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` fallback. Execution decision: `local_critical_path` because the user explicitly prohibited subagents and bounded work to this existing isolated issue branch; no parent branch or PR mutation is authorized.

Loaded skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`; additionally `golang-spf13-viper` for invocation-scoped configuration and `golang-troubleshooting` for evidence-first correction.

### Findings and planned dispositions

All overlapping findings are consolidated and **accepted** because the reviewed paths are reachable:

1. **Preview/approval gates:** initial write stores plan/token before preview success; create checks only plan ID; both cleanup stages and both sweeper paths omit preview. Add explicit preview-success state, clear tokens on failure, and use one plan → successful preview → approval → execute helper everywhere. A failed/mismatched/leaky preview must record zero execute/ledger effects.
2. **Secret-safe reports:** `ScanForSecrets` returns matched values; reasons persist them; approval argv is not semantically redacted; secret-schema config can enter argv; reports/history are `0644`. Return opaque hit metadata, redact sensitive flag operands independent of value registration, reject config on `x-secret` keys, and use atomic restrictive report writes. Serialized report tests plant marker values but never print report bodies.
3. **Crontab confinement:** schedule certification mutates process-global `PM_CRONTAB_FILE`, so parallel runs can fall through to system crontab. Replace this with context/invocation-local CLI configuration and deterministic barrier-controlled concurrent tests proving each runner only reaches its own temp file.
4. **Durable ledger/sweep authority:** write-ahead ledgers live in deleted workdirs; copy layout disagrees with sweep; sweep conflates ledger/project roots and trusts ledger connector/action/tag data. Write directly to the durable per-connector layout consumed by a fresh-process sweeper, separate ledger/workspace roots, preserve validated env references, and bind entries to connector/run/tag/action/cleanup provenance. Reject traversal, symlink, cross-connector, malformed, and unsupported ledger authority.
5. **Caller context/cancellation:** Harness calls context-free `cli.Run`, which creates `context.Background`; later stages continue after cancellation. Add a context-aware entrypoint while preserving `cli.Run`, propagate context through every in-process stage, stop before later effects, and use a bounded cleanup context only after an already-started successful mutation.
6. **Credential-file boundary:** YAML parsing is unbounded/permissive and accepts unsupported versions, empty jobs, invalid connector/env/config fields, traversal keys, and symlinks. Add bounded strict known-field decoding, version/job/count checks, registry/local identifier validation, env-reference validation, secret-schema config rejection, and no-follow regular-file checks before effects.
7. **Strict controls/resources:** boolean controls use string equality; explicit invalid values fail open; parallel can be zero/negative/huge and worker count ignores jobs. Strictly parse every supplied boolean, validate explicit parallel in a fixed positive range, cap workers by queued jobs, and preserve the existing bounded sweep age before credential loading/runners/sweep/writes; pre-telemetry argv safety validation applies to direct CLI entry.
8. **Prerequisite DAG:** preflight/manual/credential failures are observational and later live/read/write stages still run. Gate structural/preflight and credential checks before catalog/live reads and writes; successful preview gates execution.
9. **Resume compatibility:** current reuse validates only name/timestamps and accepts future schema. Persist and compare exact report schema, connector/manifest identity, and a secret-free effective-options/env-reference fingerprint; rerun on mismatch.
10. **Test pollution:** the native connector helper roots successful runs at `.`. Convert it to `t.TempDir()` and assert no source-tree `.polymetrics` artifact appears.

### RED → GREEN → refactor plan

1. Commit/push this six-artifact planning checkpoint before tests or production edits.
2. Add focused failing tests for each accepted behavior/security finding. Consolidate overlaps: one preview matrix covers initial/cleanup/sweep; one secret serialization matrix covers detector/reason/argv/config/report modes; one durable-ledger fresh-process matrix covers layout/provenance/containment; one strict-input matrix covers YAML/booleans/parallel/age/prerequisites/resume. Capture exact RED and commit/push coherent RED checkpoints before production edits.
3. Implement the smallest compatible seams and hardening in `internal/cli`, `internal/connectors/certify`, and schedule/config integration only. Preserve dynamic dispatch, `cli.Run` re-entrancy, exits 0/1/2/3, stdout/stderr/JSON, help/docs/website behavior, and valid noncredentialed sample behavior.
4. Refactor while green; run focused tests repeatedly and with `-race` for context/concurrency/crontab confinement.
5. Verify affected CLI/certify/schedule/safety packages; runtime help/bare/invalid/JSON; docs/manual/website generators and drift; connector validation; `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`. Keep verification false until full `make verify` exits 0.
6. Commit/push coherent GREEN and terminal evidence checkpoints only to the active branch. No PR, integration, parent mutation, dependency change, credentialed/live operation, system crontab, or external write/sweep.

### Fourth-cycle completion

All F1–F10 dispositions are complete. The implementation uses explicit successful-preview state for create/cleanup/sweep execution; opaque secret-hit metadata and semantic argv/error redaction; private atomic report/progress/ledger writes; invocation-local crontab configuration; durable per-connector, provenance-validated ledgers with separate temporary sweep workspaces; caller-context propagation plus bounded post-mutation cleanup; strict bounded credential files and controls; prerequisite gating; exact schema/manifest/effective-options resume identity; and temporary test roots. Additional refactoring closed final-component symlink races, file-handle lifecycles, sweep source-preparation fail-open behavior, and ledger tag control-character/provenance authority.

Planning `07d0b5a4`, RED `43acd262`, GREEN `2c0a550c`, and lint-correction `b06816ad` checkpoints are pushed. Focused/repeated/race, full CLI/certify, schedule, runtime help/bare/invalid/JSON, credential-free sample, CLI docs/goldens/website drift, connector validation, gofmt/diff/vet, explicit `go test ./...`, build, and final `make verify` all pass. The first `make verify` correctly failed on four new unchecked fixture writes; those were fixed and lint returned zero findings before the complete gate was rerun successfully. No credentials, live services, system crontab, external writes/sweeps, dependencies, PR, integration, parent mutation, or main mutation occurred.

## Fifth bounded review-correction cycle — exact HEAD `05d9c6658f52e542b6a74e87e29bdcad7275ea9d`

Cycle identity: `issue-437-fifth-review-correction-20260720`; launcher `openai-codex/gpt-5.6-sol`, thinking `high`; branch `refactor/437-connectors-certify-native-cobra`; isolated worktree and local/remote exact-head equality confirmed clean before work. Inputs are `/tmp/pm-397-437-correctness-rereview.out` and `/tmp/pm-397-437-security-rereview.out`. Parent-orchestrator policy is active for #397/#407, but this user-bounded cycle prohibits subagents and parent/PR/integration mutation; execution decision is `local_critical_path` in the already isolated #437 worktree.

GSD: `scripts/gsd doctor` and `scripts/gsd list` passed (69 commands). The required `programming-loop` remains absent (`unknown GSD command`). `scripts/gsd prompt audit-fix --phase 437-connectors-certify-native-cobra --dry-run` generated the applicable official correction prompt. The manual universal runtime loop remains the recorded programming-loop fallback.

Loaded/recorded skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`.

### Recovery-budget exception

An additional correction cycle is explicitly authorized despite the prior four cycles because the independent rereviews contain unresolved P1 security findings: forged numeric-resource cleanup authority and cleanup ledger entries becoming unrecoverable before absence verification. Stopping on the nominal recovery budget would knowingly leave destructive and leak-recovery risks. This exception is bounded to the seven consolidated findings below, strict RED-before-production, and the active issue branch only.

### Consolidated findings and dispositions

All seven are **accepted** after tracing the reachable paths; overlapping correctness/security reports are consolidated:

1. **Cleanup ledger state (P1):** normal cleanup records `cleaned_at` in `write_cleanup`, before `cleanup_verify`; sweeper records it immediately after reverse execution and performs no exact absence check. Move ledger mutation after exact verified absence. Failed/unverified cleanup remains `Uncleaned()` and retryable. Both normal and sweep cleanup retain plan → preview → approval → execute before mutation.
2. **Cleanup authority (P1):** numeric issue/milestone pairings synthesize the generic identifier `1` or tag text rather than a verified server-issued numeric ID, so a forged project ledger can target unrelated GitHub resources. Sweep authority is restricted to safely tag-addressable curated pairings (currently GitHub labels and the local outbox self-test). Numeric pairings fail closed unless a future product decision durably captures the actual server-issued identifier and verifies it against the certification tag. Cross-connector/action/entity/tag provenance remains fail closed.
3. **Ledger input bounds/nondisclosure (P1/P2):** `LoadLedger` has only a per-line scanner bound, accumulates unbounded entries, and reflects the raw malformed line. Add total-byte and entry-count bounds before accumulation; errors name only safe line/error class and never raw content or planted values.
4. **Sweep constraints (P2):** sweep loads credential files but does not apply effective default/per-connector unsupported rate/budget/limit or other constraints before workspace/telemetry/harness/cleanup effects. Validate the complete credential-file constraints immediately after bounded load and before target/workspace/effect construction.
5. **Report persistence (P2):** single and batch discard `Report.Save` errors. Propagate current/history persistence failure, fail closed on symlink/unwritable paths, and define deterministic precedence: completed report leaks/exit 3 remain dominant; otherwise persistence failure is an internal failure and no success is emitted without durable evidence.
6. **CLI prevalidation (P2):** current prevalidation handles only selected assigned forms; unknown flags and malformed space-value forms reach logger/telemetry initialization before Cobra rejects. Complete certify-only syntax/value prevalidation before logger/telemetry/config effects while preserving valid global/late flags and dynamic connector dispatch.
7. **Resume integrity (P2):** resume trusts `Passed`, leaks, and structurally incomplete reports after identity/timestamp checks. Require complete structurally valid required stages/capabilities, reject duplicate/minimal/edited/future/incompatible reports, and recompute outcome/leaks from evidence instead of serialized `Passed`. Because authenticated local-artifact provenance would require a dependency or broader product decision, fail closed by rerunning rather than adding one.

### TDD/checkpoint plan

1. Commit/push all six reopened planning artifacts before RED or production edits.
2. Add focused deterministic tests using fakes/temp/effect recorders only: cleanup success + verification failure stays uncleaned; sweep verification failure stays retryable; forged issue/milestone produce zero effects; oversized/many/malformed ledger inputs fail without marker output; sweep default/per-connector constraints produce zero effects; report current/history symlink/unwritable failures surface with leak precedence; assigned/space/unknown certify flags create no telemetry/log artifacts; minimal/edited resume artifacts rerun. Capture RED and commit/push before production.
3. Implement the smallest fail-closed corrections in certify/CLI only; no defs, dependencies, external systems, or docs behavior churn unless the canonical contract needs correction.
4. Refactor while focused/repeated/race tests stay green. Preserve command/help/exit/JSON/stdout-stderr, context/cancellation, crontab isolation, durable ledger layout, dynamic dispatch/re-entrancy, and plan → preview → approval → execute.
5. Run affected CLI/certify/schedule/safety; focused repeated and race; runtime help/bare/invalid/JSON; docs/goldens/website generation/drift; connector validation; gofmt/diff/vet/full tests/build/`make verify`; verify no go.mod/go.sum delta. `verificationPassed` remains false until a complete `make verify` exits 0.
6. Commit/push coherent GREEN and truthful terminal evidence only to the issue branch; end clean and remote-matched. No credentials, live connectors/services/system crontab, external write/sweep, credential command, connector defs, generic write tools, dependency, PR, parent, integration, or main mutation.

### Fifth-cycle completion

All seven dispositions are complete at implementation head `e9ce945e56413dbb60f5eeec2f1d6e5df688a249`. Normal and sweep ledger state changes only after exact cleanup verification; failed verification stays retryable. Sweep denies numeric GitHub pairings and permits only tag-addressable curated authority. Ledger input is capped at 1 MiB/10,000 entries and malformed-line errors are opaque. Effective sweep constraints reject before telemetry/workspace/harness effects. Report history is durable before current evidence is published; single/batch persistence errors are represented without false success and leaks remain exit 3 dominant. Certify syntax is fully prevalidated before config/logger/telemetry initialization. Resume requires strict schema/identity, complete capabilities/stages, and recomputed outcome/leak consistency; unverifiable edits rerun without a new dependency.

Planning `8acf62a9`, RED `e2559f64` plus supplemental pre-telemetry RED `3d69b7a4`, and GREEN `e9ce945e` are pushed. Focused/repeated/race, full CLI/certify, schedule/safety, runtime help/bare/invalid/JSON, docs/goldens/website drift, connector validation, gofmt/diff/vet, explicit `go test ./...`, build, and final `make verify` all pass. No module delta, credentials, live connectors/services, system crontab, external credential commands/writes/sweeps, connector defs, dependencies, generic write tools, PR, parent, integration, or main mutation occurred.

## Continuation for stacked PR delivery and parent reconcile — exact HEAD `86eea0f966814e6848e5a52143eea15dd46ff801`

Session `issue-437-continuation-20260719T211738Z`; parent branch `feat/cli-architecture-v2` is now `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa` and adds disjoint #462 TUI design docs. Execution decision: `local_critical_path` because this worker owns exactly one issue branch and isolated cwd, no subagent tool is exposed, and the remaining work is branch reconcile, verification, and stacked PR creation.

Required GSD route refreshed before production edits: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 437 --skip-research` generated the official plan prompt; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`, so the manual universal runtime loop remains the honest fallback.

Loaded skills for this continuation: `.pi/skills/gsd-core/SKILL.md`; `.agents/skills/caveman/SKILL.md`; global `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, plus already-relevant connector/certify skills `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`. Path mismatch recorded: the project prompt/developer route names `.pi/skills/go-implementation/SKILL.md`, but this checkout only has `.pi/skills/gsd-core/SKILL.md`; Go skills were loaded from `/Users/karthiksivadas/.agents/skills/cc-skills-golang/skills/*/SKILL.md` per `required-skills-routing.md`.

Continuation scope:

1. Reconcile latest parent `a5474bcb` into `refactor/437-connectors-certify-native-cobra` without dropping accepted #437 commits. Resolve only real conflicts; the known #462 design-doc parent changes are disjoint and must be retained.
2. Audit the 42-file issue diff against #437. Directly applicable certify safety corrections are justified because the native Cobra certify command exposes parsing, credentials-file, report, ledger, sweep, resume, and cleanup authority; fail-closed fixes prevent newly declared flags and native command paths from creating no-op, leak, secret, or destructive behaviors. Block rather than keep any change not tied to native connectors/certify, required docs/help/website parity, or accepted certify safety corrections.
3. No new behavior edit is planned. Existing RED evidence remains valid. If any new behavior defect is found, add a new failing test before production changes or stop for strict TDD/human-gate guidance.
4. Verify focused CLI/connectors/certify tests, runtime help (`pm help connectors`, bare `pm connectors`, `pm connectors --help`), docs/golden/website parity, safe fixture-only certify smoke, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and connector validation when feasible.
5. Push coherent continuation commits to the worker branch and open a non-draft stacked PR to `feat/cli-architecture-v2` with Conventional Commit title and body containing exactly `Refs #437`, `Refs #407`, and `Refs #397`. Claude is disabled and Copilot quota is exhausted; record human/parent fallback pending and do not retry bots.

Safety remains: no secrets, live credential checks, credentialed certification, destructive cleanup, external writes/sweeps, reverse ETL execution, new dependencies, connector defs changes, generic write tools, parent/main merge, or quality-gate weakening.

Continuation audit result before verification: the existing 42-file diff is scoped to #437: six issue phase artifacts; `cmd/pm/main.go` harness context/crontab seam required for certify in-process re-entrancy; 11 `internal/cli` native connectors/certify command, tests, and golden artifacts; 19 `internal/connectors/certify` safety/runtime/test files; and five canonical/generated CLI docs, certification design/PRD, and website parity files. No dependency, connector definition, sibling namespace, or parent-owned planning diff remains in the stacked PR diff. The certify safety corrections are directly applicable because native Cobra moved certify parsing and declared flags into the command tree; without the accepted fail-closed corrections, the new surface could silently ignore safety controls, load credentials before validation, persist/report secrets, trust forged ledger/resume artifacts, or perform cleanup/sweep effects without verified preview/absence. Parent `a5474bcb` merged cleanly with no conflicts; #462 design docs/skills/traces are retained by the merge and are no longer part of the stacked PR diff.

Continuation verification result: focused CLI/connectors/certify tests passed (`119.151s` and `7.344s`); runtime connectors help outputs are byte-identical (`8391` bytes); docs/golden tests passed (`10.347s`) and website docs generation is drift-free; fixture-only `sample` certification passed with exit 0 and empty stderr; `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and explicit `connectorgen validate` all passed. The only reverse-ETL execution was the repository Makefile's local temp-root smoke inside `make verify`, which follows plan → preview → approval → execute against sample/warehouse/outbox fixtures; no credentialed/live certification, destructive cleanup, external write/sweep, or live service ran.

Continuation delivery result: branch pushed and non-draft stacked PR opened at https://github.com/polymetrics-ai/cli/pull/466 targeting `feat/cli-architecture-v2` with body lines `Refs #437`, `Refs #407`, and `Refs #397`. Automated review route is human/parent fallback pending because Claude is disabled and Copilot quota is exhausted; no bot retry was attempted.


## Sixth bounded exact-head review-correction cycle — exact HEAD `8e7e2533c75451114c4d6ae38f89b7fd1ede6c34`

Identity: `issue-437-sixth-review-correction-20260719T220843Z`; branch `refactor/437-connectors-certify-native-cobra`; PR #466; parent #397; umbrella #407. The isolated worktree, active branch, clean status, local head, remote branch head, and PR head were confirmed equal to `8e7e2533c75451114c4d6ae38f89b7fd1ede6c34` before edits. No PR merge is authorized.

GSD route refreshed before production edits: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 437 --skip-research` generated the official prompt; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so the manual universal runtime loop remains the recorded fallback. Execution decision for this worker cycle is `local_critical_path`: the worker has one issue branch and isolated cwd, no subagent tool, and the correction is the local critical path for accepted exact-head findings. The parent coordinator records this invocation as spawned separately.

Loaded skills for this cycle: `.pi/skills/gsd-core/SKILL.md`; `.agents/skills/caveman/SKILL.md`; global `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`. `.pi/skills/go-implementation/SKILL.md` is still absent in this checkout; global cc-skills files are the actual Go implementation evidence per `required-skills-routing.md`.

### Recovery-budget exception

A sixth bounded correction cycle is explicitly authorized because unresolved accepted **High** leak-dominance and resource-authority findings can leave live resources untracked or can mask leak exit 3. Stopping on the prior recovery budget would knowingly leave destructive/recovery risk. Scope is restricted to the three accepted exact-head findings below, strict RED-before-production, and the active PR branch only.

### Accepted findings and planned dispositions

1. **High — approval idempotency after cleanup can create untracked resources.** Current write-stage ordering finalizes cleanup verification/ledger before replaying the consumed approval. If replay unexpectedly succeeds, a resource can be created after cleanup with no reopened ledger, no `leaks[]`, and no exit 3. Add deterministic RED that makes replay success mutate the fixture outbox after cleanup. Fix by ordering idempotency verification before final cleanup so any unexpected success remains covered by the existing provenance-bound cleanup/verification; if that cleanup cannot prove absence, leak evidence remains exit 3 dominant.
2. **High — report+error and progress-persistence errors mask leaks.** Batch runner `report+error` paths can replace `rep.Leaks` exit 3 with exit 2; batch progress persistence failure can return an ancillary error to the CLI before safe batch report output, mapping a leaked batch to exit 1 and hiding evidence. Add RED for runner report+error with leaks and CLI progress-persistence failure with existing leaks. Fix so `ExitCodeFor(rep)==3` dominates runner errors, aggregate batch exit remains 3, and the CLI emits safe batch report evidence when a progress persistence error occurs after leaked results exist.
3. **Medium — stale top-level leak after exact absence proof.** Cleanup call failure records a leak; later exact verification can prove absence and mark action pass/ledger cleaned while `Report.Leaks` still claims a leak. Add RED. Fix report consistency by removing the stale leak when exact cleanup verification proves absence, while preserving the cleanup stage failure and marking the write action as failed/non-leaked rather than dishonestly passed.

### RED → GREEN plan

1. Commit/push this six-artifact planning checkpoint before tests or production edits.
2. Add focused RED tests only: approval replay success appends a post-cleanup outbox row; batch runner returns leaked report plus error; CLI batch returns leaked batch plus progress persistence error; cleanup failure followed by exact absence proof. Capture failures and commit/push before production edits.
3. Implement minimal corrections in `internal/connectors/certify/stages_source.go`, `stages_write.go`, `batch.go`, and `internal/cli/certify_cli.go` only if tests require. Preserve `cli.Run` re-entrancy, stdout/stderr/JSON, dynamic connector dispatch, plan → preview → approval → execute, durable ledger provenance, and exits 0/1/2/3 with leak dominance.
4. Run focused new tests RED then GREEN, repeated and `-race` variants where applicable; affected full `internal/connectors/certify` and `internal/cli` tests; runtime help/docs/website/golden parity if semantics surface changes; connector validation; gofmt/diff/vet/full test/build/`make verify`; fixture-only sample smoke. No credentialed/live checks, services, external writes/sweeps, dependencies, PR merge, or bot review requests.
5. Commit/push GREEN/refactor and terminal evidence checkpoints; update PR #466 body with correction/TDD/gate evidence and final exact head.


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

## Seventh bounded exact-head review-correction cycle — exact HEAD `6e9e7d9422050a609306d8900d6a06c8bb1fc223`

Identity: `issue-437-seventh-bounded-correction-20260720`; branch `refactor/437-connectors-certify-native-cobra`; PR #466; parent #397; umbrella #407; base `feat/cli-architecture-v2`. Clean local head, remote branch head, and PR #466 head were confirmed equal to `6e9e7d9422050a609306d8900d6a06c8bb1fc223` before edits. The stale sixth-cycle PR-body checkbox/status is finalized: PR #466 body already records latest exact head `6e9e7d9422050a609306d8900d6a06c8bb1fc223`; sixth terminal verification remains true for that head, and seventh verification is reopened/false until new full gates pass.

GSD route refreshed before production edits: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 437 --skip-research` generated the official prompt; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so the manual `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` fallback applies. Execution decision: `local_critical_path` because this worker owns one isolated #437 branch/cwd, no subagent tool is exposed, and the parent coordinator records this worker as spawned.

Loaded skills for this cycle: `.pi/skills/gsd-core/SKILL.md`; `.agents/skills/caveman/SKILL.md`; global `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-code-style`, and `golang-naming`. `.pi/skills/go-implementation/SKILL.md` is absent in this checkout; global cc-skills files are the Go implementation evidence per `required-skills-routing.md`.

### Seventh recovery-budget exception

A seventh bounded correction is explicitly authorized because two accepted Medium findings remain unresolved at exact head: effect-before-usage validation for bare value-required certify flags and resume replay after absence-proven cleanup failure. Stopping on recovery budget would knowingly leave credential-file/read/telemetry effects before syntax rejection and repeated write/cleanup effects under `--resume`. Scope is restricted to the two behavior findings plus truthful artifact/PR-body status repair, strict RED-before-production, and the active PR branch only.

### Accepted findings and planned dispositions

1. **Medium — resume rejects absence-proven cleanup failure.** Root cause: `validResumeEvidence` treats any non-skipped `write_cleanup` or `cleanup_verify` failure with zero top-level leaks as structurally invalid. The sixth-cycle cleanup fix intentionally produces a completed, non-leaked shape when `write_cleanup` failed but `cleanup_verify` proves exact absence and the write action remains `fail`/non-leaked. Current resume therefore reruns and can repeat write/cleanup effects. Add RED seeded from this exact cleanup-failure/absence-proof report. Fix to accept only the provenance-consistent, structurally complete absence-proven shape: failed `write_cleanup`, later passed `cleanup_verify`, valid certify tag/provenance, failed/non-leaked write-action result with the exact absence-proof reason, and no leak/action contradictions. Keep all other incomplete, edited, duplicate, failed-verify, or unproven shapes fail-closed/rerun.
2. **Medium — bare required-value flags reach effects.** Root cause: native StringArray flags retain `NoOptDefVal="true"`; `prevalidateCertifySafetyArgs` normalizes space forms then validates only booleans and numeric/duration values. Bare `--stream`, `--from-env`, or `--config` can reach single runner/options; bare `--credentials-file` can initialize telemetry/logging and attempt to read path `true`; bare `--parallel` currently maps to validation exit 3 instead of usage exit 2. Add no-effect RED for bare value-required controls in applicable single/batch/sweep modes. Fix with strict value-required validation before logger, telemetry, credential loading, runner, workspace, or sweep for `from-env`, `config`, `stream`, `credentials-file`, `parallel`, and `older-than`; preserve valid `--flag=value` and `--flag value`; reject missing or empty values as usage exit 2. Audit declared certify flags: boolean controls remain bare-capable (`keep-workdir`, `write`, `full`, `all`, `resume`, `sweep`); `skip` remains value-checked (`write` only) and unsupported/hidden controls remain fail-closed before effects.
3. **Low — artifact/PR status honesty.** Finalize stale sixth-cycle delivery fields and PR-body checkbox before claiming seventh terminal evidence. Seventh fields must remain false/not-yet-verified until their actual pushed head and verification output exist.

### Seventh RED → GREEN plan

1. Commit/push this planning checkpoint before RED tests or production edits.
2. RED checkpoint: add focused tests only. Certify package: seed a completed report with failed `write_cleanup`, passed `cleanup_verify`, exact absence-proof action state, no leaks, and prove `--resume` reruns today. CLI package: use full `Run` with `PM_TELEMETRY=file` and temp roots to prove bare `--from-env`, `--config`, `--stream`, `--credentials-file`, `--parallel`, and `--older-than` either reach runner/credential path or classify incorrectly, and create no logger/telemetry/project effects after fix. Commit/push failing tests before production.
3. GREEN: minimally update `internal/connectors/certify/batch.go` resume evidence and `internal/cli/certify_cli.go` prevalidation/native validation. No docs text expected unless help/contract changes.
4. Refactor/verify: focused RED/GREEN, repeated and race relevant cases, full `internal/connectors/certify` and `internal/cli`, runtime no-effect/exit/help, fixture-only sample smoke, docs/golden/website if behavior/help changes, `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
5. Commit/push GREEN/refactor and terminal evidence checkpoints. Update PR #466 body with seventh-cycle evidence and final exact head. Do not request Claude/Copilot; human/parent review remains open. Do not merge.

Safety remains: no secrets, credentialed/live checks, live services, destructive cleanup, external writes/sweeps, new dependencies, connector defs, generic write tools, reverse ETL outside existing ordered local fixture smoke, parent/main merge, or bot review request.

### Seventh-cycle completion

Planning `82f59229`, RED `e0fb8c4b`, GREEN `2ce0e10a`, and terminal evidence `0d515e6e` were committed and pushed. PR #466 body was updated to latest exact head `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8`. The resume fix now accepts only the exact completed cleanup-failure absence-proof shape: failed `write_cleanup`, later passed `cleanup_verify`, valid tag/run provenance, one failed/non-leaked write action with reason `write_cleanup failed, but cleanup_verify proved the entity absent`, and no leak/action contradictions. Unproven cleanup failures, failed cleanup verification, malformed identity/timestamp/schema, and leak/action mismatches still rerun.

The CLI fix validates raw certify value-required syntax before config/logger/telemetry and repeats syntax validation in direct command execution. `from-env`, `config`, `stream`, `credentials-file`, `parallel`, and `older-than` require non-empty values; valid `--flag=value` and `--flag value` forms remain valid; boolean controls (`keep-workdir`, `write`, `full`, `all`, `resume`, `sweep`) retain bare semantics; `skip` remains value-checked; unsupported controls remain hidden/fail-closed. Bare value-required controls now exit 2 before telemetry/log files, credential-file reads, runner calls, sweep workspace, or project effects.

Verification passed: focused RED/GREEN, repeated, race, full `internal/connectors/certify` (`353.654s`) and `internal/cli` (`448.666s`), runtime no-effect and help byte parity (`8391` bytes), docs/golden/website drift, fixture-only sample smoke, `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...` (CLI `444.782s`, certify `349.644s`), `go build ./cmd/pm`, final `make verify` (CLI `442.999s`, certify `348.395s`, smoke/lint/docs/connectorgen green), and explicit connectorgen validation 547/0. No credentials, live services, external writes/sweeps, dependencies, connector defs, bot review, parent/main merge, or quality-gate reduction occurred.

## Eighth bounded correction — exact HEAD `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8`

Identity: `issue-437-eighth-bounded-correction-20260720`; branch `refactor/437-connectors-certify-native-cobra`; PR #466; parent #397; umbrella #407; base `feat/cli-architecture-v2`. Clean local head, remote branch head, and PR #466 head were confirmed equal to `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8` before edits. No merges, bot reviews, parent mutation, or main mutation are authorized.

GSD route refreshed before production edits: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 437 --skip-research` generated the official prompt; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so the manual `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` fallback applies. Execution decision: `local_critical_path` because this worker owns one isolated #437 branch/cwd, no subagent tool is exposed, and the parent coordinator records this worker as spawned.

Loaded skills for this cycle: `.pi/skills/gsd-core/SKILL.md`; `.agents/skills/caveman/SKILL.md`; global `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-code-style`, and `golang-naming`. `.pi/skills/go-implementation/SKILL.md` is absent in this checkout; global cc-skills files remain the Go implementation evidence per `required-skills-routing.md`.

### Eighth safety/correctness recovery exception

An eighth bounded correction is explicitly authorized because accepted Medium findings can cause future-forged resume artifacts to suppress runner execution, and whitespace-only value-required flags can reach telemetry, credential-file reads, runner, workspace, or sweep effects. Stopping on recovery budget would knowingly leave correctness and safety risk. Scope is restricted to the two behavior findings plus truthful delivery/PR-body artifact status, strict RED-before-production, and the active PR branch only.

### Accepted findings and planned dispositions

1. **Medium — future-dated completed reports are trusted.** Root cause: `completedReport` rejects zero or `CompletedAt < StartedAt` timestamps but does not reject materially future `StartedAt`/`CompletedAt`. Add deterministic RED for both ordinary completed reports and cleanup-failure/absence-proof reports with timestamps far beyond a documented clock-skew tolerance. Fix by rejecting report timestamps later than `time.Now().UTC()+smallSkew`, while preserving valid historical reports and `CompletedAt >= StartedAt`. Runner must rerun when evidence is future-invalid.
2. **Medium — space-form empty/whitespace value-required flags pass prevalidation.** Root cause: `validateCertifyRequiredValueArgs` handles assigned empty values but not `strings.TrimSpace(next)==""` for space-form values. Add no-effect RED across applicable single/batch/sweep modes for `--from-env`, `--config`, `--stream`, `--credentials-file`, `--parallel`, and `--older-than` with empty or whitespace next tokens. Fix before config/logger/telemetry, credential-file reads, runner, workspace, or sweep effects; preserve valid assigned/space values and usage exit 2 for missing/blank values.
3. **Low — artifact/PR status honesty.** Finalize seventh status truthfully and make eighth terminal fields distinguish `reviewedCodeHead`/`endingImplementationHead` from terminal evidence commit. Do not attempt impossible commit self-reference; the terminal artifact cannot include its own resulting SHA because Git computes the SHA after content is committed.

### Eighth RED → GREEN plan

1. Commit/push this planning checkpoint before RED tests or production edits.
2. RED checkpoint: add focused tests only. Certify package: seed ordinary completed and cleanup-failure/absence-proof reports with future timestamps beyond the skew tolerance and prove `--resume` reruns. CLI package: use fake runtime and full `Run`/`PM_TELEMETRY=file` temp roots to prove empty/whitespace space-form value-required flags reject as usage exit 2 with no telemetry/log/project/credential/runner/sweep effects. Commit/push failing tests before production edits.
3. GREEN: minimally update `internal/connectors/certify/batch.go` resume timestamp validation and `internal/cli/certify_cli.go` required-value prevalidation. No canonical docs text expected unless runtime help/error contract changes.
4. Refactor/verify: focused RED/GREEN, repeated and race relevant cases, full `internal/connectors/certify` and `internal/cli`, runtime no-effect/exit/help, fixture-only sample smoke, docs/golden/website drift if applicable, `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
5. Commit/push GREEN/refactor and terminal evidence checkpoints. Update PR #466 body with eighth-cycle evidence and final exact PR head. Do not request Claude/Copilot; human/parent review remains open. Do not merge.

Safety remains: no secrets, credentialed/live checks, live services, destructive cleanup, external writes/sweeps, new dependencies, connector defs, generic write tools, reverse ETL outside existing ordered local fixture smoke, parent/main merge, or bot review request.

### Eighth-cycle completion

Planning `45f190dc`, RED `ea5e412c`, and GREEN implementation `af0e4dab` were committed and pushed. RED failed before production as intended: future-dated ordinary and cleanup-failure/absence-proof reports resumed instead of rerunning, and empty/whitespace space-form value-required flags reached later validation/effects. GREEN adds a documented five-minute resume-report clock-skew tolerance and treats materially future `StartedAt`/`CompletedAt` as future-invalid evidence; valid historical completed reports and `CompletedAt >= StartedAt` remain accepted. GREEN also rejects `strings.TrimSpace(next)==""` for value-required certify flags before config/logger/telemetry, credential-file reads, runner, workspace, or sweep effects while preserving valid assigned/space values and boolean bare controls.

Verification passed: focused GREEN, repeated, race, full `internal/connectors/certify` (`347.792s`) and `internal/cli` (`443.361s`), runtime whitespace no-effect and help byte parity (`8391` bytes), docs/golden/website drift, fixture-only sample smoke, `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...` (CLI `446.588s`, certify `349.095s`), `go build ./cmd/pm`, final `make verify` (CLI `445.358s`, certify `348.425s`, smoke/lint/docs/connectorgen green), and explicit connectorgen validation 547/0. No credentials, live services, external writes/sweeps, dependencies, connector defs, bot review, parent/main merge, or quality-gate reduction occurred.

Terminal evidence and the eighth-cycle PR #466 body update are complete at parent/previous PR head `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`, while reviewed implementation head remains `af0e4dabf5be70237c02403e6ef4f003042667d6`. This docs-only closure intentionally avoids embedding its own resulting SHA; the current live PR head is authoritative from Git/GitHub, not from a self-referential artifact line.

## Docs-only evidence closure — exact start `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`

Accepted Low finding: phase artifacts carried stale terminal-commit/PR-body delivery wording after completion. No behavior RED is required and no production files are edited. Worker decision: `local_critical_path`; parent records this worker invocation as spawned. GSD route refreshed with `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase 437 --skip-research`. Loaded skills for this closure: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-documentation`, and `golang-security` per docs/CLI/safety routing. Closure validation is limited to JSON parse, stale terminal/PR marker grep, `git diff --check`, and exact diff-scope check for the six phase artifacts.

## Ninth bounded CI flake root-cause correction — exact HEAD `9f004ac5d96d84bd1f8b186496e1f594a183a18b`

Identity: `issue-437-ninth-ci-flake-correction-20260720`; branch `refactor/437-connectors-certify-native-cobra`; PR #466; parent #397; umbrella #407. Clean local head, remote branch head, and PR #466 head were confirmed equal to `9f004ac5d96d84bd1f8b186496e1f594a183a18b` before planning. No bot review, parent merge-to-main, or broad rewrite is authorized.

GSD route refreshed before test/production edits: `scripts/gsd doctor` passed; `scripts/gsd list` reported 69 commands; `scripts/gsd prompt plan-phase 437 --skip-research` generated `/tmp/pm-437-plan-phase.prompt`; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` remains unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so the manual `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` fallback applies. Execution decision: `local_critical_path` because this worker owns one isolated #437 branch/cwd, no subagent tool is exposed, and the parent coordinator records this worker as spawned.

Loaded skills for this cycle: `.pi/skills/gsd-core/SKILL.md`; `.agents/skills/caveman/SKILL.md`; global `golang-how-to`, `golang-testing`, `golang-concurrency`, `golang-context`, `golang-safety`, `golang-cli`, `golang-error-handling`, `golang-security`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-lint`, `golang-code-style`, and `golang-naming`. `.pi/skills/go-implementation/SKILL.md` is absent in this checkout and also absent at parent `c91b90cf9671b5caabc0ef4ec24d81897f870458`; global cc-skills files remain the Go implementation evidence per `required-skills-routing.md`.

### Ninth CI failure and root-cause plan

Remote RED from GitHub Actions run `29711194607`, job `verify`, head `9f004ac5d96d84bd1f8b186496e1f594a183a18b`:

```text
--- FAIL: TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit (0.25s)
    batch_test.go:372: elapsed = 252.003235ms, want well under 3x80ms serial time (parallelism not happening)
FAIL
FAIL	polymetrics.ai/internal/connectors/certify	638.060s
```

All other reported packages/checks in that failed run passed. Root cause is the test's wall-clock scheduling assertion (`elapsed > 200ms`) over three `80ms` fake jobs; loaded CI scheduling produced `252.003235ms`, slightly above 3×80ms, despite the same test already recording overlapping worker starts. The fix must not raise/remove the threshold or retry CI. It must replace wall-clock proof with deterministic bounded-concurrency evidence: barrier-controlled worker start, active-worker/max-concurrency accounting, release channel, result/order channel, and bounded timeout only for deadlock detection.

### Ninth RED → GREEN plan and completion

1. Commit/push this planning checkpoint with `verificationPassed=false` before any test edits.
2. Reconcile latest parent `origin/feat/cli-architecture-v2` at `c91b90cf9671b5caabc0ef4ec24d81897f870458`; retain disjoint #397 artifacts and do not edit parent artifacts.
3. Reproduce/stress current focused test where feasible (`go test ./internal/connectors/certify -run TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit -count=N`, optionally `-race`). If local stress does not reproduce, record that honestly; do not fabricate a local failure. The remote CI failure above remains RED evidence.
4. Test-only GREEN: rewrite `TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit` to prove bounded overlap without timing: four jobs queued, `parallel=3`, each fake runner signals started and blocks on a release channel; assert three workers start and block before release, `maxActive==3`, no active-worker limit violation, all four connectors/results complete after release, and channels/timeouts cannot leak goroutines or deadlock.
5. Run focused new test repeated and under `-race`, full `internal/connectors/certify`, full CLI via full-suite gates, `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, fixture-only sample smoke, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
6. Commit/push coherent test-fix and terminal evidence checkpoints; update PR #466 body with CI failure disposition and final head. Do not request bots or merge.

Completion: planning `0398a5d6` was committed and pushed before test edits. Parent `c91b90cf9671b5caabc0ef4ec24d81897f870458` merged cleanly as `9678d4dda2fcf331b3199f042804001c06eccf64`; only parent-owned #397 artifacts came from the parent merge and were not manually edited. Local stress did not reproduce the CI flake: focused current test `-count=50` passed (`7.601s`) and `-race -count=10` passed (`3.084s`). Remote CI run `29711194607` remains RED evidence.

Test-only GREEN `828be4de4145d1246347b820d433d52bd1e92002` replaces the elapsed-time assertion with deterministic release-barrier, active-worker, max-concurrency, violation, and order/result-channel evidence. Verification passed: focused deterministic test `-count=20` (`2.051s`), focused race `-count=10` (`2.435s`), full `go test ./internal/connectors/certify -count=1` (`358.903s`), `gofmt -w cmd internal`, `git diff --check`, `go vet ./...`, `go test ./...` (CLI `453.338s`, certify `355.074s`), `go build ./cmd/pm`, `make verify` (CLI `454.887s`, certify `357.580s`, docs validate, local smoke `smoke ok`, lint `0 issues`, connectorgen green), fixture-only `./pm connectors certify sample --root <temp> --json` (`exit=0 kind=ConnectorCertification report_kind=ConnectorCertification connector=sample passed=True stderr_bytes=0`), and explicit `go run ./cmd/connectorgen validate internal/connectors/defs` (`547 connector(s) checked, 0 findings`). No CLI/help/docs/website text changed; full CLI was covered by `go test ./...`/`make verify`.

Safety remained: no secrets, credentialed/live checks, live services, destructive cleanup, external writes/sweeps, new dependencies, connector defs, generic write tools, reverse ETL outside existing ordered local fixture smoke, parent/main merge, bot review request, or quality-gate reduction.
