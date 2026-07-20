# Phase 437 Verification

Invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

`verificationPassed`: true — full eighth-cycle local gates and `make verify` passed at reviewed implementation head `af0e4dabf5be70237c02403e6ef4f003042667d6`; terminal evidence and PR #466 body update were completed at parent/previous PR head `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`. This docs-only evidence closure keeps that truth and intentionally does not embed its own resulting SHA; current live PR head is authoritative from Git/GitHub.

## TDD / behavior

- [x] Six artifacts predate tests and production edits.
- [x] Initial RED captured before production: native connectors constructors absent.
- [x] Two focused help RED checkpoints preceded final trailing/direct action help corrections.
- [x] Native connectors actions and nested certify action/current flags/operands.
- [x] Bare/text/JSON/topic/direct/positional/trailing help; invalid action usage.
- [x] Literal `--`, malformed/legal unknown, action/operand discovery and globals.
- [x] Certify exits 0/1/2/3 and one-envelope semantics.
- [x] Fresh-tree re-entrancy, bounded concurrency, cancellation, events, telemetry.
- [x] Credential values absent from output, reports, events, and telemetry tests.
- [x] Only namespace parser calls removed; dynamic connector `parseFlags` code/diff unchanged.

## Focused / broad gates

- [x] Focused native connectors/certify tests: final `3.989s`.
- [x] Focused repeated `-count=10`: final `34.833s`.
- [x] Focused `-race`: final `40.842s`.
- [x] Router/golden/certify/telemetry focus: final `111.919s`.
- [x] Full CLI: final `431.305s` through `make verify`.
- [x] Full certify package: final `337.280s` through `make verify`.
- [x] Certify concurrency/event race focus: `2.395s`.
- [x] Required local certify smoke: exit 0, `ConnectorCertification`, sample, pass; stderr empty.
- [x] Exact-start operational differential: 21/21 unchanged; contextual action help is the reviewed intentional change.
- [x] Connector validation: 547 checked, 0 findings.
- [x] `gofmt -w cmd internal` and clean diff check.
- [x] `go vet ./...`.
- [x] `go test ./...`.
- [x] `go build ./cmd/pm`.
- [x] Final `make verify` exit 0; docs/smoke/lint/connectors all green.

## Help/manual/website parity

- [x] `pm help connectors`, bare `pm connectors`, and `pm connectors --help` are byte-equal text manuals.
- [x] Direct and trailing connectors/certify action help is contextual and side-effect free.
- [x] Bare JSON is one `CommandManual` envelope for `connectors`.
- [x] Invalid action remains usage exit 2.
- [x] Certify examples, credential-reference safety, envelopes, and exit 0/1/2/3 are documented.
- [x] `docs/cli/connectors.md` regenerated from the canonical manual.
- [x] Website CLI reference mirrored and `website/lib/docs.generated.ts` regenerated.
- [x] Golden transcripts regenerated only for the reviewed connectors-manual change.
- [x] Docs generation/drift and website generation pass.
- [x] Website typecheck not applicable: existing `node_modules/.bin/tsc` is absent; no prohibited dependency install was attempted.
- [x] Completion metadata remains registered through the native tree with `NoFileComp`; Phase 15/19 broad churn remains deferred.

## Safety/scope/delivery

- [x] Correct isolated branch and exact start.
- [x] GSD adapter/manual fallback and required skills recorded.
- [x] Fixture/replay/local-only tests and smoke; no live credential check or external write.
- [x] No real credential value requested, printed, summarized, or stored.
- [x] No connector definitions, dependency files, or legacy dynamic parser changes.
- [x] No services, generic tools, destructive/admin/production action, or quality-gate reduction.
- [x] Planning, RED, GREEN, help-correction, and direct-help checkpoints committed/pushed.
- [x] No PR/review per user instruction.

## Accepted review correction checklist

Correction start: `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`; session `issue-437-review-correction-20260719T113319Z`.

- [x] Read and accept all five findings in `/tmp/pm-397-review-437.log`.
- [x] Reopen all phase artifacts before test or production edits.
- [x] RED: unsupported `record`/`replay`/production-write/rate-limit/budget/live-all-modes controls reject before runner invocation.
- [x] RED: single certify emits exactly one connector span and preserves connector-validation-before-options precedence.
- [x] RED: batch credential-file load precedes parallel parsing and preserves exact load/run error wrappers and bytes.
- [x] RED: only `--help`, `-h`, and intentional positional `help` render connectors manuals; false/assigned malformed/unknown short clusters do not.
- [x] RED: CLI and website docs separate pre-report usage/validation/runtime exits from completed report outcomes.
- [x] Focused differential and repeated/race tests: base/current 5/5 exact; race `29.046s`; ×10 `24.991s`.
- [x] Certify exits, redaction, and unsupported replay no-live/no-write runner test: package race `349.263s`; exit focus `21.618s`.
- [x] Local sample fixture smoke with temp root only: exit 0, `ConnectorCertification`, sample passed, stderr empty, planted value absent.
- [x] Runtime help, golden, generated CLI docs, website generation/parity: CLI docs/golden `24.275s`; website regeneration hash-stable.
- [x] Full CLI and certify packages: CLI `435.572s`; certify `338.846s`.
- [x] `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`: final verify exit 0, real `468.36s`.
- [x] `go run ./cmd/connectorgen validate` reports 547 checked, zero findings.
- [x] Final artifacts committed and pushed (`2987f21b`); no dependencies/services/credentials/PR/review.

## Correction result

Implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`. Exact-base differential matched stdout, stderr, and exit codes in all five reviewed precedence/help cases. Unsupported replay and the other five controls return usage errors without any runtime call; no credential resolution, live stage, or write stage occurs. All verification used temporary roots, local sample behavior, synthetic planted redaction values, and existing replay fixtures only.

## Second accepted safety correction checklist

Second-correction start: `0d743e54e06c9e27e550eacce9be7899a9e23d19`; session `issue-437-second-safety-correction-20260719`.

- [x] Read and accept every P1/P2/P3 finding in `/tmp/pm-397-rereview-437.log`.
- [x] Reopen plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state before tests or production edits.
- [x] Commit/push the planning checkpoint before RED tests (`aa39fd9d`).
- [x] RED effect-recorder tests expose that `--write=false` and `--skip=write` do not override credential-file `write: true`.
- [x] RED exposes configured credential-file sandbox/rate/budget/limit reaching batch/runner rather than failing closed.
- [x] RED exposes visible/accepted `--credential`, `--limit`, and `--modes`, unrestricted skip values, and mode-inapplicable controls running effects.
- [x] RED audit enumerated every declared certify flag by supported mode or explicit fail-closed expectation; GREEN mapping later completed.
- [x] P3 stale certification architecture/PRD examples and claims removed; connector-help name claim made accurate.
- [x] Runtime help source, CLI docs, goldens, website docs, and generated data updated; final binary help checks remain in full verification.
- [x] Focused effect/no-op audit tests pass, repeated ×10, and under race (CLI `1.726s`, certify `2.535s`).
- [x] Full CLI (`440.910s`) and certify (`346.271s`) package suites preserve exits, redaction, dynamic dispatch, and valid base behavior.
- [x] Local sample smoke passed under a temporary root with no credentials/services; JSON kind, connector, and pass status asserted without printing report data.
- [x] Runtime topic/bare/flag help byte parity, certify help, invalid action exit 2, CLI docs/goldens, website full data regeneration, and drift checks pass; website typecheck tool is absent and no dependency install was attempted.
- [x] `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` pass; verify exit 0 in `7m36.852s` (CLI `434.190s`, certify `337.470s`, smoke/lint/docs/build/vet green).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` reports 547 checked, zero findings (both verify and explicit rerun).
- [x] Final verification artifacts committed and active issue branch pushed (`974495d5`); no dependencies/services/credentials/PR/review.

## Third accepted safety/correctness correction checklist

Third-correction start: `437d13cf`; session `issue-437-third-safety-correction-20260719`.

- [x] Read and accept every finding in `/tmp/pm-397-rereview2-437.log`.
- [x] Reopen plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state before tests or production edits.
- [x] Commit/push the planning checkpoint before RED tests (`766e2d9d`).
- [x] RED: every unknown certify flag, including write-like typos, rejects before credential loading, runner, sweep, or other effects.
- [x] RED: `--older-than` rejects zero, negative, and unreasonably large values before sweep.
- [x] RED: ordinary two-run `--resume` reuses completed prior reports without future timestamps; incomplete reports rerun.
- [x] RED: any credential-file `exec` entry rejects before runner effects; no test executes an external command.
- [x] Remove external credential execution code and all docs/examples/claims.
- [x] Re-audit every certify flag across modes and every CLI/manual/website claim.
- [x] Correct usage exit documentation and release-stage token examples to `ga`.
- [x] Focused and repeated unknown-flag/no-effect/exec/resume/sweep tests pass under race (CLI ×10 `1.283s`, certify ×10 `1.393s`, CLI race `1.713s`, certify race `2.261s`).
- [x] Runtime topic/bare/flag/certify help is byte-equal, invalid action and unknown typo exit 2, docs generation matches tracked CLI manuals, golden transcripts pass, and website docs regeneration is hash-stable.
- [x] Credential-free local sample smoke passed under a temporary root; `ConnectorCertification`, connector `sample`, pass, and empty stderr asserted without printing report data.
- [x] Full CLI (`446.382s`, real `7m49.747s`) and certify (`350.637s`, real `6m13.463s`) suites pass.
- [x] `gofmt -w cmd internal`, clean diff, `go vet ./...`, `go test ./...`, and `go build ./cmd/pm` pass.
- [x] `make verify` exits 0 in real `14m58.384s`; full tests, docs validation, ordered local smoke, lint 0, build/vet/tidy checks, and connector validation pass.
- [x] Explicit `go run ./cmd/connectorgen validate internal/connectors/defs` reports 547 checked, zero findings.
- [x] Final artifacts record only actual terminal evidence and align `verificationPassed` with the full gate.
- [x] Commit/push the final truthful verification artifacts only to the active issue branch (`3854295b`); no credentials, external credential commands, services, dependencies, PR, or review.

## Fourth bounded review-correction checklist

Exact start `1e27b14012f65ffa24c01ed855d0405c24401eee`; launcher `openai-codex/gpt-5.6-sol` / `high`.

- [x] Confirm isolated worktree, active branch, clean tree, exact HEAD, and matching local/remote branch heads.
- [x] Read full issues #437/#397/#407, contracts, GSD/manual loop, CLI parity/runtime references, ADR-0002, certification design, CLI Architecture v2 sources, migration context, all phase artifacts, both review outputs, and required skills.
- [x] Run GSD doctor/list/plan prompt; record unavailable programming-loop and manual fallback.
- [x] Consolidate every overlapping correctness/security finding into F1–F10 with accepted reasoned dispositions.
- [x] Commit/push planning checkpoint before tests or production edits (`07d0b5a4`).
- [x] Capture focused RED for every accepted finding before production changes; certify focus failed in `16.755s` across preview, secret, report, YAML, ledger, crontab, context, prerequisite, and resume tests; CLI focus failed in `1.051s` across booleans, parallel, secret config, and source-root isolation; RED `43acd262` pushed.
- [x] Failed/mismatched/leaky preview blocks execution and ledger in initial write, both cleanup paths, and both sweep paths.
- [x] Secret detector returns opaque metadata; approval/config secrets never enter argv/reasons/serialized reports; secret-schema config injection rejects; reports/history/progress/ledgers use restrictive atomic/no-follow writes as applicable.
- [x] Parallel certification uses invocation-local crontab confinement; deterministic concurrent and race tests prove system backend unreachable.
- [x] Durable per-connector ledger is directly consumed after process restart; workspace/ledger roots are separate; connector/run/tag/action/cleanup provenance and containment reject forged entries.
- [x] Caller context reaches in-process CLI stages; pre/mid-cancel stops later effects; already-started successful mutation gets bounded cleanup only.
- [x] Credential files are size-bounded, strict-known-field, supported-version, nonempty, count-bounded, registry/local-identifier and env-reference validated, secret-config rejecting, regular/no-symlink inputs.
- [x] Every boolean/safety control strictly parses; explicit parallel is positive/bounded, workers capped by jobs; sweep age remains bounded; failures occur before telemetry/credentials/runners/sweep/writes.
- [x] Structural/preflight/credential prerequisite failures prevent later live reads/writes.
- [x] Resume requires exact schema, connector/manifest identity, and secret-free options/reference fingerprint.
- [x] Native test helpers use `t.TempDir()` and assert no source-tree artifacts.
- [x] Focused tests repeated and `-race` for context/concurrency/crontab: certify ×10 `20.504s`, CLI ×10 `15.659s`, context/crontab race `1.743s`, batch race `4.216s`, CLI race `165.608s`.
- [x] Affected `internal/cli`, `internal/connectors/certify`, and schedule packages pass; standalone checkpoints were CLI `442.615s` and certify `328.909s`, and final-code package results were CLI `452.912s` / certify `346.633s` in explicit full tests and CLI `439.981s` / certify `330.355s` in verify.
- [x] Runtime `pm help connectors`, bare `pm connectors`, and `pm connectors --help` are byte-equal; certify help/JSON pass and invalid action exits 2.
- [x] CLI docs generated into a temporary root match `docs/cli`; goldens pass; website full-data regeneration is drift-free; docs validation passes.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` reports 547 checked, zero findings.
- [x] `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; explicit `go test ./...` (real `456.93s`); and `go build ./cmd/pm` pass.
- [x] Full `make verify` exits 0 on the required complete rerun (real `464.41s`; CLI `439.981s`, certify `330.355s`, lint 0, docs/smoke/connectorgen green).
- [x] No go.mod/go.sum delta and no unjustified files outside #437 correction scope.
- [x] Final commits pushed only to active branch; terminal artifact was later committed/pushed in cycle history; no PR/integration/parent/main mutation.

The first `make verify` attempt failed only at lint after tests because four new fixture writes did not check `fmt` return values. Commit `b06816ad` fixed all 17 fixture writes consistently; focused tests and targeted certify lint passed before the complete successful rerun. All runtime checks used temporary roots and the dependency-free sample/local smoke only.

## Fifth bounded review-correction checklist

Exact start `05d9c6658f52e542b6a74e87e29bdcad7275ea9d`; identity `issue-437-fifth-review-correction-20260720`; launcher `openai-codex/gpt-5.6-sol` / `high`.

`verificationPassed`: true — complete final gate exited 0.

- [x] Confirm isolated branch, clean exact start, and matching local/remote branch head.
- [x] Read issues #437/#397/#407, required contracts/workflows, CLI parity, ADR/design/migration context, all required skills, all six artifacts, and both rereview files.
- [x] Run GSD doctor/list; capture missing programming-loop; generate applicable `audit-fix --dry-run` prompt; record manual fallback and `local_critical_path` no-subagent decision.
- [x] Record explicit recovery-budget exception for unresolved P1 security findings.
- [x] Consolidate overlaps and accept all seven findings with root-cause dispositions.
- [x] Commit/push planning before RED (`8acf62a9`).
- [x] RED: normal cleanup and sweep execute-success/verify-failure became cleaned/non-retryable.
- [x] RED: validly shaped forged GitHub issue/milestone ledger entries each reached a CLI effect.
- [x] RED: oversized/many ledgers lacked total/count bounds and malformed input reflected the planted marker.
- [x] RED: seven effective sweep constraints returned nil and created project state.
- [x] RED: current/history symlink/unwritable save failures produced false success; leak exit remained 3 but persistence failure was silent.
- [x] RED: unknown/malformed assigned and space-value certify flags created telemetry; malformed space parallel was late validation.
- [x] RED: minimal/edited/duplicate resume artifacts were trusted without rerun.
- [x] Commit/push coherent RED before production (`e2559f64`, supplemental pre-telemetry RED `3d69b7a4`).
- [x] GREEN accepted corrections with focused tests.
- [x] Focused tests repeated ×10 and under `-race` for cleanup/sweep/concurrency/prevalidation.
- [x] Affected full CLI (`443.427s`), full certify (`327.840s`), schedule, safety, and scoped vet pass.
- [x] Runtime `pm help connectors`, bare `pm connectors`, and `pm connectors --help` are byte-equal; certify help/JSON and invalid usage exit 2 pass.
- [x] CLI docs/goldens pass (`16.600s`); website regeneration is drift-free; no canonical docs change was needed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` reports 547 checked, zero findings.
- [x] `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; explicit `go test ./...` real `7m34.316s`; `go build ./cmd/pm` pass.
- [x] Full `make verify` exits 0 in real `7m52.496s` (CLI `449.572s`, certify `332.793s`, docs/smoke/lint/connectorgen green).
- [x] No go.mod/go.sum delta, connector defs, credentials, services, system crontab, external writes/sweeps, generic write tools, dependencies, PR, parent, integration, or main mutation.
- [x] GREEN `e9ce945e` pushed; truthful terminal artifact was later committed/pushed clean and remote-matched.

## Continuation checklist — parent reconcile and stacked PR

Exact start: `86eea0f966814e6848e5a52143eea15dd46ff801`; parent target: `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa`.

- [x] Reread required issue/contract/GSD/parity/ADR/phase/skill context before mutation.
- [x] Run `scripts/gsd doctor` and `scripts/gsd prompt plan-phase 437 --skip-research`.
- [x] Confirm `scripts/gsd prompt programming-loop init --phase 437 --dry-run` is unavailable and record manual fallback.
- [x] Record actual skills loaded and `.pi/skills/go-implementation/SKILL.md` absence/path mismatch.
- [x] Refresh PLAN/TDD/VERIFICATION before parent merge or production edits.
- [x] Audit the existing 42-file issue diff against #437 and justify certify safety correction scope: all files are phase artifacts, native connectors/certify CLI code/tests/goldens, certify safety code/tests, command harness seam, or required docs/website parity; no unrelated file, dependency, or connector-def delta.
- [x] Merge latest parent `origin/feat/cli-architecture-v2` without losing #437 commits or #462 design-doc additions: clean merge commit `dc4aed23dcc42878f48da62fe7f1a236e2103ed1`, no conflicts.
- [x] Focused CLI/connectors/certify tests pass: `go test ./internal/cli -run 'TestConnectorsCommandIsNativeCobraSubtree|TestNativeConnectors|TestNativeCertify|TestConnectorsManual|TestCertifyCLI|TestTelemetryCertify' -count=1` => `ok ... 119.151s`; `go test ./internal/connectors/certify -run 'TestRereview|TestReviewCorrection|TestBatch|TestCredsFile|TestLedger|TestReport|TestSweeper|TestStage' -count=1` => `ok ... 7.344s`.
- [x] Runtime help checks pass: built `./pm`; `./pm help connectors`, bare `./pm connectors`, and `./pm connectors --help` are byte-identical (`8391` bytes).
- [x] Golden/docs/website parity checks pass: focused CLI docs/golden tests `ok ... 10.347s`; `cd website && node scripts/gen-docs-data.mjs` wrote 11 pages; tracked docs/website diff clean.
- [x] Safe fixture-only certify smoke passes without credentials/live services/external writes: `./pm connectors certify sample --root $(mktemp -d) --json` exit 0, `kind=ConnectorCertification`, report connector `sample`, passed `true`, stderr bytes `0`.
- [x] `gofmt -w cmd internal` produces no unintended drift; `git diff --check` passes.
- [x] `go vet ./...` passes.
- [x] `go test ./...` passes; full CLI package `457.441s`.
- [x] `go build ./cmd/pm` passes.
- [x] `make verify` passes; full CLI package `461.542s`, docs validate, local Makefile smoke passed, lint `0 issues`, connectorgen validation green.
- [x] Connector validation remains green when run: `go run ./cmd/connectorgen validate internal/connectors/defs` => `547 connector(s) checked, 0 findings`.
- [x] No go.mod/go.sum delta, connector defs, credentials, live checks, destructive cleanup, external writes/sweeps, generic write tools, or main/parent merge.
- [x] Branch pushed and non-draft stacked PR opened to `feat/cli-architecture-v2` with exactly `Refs #437`, `Refs #407`, and `Refs #397`: https://github.com/polymetrics-ai/cli/pull/466.
- [x] Automated external review route recorded: Claude disabled, Copilot quota exhausted, human/parent fallback pending; no bot retries.


## Sixth bounded exact-head review-correction checklist

Exact start: `8e7e2533c75451114c4d6ae38f89b7fd1ede6c34`; identity `issue-437-sixth-review-correction-20260719T220843Z`.

`verificationPassed`: true — complete final `make verify` rerun exited 0 after the lint correction.

- [x] Confirm clean branch, local head, remote head, and PR #466 head all equal `8e7e2533c75451114c4d6ae38f89b7fd1ede6c34` before edits.
- [x] Reread required issue, contracts, GSD/manual loop, skill routing, CLI parity, review routing, ADR/design/migration context, and all phase artifacts.
- [x] Run `scripts/gsd doctor`; capture missing `programming-loop`; generate/read `scripts/gsd prompt plan-phase 437 --skip-research`; record manual fallback.
- [x] Record actual execution decision `local_critical_path` and parent-spawn note.
- [x] Record recovery-budget exception for unresolved High leak-dominance/resource-authority risks.
- [x] RED: approval replay success cannot leave a post-cleanup untracked resource — failed with final outbox action `"create"` after replay success.
- [x] RED: runner report+error with leaks remains exit 3 and preserves report evidence — failed with batch exit `2`.
- [x] RED: progress persistence failure with existing leaks remains exit 3 and preserves safe batch report output — failed with exit `1` and Error envelope.
- [x] RED: cleanup call failure followed by exact absence proof clears stale top-level leak without marking the stage honestly passing — failed with stale `Report.Leaks`.
- [x] GREEN: minimal fixes pass focused new tests (`certify` 21.288s; `cli` 0.565s).
- [x] Repeated focused tests and race variants pass (`certify` count=3 62.605s, `cli` count=3 0.567s, `certify -race` 229.386s, `cli -race` 1.666s).
- [x] Affected full `go test ./internal/connectors/certify` (`351.189s`) and `go test ./internal/cli` (`445.489s`).
- [x] Runtime help/docs/golden/website parity: help byte-identical `8391` bytes; focused docs/golden `11.149s`; website generation drift-free; no canonical text update required.
- [x] Fixture-only `./pm connectors certify sample --root <temp> --json` smoke: exit 0, `ConnectorCertification`, sample, passed true, stderr bytes 0.
- [x] `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; `go test ./...` (CLI `445.908s`, certify `350.792s`); `go build ./cmd/pm`; final `make verify` (CLI `453.202s`, certify `357.136s`, lint 0, smoke ok); explicit `go run ./cmd/connectorgen validate internal/connectors/defs` => 547/0.
- [x] PR #466 body update completed for sixth cycle and records exact head `6e9e7d9422050a609306d8900d6a06c8bb1fc223`.
- [x] No credentials, live certification, services, external writes/sweeps, dependencies, generic write tools, bot review request, parent/main merge, or quality-gate reduction.

## Seventh bounded exact-head correction checklist

Exact start: `6e9e7d9422050a609306d8900d6a06c8bb1fc223`; identity `issue-437-seventh-bounded-correction-20260720`.

`verificationPassed`: true — complete final `make verify` exited 0 for seventh cycle.

- [x] Confirm clean branch, local head, remote head, and PR #466 head all equal `6e9e7d9422050a609306d8900d6a06c8bb1fc223` before edits.
- [x] Reread required issue, contracts, GSD/manual loop, skill routing, CLI parity, ADR/design/migration context, and all phase artifacts.
- [x] Run `scripts/gsd doctor`; capture missing `programming-loop`; generate/read `scripts/gsd prompt plan-phase 437 --skip-research`; record manual fallback.
- [x] Record execution decision `local_critical_path` and parent-spawn note.
- [x] Record recovery-budget exception for unresolved effect-before-usage and resume-replay findings.
- [x] Planning checkpoint committed and pushed before RED tests or production edits (`82f59229`).
- [x] RED: absence-proven cleanup-failure report resumes without rerunning; current code failed by invoking runner (`0.698s`).
- [x] RED: bare value-required flags reject as usage exit 2 before logger/telemetry/credentials/runner/workspace/sweep effects; current code failed by reaching stream `true`, path `true`, fake runtime effects, or validation exit 3 (`5.853s`).
- [x] GREEN: strict value-required validation preserves `--flag=value` and `--flag value`, boolean bare flags, fail-closed unsupported controls, and mode-specific no-effect behavior.
- [x] GREEN: resume accepts only structurally complete provenance-consistent cleanup-failure absence proof; incomplete/edited/leaky/unproven reports still rerun.
- [x] Focused repeated and race variants pass: certify count=3 `0.660s`, CLI count=3 `0.575s`, certify race `1.704s`, CLI race `1.729s`.
- [x] Full `go test ./internal/connectors/certify -count=1` (`353.654s`) and `go test ./internal/cli -count=1` (`448.666s`) pass.
- [x] Runtime help/no-effect/exit checks pass: `./pm help connectors`, bare `./pm connectors`, `./pm connectors --help` byte-identical (`8391` bytes); bare `--stream` exits 2, says `--stream requires a value`, and creates no `.polymetrics`.
- [x] Fixture-only `./pm connectors certify sample --root <temp> --json` smoke passes without credentials/live services/external writes: exit 0, `ConnectorCertification`, sample, passed true, stderr bytes 0.
- [x] Docs/golden/website parity checked: focused CLI docs/golden `11.210s`; `cd website && node scripts/gen-docs-data.mjs` wrote 11 pages; no tracked drift.
- [x] `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; `go test ./...` (CLI `444.782s`, certify `349.644s`); `go build ./cmd/pm`; final `make verify` (CLI `442.999s`, certify `348.395s`, docs validate, ordered local smoke `smoke ok`, lint `0 issues`, connectorgen green); explicit `go run ./cmd/connectorgen validate internal/connectors/defs` => 547/0 all pass.
- [x] PR #466 body update completed after terminal artifact commit; final seventh exact head is `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8`.
- [x] No credentials, live certification, services, external writes/sweeps, dependencies, generic write tools, bot review request, parent/main merge, or quality-gate reduction.

## Eighth bounded exact-head correction checklist

Exact start: `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8`; identity `issue-437-eighth-bounded-correction-20260720`.

`verificationPassed`: true — complete final local gates passed for the eighth correction.

- [x] Confirm clean branch, local head, remote head, and PR #466 head all equal `0d515e6ec8ac11a6e049f8f7f8390725dc5b5dd8` before edits.
- [x] Reread required issue, contracts, GSD/manual loop, skill routing, CLI parity, ADR/design/migration context, and all phase artifacts.
- [x] Run `scripts/gsd doctor`; capture missing `programming-loop`; generate/read `scripts/gsd prompt plan-phase 437 --skip-research`; record manual fallback.
- [x] Record execution decision `local_critical_path` and parent-spawn note.
- [x] Record eighth recovery exception for future-evidence and whitespace effect-before-usage findings.
- [x] Planning checkpoint committed and pushed before RED tests or production edits (`45f190dc`).
- [x] RED: future-dated ordinary completed report reruns instead of resuming; current code failed by resuming.
- [x] RED: future-dated cleanup-failure absence-proof report reruns instead of resuming; current code failed by resuming.
- [x] RED: empty/whitespace space-form value-required flags reject as usage exit 2 before logger/telemetry/credentials/runner/workspace/sweep effects; current code failed by reaching later validation/effects.
- [x] GREEN: report timestamps materially beyond documented skew tolerance are future-invalid, while historical reports and `CompletedAt >= StartedAt` remain valid.
- [x] GREEN: `strings.TrimSpace(next)==""` is value-missing for all value-required certify flags; valid assigned/space values still pass and boolean bare flags remain valid.
- [x] Focused repeated and race variants pass: certify count=3 `1.192s`, CLI count=3 `6.887s`, certify race `2.450s`, CLI race `23.441s`.
- [x] Full `go test ./internal/connectors/certify -count=1` (`347.792s`) and `go test ./internal/cli -count=1` (`443.361s`) pass.
- [x] Runtime help/no-effect/exit checks pass: `./pm help connectors`, bare `./pm connectors`, and `./pm connectors --help` byte-identical (`8391` bytes); whitespace `--stream` exits 2, says `--stream requires a value`, and creates no `.polymetrics`.
- [x] Fixture-only `./pm connectors certify sample --root <temp> --json` smoke passes without credentials/live services/external writes: exit 0, `ConnectorCertification`, sample, passed true, stderr bytes 0.
- [x] Docs/golden/website parity checked: focused CLI docs/golden `11.208s`; `cd website && node scripts/gen-docs-data.mjs` wrote 11 pages; tracked docs/website diff clean.
- [x] `gofmt -w cmd internal`; `git diff --check`; `go vet ./...`; `go test ./...` (CLI `446.588s`, certify `349.095s`); `go build ./cmd/pm`; final `make verify` (CLI `445.358s`, certify `348.425s`, docs validate, ordered local smoke `smoke ok`, lint `0 issues`, connectorgen green); explicit `go run ./cmd/connectorgen validate internal/connectors/defs` => 547/0 all pass.
- [x] PR #466 body update completed at terminal-evidence parent/previous PR head `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`; current live PR head is intentionally sourced from Git/GitHub instead of a self-referential artifact line.
- [x] No credentials, live certification, services, external writes/sweeps, dependencies, generic write tools, bot review request, parent/main merge, or quality-gate reduction.

## Docs-only evidence closure checklist

Exact clean local/remote/PR start: `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`. No production edit and no behavior RED required.

- [x] Read issue #437 acceptance criteria, required contracts/workflows, phase artifacts, GSD adapter, skill routing, CLI parity, and required docs/architecture context.
- [x] GSD route refreshed: `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase 437 --skip-research`.
- [x] Execution decision recorded as `local_critical_path`; parent records spawned.
- [x] Reviewed implementation head retained as `af0e4dabf5be70237c02403e6ef4f003042667d6`.
- [x] Terminal-evidence parent/previous PR head recorded as `f211562ef4fd64ee7d7de4f274a3facf6ff44f51`.
- [x] Current live PR head declared authoritative from Git/GitHub and intentionally not embedded self-referentially.
- [x] JSON parse, stale terminal/PR marker grep, `git diff --check`, and exact diff-scope check only phase artifacts pass.
