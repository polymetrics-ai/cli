# Issue 389 TDD Ledger — proof-recovery repair

## Baseline for this repair run

- Branch: `fix/389-shepherd-proof-recovery`
- Start head: `db13cbaa8e27cbc86130ce2547f3e60b82b5217c`
- GSD command: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`
- GSD adapter checks run: `scripts/gsd doctor`, `scripts/gsd list`
- Current evidence status: prior claims of independent validation, ratification, recovery planning,
  final verification, and canary readiness are invalid for this repair run until re-proven against a
  new exact candidate head.

## Skills recorded

`gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-context`,
`golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-observability`, `golang-lint`.

## Read-only recon evidence

- Scout finding: `cmd/shepherd/main.go:persistSuccessProof` currently writes
  `Validator="openai-codex/gpt-5.6-sol"`, `Thinking="high"`, `Verdict="PROCEED"`, and
  `Ratified=true` directly.
- Scout finding: `internal/authority/ratification.go` contains `authority.Ratify`, but production
  code does not call it.
- Scout finding: store proof tests do not yet reject missing ratification or non-PROCEED verdicts.
- Orchestration decision: start with RED tests around proof/attestation before any production edit.

## Slice A — Real independent validation and ratification

RED tests:
- [x] `internal/store/proof_test.go`: reject `ArtifactProof{Ratified:false}`.
- [x] `internal/store/proof_test.go`: reject attestation verdicts other than real successful
      ratification (for example `RETRY`/`HALT`).

RED evidence:
- FAIL (expected) `cd agent-runtime/shepherd && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts' -count=1`
  - `TestArtifactProofRejectsUnratifiedResult`: `unratified artifact proof accepted`
- [x] `cmd/shepherd/main_test.go`: successful supervise path must not create a ratified proof when
      independent validator evidence is absent.
- [x] `cmd/shepherd/main_test.go`: stale candidate head leaves canonical branch unchanged.
- [x] `cmd/shepherd/main_test.go`: GPT-5.5 validator evidence is rejected.
- [x] `cmd/shepherd/main_test.go`: `RETRY` and `HALT` verdicts are rejected.
- [x] `cmd/shepherd/main_test.go`: every rejected path leaves canonical HEAD and canonical `.gsd`
      unchanged.
- [x] `cmd/shepherd/main_test.go`: successful path validates before promotion and persists proof plus
      attestation after ratification.

RED evidence:
- FAIL (expected) `cd agent-runtime/shepherd && go test ./cmd/shepherd -run 'TestSuperviseRejectsInvalidIndependentValidationWithoutPromotion|TestSuperviseRatifiesBeforePromotingCandidate' -count=1`
  - compile failed because `internal/validation` and `independentValidatorFactory` did not exist yet.

GREEN evidence (partial store hardening only):
- PASS `cd agent-runtime/shepherd && gofmt -w internal/store/proof_test.go internal/store/store.go && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts|TestArtifactProofBindsExactHeadsAndRatification|TestAttestationPersistsValidatorProof' -count=1`
- Production change: `PutArtifactProof` now rejects unratified proofs; `PutAttestation` now rejects non-`PROCEED` verdicts.

False-green evidence for commit `19d051c6`:
- The previous Slice A completion claim is invalid. Tests only proved fake-validator port behavior and
  did not test the real production validator implementation.
- Production validation had no real result producer; it invoked canonical `validate-milestone`, which can
  mutate GSD workflow state and does not produce Shepherd validation results.
- Production trusted worker-controlled `.gsd/shepherd-validation.json`, fabricated fallback session IDs,
  used generation as state version, hard-coded PR base as `main`, and blindly required/claimed UAT.
- Slice A remains open; Slice B, PR creation, final Sol review, and canaries remain blocked.

Retry RED tests added against the real production validator:
- [x] `internal/validation`: no validation-result producer exists.
- [x] `internal/validation`: stale pre-existing result is ignored/rejected.
- [x] `internal/validation`: no new validator session is rejected.
- [x] `internal/validation`: validator session model GPT-5.5 is rejected.
- [x] `internal/validation`: thinking below high is rejected.
- [x] `internal/validation`: result head/evidence/request nonce mismatch is rejected.
- [x] `internal/validation`: candidate moving during validation is rejected.
- [x] `internal/validation`: stale base branch/governance state version is rejected.
- [x] `internal/validation`/`internal/authority`: RETRY/HALT or missing required gates are rejected.
- [x] `internal/validation`: failed validation paths leave candidate `.gsd` unchanged and do not mutate Git except the explicit candidate-move fixture.

RED evidence: these tests target production `validation.GSDValidator` with a helper process and would fail against `19d051c6` because it used canonical `validate-milestone`, trusted worker-local `.gsd/shepherd-validation.json`, accepted derived session IDs, lacked protected nonce-bound result transport, and did not bind state/base/gates.

Second false-green evidence for commit `99604d48`:
- The claimed dedicated subprocess was `gsd headless shepherd-validate`, but neither pinned official GSD
  nor the repository adapter registers that command.
- Helper tests substituted the Go test executable and manually wrote sessions/results, proving transport
  parsing but not that the production validator was callable.

Final corrected Slice A GREEN evidence:
- Added a contract test proving the former GSD executable fails the required Pi capability probe.
- Added a separately configured `pi_command`; production invokes real Pi with `--mode json --print`,
  `openai-codex/gpt-5.6-sol`, `--thinking high`, and only `read,bash,grep,find,ls`.
- Disabled extensions, skills, templates, themes, context files, and project trust for the validator.
- Added exact fake-Pi process-boundary tests for model/thinking, candidate cwd, tool allowlist, dedicated
  session directory, fresh session, nonce/head/hash/repository/PR/base binding, malformed/stale/replayed/
  mismatched results, nonzero exit, timeout, missing result, and startup capability failure.
- Validation retries now use a stable audit request identity bound to generation/attempt/state version plus
  a fresh cryptographic nonce subdirectory, preventing collisions while retaining replay evidence.
- Pi JSON output is bounded in memory, parsed only from final assistant messages, and never included raw
  in durable errors; only redacted classifications/counts are exposed.
- Preserved protected nonce-bound transport, exact-head rereads, required-gate derivation, ratification,
  full durable attestation, and delayed promotion.
- Shepherd now writes nonce-bound validation requests under protected state, outside the candidate worktree, and rejects stale/reused/malformed result files.
- Removed derived/fabricated validator session identity; a new session after validator start with exact worktree, Sol/high model, and high thinking is mandatory.
- Result proof binds request ID, nonce, base branch/head, candidate/observed head, durable governance state version, contract/evidence hash, verdict, gates, issue time, and expiry.
- Extended durable attestation schema/migration to persist repository, PR, base branch, base/candidate/observed head, delivery, generation, unit, attempt, state version, hashes, session ID, model/thinking, verdict, gates, issued, and expiry.
- Required gates are derived from official unit metadata; UAT is required only for UAT phase metadata.
- PASS `cd agent-runtime/shepherd && go test ./internal/validation ./internal/store ./internal/authority ./internal/workspace ./cmd/shepherd`.
- PASS `cd agent-runtime/shepherd && go test ./...`.
- PASS `cd agent-runtime/shepherd && go test -race ./...`.
- PASS `cd agent-runtime/shepherd && go vet ./...`.
- FAIL `cd agent-runtime/shepherd && golangci-lint run ./...` with the same 30 pre-existing findings; the earlier new `validation/validator.go` staticcheck finding was fixed.
- PASS `cd agent-runtime/shepherd && go build ./cmd/shepherd && make verify && cd ../.. && scripts/tests/shepherd-module-boundary.sh && git diff --check && go list ./...`.
- PASS live smoke: `POLYMETRICS_SHEPHERD_LIVE_VALIDATOR=1 go test ./internal/validation -run TestLivePiValidatorSmoke -count=1 -v` observed `openai-codex/gpt-5.6-sol`, `high`, fresh session `019f62b3-9830-7129-9c93-2104ed54a10e`, bound fixture head `6650f5e18ecbbf15c18739a8422fa1ba663a0635`, bound evidence hash, and `PROCEED`.

## Slice B — Durable attempt lifecycle and crash recovery — COMPLETE / GREEN

Slice A accepted GREEN at `95a17f18274c87ed0e3fde825b41257039b757de`.
Slice B independently accepted GREEN at `1a050692f9e47b5b4d3d74cfb38e56c67d461399`.
Orchestration: `local_critical_path`; no overlapping mutating workers.

Completed RED/GREEN tests:
- [x] all 11 lifecycle states and immutable fields persist and survive SQLite reopen.
- [x] duplicate attempt identity cannot rebind branch, path, or base head.
- [x] illegal, backward, terminal, generic-ratification, ambiguous-promotion, and stale-owner transitions fail closed.
- [x] worktree creation is positively confirmed before `prepared`; dispatch persists `running`.
- [x] candidate/validated heads persist before validation and proof/attestation/ratification commit atomically.
- [x] preparation, pre-dispatch query, and runtime failures persist explicit bounded classifications.
- [x] runtime failure becomes `retained_for_recovery`; cleanup failure becomes `cleanup_blocked`.
- [x] startup reconciliation removes only exact database-owned, confirmed, non-live worktrees/branches.
- [x] unknown/mismatched/live/checked-out/unconfirmed paths and branches remain untouched.
- [x] retry after retained failure creates a fresh branch/path and never reuses the prior worktree.
- [x] reconciliation is idempotent across repeated supervisor/database restarts.
- [x] pre-Slice-B schema migration preserves delivery-run, proof, and attestation records.
- [x] hard-crash reopen fences the old lease, interrupts running unit state, and restores delivery readiness.
- [x] ambiguous running/promoting attempts durably await human recovery; resume succeeds only after exact resources are proven absent.

RED evidence (2026-07-15):
- `cd agent-runtime/shepherd && go test ./internal/store ./internal/workspace ./cmd/shepherd` failed as expected.
- Store compile failures: missing durable `AttemptWorktreeState`, record/update types, and lifecycle APIs.
- Workspace compile failures: missing branch identity, owned-attempt reconciliation types, and reconciliation API.
- `cmd/shepherd` remained green before integration, confirming the initial RED boundary was the new Slice B behavior.
- After store/workspace GREEN, focused supervise tests failed with zero `cleanup_complete` and zero
  `retained_for_recovery` records, proving the real supervise path was not yet integrated.

GREEN evidence (2026-07-15):
- PASS `cd agent-runtime/shepherd && go test ./internal/store ./internal/workspace ./cmd/shepherd`.
- PASS `cd agent-runtime/shepherd && make verify` including `go vet`, full tests, and `go test -race ./...`.
- PASS repository gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.
- `golangci-lint run ./...`: exactly 30 findings, matching HEAD baseline; no Slice B file finding.

Refactor evidence:
- Repository-global flock and SQLite lease epochs fence every delivery lease, bootstrap/query, promotion, and cleanup.
- Attempt workspaces use an explicit disjoint `attempt_root`; exact path/branch/head/common-dir and no-symlink checks precede cleanup.
- Independent reviewer/security passes drove fail-closed running/promoting recovery, positive resource confirmation,
  atomic ratification, authorized-branch checks, bounded output, and human-gated convergence fixes.
- Slice C promotion/state-swap journaling remains explicitly excluded.

## Slice C — Crash-safe GSD-state promotion — COMPLETE / GREEN at `f0fbf47f`

Orchestration: `local_critical_path`; no overlapping mutating workers.

Required RED tests:
- [x] promotion journal persists across store close/reopen.
- [x] failure before Git promotion.
- [x] failure immediately after Git promotion.
- [x] failure before canonical GSD backup rename.
- [x] failure after backup rename but before staged-state install.
- [x] failure after staged-state install but before journal completion.
- [x] restart at every boundary converges idempotently.
- [x] repeated recovery produces no duplicate effects.
- [x] canonical Git and `.gsd` never finish mixed.
- [x] moved canonical head fails closed.
- [x] dirty canonical worktree fails closed.
- [x] missing, changed, corrupt stage/backup blocks without deletion.
- [x] unknown staging/backup directories are never deleted.
- [x] expired or mismatched proof before Git promotion blocks promotion.
- [x] already-promoted valid journal completes forward without a new verdict.
- [x] SQLite data committed through WAL survives staging and installation.
- [x] installed `gsd.db` passes integrity check without stale WAL/SHM files.
- [x] existing Slice A/B validation, ratification, lifecycle, cleanup, and restart tests remain green.

RED evidence (2026-07-15):
- `cd agent-runtime/shepherd && go test ./internal/store ./internal/workspace ./cmd/shepherd` failed as expected.
- Store failures: missing durable promotion journal types/states and create/get/transition/recovery-claim APIs.
- Workspace failures: missing bounded manifest, WAL-safe staging, rename install/recovery, cleanup, and failpoint boundaries.
- Command failures: missing real promotion coordinator, failure boundaries, startup recovery, and journal state integration.
- The RED suite covers all requested pre/post-Git and pre/between/post-state-swap boundaries, repeated
  recovery, moved/dirty heads, corrupt/missing/unknown resources, stale authority, forward recovery,
  WAL survival/integrity, and Slice B no-journal human gating.

GREEN evidence (2026-07-15): focused store/workspace/command promotion suites pass; full nested
`go test ./...` passes; final store race rerun passes; earlier full command/workspace race pass remains
valid for the crash-boundary suite; `go vet ./...`, `go build ./cmd/shepherd`, root `make verify`, module
boundary, root `go list`, gofmt, and diff checks pass. Nested lint reports 29 baseline findings and no
Slice C production finding.
All nine injected journal/Git/swap/completion boundaries converge on restart. WAL-backed SQLite data
installs standalone and passes integrity checking.

Refactor evidence: two independent reviewer/security cycles drove pre-stage durable intent,
pre-Git resource and attempt revalidation, immutable full attestation binding, blocked-journal gates,
bounded root-confined copying, and crash-resumable rooted cleanup tombstones. The stronger request to
cryptographically include the complete staged SQLite snapshot in the independent model attestation is
deferred: Slice C's approved contract binds both ratified proof and deterministic full-state manifest
inside the protected journal after worker quiescence; it does not redefine Slice A's evidence schema.
Slice D onward is excluded.

## Slice D — Official GSD 1.11 registry loading — COMPLETE / GREEN

Orchestration: `local_critical_path`; read-only recon/review sidecars only. Skills loaded:
`gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`,
`golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-observability`, and `golang-lint`.

Planned RED tests:
- [x] realistic pinned 1.11 registry fixture with array spreads loads through normalized JSON.
- [x] exact allowed and required tools survive normalization.
- [x] forbidden tools and reason strings survive normalization.
- [x] missing/malformed/partial/oversized/duplicate/unexpected fields return `runtime_contract_mismatch`.
- [x] missing registry, wrong version, symlink, path escape, and source drift fail closed.
- [x] null phase/tool metadata never receives built-in fallback.
- [x] unknown units and phases fail closed.
- [x] versioned sidecar policy is separate from official metadata and narrowly allowlisted.
- [x] official coordination phases route Sol/high; execution phases route GPT-5.5/high.
- [x] unit names cannot substitute for phase routing.
- [x] subagent events cannot overwrite top-level observed model/thinking evidence.
- [x] prompt-advertised tools must match normalized official tool contracts.
- [x] production startup has no `BuiltinUnitRegistry` fallback.
- [x] complete host package snapshot and absolute Node hash pin reject drift and worktree roots.
- [x] verified-byte registry import removes filesystem verify/import TOCTOU.
- [x] every unexpected current-run top-level identity transition fails immediately.
- [x] stale/symlinked/duplicate/unknown/trailing session headers fail closed.
- [x] observed forbidden or disallowed `gsd_*` tool starts fail closed.
- [x] exporter and validator descendants are terminated through process-group cancellation and WaitDelay.
- [x] bounded no-follow hashing and policy reads detect pre/post pathname or inode replacement.
- [x] Podman tags resolve once to an immutable image ID and fail admission without an approved full-image digest.

RED evidence (2026-07-15): `cd agent-runtime/shepherd && go test ./internal/gsd -run
'Test(LoadPinnedUnitRegistry|NormalizedRegistry|SidecarPolicy|DecodeNormalized)' -count=1` failed at compile time on the new normalized-document schema, bounded loader, strict decoder, and sidecar-policy APIs. This proves the pre-existing regex parser and built-in null substitution cannot satisfy the Slice D contract.
A second RED run of `go test ./internal/gsd ./cmd/shepherd -run
'Test(ReadSessionIdentityExcludesPersistedSubagentSessions|ObservedRuntimeIdentityIgnoresSubagentEvents)'`
failed because the newest delegated session supplied GPT-5.5/medium and the top-level observer had no
provenance filter.
GREEN evidence (2026-07-15): focused registry/runtime/session/tool/settings/process/store/workspace/
validator/command suites pass; the installed official GSD Pi 1.11.0 registry export and complete host
runtime snapshot qualification pass with the pinned Node binary. Full nested `go test ./...`, full
`go test -race ./...`, `go vet ./...`, `go build ./cmd/shepherd`, nested and root `make verify`, module
boundary, root `go list ./...`, gofmt, and `git diff --check` pass. Nested lint is 28 known findings,
one below the 29-finding Slice D baseline, with no new Slice D production finding.

Refactor evidence: four independent read-only reviewer/security cycles were dispositioned. Fixes include
v3 root/mode/owner manifests, per-launch settings/preferences/prompt revalidation, exact canonical
`next`/`discuss` IDs, fresh empty checkpoints for state-only units, durable attempt/session identity,
strict current-run session deltas, trusted MCP workflow namespaces without blocking non-GSD MCP tools,
fail-closed uncontracted workflow calls, process-tree cleanup on cancellation and normal validator exit,
and bounded stable policy/session reads. Retained Podman fails closed. `--continue-unit` also fails
closed until disposable-worktree continuation is qualified. The explicitly selected same-UID host
model remains a documented architecture trust assumption requiring a future separate UID, OS sandbox,
or human-qualified container; it is not represented as isolation. Slice E onward is excluded.

## Slice E — Real Sol/high recovery planning

Planned RED tests:
- [ ] static recovery-planning text is insufficient.
- [ ] planner result must include observed model/thinking, evidence hash, typed action, and bounded
      plan.
- [ ] unallowlisted action fails closed.
- [ ] budget exhaustion persists `awaiting_decision` across restart.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice F — Authority-gated external effects

Planned RED tests:
- [ ] no direct `SyncDecisionComment` production path.
- [ ] outbox pending, claim, send, failure, restart, and idempotent replay.
- [ ] worker ports cannot directly mutate GitHub.

GREEN evidence: pending.
Refactor evidence: pending.

## Slice G — Real supervise integration coverage

Planned RED tests:
- [ ] success path reaches `final_human_gate` only after independent Sol/high validation,
      ratification, and promotion.
- [ ] missing/GPT-5.5 validator evidence fails without canonical mutation.
- [ ] stale candidate head fails without canonical mutation.
- [ ] validator `RETRY`/`HALT` fails without canonical mutation.
- [ ] crash/restart at each promotion boundary.
- [ ] retained failed attempt followed by fresh attempt.
- [ ] recovery planning and `awaiting_decision` restart.
- [ ] outbox restart and duplicate suppression.
- [ ] official registry spread metadata.
- [ ] canonical worktree unchanged on every failed path.

GREEN evidence: pending.
Refactor evidence: pending.

## Verification log

No production-code verification is claimed yet for this repair run. Focused RED tests will be recorded
before the first production edit, followed by focused GREEN and full nested-module gates after each
coherent slice.
