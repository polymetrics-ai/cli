# Phase 437 TDD Ledger

Issue #437; invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

GSD doctor/list and plan-phase prompt passed. Adapter has no `programming-loop`; manual universal-runtime-loop fallback was used. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

| Step | Kind | Evidence | Status |
|---:|---|---|---|
| 0 | Planning | Six issue-local artifacts written before tests/production | Complete |
| 1 | Initial RED | `go test ./internal/cli -run 'TestConnectorsCommandIsNativeCobraSubtree|TestNativeConnectors|TestNativeCertify' -count=1` | Failed before production: `newConnectorsCobraCommand` and `newConnectorsCobraCommandWithRuntime` undefined |
| 2 | GREEN | Native subtree, typed flags/runtime seam, compatibility normalization, namespace-only parser removal | Focused `3.876s`; router/golden/certify/telemetry `111.507s`; repeated ×10 `34.915s`; race `40.844s`; certify package `336.422s` |
| 3 | Help RED | Trailing list/catalog/inspect/certify help contract | Failed before final help edit: operations ran, certify positional help returned usage 2, JSON trailing help returned certification output |
| 4 | Help GREEN | Contextual direct/trailing help before effects | Focused `3.884s` |
| 5 | Operand-help RED | Direct `connectors inspect --help|help` | Failed before correction: help token captured as connector name and returned internal connector-not-found |
| 6 | Operand-help GREEN | Help-aware private operand capture | Focused `3.989s` |
| 7 | Refactor | Repeated/race/router/golden/differential/parity | Final native ×10 `34.833s`; native race `40.842s`; router/golden/certify/telemetry `111.919s`; exact-start operations 21/21 |
| 8 | Full gate | certify smoke, connector validation, docs/website, gofmt/vet/test/build/make verify | Complete: smoke exit 0/pass; validation 547/0; final `make verify` exit 0 |
| 9 | Delivery | Finalize six artifacts, commit/push, no PR/review | Complete |

## Contract proven

- `connectors` is native Cobra and absent from legacy wrappers; native list/catalog/inspect/manual aliases/help/certify are declared.
- Certify supports single connector, `--all`, and `--sweep`, with the current command-contract flags represented as repeatable `StringArray` flags with `NoOptDefVal=true`.
- Repeated values, bare/assigned/space forms, positional connector selection, ignored tails, unknowns, literal `--`, malformed action heads, no later discovery, and globals retain compatibility.
- Bare/topic/direct/long/short/positional/trailing help is canonical text/JSON and side-effect free; invalid actions remain usage errors.
- Certification reports retain exits 0 pass, 1 internal, 2 certification failure, 3 leak dominance and one-envelope output.
- Fresh command trees, context cancellation, bounded batch concurrency, event sequence, telemetry spans, and credential-value exclusion remain covered.
- Dynamic connector dispatch remains the only CLI `parseFlags` consumer pair; `parse.go` is unchanged.

## Final evidence

- Full `make verify`: CLI `431.305s`, certify `337.280s`, docs, ordered local smoke, lint 0, connector validation 547/0.
- Required explicit local certify smoke: exit 0, sample `ConnectorCertification`, passed, empty stderr.
- No live credential check/write/sweep, service, dependency, connector definition, or private material used.

## Accepted review correction ledger

Session `issue-437-review-correction-20260719T113319Z`; exact correction start `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`; all five findings in `/tmp/pm-397-review-437.log` accepted.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| C0 | Planning | Six artifacts reopened before tests/production; GSD manual fallback, skills, safety scope, and verification commands recorded | Complete |
| C1 | RED | `go test ./internal/cli -run 'TestNativeCertifyRejectsUnsupportedSafetyAndModeControls|TestNativeConnectorsOnlyExactHelpFormsRenderManual|TestConnectorsManualSeparatesCLIAndCompletedCertificationExits|TestCertifyCLIBatchLoadsCredentialsBeforeValidatingParallelAndPreservesErrorBytes|TestTelemetryCertifyInvalidOptionsPreserveSingleSpanAndConnectorValidationPrecedence' -count=1` | Complete: unsupported controls invoked runtime and were visible; false/malformed/cluster help rendered manuals; docs phrases absent; missing creds gained batch wrapper and invalid parallel won; invalid connector lost to option usage and emitted no certify span |
| C2 | GREEN | Hidden fail-closed controls; single connector span/validation/options ordering; batch injected load then parallel then wrapped run; exact-only connectors help normalization; canonical/generated/website docs | Complete: focused `3.004s`; native/certify/telemetry `108.532s` |
| C3 | Refactor | Base/current differential; repeated/race; certify exit/redaction/replay-no-live; local sample; docs/golden/website parity | Complete: differential 5/5 byte-identical; focused race `29.046s`; ×10 `24.991s`; certify redaction/replay/concurrency race `349.263s`; exit focus `21.618s`; sample exit 0/pass/redacted; docs/golden `24.275s`; website regeneration hash-stable |
| C4 | Full gate | gofmt, vet, full tests, build, `make verify`, connector validation | Complete: full CLI `435.572s`; certify `338.846s`; vet/build pass; validation 547/0; `make verify` exit 0 (`468.36s`, CLI `444.436s`, certify `346.018s`, smoke/lint/docs/validation green) |
| C5 | Delivery | Finalize artifacts, commit/push, no services/credentials/PR/review | Complete: terminal verification artifact commit `2987f21b` pushed; final delivery marker follows |

Correction RED must be captured and committed before any production edit. Fixture/temp data only; no live credentials, HTTP writes, reverse ETL execution, or external sweep.

## Second accepted safety correction ledger

Session `issue-437-second-safety-correction-20260719`; exact start `0d743e54e06c9e27e550eacce9be7899a9e23d19`; all P1/P2/P3 findings in `/tmp/pm-397-rereview-437.log` accepted.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| S0 | Planning | Six artifacts reopened with verification false, accepted findings, GSD fallback, skills, safety scope, flag audit, and checkpoint plan | Complete: commit `aa39fd9d` pushed |
| S1 | RED | Effect recorder: batch write-disable dominance; credential constraint fail-closed; unsupported/mode-inapplicable controls and skip values rejected before effects; docs/help claims fail | Complete: focused command failed as required—unsupported single flags visible and ran; 18 skip/mode no-ops recorded runtime effects; both write-disable forms preserved entry `write:true`; seven credential constraints reached batch/runner; sandbox/rate/budget/limit direct batch cases invoked factory; stale docs/help claims remained |
| S2 | GREEN P1/P2 | Batch safe overrides; no discarded credential constraints; every declared certify flag used or fail-closed; only implemented skip values accepted | Complete: effect/no-op focus passed; unsupported single controls hidden/rejected; mode validation runs before runtime effects; batch disable overrides precede credential validation; sandbox gates writes; unsupported rate/budget/limit fail before factory |
| S3 | GREEN P3 | Architecture/PRD/help claims corrected; CLI and website artifacts regenerated | Complete: stale rejected controls removed; help accurately names namespace manual; CLI docs/goldens and website docs data regenerated |
| S4 | Verify | Focused/repeated/race/no-op audit/sample smoke/full CLI+certify/docs+website/gofmt+vet+test+build+make verify+connectorgen | Complete: repeated `0.661s`; race CLI `1.726s`/certify `2.535s`; full CLI `440.910s` then verify `434.190s`; full certify `346.271s` then verify `337.470s`; no-op/help/sample/docs/website pass; `make verify` exit 0 in `7m36.852s`; connectorgen 547/0 |
| S5 | Delivery | Finalize artifacts and commit/push; no credentials/services/dependencies/PR/review | Complete: verification artifact `974495d5` pushed; final delivery marker follows |

Strict TDD gate: S1 failing tests and observed failures must be captured and committed before any production edit. All data remains fixture/synthetic-reference/temp-only; no secret values, credential resolution, live checks, external writes, services, dependency changes, PR, or review.

## Third accepted safety/correctness correction ledger

Session `issue-437-third-safety-correction-20260719`; exact start `437d13cf`; all findings in `/tmp/pm-397-rereview2-437.log` accepted. The adapter's programming-loop command is unavailable, so the recorded fallback is the manual universal runtime loop.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| T0 | Planning | Six artifacts reopened with verification false, accepted findings, GSD fallback, skills, safety scope, flag/docs audit, RED/GREEN plan, and checkpoints | Complete: commit `766e2d9d` pushed |
| T1 | RED | Unknown certify flags and write-like typos reject before credential load/runner/sweep; non-positive and excessive sweep age reject before effects; credential-file exec rejects before effects; ordinary two-run resume skips completed reports and reruns incomplete reports | Complete: focused test failed as required—certify retained unknown-flag whitelist and all seven typo cases reached mode handling/runtime; five unsafe ages swept; exec reached batch and parsed from file; second ordinary resume reran; direct exec constraint reached runner; stale `generally_available` docs remained |
| T2 | GREEN safety | Strict certify flag parsing; bounded positive sweep age; exec removed/rejected; no generic external execution path | Complete: certify unknown whitelist disabled; sweep accepts only `(0, 8760h]`; loader, resolver, and batch reject exec; `os/exec` path removed |
| T3 | GREEN correctness/docs | Completed-report resume semantics; incomplete reports rerun; usage exit, `ga`, resume/sweep/exec docs and generated artifacts corrected | Complete: resumed results reuse prior report/exit; malformed or incomplete artifacts rerun; canonical/generated/golden/website docs corrected |
| T4 | Verify | Focused/repeated/race/resume/sweep/no-effect/audit/docs generation/local smoke/full CLI/certify/gofmt/vet/test/build/make verify/connectorgen | Complete: focused/repeated/race green; CLI `446.382s`; certify `350.637s`; `make verify` exit 0 in `14m58.384s`; connectorgen 547/0; docs and website drift-free |
| T5 | Delivery | Truthful final artifacts, coherent commits, active-branch push; no credentials/external credential commands/services/dependencies/PR/review | Complete: terminal verification artifact `3854295b` pushed; final delivery marker records closure |

Strict TDD gate: T1 failures and effect-recorder observations must be captured and committed before production edits. Tests use only fixtures, synthetic references, in-process fakes, and `t.TempDir`; no test may invoke an external credential command.

## Fourth bounded review-correction ledger

Session `issue-437-sol-high-review-correction-20260719`; exact start `1e27b14012f65ffa24c01ed855d0405c24401eee`; model `openai-codex/gpt-5.6-sol`, thinking `high`. Both independent review files were read and traced. Consolidated dispositions F1–F10 in PLAN.md are accepted. Programming-loop is absent; manual universal-loop fallback applies. Required and added skills are recorded in PLAN.md.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| R4-0 | Planning | Six artifacts reopened before tests/production; exact branch/head/remote/clean checks, reviews, dispositions, GSD route, skills, safety scope, RED/GREEN plan, gates, checkpoints, and execution decision recorded | Complete; `07d0b5a4` pushed before tests/production |
| R4-1 | RED preview/security | Focused effects prove failed initial/cleanup/sweep preview cannot run or ledger; opaque secret hits and semantic argv/config/report nondisclosure fail on current code | Captured: create executed once without validated preview; cleanup order was `[plan run]`; detector metadata copied marker; report mode was `0644`; malformed booleans and secret config reached runtime with exit 0 |
| R4-2 | RED confinement/context | Deterministic concurrent crontab test proves current global env race; pre/mid-cancel tests prove context is dropped and later effects continue | Captured: barrier test observed 4 cross-run/operator crontab selections; pre-cancelled Runner returned nil and invoked CLI |
| R4-3 | RED durability/input | Fresh-process durable sweep/layout/provenance, strict YAML/version/jobs/identifier/env/secret-config/path/symlink/atomic-mode, boolean/parallel/age/prerequisite/resume compatibility, and source-tree pollution tests fail on current code | Captured: strict YAML cases and symlink accepted; durable expected layout absent; failed preflight reached credentials; future/incompatible resume accepted; explicit parallel 0/-1/33 ran; helper root was `.`; RED `43acd262` pushed |
| R4-4 | GREEN | Smallest context-aware, invocation-local, durable, strict, fail-closed implementation passes every focused test | Complete: `2c0a550c` pushed; opaque redaction, preview gates, local crontab, durable/provenance ledgers, strict files/controls, cancellation cleanup, prerequisites, identity resume, private atomic artifacts, and temp roots green |
| R4-5 | Refactor/verify | Repeated/race variants, affected packages, CLI parity, docs/goldens/website drift, connector validation, fmt/diff/vet/test/build/`make verify` | Complete: ×10 certify `20.504s`, CLI `15.659s`; race context/crontab `1.743s`, batch `4.216s`, CLI `165.608s`; standalone package checkpoints CLI `442.615s`, certify `328.909s`; final-code full tests CLI `452.912s`, certify `346.633s`, real `456.93s`; final `make verify` exit 0 real `464.41s` |
| R4-6 | Delivery | Truthful terminal artifacts, coherent GREEN/evidence commits and active-branch pushes; clean worktree | Complete after terminal artifact commit/push; GREEN `2c0a550c` and lint fix `b06816ad` already pushed |

The first full `make verify` ran all tests but failed at lint on four unchecked `fmt.Fprint/Fprintf` fixture writes. The resource-handling correction was committed as `b06816ad`; focused tests and certify lint passed, then the entire `make verify` gate was rerun and exited 0. Final verification also includes help/bare/flag-help byte parity, invalid action exit 2, JSON manual kind, credential-free sample pass, temporary CLI-doc generation diff, golden pass, website regeneration with no drift, docs validation, and connectorgen 547/0.

Strict TDD gate: R4-1 through R4-3 exact failing outputs must be captured before any production edit. Tests use fixtures, fake runners/backends, in-process seams, synthetic markers, and temporary roots only; they must assert marker absence without printing report bodies.

## Fifth bounded review-correction ledger

Identity `issue-437-fifth-review-correction-20260720`; exact start `05d9c6658f52e542b6a74e87e29bdcad7275ea9d`; both rereviews traced and consolidated into seven accepted findings in PLAN.md. The explicit recovery-budget exception applies because unresolved P1 security risks require another correction. Programming-loop remains unavailable; applicable `audit-fix --dry-run` prompt plus manual universal loop applies. No subagents by user order; execution is `local_critical_path`.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| R5-0 | Planning | Reopen all six artifacts with exact identity, findings/dispositions, skills, recovery exception, RED/GREEN/gates/checkpoints; commit/push before tests or production | Complete: `8acf62a9` pushed before RED |
| R5-1 | RED cleanup/authority | Cleanup execute success + verify failure remains uncleaned; sweep verify failure remains retryable; forged issue/milestone ledger entries produce zero plan/preview/run/cleanup effects | Captured: normal and sweep marked cleaned; both numeric pairings reached one CLI effect |
| R5-2 | RED inputs/constraints | Oversized, too-many-entry, and malformed-marker ledgers fail before effects without marker output; sweep effective default/per-connector rate/budget/limit/other constraints fail before workspace/telemetry/harness/cleanup | Captured: 1 MiB input lacked bound error; 10,001 entries accepted; malformed marker reflected; seven sweep constraints returned nil and created project state |
| R5-3 | RED persistence/prevalidation | Current/history symlink or unwritable report save surfaces deterministically with leak exit 3 dominant; assigned/space/unknown malformed certify flags produce no logger/telemetry files or runtime effects | Captured: three passing save failures exited 0; leaked save failure was silent; invalid assigned/space forms created telemetry; bad space parallel classified validation after effects |
| R5-4 | RED resume | Minimal, duplicate-stage, edited outcome/leaks, future, and incompatible reports rerun; valid complete evidence resumes with recomputed result | Captured: all four new invalid/edited reports resumed without runner invocation |
| R5-5 | GREEN | Smallest fail-closed implementation satisfies R5-1 through R5-4 while preserving plan → preview → approval → execute, context, crontab isolation, dynamic dispatch, re-entrancy, outputs/exits, docs, and no dependencies | Complete: focused certify `7.288s`, CLI `8.389s`; ×10 certify `48.652s`, CLI `77.313s`; race certify `51.469s`, CLI `89.921s`; full certify `327.840s`, full CLI final `443.427s`; schedule/safety and scoped vet pass |
| R5-6 | Verify | Focused/repeated/race; CLI/certify/schedule/safety; help/bare/invalid/JSON; docs/goldens/website; connector validation; fmt/diff/vet/full tests/build/`make verify`; no module delta | Complete: explicit full tests real `7m34.316s`; final `make verify` exit 0 real `7m52.496s`; all listed gates green |
| R5-7 | Delivery | Coherent planning/RED/GREEN/evidence commits pushed only to active issue branch; clean and remote-matched; no PR/parent/integration/main mutation | Complete after this terminal artifact commit/push; implementation head `e9ce945e` already pushed |

Strict TDD gate: R5-1 through R5-4 failures must be captured and committed/pushed before any production edit. Tests use only deterministic fakes, temporary paths, synthetic non-secret markers, and effect recorders; raw malformed lines/report bodies/credential values must never be printed.

## Continuation ledger — parent reconcile, verification, stacked PR

Identity `issue-437-continuation-20260719T211738Z`; exact start `86eea0f966814e6848e5a52143eea15dd46ff801`; latest parent `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa`. `scripts/gsd doctor` and `scripts/gsd prompt plan-phase 437 --skip-research` ran; `programming-loop` remains unavailable, so manual universal-loop fallback applies. Execution decision: `local_critical_path`.

Loaded skills: `.pi/skills/gsd-core/SKILL.md`, `.agents/skills/caveman/SKILL.md`, and global cc-skills paths for `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`. `.pi/skills/go-implementation/SKILL.md` is absent; global Go skill files are the actual evidence.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| C6-0 | Planning | Refresh PLAN/TDD/VERIFICATION before parent merge or production edits; record GSD route, skills, mismatch, safety, and no-new-behavior TDD stance | Complete in this artifact edit |
| C6-1 | Diff audit | Inspect existing 42-file diff against #437; justify directly applicable certify safety corrections; block unrelated changes | Complete: 42 files classify as 6 issue phase artifacts, 1 `cmd/pm` certify harness seam, 11 `internal/cli` native connectors/certify files/tests/goldens, 19 `internal/connectors/certify` safety files/tests, and 5 docs/website parity files; no go.mod/go.sum or connector defs delta; no unrelated files kept |
| C6-2 | Parent reconcile | Merge latest `origin/feat/cli-architecture-v2` (`a5474bcb`) without dropping #437 commits; resolve only real conflicts | Complete: clean `ort` merge, no conflicts; parent #462 design docs/skills/traces retained; post-merge `HEAD=dc4aed23dcc42878f48da62fe7f1a236e2103ed1`, ahead 36 / behind 0 versus parent |
| C6-3 | RED gate | No new behavior edit planned; do not fabricate RED. Existing RED remains valid. Any new behavior fix requires a failing test first or a stop | Complete: no new behavior defect or production edit was introduced in continuation; no new RED fabricated |
| C6-4 | Verification | Focused CLI/connectors/certify, runtime help, docs/golden/website parity, safe fixture smoke, gofmt, vet, full tests, build, `make verify`, connector validation | Complete: focused CLI `119.151s`, certify `7.344s`; help byte-equal `8391` bytes; docs/golden `10.347s`; website gen clean; certify sample exit 0/pass; gofmt/diff/vet/full tests/build; `make verify` exit 0; connectorgen 547/0 |
| C6-5 | Delivery | Commit/push coherent continuation evidence and open non-draft stacked PR to parent; record Claude disabled and Copilot quota exhausted with human/parent fallback pending | Pending |
