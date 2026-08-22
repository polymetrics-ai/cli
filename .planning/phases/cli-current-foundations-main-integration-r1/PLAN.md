# Current Foundations Main Integration r1

## Task Delivery Header

- Issue: Refs #4302 — status/export foundation; Refs #4303 — declarative reverse ETL; Refs #4305 — structured REST body; Refs #4306 — source import; Refs #4307 — closed operation runtime.
- Provisional integration base: `114a67727f2ef60b132054091c73987be4118a9b` (Firstmate-approved clean branch state; preserved parent `9e3cd99b7ebd2ebac2303ad8770e50fee85c92c6`).
- Merges into: `fm/cli-current-foundations-main-integration-r1 → main`.
- Delivery: One human-gated rollup PR against `main`, created only by the later no-mistakes stage after this branch has the exact qualified component heads, local production gates, actual-provider qualification, and a complete evidence manifest.
- Working branch: `fm/cli-current-foundations-main-integration-r1`.
- Task: Preserve the exact five component histories and compose the source-import, closed-runtime, status/export, structured-body, and declarative reverse-ETL foundations without broadening API authority or dropping generated, provider-owned, safety, App, CLI, documentation, or result surfaces.
- Verification: Targeted red-green tests for every conflict correction; focused engine, commandrunner, App, CLI, sync-transport, source-import, generator, and regression tests; `go vet ./...`; `go build ./cmd/pm`; `connectorgen validate`; `surface-sync --check`; generated CLI/help/manual/website checks; `connector-boundary`; `make verify`; and the specified real-provider qualifications.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Exact component heads are immutable and approved before composition | live | `data/cli-current-foundations-main-integration-r1/input-manifest.json` records a matching remote pull ref and Firstmate qualification for every merged head. |
| No component contract is lost in conflict resolution | live | Production-shaped red tests exercise all cross-feature routes and remain green from the final rollup SHA. |
| Installed CLI and persisted App preserve declared operations and provider results | live | Installed-binary and App tests assert reachable declared surfaces, plan/apply/ack behaviour, result fields, and generated help. |
| Actual-provider claims are not synthetic-only | live | Qualification report records sanitized real-provider status/export, declared write, source-import, and reverse-ETL assertions from the integrated SHA. |

## Scope and safety invariants

- Preserve component commits with merge commits; do not cherry-pick, rewrite, reset, force-push, close, retarget, or merge any component PR.
- Every integration or conflict-resolution commit includes `Refs #4302`, `Refs #4303`, `Refs #4305`, `Refs #4306`, and `Refs #4307`.
- No connector-name branch, generic HTTP method/path/header/body/action authority, generic shell, or generic SQL write capability is permitted.
- Declarations own operation identifiers, path/query/header/body mappings, bounded transfer rules, status metadata, and all non-secret provider result fields. Credentials stay in existing encrypted/masked boundaries and never enter evidence.
- Reverse ETL remains plan → preview → approval → execute → durable acknowledgement.

## Inputs and merge order

1. Confirm the immutable source-import head for #4306 / #4312, then merge it without rewriting its history.
2. Confirm the immutable closed-runtime head for #4307 / #4311, inspect source-import/runtime overlap, add red tests before any correction, and merge it.
3. The exact #4302 / #4308 head has Firstmate's terminal qualification; inspect and merge it after #4306 and #4307 without rewriting its history.
4. Compose the captain-authorized provisional local no-mistakes structured-body head #4305 / `55ddb650aa5594ddd156b0939cb1df6027a31d56` through its preserved merge `0eb98d3844da7b48d0ca27f51ba7deb46d8f5d1b`; do not substitute the older visible branch head.
5. Compose the captain-authorized provisional local no-mistakes reverse-ETL head #4303 / `e7f474375af969555efd82f684ad6d0b8a26cfc0` through its preserved merge `808896a28873c5f0479fa10e2f798da56f885b5e`; do not substitute the currently failing #4304 head.
6. Stop at an unpushed composite after focused combined checks and independent-review evidence. Later component commits are additive follow-ups, not substitutions.

## TDD integration slices

| Slice | Red | Green | Refactor / regression guard |
| --- | --- | --- | --- |
| Declaration-bound request shaping | A real declaration combining structured body, typed header, exact query, and exact path must fail before I/O for malformed, unknown, oversized, duplicate, CR/LF, or cross-bound values. | The request materializes only declaration-owned fields and reaches the declared operation exactly. | Re-run scalar, form, SCIM, binary, and specialized GitHub cases unchanged. |
| Bounded response/status composition | A terminal status-only 4xx/5xx after normal retry handling must retain final metadata, while binary/text GET errors must remain errors and produce no output file. | Closed status/text/binary operations preserve typed status, headers, body bytes, and bounds. | Re-run loader, output, retry, and download regressions. |
| Source declaration reachability | An accepted lock-verified source operation must fail if it cannot reach generated command/help surfaces. | Exact source bytes/count/SHA import produces the fixed declaration and installed surface. | Re-run lossless/oversized/malformed source and generator validation cases. |
| Typed reverse-ETL composition | Multiple named actions must fail if plan/apply/ack or any provider result field is not persisted through App and installed CLI. | Independently selectable actions complete plan, preview, approval, apply, durable acknowledgement, and provider-result preservation. | Re-run existing typed destinations and direct-write bindings. |

## GSD and skills record

- GSD adapter health and command resolution: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`; `go run ./cmd/agentcontractgen check`.
- Inline/manual fallback: the task's canonical single-worker contract and this runtime forbid spawning the generated GSD roles, so the generated prompts are executed and recorded inline.
- Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-documentation`.
- CLI parity is mandatory: runtime help, bare namespace behavior, docs/cli, website, generated help/manual/discovery surfaces, JSON/stdout-stderr contract, and reverse-ETL safety text are verified together.

## Checkpoints

- [x] Isolated worktree, clean `main` base, no-mistakes daemon health, GSD adapter, skills, base commit, and published-ref intake verified.
- [x] All three published heads were re-read through `gh-axi`; #4305/#4303 use the captain-authorized exact local-pipeline provenance recorded in the input manifest.
- [x] Component history is preserved through exact merge ancestry, with focused red-green-refactor evidence for both shared seams.
- [ ] Focused local combined checks pass from one provisional composite SHA (in progress).
- [ ] Independent-review evidence manifest/report is complete, branch is clean, and the checkpoint is committed and unpushed.

---

## Foundation 0.3.0 release-candidate r1 integration contract

### Task Delivery Header

- Issue: Refs #4305 — declaration-bound structured REST bodies; Refs #4306 — hash-locked provider operation contracts; Refs #4307 — closed operation runtime. This release-candidate only composes their already-reachable inputs plus the audited canonical, reverse-action, and public-output successors.
- Base branch: `fm/cli-current-foundations-main-integration-r1`.
- Merges into: `fm/cli-foundation-release-candidate-r1` → `fm/cli-current-foundations-main-integration-r1` → `main` (human-gated); this task opens no merge and never pushes a default branch.
- Delivery: An unmerged Foundation PR against `fm/cli-current-foundations-main-integration-r1`, whose head is the immutable schema-3 evidence closure commit after all specified local gates pass.
- Working branch: `fm/cli-foundation-release-candidate-r1`, created at `041d2ec7ed986aea15d2d3d64f2076b484c3f999`.
- Task: Compose exactly core `041d2ec7ed986aea15d2d3d64f2076b484c3f999`, reverse action `487ec14e01c90f31a71b1cb5de060b8c66a203e9`, and public output `7fdef00d7e758cb4a3c413a16f8452ee0615f0d5` using ancestry-preserving merges; reproduce and close only FND-B09 with an evidence-only schema-3 commit; record FND-W01 without speculative remediation.
- Verification: Exact ref/ancestry and remote read-back proofs; reverse and public conflict-adjacent test cohorts; all prescribed generation, lint, build, vet, race, documentation, certification, and strict evidence gates; PR API base/head read-back.
- RC-07 correction: Preserve the pre-existing 120 GitHub direct-read certification candidates (23 manual plus 97 generated). At exact evidence head `596c90c`, the source-projection conservative downgrade changed 77 declared generated candidates to `partial`. Restore only the declaration-generated candidate projection and prove it through the fresh binary with a fixture transport; do not lower expectations, add a connector family, or make a live-certification claim.
- RC-08 confirmed Foundation regression (captain evidence, 2026-08-22): `v0.2.1` (`d842c815a`), `origin/main`, and the release candidate each contain 1,225 GitHub API-surface endpoints, but the released/main baseline has exactly one blocked retired POST route while candidate `a750e2c4e` has 371 blocked. The exact 370 newly blocked GET endpoints equal the inventory deficit (`1,224 - 854`) and each reports `Locked source operation … has no field-complete declaration-owned executable route`; the locked REST count remains 1,220. The conservative tightening introduced by `748865f1b` is the culprit; branch counts have oscillated, so measure the actual branch head after every generated change. Repair the common source-projection path so only existing, field-complete declaration-owned direct-read routes are restored into both CLI and API surface. Do not hand-edit generated artifacts, relax the inventory test, exclude endpoints, or manufacture routes. Completion requires exact head-vs-`v0.2.1` parity: endpoints=1,225, blocked=1, candidates=120, and canonical blocked endpoint identity equal to the same retired POST route. Prove restored routes through a fresh `pm` binary against a hermetic fixture: every exercised command must send its declared, resolved request and meet its declared output assertion. This is local behavioral proof only, not live certification.
- RC-09 CodeQL correction (captain evidence, 2026-08-22): exact head `25b2f8447…` has three PR-introduced alerts: high `go/allocation-size-overflow` at `engine/write.go:267` from `len(values)*4`, plus two `go/useless-assignment-to-local` warnings at `rate_parking.go:876` and `connsdk/http.go:1312`. Restore safe allocation without multiplication and remove only the proven dead assignments. Do not suppress, baseline, annotate, or alter the pre-existing JavaScript/ledger alerts. The static alert is a direct red witness; retain behavioral redaction, parked-resume, and terminal-receipt tests, then run CodeQL/targeted package checks. This code change supersedes E: freeze a new implementation I2 and create a new schema-3 evidence-only closure E2 whose parent and `code_sha` are I2.
- RC-10 verified-release correction (CI `Verify`, 2026-08-22): at immutable E2 `8ed5ab93…`, the full exact-SHA gate fails hard-coded `partial` expectations for `issue status`, `pr checks`, `ruleset check`, `search prs`, plus eleven legacy list commands. The captain's comparison shows the eleven declarations are identical and executable at `v0.2.1`, `origin/main`, and E2: they bind valid API-surface routes even where no `operation`/`stream`/`write` field is needed (376 such implemented commands exist on both baselines). Do not change declarations or make working commands refuse. First run every disputed command through the real `pm` binary against a faithful fixture, asserting its declared request and output; only if that proof passes may the stale expectation be corrected with the execution evidence. Independently reproduce the `github_direct_read_certification_binary_test.go:110` non-JSON certification-output report before fixing it; a local pass is not a reason to change it. Preserve endpoints=1,225, blocked=1, candidates=120 and do not hand-edit `api_surface.json`.
- RC-11 CI-containment correction (Verify run `32589551092`, 2026-08-22): the 97-stage GitHub candidate fixture is a real release-gate failure under the aggregate `internal/cli` run. A concurrent test exports an unreachable `PM_DRAGONFLY_ADDR`; the fresh child process inherits it and the certification tier correctly refuses `shared_coordinator_unavailable` before fixture provider I/O. The fixture must remove only that inherited coordinator endpoint from its child environment, preserving the fixture token and all runtime behavior. First reproduce the refusal with the same invalid environment, then prove the isolated child completes all 97 authenticated declared request/output stages even while the parent has that invalid coordinator value. Do not weaken rate-limit runtime behavior, provider assertion, candidate count, or source artifacts.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Audited branch inputs are exact and publishable without rewriting history | live | `git merge-base --is-ancestor` succeeds for each lagging origin head and `git ls-remote` returns the expected private SHA after non-force push. |
| Reverse merge retains both transport invariant sets | live | Focused App, transport, Postgres, and CDC tests pass from the merge commit; the merge diff retains empty-publication/auth/deadline/acknowledgement fences and action-owned receipt/readback binding. |
| Public merge retains masking and declaration authority | live | Engine, connector, commandrunner, SQS, and output tests pass from the merge commit, including exact-secret masking and raw value/key preservation. |
| FND-B09 is closed against the exact implementation | live | The pre-closure schema-2 manifest fails strict decode; the evidence-only closure has `code_sha` equal to its parent and `connectorgen evidence-gate` reports graph, artifact, ledger, and review agreement. |
| FND-W01 is not silently expanded into release code | live | Current-code write ordering and existing source-import test evidence are inspected; the review artifact records a deferred/non-reproducing disposition unless a deterministic release failure is demonstrated. |
| RC-09 removes the PR-introduced CodeQL alerts without weakening runtime behavior | live | CodeQL reports zero alerts at the three exact paths; redaction still masks all literal forms, rate parking preserves reconciliation side effects, and terminal provider receipts remain retained. |
| RC-10 distinguishes a stale hard-coded partial verdict from a genuine execution defect | live | A faithful real-binary fixture executes each disputed command, observes its exact declared method/resolved path/authentication, and validates the command's declared result. Only successful evidence supports correcting a test expectation; otherwise the exact failed command/reason is escalated. |
| RC-11 makes the 97-stage fixture hermetic against unrelated test-process coordination state | live | With `PM_DRAGONFLY_ADDR=127.0.0.1:1` in the parent, the fresh fixture child must still execute all 97 authenticated declared direct-read stages; the invalid endpoint must never reach the fixture child. |

### Composition guardrails

1. Advance the canonical and public origin refs only after exact identity and fast-forward ancestry proof; use normal `git push` and `git ls-remote`, never force push. The reverse-action origin ref already equals its audited head.
2. Merge reverse action first. At `internal/app/transport_dispatch.go`, `internal/connectors/native/postgres/arrow_full_overwrite_transport.go`, and `internal/synctransport/types.go`, retain both core empty-publication/auth/deadline/acknowledgement fences and reverse action-owned receipt/readback binding.
3. Merge public output second. At `internal/connectors/connectors.go`, `internal/connectors/engine/direct_read.go`, `internal/connectors/engine/operation_headers.go`, `internal/connectors/native/amazon-sqs/direct_read.go`, and `internal/connectors/write_result_output_test.go`, retain exact configured-secret masking without ordinary-key, identifier, raw-receipt, or declaration-authority corruption.
4. The implementation/generated composition is frozen as SHA I before evidence work. The only commit above I may change `data/cli-current-foundations-main-integration-r1/**` and `.planning/phases/cli-current-foundations-main-integration-r1/**`; it binds `code_sha=I` and includes no production/generated implementation change.
5. Inline/manual GSD execution is the documented fallback: this single-worker RC contract and the current runtime prohibit generated role spawning. Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-lint`.
