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

## Slice E — Real Sol/high recovery planning — COMPLETE / GREEN

Accepted checkpoint: `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`.
Slice D accepted GREEN at `cacb32e8e16b7ba70742cc5365cb83fffd74ca35`.
Orchestration: `local_critical_path`; no overlapping mutating worker. GSD command:
`scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
Skills loaded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-observability`, `golang-lint`, `golang-code-style`,
`golang-naming`, and `golang-documentation`.

Planned RED tests:
- [x] static recovery text is rejected.
- [x] required typed failure classes are exhaustive and unknown input fails closed.
- [x] dead worker retries only within its independent class budget.
- [x] silent tool persists and enforces deterministic backoff.
- [x] artifact missing/invalid and reversible validation failure invoke a real Sol/high planner.
- [x] model/thinking mismatch, scope breach, dirty tree, stale head, and ratification failure block
      without planner invocation.
- [x] GitHub/outbox uncertainty blocks without repeating an external effect.
- [x] unknown failure safely awaits decision or blocks.
- [x] exhaustion durably enters awaiting decision or blocked according to policy.
- [x] restart preserves counts, policy, action, plan, session evidence, retry time, and exhaustion.
- [x] concurrent owner/lease-epoch claims cannot consume one budget attempt twice.
- [x] GPT-5.5/non-high/stale/fabricated/replayed/mismatched planner evidence is rejected.
- [x] unallowlisted and class-forbidden actions are rejected.
- [x] planner cannot widen authority, emit executable commands, broad paths, tools, or external writes.
- [x] planner timeout/cancellation and normal parent exit synchronously terminate descendants.
- [x] malformed/oversized/duplicate/case-duplicate/unknown/partial/trailing output fails closed.
- [x] failed planning/rejection leaves canonical Git and `.gsd` unchanged.
- [x] accepted retries use fresh Slice B attempt resources.
- [x] Slices A-D remain green.

RED evidence (2026-07-15): `cd agent-runtime/shepherd && go test ./internal/recovery ./internal/store
./internal/supervisor ./cmd/shepherd -count=1` failed as expected. `internal/recovery` had no production
files and failed on missing typed classes/actions/policy, strict planner request/result/session evidence,
and bounded process APIs. `internal/store` failed on missing fenced structured reservation/outcome,
backoff/dispatch, replay, and durable evidence APIs. Existing supervisor/command packages remained
green, isolating the RED boundary to Slice E behavior. The RED suite explicitly rejects the static
recovery sentence and covers typed policy, real Sol/high identity, malformed/replayed evidence,
class-independent concurrency-safe budgets, restart/backoff/exhaustion, and dispatch claims.

GREEN evidence (2026-07-15): focused recovery/store/supervisor/command tests pass. A live Pi smoke
observed `openai-codex/gpt-5.6-sol`, `high`, fresh session
`019f6721-1e33-7dea-9b20-991a2e004715`, a strict bound result, and no tools. Full nested tests,
race, vet, build, nested/root `make verify`, module boundary, root package listing, formatting, and diff
checks pass. Nested lint remains the exact accepted 28 findings with no `internal/recovery` finding and
no new Slice E production finding.

Refactor evidence: independent GPT-5.6 Sol/high correctness and security passes were run repeatedly
against the complete working diff from Slice D. Every actionable finding was dispositioned, including
reversible gating, typed silent/dead-worker sentinels, unsafe joined-error dominance, no-tool stream and
durable-session proof, branch/scope binding, legacy migration, globally ordered crash-safe dispatch,
plan execution, evidence/action policy binding, owner plus lease-epoch fencing, symlink containment,
actual-clock expiry, external-effect uncertainty blocking, mutating-skip preservation, noncanonical
retry blocking, no-decision crash gating, dispatch-owner/epoch-fenced unit disposition, crash
reconciliation of claimed dispatch, mutating-skip consumption without historical masking, durable
assistant model proof, pre-unit dispatch failure disposition, persisted dispatch epoch, removal of the
unfenced unit-finish API, typed decision-comment uncertainty, historical-consumed decision ordering,
deferred finish-error propagation, complete error-chain preservation, unsafe noncanonical policy
preservation, unconditional post-backoff authority/GSD/lock/lease revalidation, transactionally
fenced unit starts, retained runner causes, finish-error chain preservation, and durable outbox-failure
classification, and deferred delivery-finish error propagation. The accepted
same-UID host assumption remains documented; unsupported non-Unix process-tree cleanup now fails
planner construction closed.

## Slice F — Authority-gated external effects — COMPLETE / GREEN at `ea88c92f`

Slice E accepted GREEN at `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`.
Orchestration: `local_critical_path`; no overlapping mutating worker. Read-only tester/reliability/
review/security sidecars may overlap. GSD command:
`scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
No live GitHub mutation is allowed; tests use fakes, in-memory runners, or `httptest`.

Planned RED tests:
- [x] architecture rejects direct `SyncDecisionComment`, `SyncQuestionComment`, and other
      write-capable GitHub calls outside the outbox GitHub executor.
- [x] reply polling exposes only a read-only GitHub port; only the executor receives write capability.
- [x] workers, units, validators, recovery planners, reviewers, enqueue helpers, and executors cannot
      create grants or expand controller-derived authority.
- [x] immutable grants bind delivery, repository, issue/PR, capability, generation, head, epoch, and target.
- [x] merge capabilities, unsupported effect types, changed targets, stale heads/epochs, and capability
      mismatches fail closed.
- [x] strict versioned payloads reject missing, duplicate/case-duplicate, unknown, partial, oversized,
      trailing, secret-bearing, arbitrary-command, and unbounded diagnostic fields.
- [x] canonical payload/hash, stable idempotency key, grant identity, claim identity, and bounded typed
      results/errors persist unchanged across SQLite reopen.
- [x] `pending`, `claimed`, `sent`, `failed`, `uncertain`, `blocked`, and `cancelled` enforce legal,
      owner-fenced, epoch-fenced, and terminal transitions.
- [x] claim expiry/recovery is deterministic and fenced; definite pre-send failures may retry only by
      policy; ambiguous post-send failures become `uncertain` and are never blindly replayed.
- [x] exact stable comment marker/payload reconciliation suppresses duplicates and blocks duplicate
      markers, payload conflicts, target changes, and ambiguous reads.
- [x] older decision-summary revisions cannot overwrite newer ledger revisions.
- [x] question effects bind request ID, generation, unit, head, and external comment ID.
- [x] startup reconciles every decision-ledger/outbox crash boundary without pretending the two SQLite
      stores are atomic.
- [x] all required effect telemetry is typed, bounded, secret-free, and emitted at correct transitions.
- [x] existing Slices A-E remain green.

RED evidence (2026-07-15): test-only `internal/outbox` architecture, strict payload/policy,
durable state-machine, fenced claim recovery, marker reconciliation, duplicate/conflict, and post-send
uncertainty tests were added before production code. `cd agent-runtime/shepherd && go test
./internal/outbox -count=1` fails as expected because `Comment`, `Target`, `Controller`, `Authorization`,
`Event`, and the durable outbox/executor APIs do not exist. The broader focused command also reached the
same outbox build failure while existing store/GitHub/domain/recovery packages remained green; its
command package tail exceeded the 120-second capture, so that run is RED evidence only, not verification.
GREEN evidence (2026-07-16): focused and full nested tests pass; full race, vet, build, nested/root
`make verify`, module boundary, root `go list`, formatting, and diff checks pass. Nested lint remains
exactly the accepted 28 findings with no Slice F package finding. No live GitHub mutation occurred.
Refactor evidence: repeated independent GPT-5.6 Sol/xhigh correctness, security, and restart/reliability
passes dispositioned all actionable findings. Fixes include immutable delivery targets, current-store
controller revalidation, fresh Git/lease fences, summary slot serialization, exact uncertainty
reconciliation, atomic reply/run transitions with immutable actor/comment/time provenance, safe expiry,
promotion-decision resolution, exact-head final-gate proof under repository lock, crash-safe ledger-tail
repair, mandatory execution fences, bounded secret scanning, and old-generation uncertainty settlement
before manual resume.

## Slice G — Real supervise integration coverage — COMPLETE / GREEN at `ee474811`

Base: accepted Slice F checkpoint `ea88c92f5f3c0b1c5f3f434fa52efba24624f803`.
Orchestration: `local_critical_path`; one sequential harness/command/storage trust boundary, no
mutating sidecar. Read-only recon/reliability/security/review sidecars may overlap.
GSD command: `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
Skills: `gsd-core`, `polymetrics-issue-delivery`, `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`,
`golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and
`golang-troubleshooting`. Requested `.pi/skills/go-implementation/SKILL.md` is absent; no manual-GSD
fallback is used because the repo-local programming-loop adapter is healthy.

Process-level tests (`//go:build integration`):
- [x] real built CLI parses supervise args/config and reaches `final_human_gate` only after GPT-5.5/high
      implementation, fresh GPT-5.6 Sol/high validation, exact candidate proof, ratification, journaled
      Git/GSD promotion, attempt cleanup, and outbox convergence; merge is never attempted.
- [x] missing validator result, GPT-5.5 validator identity, non-high thinking, stale/moving candidate,
      `RETRY`, `HALT`, missing required gate, no governed delta, and missing/changed artifact each preserve
      the complete canonical branch/head/status/GSD manifest and retain a bounded durable failure.
- [x] subprocess termination/restart covers journal-created, normalized state staging, pre-Git, post-Git,
      both state-swap rename boundaries, post-install/pre-complete, and final-gate projection without
      duplicate promotion or mixed canonical Git/GSD.
- [x] termination while running, validation timeout then fresh attempt, exact branch/path/session identity,
      no old candidate/validator proof reuse, and unknown/ambiguous worktree preservation/blocking pass.
- [x] typed recovery planning uses a fresh GPT-5.6 Sol/high process, fenced budget reservation,
      deterministic exhaustion/`awaiting_decision`, restart-safe exact reply acceptance, and unauthorized,
      edited, stale, or duplicate reply rejection without lost/duplicate requests.
- [x] pending, expired-claim, post-write/pre-sent crash, exact-marker reconciliation, duplicate suppression,
      duplicate-marker collision blocking, and markerless uncertainty never blindly replay.
- [x] official GSD 1.11 spread metadata preserves exact phase chains/tool contracts and routes execution to
      GPT-5.5/high and planning to GPT-5.6 Sol/high; unknown/partial/stale metadata fails closed.
- [x] every test invokes the actual executable/config/process boundary, uses isolated real Git/SQLite
      state, strict deterministic argv fakes, bounded heartbeat/output, sanitized environment, no network/
      credential access, race-instrumented child binaries, and no test order.

RED evidence (2026-07-16, before production edits):
- `cd agent-runtime/shepherd && go test -tags=integration ./integration/... -run TestSuperviseFakeRuntime -count=1` fails.
- Success fails at the actual built CLI with `runtime admission: runtime_contract_mismatch: complete installed GSD runtime tree differs from the qualified package`; the fake executor is never observed.
- All seven validator-rejection cases fail before `authority.db` exists, proving direct-call package-main helpers do not exercise the real command/admission boundary and cannot satisfy Slice G.
- Root cause: the built command couples full pinned host runtime preparation to the child executor. A compile-tagged integration binary needs a narrow bounded fake executor seam while retaining exact official registry export/admission, metadata routing, command/config parsing, stores, validation, ratification, promotion, outbox, and final-gate paths. Release builds must have no reachable override.
GREEN evidence (2026-07-16):
- actual-CLI success/rejection/registry/recovery tests pass with real local Git and SQLite;
- normalized SQLite/WAL-aware GSD snapshots bind validation evidence to the exact installed stage;
- candidate no-delta, post-validator mutation, and post-final-gate canonical GSD drift fail closed;
- Slice-F-format post-Git journals retain forward recovery compatibility;
- outbox pending/claimed/sent/uncertain and duplicate-marker boundaries converge without blind replay;
- implementation and planner/validator sessions are fresh, model/thinking bound, and joined to exact
  proof/promotion identities in process assertions;
- `go test -tags=integration ./integration/... -count=1` passes;
- `go test -race -tags=integration ./integration/... -count=1` passes with race-built Shepherd/helper children;
- full nested unit/race/vet/build and nested `make verify` pass;
- root `make verify`, module boundary, root package listing, formatting, JSON, and diff checks pass;
- default and integration-tagged lint each remain exactly the accepted 28 findings (25 `errcheck`,
  2 `staticcheck`, 1 `unused`) with zero differential.

Refactor evidence:
- two independent GPT-5.6 Sol/xhigh read-only cycles found and drove fixes for fake-GitHub target/payload
  strictness, credential-free environments, session binding, candidate-local workflow state, complete
  canonical/proof oracles, child race instrumentation, normalized staged-state binding, legacy post-Git
  recovery, continuous heartbeat evidence, reply ambiguity, and final-gate GSD revalidation;
- the first exact-head review of `4592734803f20e1b4893efae2ebd900525a92868` found missing continuous
  decision polling/expiry, SIGINT and pre-send outbox boundaries, terminal Pi lifecycle handling, and
  ordinary-exit descendant cleanup; sequential fixes add exact branch/head/lock/lease application fences,
  same-process decision tests, complete ordered lifecycle/tool pairing, controller-owned process pipes,
  and inherited-output descendant regressions;
- replacement exact-head review at `b08c93cc6b1de6a6c89d57c14da6c14d01d7e420` found validator
  turn/session provenance gaps; sequential fixes require complete successful non-retrying Pi lifecycle,
  exact stream-to-final-durable-proof hashing, and an exclusively created fresh session directory, with
  missing/duplicate/out-of-order/error/retry/tool-before-message/session/proof-reuse regressions;
- exact-head review at `c1a34d23585329a9eb7f64a1ef687e0268c17666` found implementation-turn,
  lifecycle-alias, durable assistant-row, and escaped-output-drain gaps. Fixes require final successful
  implementation turns and durable stops, recursive case-fold/Unicode-safe canonical JSON fields with a
  strict object-field bound, type=`message` assistant provenance, and bounded controller pipe draining;
- exact-head review at `ee8f1fa785a8a44295d839b3bac9c970a81f37cd` found missing positive
  workflow-transition and validator evidence-tool provenance. Fixes require every official required tool
  to complete successfully, persist observed transitions in the validator-bound manifest, recheck them at
  promotion, and reject zero/missing/failed validator evidence-tool outcomes;
- exact-head review at `3542ee007df66648c1f1292e2f0d58d04a8dada5` found explicit errored
  implementation/validator `agent_end` statuses were ignored; both paths now reject every explicit
  non-success terminal status with process-level regressions;
- final immutable exact-head GPT-5.6 Sol/high correctness/security/restart/verification/test-realism
  review at `ee474811378edd604e1e86e413f0bcafeced452b` reports no findings; local/remote equality,
  cleanliness, and generated-binary absence were confirmed before parent synchronization.

## Post-Slice-G parent synchronization

No behavior change or new production TDD slice is introduced. This stage repairs Git ancestry and
canonicalizes accepted evidence only.

- [x] Start state was clean on `fix/389-shepherd-proof-recovery` at accepted/local/remote Slice G head
      `ee474811378edd604e1e86e413f0bcafeced452b`, with no generated `pm` or `shepherd` binary.
- [x] Repo-local GSD adapter doctor/list/source checks and the programming-loop prompt passed.
- [x] Parent PR #390 remained open/draft from `feat/372-gsd-pi-go-shepherd` to `main` at exact head
      `d72e597e35b5104cf58936612053705c280fc2b1`.
- [x] Pre-squash head `c539b49bd767b0839f0989d52bd69da80c30843e` is an ancestor of Slice G.
- [x] Pre-squash and squashed-parent tree IDs are both
      `9c9ffd9a0e0f6d76955cd048978662d57e888291`.
- [x] Guarded ancestry-only merge `17ca31f6d04def71d55137d25d8194feaea10829` has parents Slice G
      and exact parent head; `git diff ee474811... HEAD` is empty immediately after the merge.
- [ ] Fresh full synchronization verification and exact-head GPT-5.6 Sol/high review after this separate
      planning checkpoint.
- [ ] Authorized branch push, exact local/remote equality, draft stacked PR, and CI monitoring.

## Post-Slice-G exact-head review fix — FOCUSED GREEN / INTEGRATION LOADER BLOCKED

Authorization: change only deletion-proof identity/revalidation and bounded Git execution plus minimum
evidence/process-test propagation. Preserve `ee474811`, `17ca31f6`, and `e53e9e56`; no amend, rebase,
rewrite, dependency, credential, canary, cleanup/migration, merge, or `main` mutation.

Confirmed review findings:

- **High:** `ArtifactManifest` creates a zero-hash deletion marker, but `verifyArtifactHashes` always
  opens the path, so an authorized tracked deletion cannot be independently validated or promoted.
- **Medium:** the Git runner buffers stdout/stderr without execution-time bounds, checks stdout only
  after process completion, and converts every `git show` failure—including size and Git errors—into
  a deletion marker.

RED tests before production edits:

- [x] Scoped tracked deletion yields explicit `Deleted=true` plus exact sentinel.
- [x] Out-of-scope deletion remains `ErrWriteScopeBreach`.
- [x] Rename is deterministic deletion plus addition with rename detection disabled.
- [x] Unknown/malformed status, Git failure, and oversized object never become deletion.
- [x] Validator accepts deletion only when the exact candidate path and symlink-free containing path
      are absent; recreated file/directory/symlink and symlinked parent fail.
- [x] Deletion flag/sentinel mismatch and present-artifact deletion sentinel fail.
- [x] Present artifacts that disappear or become symlinks fail before/post validation.
- [x] Git stdout/stderr are bounded while running; exact limit passes, over-limit has typed identity,
      errors are bounded/redacted, cancellation remains classifiable, and Git environment is sanitized.
- [x] Actual built `shepherd supervise` deletion succeeds through validator, ratification, promotion,
      and `final_human_gate`; rejected variants preserve canonical Git/GSD.

RED evidence captured before production edits:

```bash
cd agent-runtime/shepherd && go test ./internal/git -count=1
```
Expected FAIL: `Artifact.Deleted`, `DeletionSentinelHash`, `ErrOutputLimit`, `maxGitStdoutBytes`, and
`maxGitStderrBytes` are undefined, proving deletion identity and typed bounded Git output do not exist.

```bash
cd agent-runtime/shepherd && go test ./internal/validation -count=1
```
Expected FAIL: `ArtifactHash.Deleted` and `DeletionSentinelHash` are undefined, proving protected
validator requests cannot express typed deletion proofs.

```bash
cd agent-runtime/shepherd && go test ./cmd/shepherd -count=1
```
Control PASS in 43.465s; command package still compiles before production rewiring.

```bash
cd agent-runtime/shepherd && go test -tags=integration ./integration/... -run 'TestSupervise.*Delet|TestArtifact.*Delet' -count=1
```
Environment-blocked RED: built deletion tests could not reach runtime because this isolated checkout lacks
the packaged official GSD loader at `/tmp/.tools/gsd-pi-1.11.0/.../loader.js`; retrying with the only local
loader failed admission with `runtime_contract_mismatch: GSD command does not target the packaged loader`.
The process tests are present and remain a required GREEN gate when the loader is available.

GREEN/refactor evidence:

- Added typed `Deleted` JSON identity with `omitempty` to Git artifacts and validator artifact hashes,
  using the exact shared deletion sentinel string.
- `ArtifactManifest` now parses `git diff --name-status -z --no-renames`, treats `D` as deletion,
  `A/M/T` as present object hashes, rejects malformed/unknown status, and propagates Git/hash/limit
  failures without manufacturing deletion.
- Git stdout/stderr use bounded draining buffers with typed `ErrOutputLimit`, bounded sanitized
  diagnostics, argv execution, sanitized Git environment, and context cancellation identity.
- Present Git objects are hashed as a stream through `git cat-file blob` with an 8 MiB governed
  maximum; oversized blobs return `ErrOutputLimit` and never become deletion.
- Validator pre/post revalidation now requires consistent flag/sentinel pairs. Deleted paths pass only
  when absent through a lexical, no-follow existing-component walk; present paths must remain regular
  non-symlink files with the bound digest.
- Command evidence conversion, promotion proof decoding, integration proof assertions, and fake-runtime
  deletion scenarios propagate the typed deletion bit.

GREEN commands:

```bash
cd agent-runtime/shepherd && go test ./internal/git -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git 1.078s
cd agent-runtime/shepherd && go test ./internal/validation -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation 4.585s
cd agent-runtime/shepherd && go test ./cmd/shepherd -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/cmd/shepherd 45.426s
cd agent-runtime/shepherd && go test ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 0.917s, internal/validation 4.930s, cmd/shepherd 44.446s
cd agent-runtime/shepherd && go test -race ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 8.417s, internal/validation 100.724s, cmd/shepherd 253.859s
cd agent-runtime/shepherd && go test ./...
# PASS all Shepherd nested packages
cd agent-runtime/shepherd && go test -race ./...
# PASS all Shepherd nested packages
cd agent-runtime/shepherd && go vet ./... && go build ./cmd/shepherd && make verify
# PASS; generated agent-runtime/shepherd/shepherd removed afterwards
scripts/tests/shepherd-module-boundary.sh && git diff --check && go list ./...
# PASS boundary; PASS diff check; PASS root package list
make verify
# PASS root verify/smoke/connectorgen; generated ./pm removed afterwards
cd agent-runtime/shepherd && golangci-lint run ./...
# Expected nonzero accepted baseline: exactly 28 issues (25 errcheck, 2 staticcheck, 1 unused)
cd agent-runtime/shepherd && golangci-lint run --build-tags=integration ./...
# Expected nonzero accepted baseline: exactly 28 issues (25 errcheck, 2 staticcheck, 1 unused)
```

Integration environment note:

```bash
cd agent-runtime/shepherd && go test -tags=integration ./integration/... -run 'TestSupervise.*Delet|TestArtifact.*Delet' -count=1
```
The isolated worker could not execute this gate because its temporary checkout lacked the packaged
loader at `/tmp/.tools/gsd-pi-1.11.0/.../loader.js`; no production fallback or fake admission bypass
was added. The coordinator later ran the deletion process tests with the canonical packaged loader in
normal and race modes, both passing.

Orchestration decision: `spawned` for RED/GREEN implementation: one isolated GPT-5.5/high worker owned
Git artifact parsing, validator request identity, command evidence hashing, and process integration as
one sequential trust boundary. Planning and canonical verification remained coordinator-owned.

## Verification log

Slices A-G remain accepted at their exact checkpoints. Slice G is accepted at
`ee474811378edd604e1e86e413f0bcafeced452b`; parent ancestry is reconciled without altering that
checkpoint's content. The replacement exact-head review is blocked only on the two findings above.
No credential access, connector operation, canary, cleanup/migration, PR merge, or `main` mutation ran.
The only GitHub access in this stage so far was read-only parent-PR metadata.

## Post-Slice-G bounded Git / descriptor-root follow-up — FOCUSED GREEN

Start head: `bfc937ef2bc523950c14929b73b00d9e054957d6` with a clean worktree.
GSD checks: `scripts/gsd doctor`, `scripts/gsd list`, and
`scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto --dry-run` passed.
Skills recorded: `gsd-core`, `polymetrics-issue-delivery`, `gsd-programming-loop`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-context`, `golang-concurrency`, and `golang-lint`.

Confirmed adjacent RED tests added before production edits:

- strict full `--name-status -z --no-renames` parsing rejects leading, interior, extra, missing-final,
  malformed, and unknown records;
- a 129-record status stream is rejected before any `cat-file` process is spawned;
- `hashGitObject` rejects declared sizes above 8 MiB before streaming, allows exact 8 MiB, rejects
  streamed-byte mismatch, classifies nonzero cat-file separately from `ErrOutputLimit`, detects stderr
  overflow, and returns `ErrOutputLimit` on stream overflow;
- generic Git stdout/stderr exact limits pass, overflow returns `ErrOutputLimit`, parent cancellation
  remains `context.Canceled`, and a Unix process-tree regression proves endless descendants are reaped.

Exact RED evidence captured before production edits:

```bash
cd agent-runtime/shepherd && go test ./internal/git -count=1
```
Expected FAIL with:

- `TestArtifactManifestRejectsUnknownAndMalformedStatus/leading-terminator`: `status err=<nil>` and a
  manufactured present artifact;
- `TestArtifactManifestRejectsUnknownAndMalformedStatus/interior-terminator`: `status err=<nil>` and
  manufactured artifacts;
- `TestArtifactManifestRejectsUnknownAndMalformedStatus/extra-terminator`: `status err=<nil>`;
- `TestArtifactManifestRejectsUnknownAndMalformedStatus/missing-final-terminator`: `status err=<nil>`;
- `TestArtifactManifestParsesAllStatusesBeforeHashing`: 129 artifacts accepted and hashed;
- `TestHashGitObjectChecksDeclaredSizeBeforeStreaming`: oversized declared object returned a hash and
  `err=<nil>`;
- `TestHashGitObjectExactLimitPassesAndSizeMustMatch`: declared-size mismatch returned a hash and
  `err=<nil>`;
- `TestHashGitObjectClassifiesCatFileFailuresAndOverflow`: stderr overflow returned `err=<nil>`;
- `TestRunOutputLimitTerminatesProcessGroupDescendants`: endless output returned
  `context deadline exceeded` instead of `ErrOutputLimit` and process-group cleanup.

GREEN/refactor evidence:

- Added strict name-status parsing and a 128-artifact pre-hash ceiling.
- Added bounded argv Git process execution with internal `ErrOutputLimit` cancellation, parent-context
  precedence, finite `WaitDelay`, and Unix process-group cleanup; no shell was introduced.
- `hashGitObject` now uses bounded `git cat-file -s` before streaming, exact byte-count verification,
  bounded stderr detection, and immediate process cleanup on stream/stderr overflow.
- Validator deletion/present checks now open through descriptor-relative `os.Root`, pin existing parent
  roots, reject symlinks/replacement by stable file identity, check deleted absence relative to the
  pinned parent, and hash the opened descriptor for present files.
- Validation and promotion now alias the Git package deletion sentinel as the source of truth.
- Declined disposition: shared-repository Git config/environment policy remains out of scope for this
  bounded follow-up under the documented accepted same-UID host trust assumption; no environment
  allowlist/config policy was changed.

GREEN commands:

```bash
cd agent-runtime/shepherd && go test ./internal/git -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git 1.205s
cd agent-runtime/shepherd && go test ./internal/validation -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation 4.758s
cd agent-runtime/shepherd && go test ./cmd/shepherd -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/cmd/shepherd 44.967s
cd agent-runtime/shepherd && go test ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 1.160s, internal/validation 4.840s, cmd/shepherd 44.643s
cd agent-runtime/shepherd && go test -race ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 20.773s, internal/validation 100.993s, cmd/shepherd 254.629s
cd agent-runtime/shepherd && go test ./...
# PASS all Shepherd nested packages
cd agent-runtime/shepherd && go vet ./...
# PASS
cd agent-runtime/shepherd && go build ./cmd/shepherd && rm -f shepherd && git diff --check
# PASS
```

## Post-Slice-G cleanup/UTF-8 residual follow-up — RED

Start head: `ec8c2dc523a2ce55c0d4a4bcbd9b5739df541fad` with a clean worktree. Current phase artifacts/diff were read before test edits; deletion integration remains coordinator/loader-owned.

Confirmed adjacent RED tests added before production edits:

- `internal/git`: invalid UTF-8 in `--name-status -z` status/path records is rejected before string conversion, including deletion records.
- `internal/git` Unix: successful fake Git parents that leave same-process-group sleeping descendants are synchronously cleaned after ordinary `run` and `hashGitObject` parent exit.
- `internal/validation` Unix: artifact verification closes each descriptor root/file before the next artifact under a bounded 128-item set and low file-descriptor limit.

Exact RED evidence captured before production edits:

```bash
cd agent-runtime/shepherd && go test ./internal/git -run 'TestArtifactManifestRejectsUnknownAndMalformedStatus|TestRunCleansProcessGroupDescendantsAfterSuccessfulParentExit|TestHashGitObjectCleansProcessGroupDescendantsAfterSuccessfulParentExit' -count=1
```
FAIL:
- `TestArtifactManifestRejectsUnknownAndMalformedStatus/invalid-utf8-deletion-path`: `status err=<nil>` and deletion artifact path converted to `�`.
- `TestRunCleansProcessGroupDescendantsAfterSuccessfulParentExit`: sleeping descendant survived process-group cleanup.
- `TestHashGitObjectCleansProcessGroupDescendantsAfterSuccessfulParentExit`: sleeping descendant survived process-group cleanup.

```bash
cd agent-runtime/shepherd && go test ./internal/validation -run TestVerifyArtifactHashesClosesArtifactRootPerItem -count=1
```
FAIL: helper exited with `verify retained roots: openat agent-runtime: too many open files`.

GREEN/refactor evidence:

- `run` and `hashGitObject` now explicitly call `cleanupGitProcessTree` after `Run`/`Wait`, while `Cmd.Cancel` only signals the process group; parent-context errors retain priority over output-limit errors, and output-limit errors retain priority over cleanup errors.
- Unix process cleanup now synchronously kills and polls the Git process group after parent exit, covering both generic Git commands and `cat-file blob` hashing.
- Git name-status parsing rejects invalid UTF-8 status/path records before string conversion, scope checks, deletion sentinel construction, object hashing, or JSON evidence conversion.
- Validator artifact verification is extracted into per-artifact verification so each `os.Root`/file closes before the next bounded artifact, with close errors propagated.

GREEN commands:

```bash
cd agent-runtime/shepherd && go test ./internal/git -run 'TestArtifactManifestRejectsUnknownAndMalformedStatus|TestRunCleansProcessGroupDescendantsAfterSuccessfulParentExit|TestHashGitObjectCleansProcessGroupDescendantsAfterSuccessfulParentExit' -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git 0.525s
cd agent-runtime/shepherd && go test ./internal/validation -run TestVerifyArtifactHashesClosesArtifactRootPerItem -count=1
# PASS ok github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation 0.422s
cd agent-runtime/shepherd && go test ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 1.310s, internal/validation 5.546s, cmd/shepherd 44.649s
cd agent-runtime/shepherd && go test -race ./internal/git ./internal/validation ./cmd/shepherd -count=1
# PASS internal/git 26.482s, internal/validation 102.176s, cmd/shepherd 255.966s
git diff --check
# PASS
```

Isolated-worker integration gate was environment-blocked:

```bash
cd agent-runtime/shepherd && go test -tags=integration ./integration/... -run 'TestSupervise.*Delet|TestArtifact.*Delet' -count=1
```
FAIL in the temporary worker only: `official GSD loader is unavailable at /tmp/.tools/gsd-pi-1.11.0/node_modules/@opengsd/gsd-pi/dist/loader.js`.
No production fallback or fake admission bypass was added. Canonical packaged-loader runs later pass.

## Canonical review-fix verification — GREEN / REVIEW PENDING

Coordinator canonical-checkout evidence after fast-forwarding all isolated GPT-5.5/high commits:

```bash
cd agent-runtime/shepherd
gofmt -w internal/git internal/validation cmd/shepherd integration
go test ./internal/git -count=1
go test ./internal/validation -count=1
go test ./cmd/shepherd -count=1
go test -tags=integration ./integration/... -run 'TestSupervise.*Delet|TestArtifact.*Delet' -count=1
go test -race -tags=integration ./integration/... -run 'TestSupervise.*Delet|TestArtifact.*Delet' -count=1
go test ./internal/git ./internal/validation ./cmd/shepherd -count=1
go test -race ./internal/git ./internal/validation ./cmd/shepherd -count=1
go test -tags=integration ./integration/... -count=1
go test -race -tags=integration ./integration/... -count=1
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
```

All pass. Root `make verify`, module boundary, root `go list ./...`, planning JSON, diff hygiene, and
binary cleanup also pass. First lint run found one new test-only unchecked `os.RemoveAll`; after the
scoped test cleanup fix and affected normal/race reruns, default and integration-tagged lint both return
exactly 28 accepted findings: 25 `errcheck`, 2 `staticcheck`, 1 `unused`, zero differential.
`verificationPassed=true`. Two fresh independent read-only GPT-5.6 Sol/high processes reviewed exact
head `c72778def85ddccdee91bd648d7c0d569eb5fa94` against
`d72e597e35b5104cf58936612053705c280fc2b1`: correctness/recovery/test-realism PASS with no findings;
security PASS with no findings. Both confirm typed deletion/evidence/promotion binding, bounded Git
process/output behavior, descriptor-relative path stability, actual-CLI realism, and authorized scope.
The shared Git-config item remains a nonblocking documented same-UID trust assumption. This final
evidence-only commit will receive replacement exact-head review before push. No credential,
connector/API operation, canary, cleanup/migration, PR/merge, or `main` mutation occurred.

## Draft PR #456 nested-module test-infrastructure fix — RED

Start head/local/remote: `7432f0a5da90f255b74307d12c26863b61c1a16f`, clean. Draft PR #456 is
open from `fix/389-shepherd-proof-recovery` to `feat/372-gsd-pi-go-shepherd`; parent draft PR #390
remains the human-gated integration target. Failed GitHub evidence: workflow run `29578379908`, job
`87877981167`, check `nested-module`.

GSD command: `scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`.
Adapter doctor/list/source checks pass. Skills: `gsd-core`, `polymetrics-issue-delivery`,
`gsd-programming-loop`, `golang-how-to`, `golang-testing`, `golang-troubleshooting`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`,
and `golang-lint`. `.pi/skills/go-implementation/SKILL.md` is absent; no manual fallback.

### RED A — Node fixture portability

The requested ordinary local reproduction did not fail because the current NVM Node command
`/Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/node` is a regular canonical file, not a
terminal symlink. This mismatch is recorded rather than overwritten. The same pre-edit test passes
with ordinary `PATH`.

A deterministic temporary symlinked `node` placed first on `PATH` reproduces the CI class exactly:

```text
registry_test.go:35: runtime_contract_mismatch: open bounded runtime source
FAIL github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd
```

The unedited control with the symlink target's canonical directory prepended to `PATH` passes. Root
cause: test fixtures hash/qualify `exec.LookPath("node")` after lexical `filepath.Abs`, so production
`O_NOFOLLOW` correctly rejects a symlink. Fixture canonicalization must use `filepath.EvalSymlinks`;
production no-symlink admission must not change.

### RED B — descendant assertion timing

CI failed:

```text
TestRunnerKillsInheritedOutputDescendantAfterSuccessfulParentExit
runner_test.go:444: supervised descendant pid 6847 survived cleanup: <nil>
```

That test performs one immediate `syscall.Kill(pid, 0)` check. Adjacent process-cleanup tests already
poll for `ESRCH` with a five-second maximum and 10 ms interval. The test-only fix must share bounded
eventual verification, fail on a genuinely live process, and introduce no platform skip or production
cleanup change.

`verificationPassed=false` until all requested targeted/full gates, exact-head review, push, and fresh
PR CI complete. No test source or production source was edited before this RED evidence.

### GREEN — portable fixtures and bounded descendant assertion

Isolated `openai-codex/gpt-5.5`/high worker commit:
`20540e79bf8929390e64fcd165046d6704199e6b` (`test(shepherd): make GSD runtime checks portable`).
Only these test files changed:

- `internal/gsd/process_unix_test.go`
- `internal/gsd/registry_test.go`
- `internal/gsd/runner_test.go`
- `internal/gsd/runtime_snapshot_test.go`
- `internal/gsd/test_helpers_test.go`

`qualifiedNodePathForTest` now applies `exec.LookPath`, `filepath.Abs`, and
`filepath.EvalSymlinks`; registry/snapshot commands and hashes use that canonical fixture binary.
`TestResolveQualifiedNodeRejectsSymlinkedExecutable` first accepts the exact canonical binary/hash,
then requires `ErrRuntimeContractMismatch` for a symlink to the same binary. Production files and
qualification policy are unchanged.

`waitForPIDExitForTest` polls `syscall.Kill(pid, 0)` for `ESRCH` with a five-second deadline and 10 ms
interval. The formerly immediate runner assertion and adjacent duplicated ordinary-exit/maintenance
loops use it. Any process that remains live through the bound fails; no platform skip or production
cleanup change was added.

Coordinator GREEN evidence:

```text
Node fixture target set -count=10: PASS (6.146s)
Requested descendant target set -count=10: PASS (0.398s)
Adjacent maintenance/process descendant set -count=10: PASS (0.453s)
Symlink fixture/security regressions -count=10: PASS (2.823s)
go test -race ./internal/gsd -count=3: PASS (37.554s)
go test ./internal/gsd -count=5: PASS (10.462s)
go test ./...: PASS
go test -race ./...: PASS
go test -tags=integration ./integration/... -count=1: PASS (122.417s)
go test -race -tags=integration ./integration/... -count=1: PASS (583.931s)
go vet ./...: PASS
go build ./cmd/shepherd: PASS
make verify (nested): PASS
make verify (root): PASS
scripts/tests/shepherd-module-boundary.sh: PASS
go list ./...: PASS (145 packages)
git diff --check and planning JSON: PASS
generated pm/shepherd binaries absent: PASS
```

`golangci-lint run ./...` returns the accepted nonzero baseline exactly: 25 `errcheck`, 2
`staticcheck`, 1 `unused`, with no finding in any changed test file. `verificationPassed=true`; fresh
exact-head GPT-5.6 Sol/high review, normal push, and fresh PR #456 CI remain.
