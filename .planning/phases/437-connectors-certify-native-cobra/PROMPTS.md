# Phase 437 Prompts

## Kickoff snapshot

Task: implement issue #437, final serialized Phase 9 unit under #397/#407, from exact `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` on isolated `refactor/437-connectors-certify-native-cobra`; Sol/high; no dependencies/services/credentials/PR/review.

Identity: `issue-437-pi-sol-high-20260719T095145Z`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 437 --skip-research
scripts/gsd prompt programming-loop init --phase 437 --dry-run
```

Doctor/list and plan generation passed. `programming-loop` is absent from the adapter registry, so the manual universal-runtime-loop fallback applied.

Execution decision: `local_critical_path` — final assigned serialized namespace; central router scope collides; no subagent tool exposed; user bounded delivery to implementation/commit/push.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

Safety prompt: preserve `cli.Run`, dynamic connector dispatch, exits 0/1/2/3, context/events/telemetry, stdout/stderr, and credential-value exclusion. Use only fixture/replay/local tests and temp roots. Do not touch defs, call live credentials, perform external writes, add dependencies/services, or create PR/review.

Downstream artifact: native connectors/certify subtree, declared current flags and runtime seam, compatibility normalization, contextual direct/trailing help, focused parity tests, and canonical manual/docs/website updates.

Verification result: superseded by the accepted correction cycle below; current phase verification is pending.

## Accepted correction kickoff snapshot

Task: at exact HEAD `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`, accept all five findings in `/tmp/pm-397-review-437.log`; update artifacts and commit/push them, then add/commit RED tests before production. Fail closed on unsupported safety/mode flags; restore legacy single span/validation/options and batch file/parallel/error precedence; exact-only connectors help; accurate pre-report versus completed-report exits across canonical/generated/website docs. Sol/high; fixture/temp only; no dependencies/services/credentials/writes/PR/review.

Commands: `scripts/gsd doctor` passed; both `scripts/gsd prompt programming-loop ...` forms returned `unknown GSD command`, so the manual universal runtime loop applies. Execution decision: `local_critical_path` (bounded correction, isolated worktree, no subagent tool).

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-spf13-cobra`.

Downstream artifact: complete at implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`; terminal verification artifact `2987f21b` is pushed and this final delivery marker closes the phase.

Verification result: pass. GREEN evidence: focused `3.004s`; native/certify/telemetry `108.532s`; base/current differential 5/5 byte-identical; focused race `29.046s`; ×10 `24.991s`; certify redaction/replay/concurrency race `349.263s`; exit focus `21.618s`; local sample exit 0/pass/redacted; docs/golden `24.275s`; website regeneration hash-stable; full CLI `435.572s`; certify `338.846s`; validation 547/0; final `make verify` exit 0 in `468.36s`. No live credentials/writes, services, dependencies, PR, or review.

## Second accepted safety correction kickoff snapshot

Task: at exact HEAD `0d743e54e06c9e27e550eacce9be7899a9e23d19`, accept every finding in `/tmp/pm-397-rereview-437.log`; update and push artifacts, then add/commit/push RED effect-recorder tests before production. P1: batch write-disable controls dominate credential-file writes and credential constraints are never discarded. P2: reject/hide unsupported single controls, constrain skip, reject mode no-ops, and audit every flag. P3: remove stale architecture/PRD claims, correct connector help naming, regenerate CLI/website data. Preserve base behavior, dynamic dispatch, exits, and redaction. Fixture/temp only; no credentials/services/dependencies/PR/review.

Commands: `scripts/gsd doctor` and `scripts/gsd prompt plan-phase 437 --skip-research` passed; `scripts/gsd prompt programming-loop ...` returned `unknown GSD command`, so the manual universal runtime loop applies. Execution decision: `local_critical_path` (bounded correction, existing isolated issue worktree, no subagent tool).

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-spf13-cobra`.

Downstream artifact: planning `aa39fd9d`, RED `9a47ff3d`, GREEN `7b6eaa58`, and verification `974495d5` checkpoints pushed; this final delivery marker closes the correction.

Verification result: pass. Focus/effect/no-op, repeated ×10 (`0.661s`), race (CLI `1.726s`, certify `2.535s`), full CLI (`440.910s`), full certify (`346.271s`), runtime help/no-effect/sample smoke, CLI docs/goldens, website full-data regeneration, gofmt/vet/test/build, `make verify` (`7m36.852s`), and connectorgen 547/0 pass. No credentials, services, dependencies, PR, or review.

## Third accepted safety/correctness correction kickoff snapshot

Task: at exact HEAD `437d13cf`, accept every finding in `/tmp/pm-397-rereview2-437.log`; update and push artifacts, then add/commit/push RED tests before production. Reject every unknown certify flag and write-like typo before effects; require a positive, reasonably bounded sweep age; make ordinary completed reports reusable by `--resume` while rerunning incomplete reports; reject credential-file `exec` before effects and remove generic external execution code/docs; correct usage exits, release-stage token `ga`, flags/docs audit, and terminal artifact honesty. Fixture/temp and in-process fakes only; no external commands, credentials, services, dependencies, PR, or review.

Commands: `scripts/gsd doctor` and `scripts/gsd list` passed; `scripts/gsd prompt programming-loop ...` returned `unknown GSD command`, so the manual universal runtime loop applies. Execution decision: `local_critical_path` (bounded correction, existing isolated issue worktree, no subagent tool).

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-context`, `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-spf13-cobra`, and `golang-lint`.

Downstream artifact: complete at implementation head `f56bc825`; strict certify parsing, bounded sweep age, prohibited-exec rejection/removal, completed-report resume reuse, incomplete-report rerun, and corrected generated docs are committed and pushed. Terminal verification artifact `3854295b` is also pushed.

Verification result: pass. Focused/repeated/race/resume/sweep/no-effect/audit tests, runtime help, docs/goldens, hash-stable website generation, credential-free local sample smoke, full CLI (`446.382s`), full certify (`350.637s`), gofmt/vet/full tests/build, `make verify` (exit 0 in `14m58.384s`), and connectorgen 547/0 pass. No external credential command, live credential, service, dependency, PR, or review was used.

## Fourth bounded review-correction kickoff snapshot

Task: at exact clean/matched HEAD `1e27b14012f65ffa24c01ed855d0405c24401eee`, trace and disposition both independent review outputs, consolidate overlaps, commit/push planning, capture and commit RED before production, then correct preview/approval ordering, secret rendering/config/report modes, crontab concurrency confinement, durable provenance-bound ledger/sweep, context cancellation, strict bounded credentials and safety controls, prerequisite gating, resume compatibility, and test artifact pollution. Preserve dynamic dispatch, `cli.Run`, exits 0/1/2/3, output/docs parity, and valid noncredentialed behavior.

Commands: `scripts/gsd doctor`, `scripts/gsd list`, `scripts/gsd prompt plan-phase 437 --skip-research`, and `scripts/gsd sources plan-phase` pass. `scripts/gsd prompt programming-loop init --phase 437 --dry-run` returns `unknown GSD command`; manual universal-loop fallback recorded. Execution decision: `local_critical_path` because user explicitly prohibited subagents and any work outside the isolated issue branch.

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`; added `golang-spf13-viper`, `golang-troubleshooting`.

Downstream artifact: complete. Planning `07d0b5a4`, RED `43acd262`, GREEN `2c0a550c`, and lint/resource-fix `b06816ad` are pushed. The implementation provides successful-preview mutation gates; opaque and semantic redaction; strict private artifacts and credential references; invocation-local crontab selection; durable provenance-bound ledgers with fresh sweep workspaces; context cancellation with bounded post-mutation cleanup; strict controls/resources/prerequisites; exact identity-bound resume; and temporary test roots.

Verification result: pass. Focused/repeated/race matrices, full CLI/certify/schedule, runtime help/bare/invalid/JSON and credential-free sample, temporary CLI docs/goldens/website drift, docs and connector validation, gofmt/diff/vet, explicit full tests/build, and final `make verify` exit 0 (`464.41s`) pass. An initial verify run correctly stopped at lint on unchecked synthetic fixture writes; `b06816ad` fixed the resource handling and the complete gate reran successfully. No credentials, live services, system crontab, external writes/sweeps, dependencies, PR, integration, parent mutation, or main mutation occurred.

## Fifth bounded review-correction kickoff snapshot

Identity: `issue-437-fifth-review-correction-20260720`; exact clean/matched start `05d9c6658f52e542b6a74e87e29bdcad7275ea9d`; launcher `openai-codex/gpt-5.6-sol`, thinking `high`; no subagents. Correct seven consolidated rereview findings: defer ledger cleaning until exact absence; deny forged numeric-resource sweep authority; bound and sanitize ledger loading; reject all unsupported sweep credential constraints before effects; propagate report/history persistence with leak-dominant precedence; reject malformed/unknown certify flags before logger/telemetry/files/network; and structurally validate/recompute resume evidence or rerun.

Commands: `scripts/gsd doctor` and `scripts/gsd list` passed; programming-loop remains unavailable; `scripts/gsd prompt audit-fix --phase 437-connectors-certify-native-cobra --dry-run` generated the applicable correction prompt; manual universal-loop fallback recorded. Execution decision: `local_critical_path` because the user prohibited subagents and all parent/PR/integration mutation while assigning the isolated issue worktree.

Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`.

Recovery-budget exception: unresolved P1 destructive cleanup authority and unrecoverable-ledger findings require this extra bounded correction rather than stopping after four cycles.

Downstream artifact: complete at implementation head `e9ce945e56413dbb60f5eeec2f1d6e5df688a249`. Planning `8acf62a9`, RED `e2559f64`/`3d69b7a4`, and GREEN `e9ce945e` are pushed. Cleanup retryability/authority, bounded opaque ledgers, pre-effect sweep constraints, durable persistence precedence, pre-telemetry parsing, and strict recomputed resume evidence are corrected.

Verification result: pass. Focused final `7.288s` certify / `8.389s` CLI; ×10 `48.652s` / `77.313s`; race `51.469s` / `89.921s`; full package checkpoints certify `327.840s`, CLI `443.427s`; schedule/safety and vet; help/bare/invalid/JSON; docs/golden `16.600s`; website drift-free; connectorgen 547/0; explicit `go test ./...` real `7m34.316s`; build; final `make verify` exit 0 real `7m52.496s` (CLI `449.572s`, certify `332.793s`, lint 0). No credentials, live services/connectors, system crontab, external writes/sweeps, dependencies, PR, parent, integration, or main mutation.

## Continuation kickoff snapshot

Identity: `issue-437-continuation-20260719T211738Z`; exact clean start `86eea0f966814e6848e5a52143eea15dd46ff801`; latest parent `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa`; worker branch `refactor/437-connectors-certify-native-cobra`; base branch `feat/cli-architecture-v2`; no subagents.

Task: reconcile latest parent, audit existing 42-file #437 diff, justify directly applicable certify safety corrections, run full local gates and safe fixture-only certify smoke, push continuation commits, and open non-draft stacked PR to parent with exactly `Refs #437`, `Refs #407`, `Refs #397`. Claude disabled and Copilot quota exhausted; record human/parent fallback pending without bot retries.

Commands refreshed before production edits: `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 437 --skip-research`; `scripts/gsd prompt programming-loop init --phase 437 --dry-run` (unavailable: `unknown GSD command`). Execution decision: `local_critical_path`.

Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, and connector/certify support skills `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`. `.pi/skills/go-implementation/SKILL.md` is missing; global cc-skills files are the actual Go implementation evidence.

Downstream artifact: parent `a5474bcb` merged cleanly at `dc4aed23`; 42-file #437 diff audited as scoped; continuation artifacts updated; verification passed; non-draft stacked PR opened at https://github.com/polymetrics-ai/cli/pull/466; human/parent review fallback pending.

Verification result: pass. Focused CLI `119.151s`, certify `7.344s`; connectors help byte-equal `8391` bytes; docs/golden `10.347s`; website docs generation drift-free; fixture-only sample certify exit 0/pass/stderr 0; gofmt/diff/vet/full tests/build; `make verify` exit 0; connectorgen 547/0.
