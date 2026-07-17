# Issue 389 Verification — proof-recovery repair

## Evidence policy

All prior PASS/checkmark claims for independent validation, ratification, recovery planning, canaries,
and final readiness are superseded for this repair run. A gate is checked only after it is rerun against
the current branch and exact candidate head.

## Focused gates to add/run by slice

### A. Independent validation and ratification

- [x] Missing validator evidence fails closed against the real production validator.
- [x] GPT-5.5 validator evidence fails closed against the real production validator.
- [x] Stale candidate head fails closed against the real production validator.
- [x] Validator `RETRY`/`HALT` fails closed against the real production validator.
- [x] `authority.Ratify` is called with real validation result and stored evidence.
- [x] Canonical branch remains unchanged on every failed validation/ratification path in command tests; production validator unit tests also prove candidate `.gsd` remains unchanged on rejected paths.
- [x] Dedicated validator execution invokes configured Pi directly and uses neither `validate-milestone` nor invented GSD verbs.
- [x] Exact Pi flags enforce Sol/high, JSON print mode, dedicated sessions, read-only tools, disabled project resources, and bounded capability probing.
- [x] Opt-in actual Pi smoke produced a fresh bound Sol/high session/result.
- [x] Result transport is nonce-bound under protected Shepherd state outside the candidate worktree.
- [x] New validator session, exact model, high thinking, request nonce, candidate head, evidence hash,
      base branch, and durable state version are all verified.

### B. Attempt lifecycle and crash recovery — COMPLETE / GREEN `1a050692`

- [x] Attempt identity, confirmed branch/path ownership, controller owner/epoch, base/candidate/validated
      heads, bounded diagnostics, timestamps, and all exact lifecycle states persist in SQLite/reopen.
- [x] Legal graph rejects skipped/backward/terminal/stale-owner transitions and reserves ratification for
      the atomic proof+attestation transaction; `promoting` cannot become cleanup-eligible.
- [x] Real supervise create/prepare/query/dispatch/validate/ratify/promote/failure/cleanup paths update lifecycle.
- [x] Startup reconciliation runs before the first supervise query and is idempotent across restarts.
- [x] Cleanup requires confirmed database ownership plus exact path/branch/head/common-dir/non-live proof.
- [x] Unknown, unconfirmed, mismatched, checked-out, live-running, and promoting resources are preserved.
- [x] Preparation/query/runtime/cleanup failures transition explicitly with bounded diagnostics.
- [x] Retry after retained failure receives a fresh branch/worktree.
- [x] Pre-Slice-B migration preserves delivery-run, proof, and attestation records.
- [x] Repository-global lock, lease takeover, stale delivery/unit interruption, and human-gated absent-resource
      resolution provide hard-crash convergence without broad prune, deletion, or `RemoveAll`.
- [x] `attempt_root` is explicitly disjoint from canonical worktree and protected `state_dir`.
- [x] Slice A validation, ratification, exact-head, write-scope, and delayed-promotion regressions remain green.
- [x] Focused tests, full/race module verification, root repository gates, and exact 30-finding lint baseline passed.

### C. GSD-state promotion — COMPLETE / GREEN `f0fbf47f`

- [x] Journal identity binds delivery/generation/unit/attempt, base/candidate/validated heads,
      proof/attestation, governance version, and staged GSD manifest/hash.
- [x] Journal close/reopen persistence and the eight required journal states are covered.
- [x] Candidate `.gsd` is staged and fully validated before any canonical Git mutation.
- [x] `gsd.db` uses a consistent SQLite snapshot/checkpoint/backup path; integrity passes and installed
      state has no stale WAL/SHM dependency.
- [x] Manifest is deterministic and bounded by relative path, file type, count, size, and content hash;
      symlinks, special files, traversal, and unexpected paths fail closed.
- [x] No `RemoveAll` or remove-and-repopulate operation touches canonical `.gsd`.
- [x] Same-filesystem stage/backup renames and parent-directory fsyncs make both rename boundaries recoverable.
- [x] Failure injection covers before/after Git promotion, before backup rename, between renames, and
      after install before completion.
- [x] Restart at every boundary converges idempotently without duplicate effects or mixed Git/GSD state.
- [x] Base head may resume promotion; candidate head completes forward; any other head blocks.
- [x] Dirty canonical worktree, stale lease, unauthorized branch, or invalid proof blocks before mutation.
- [x] Missing/changed/corrupt/mismatched/symlinked stage or backup is preserved and journal becomes blocked.
- [x] Unknown stage/backup directories remain untouched.
- [x] `promoting` with a complete valid journal is recoverable; without one remains human-gated.
- [x] Real supervise startup recovery runs before canonical GSD query or new dispatch.
- [x] Existing Slice A/B tests remain green; final lint-baseline/differential command is recorded after checkpoint.

Slice C local evidence: focused promotion tests PASS; all nine journal/Git/swap failpoints PASS;
full nested tests PASS; store race PASS after final attestation hardening; full command/workspace race
PASS before the final store-only binding change; vet/build/gofmt/diff checks PASS. Root `make verify`, module boundary, root `go list`, and diff checks PASS. Nested lint reports 29
baseline findings and zero findings in Slice C production files (no differential regression).

### D. Official GSD 1.11 registry loading — COMPLETE / GREEN

- [x] Validated pinned runtime root/version/source identity is required before export.
- [x] Export uses argv execution without shell interpolation, timeout/cancellation, and bounded output.
- [x] Strict normalized JSON rejects missing, malformed, partial, oversized, duplicate, and unknown fields.
- [x] Realistic 1.11 array spreads such as `RUN_UAT_WORKFLOW_TOOL_NAMES` resolve.
- [x] Kind, scope class, phase chain, allowed tools, required tools, forbidden tools, and reasons are exact.
- [x] Missing registry, wrong version, symlink/path escape, and source drift return `runtime_contract_mismatch`.
- [x] Null phase/tool contracts receive no built-in fallback.
- [x] Unknown units/phases fail closed; versioned sidecars remain separate from official metadata.
- [x] Official coordination phases route Sol/high and execution phases route GPT-5.5/high.
- [x] Unit names do not influence model routing.
- [x] Child/subagent model events cannot overwrite top-level observed evidence.
- [x] Prompt-advertised tools are checked against normalized official contracts.
- [x] Production startup cannot use `BuiltinUnitRegistry` or another sample fallback.
- [x] Complete host runtime tree and absolute Node executable are hash-pinned and privately snapshotted.
- [x] Runtime roots under canonical/attempt worktrees fail closed.
- [x] Registry import uses verified immutable bytes rather than mutable runtime paths.
- [x] Current-run identity rejects every unexpected top-level transition and stale fallback session.
- [x] Session headers reject symlinks, duplicates, unknown/trailing fields, replacement, and ambiguity.
- [x] Every observed `gsd_*` call is allowed and not forbidden by normalized unit metadata.
- [x] Exporter and validator process groups synchronously cancel descendants with WaitDelay and bounded cleanup.
- [x] Runtime/policy hashing is bounded, no-follow, and checks inode/path identity before/after reads.
- [x] Podman resolves immutable inspected image IDs and fails closed without full-image qualification.
- [x] Focused GSD, validation, store, supervisor, workspace, and command tests pass.
- [x] Full/race/vet/build/root verify/module-boundary/diff/go-list gates pass.
- [x] Lint is 28 findings, below the 29-finding baseline, with zero Slice D differential findings.

Security-auditor finding disposition (first read-only pass, working diff):
- Host runtime not fully pinned / candidate-root execution — **resolved by complete private v3 snapshot and worktree rejection**.
- Mutable Podman tag — **resolved by immutable image-ID resolution plus fail-closed image qualification**.
- Model identity not current-run bound / later transitions ignored — **resolved**.
- Ambiguous session/event provenance — **resolved for governed current-run evidence and durable attempt binding**.
- Verify/import and source-hash TOCTOU — **resolved through immutable verified-byte import and stable private snapshot reads**.
- Incomplete prompt/tool enforcement and Podman patch parity — **resolved by all-observed-tool checks, exact active prompt-tree validation, and fail-closed Podman admission**.
- Exporter descendant leakage — **resolved with synchronous process-group cleanup and WaitDelay**.
- Unbounded source hashing — **resolved with bounded no-follow pre/post inode and pathname identity reads**.

Final local review disposition (four independent read-only reviewer/security cycles):
- Runtime root/mode/owner and verify/import TOCTOU — **resolved** with the complete v3 manifest,
  private read-only snapshot, per-launch digest/owner/settings/prompt guards, and immutable data imports.
- Canonical aliases, state-only heads, and attempt identity — **resolved** with canonical type/ID binding,
  discuss milestone equality, disposable worktrees, fresh hook-disabled checkpoints, and durable
  generation/head/attempt/session fingerprints.
- Current-run identity and continuation — **resolved** for governed runs via strict one-session deltas,
  every-transition validation, live/durable equality, and persisted evidence; unqualified disposable
  continuation now fails closed.
- Tool enforcement — **resolved** for every observed `gsd_*` start, including uncontracted commands and
  foreign MCP spoofing, while normal non-GSD MCP tools remain available.
- Process cleanup — **resolved** for GSD exporters/runners and independent validator probe/run on
  cancellation and normal parent exit; dedicated descendant-termination regressions pass.
- Settings/preferences/session bounds — **resolved** with owned regular no-symlink stable reads and
  bounded session enumeration.
- Podman parity — **resolved by fail-closed admission** until a complete image digest is approved.
- Same-UID host isolation — **documented accepted architecture trust assumption**, not an isolation
  claim; elimination requires a future separate UID, OS sandbox, or human-qualified container.
- Accepted Slice A validator live-event/session correlation schema remains unchanged; Slice D replaced
  its multi-selection filesystem fallback with one strict current-run session evidence path.

### E. Sol/high recovery planning — COMPLETE / GREEN

- [x] Static recovery text is rejected and removed from production.
- [x] Required failure classes and typed actions are exhaustive; unknown values fail closed.
- [x] Unsafe classes and authority/security/dependency/secret/destructive changes never invoke planner retry.
- [x] GPT-5.6 Sol/high recovery planner is a separate protected Pi process with fresh nonce/session.
- [x] One fresh top-level session proves exact model, high thinking, session ID, and current-run identity.
- [x] Strict bounded request/result JSON binds nonce, delivery/generation/unit/attempt/head/class,
      evidence/authority hashes, action, typed steps, backoff, issued/expiry times, and replay state.
- [x] Duplicate/case-duplicate/unknown/partial/oversized/trailing/mismatched/stale/replayed output fails closed.
- [x] Action allowlist and per-class action policy are enforced by the controller and durable store.
- [x] Plan steps are bounded typed primitives; no arbitrary command/path/tool/external write executes.
- [x] Durable budgets are independently keyed by delivery/generation/unit/head/failure class.
- [x] Atomic owner/lease-epoch-fenced reservation prevents duplicate consumption and policy changes.
- [x] Deterministic exponential backoff and `next_retry_at` survive restart and block early dispatch.
- [x] Exhaustion durably enters awaiting decision or blocked without redispatch/duplicate decision.
- [x] Planner failures cannot create an unbounded planner loop.
- [x] External-effect uncertainty is typed, blocks, and never blindly repeats writes.
- [x] Planner process groups clean descendants on timeout, cancellation, and normal exit; unsupported
      non-Unix cleanup fails planner construction closed.
- [x] Planning/rejection leaves canonical Git and `.gsd` unchanged.
- [x] Accepted retries use fresh attempt resources.
- [x] Focused recovery/store/supervisor/command tests pass.
- [x] Full/race/vet/build/root verify/module-boundary/diff/go-list gates pass.
- [x] Lint equals the accepted 28-finding baseline with zero Slice E findings.
- [x] Independent Sol/high correctness and security reviews have no unresolved actionable findings.

### F. Authority-gated external effects — COMPLETE / GREEN `ea88c92f`

Architecture and authority:
- [x] Decision summaries, questions, statuses, and future governed mutations route only through the
      fenced durable outbox.
- [x] Direct `SyncDecisionComment`, `SyncQuestionComment`, and other write-capable GitHub production
      paths outside the outbox executor are absent by architecture test.
- [x] Reply polling depends on a read-only GitHub port; only the executor receives write capability.
- [x] Only controller policy derives immutable narrow grants; workers/helpers/executors cannot self-grant.
- [x] Grants bind delivery/repository/issue-or-PR/capability/generation/head/epoch/target.
- [x] `forbidden_main_merge`, `merge.main`, `pr.merge`, unsupported effects, stale heads/epochs, and
      changed targets fail closed.

Persistence, state, and recovery:
- [x] Strict versioned payload decoding and canonical payload/hash identity are covered.
- [x] Persisted records contain no credentials, tokens, raw environment values, arbitrary commands,
      payload secrets, or unbounded diagnostics.
- [x] Pending/claimed/sent/failed/uncertain/blocked/cancelled legal transitions and terminality pass.
- [x] Claim owner/epoch fencing, expiration, crash recovery, restart, and idempotent replay pass.
- [x] Definite pre-send failure follows bounded policy; ambiguous post-send failure becomes uncertain
      and is never blindly replayed.
- [x] Immutable request, authority, claim, result, and bounded typed error identity survive SQLite reopen.

GitHub reconciliation and cross-store honesty:
- [x] Stable Shepherd markers reconcile exact marker/payload identity before write.
- [x] Exact duplicates suppress writes; duplicate markers, payload conflicts, changed targets, and
      ambiguous read results block.
- [x] Older claimed summary revisions cannot overwrite newer ledger revisions.
- [x] Question effects bind request ID, generation, unit, head, and external comment ID.
- [x] Deterministic startup reconciliation covers every decision-ledger/outbox crash boundary without
      claiming cross-database atomicity.
- [x] All required typed bounded telemetry transitions are covered.

Gates and review:
- [x] Focused outbox/store/GitHub/domain/recovery/command tests pass.
- [x] Full/race/vet/build/nested and root verify/module-boundary/diff/go-list gates pass.
- [x] Lint remains exactly the accepted 28 findings with zero Slice F differential.
- [x] Independent exact-head GPT-5.6 Sol/high correctness and security reviews have no unresolved
      actionable findings.
- [x] No live GitHub mutation occurred.

### G. Real supervise integration coverage — COMPLETE / GREEN `ee474811`

Harness integrity:
- [x] `//go:build integration` tests build a temporary Shepherd executable and invoke actual
      `shepherd supervise` argv/config parsing; no direct `runSupervise` shortcut.
- [x] Each test owns isolated real Git, SQLite, GSD home/state, attempt root, process logs, and fakes.
- [x] GSD/Pi/GitHub fakes are bounded strict-argv processes; no network, credentials, shared user state,
      production-reachable crash flag, or order dependence.

Lifecycle and rejection:
- [x] Success proves metadata-routed GPT-5.5/high implementation, fresh GPT-5.6 Sol/high validation,
      exact-head/hash artifact proof, attestation/ratification, normalized Git/GSD promotion, safe attempt
      cleanup, durable outbox convergence, terminal JSON `final_human_gate`, and no merge attempt.
- [x] Missing/GPT-5.5/non-high validator evidence, stale/moving candidate, `RETRY`, `HALT`, missing gate,
      no governed delta, and missing/changed artifact preserve canonical branch/head/status/GSD and persist
      correct bounded failure/retention.
- [x] Promotion restart covers every durable journal/Git/state-swap/final-projection boundary with
      idempotent forward convergence, WAL-normalized proof binding, legacy post-Git recovery, and no mixed state.
- [x] Running termination, validation failure/fresh attempt, exact session/path identity, stale evidence
      rejection, and unknown/ambiguous worktree preservation/blocking pass.
- [x] Recovery planning/budget/awaiting-decision/reply restart and unauthorized/edited/stale/duplicate reply
      rejection pass.
- [x] Outbox pending/expired claim/post-write uncertainty/reconciliation/duplicate/collision cases suppress
      duplicate writes and never blindly replay.
- [x] Official registry spread metadata preserves exact phase/tool contracts and routes by phase; unknown/
      partial/stale metadata fails closed.

Required gates:
- [x] `go test -tags=integration ./integration/... -run TestSuperviseFakeRuntime -count=1`.
- [x] `go test -tags=integration ./integration/... -run TestSuperviseRestart -count=1`.
- [x] `go test -tags=integration ./integration/... -count=1`.
- [x] `go test -race -tags=integration ./integration/... -count=1` (race-built child processes).
- [x] Focused component regression command passes.
- [x] Full nested tests/race/vet/build/`make verify`, root `make verify`, boundary, diff, and root list pass.
- [x] Default and integration-tagged lint remain exactly 28 accepted findings with zero Slice G differential.
- [x] Exact-head GPT-5.6 Sol/high correctness/restart/security/test-realism review at
      `ee474811378edd604e1e86e413f0bcafeced452b` has no findings.

### Post-Slice-G parent synchronization and stacked-PR preparation

- [x] Clean local/remote Slice G equality and generated-binary absence confirmed before synchronization.
- [x] Parent PR #390 is open/draft from `feat/372-gsd-pi-go-shepherd` to `main` at exact expected head
      `d72e597e35b5104cf58936612053705c280fc2b1`.
- [x] `c539b49bd767b0839f0989d52bd69da80c30843e` is a Slice G ancestor.
- [x] Pre-squash and squashed-parent tree IDs are exactly equal:
      `9c9ffd9a0e0f6d76955cd048978662d57e888291`.
- [x] Ancestry-only merge `17ca31f6d04def71d55137d25d8194feaea10829` has the required two parents.
- [x] The ancestry merge has zero diff from accepted Slice G and leaves a clean worktree.
- [x] Planning reconciliation touches only `.planning/phases/issue-389-shepherd-hardening/`.
- [ ] Fresh full normal/race/integration/vet/build/verify/lint/hygiene gates pass after planning commit.
- [ ] Fresh exact-head GPT-5.6 Sol/high correctness/security/recovery/test-realism review has no findings.
- [ ] Branch push, exact local/remote equality, clean tree, draft stacked PR, and required CI checks pass.

### Post-Slice-G exact-head review fix

Deletion proof:
- [x] Explicit typed deletion and exact sentinel are jointly required and hash-bound.
- [x] Present artifacts require `Deleted=false`, a normal content hash, regular no-follow path, and
      exact bounded content digest.
- [x] Deleted artifacts require exact absence through a symlink-free contained parent chain before and
      after validation.
- [x] Scoped deletion, out-of-scope deletion, deterministic rename, malformed/unknown status, recreated
      path variants, flag/sentinel mismatch, and disappearing present artifacts are covered.
- [ ] Actual built-CLI deletion reaches ratification, promotion, and `final_human_gate`; rejected variants
      leave canonical Git/GSD unchanged. Test exists but this isolated checkout lacks the packaged
      official GSD loader required to execute the integration harness.

Bounded Git execution:
- [x] Git diff status, stdout, stderr, object bytes, and errors are bounded during execution.
- [x] Exact limits pass; over-limit stdout/stderr return typed/sentinel errors with bounded diagnostics.
- [x] Context cancellation and nonzero Git exit remain distinguishable and never create deletion records.
- [x] Argv execution and sanitized Git environment remain unchanged; no shell or dependency is added.

Gates/review:
- [x] Focused Git/validator/command tests pass.
- [x] Focused race and full normal/race nested Shepherd tests pass.
- [ ] Deletion integration tests are blocked by missing packaged official GSD loader in this checkout.
- [x] Full nested tests/race/vet/build/`make verify`, root `make verify`, boundary/list/diff/hygiene pass.
- [x] Default and integration-tagged lint remain exactly 25 `errcheck`, 2 `staticcheck`, 1 `unused`.
- [ ] Fresh exact-head GPT-5.6 Sol/high correctness and security reviews have no unresolved findings.
- [ ] Only then: fast-forward push, draft stacked PR, CI monitoring, and stop before canaries.

## Required command gates after each coherent slice

```bash
cd agent-runtime/shepherd
gofmt -w cmd internal
go test <focused packages>
go test ./...
go test -race ./...
go vet ./...
golangci-lint run ./...
go build ./cmd/shepherd
make verify
cd ../..
scripts/tests/shepherd-module-boundary.sh
git diff --check
go list ./...
```

## Deferred/human-gated checks

- [ ] Merge-disabled Twenty canary to `final_human_gate` — deferred until exact candidate head has
      independent Sol/high validation and human approval.
- [ ] Merge-disabled Asana canary to `final_human_gate` — deferred until exact candidate head has
      independent Sol/high validation and human approval.
- [ ] Parent PR #390 merge — human-only, not executable by this agent.
- [ ] Any issue-389 stacked PR merge or ready-for-review transition — human-only for this stage.
- [ ] Cleanup/migration and all `main` mutation — blocked.

## Current verification status

- GSD adapter health: PASS (`scripts/gsd doctor`, `scripts/gsd list`).
- Slice F programming-loop activation: PASS (`scripts/gsd prompt programming-loop run --phase issue-389-shepherd-hardening --mode auto`); `scripts/gsd sources programming-loop` resolved the repo-local adapter and pinned official sources.
- Slice A store hardening focused gate: PASS `cd agent-runtime/shepherd && go test ./internal/store -run 'TestArtifactProofRejectsUnratifiedResult|TestAttestationRejectsNonProceedVerdicts|TestArtifactProofBindsExactHeadsAndRatification|TestAttestationPersistsValidatorProof' -count=1`.
- Nested module unit gate after partial store hardening: PASS `cd agent-runtime/shepherd && go test ./...`.
- Race gate: PASS `cd agent-runtime/shepherd && go test -race ./...`.
- Vet/build/make/boundary/root listing: PASS `cd agent-runtime/shepherd && go vet ./... && go build ./cmd/shepherd && make verify && cd ../.. && scripts/tests/shepherd-module-boundary.sh && git diff --check && go list ./...`.
- Lint gate: FAIL `cd agent-runtime/shepherd && golangci-lint run ./...` with existing `errcheck`, `ineffassign`, `staticcheck`, and `unused` findings outside the focused proof hardening. This repair did not claim lint green.
- Previous Slice A completion evidence at `19d051c6`: false green. It did not test the real production validator and left no real proof producer.
- `99604d48` Slice A evidence: second false green because production called unsupported `gsd headless shepherd-validate`; helper tests did not prove production callability.
- Corrected Pi production-validator focused gate: PASS `cd agent-runtime/shepherd && go test ./internal/validation ./internal/store ./internal/authority ./internal/workspace ./cmd/shepherd`.
- Exact live Pi smoke: PASS `POLYMETRICS_SHEPHERD_LIVE_VALIDATOR=1 go test ./internal/validation -run TestLivePiValidatorSmoke -count=1 -v`; observed Sol/high, fresh session `019f62b3-9830-7129-9c93-2104ed54a10e`, fixture head `6650f5e18ecbbf15c18739a8422fa1ba663a0635`, bound evidence hash, verdict `PROCEED`.
- Full nested test: PASS `cd agent-runtime/shepherd && go test ./...`.
- Full nested race: PASS `cd agent-runtime/shepherd && go test -race ./...`.
- Vet: PASS `cd agent-runtime/shepherd && go vet ./...`.
- Build/make/module-boundary/root list: PASS `cd agent-runtime/shepherd && go build ./cmd/shepherd && make verify && cd ../.. && scripts/tests/shepherd-module-boundary.sh && git diff --check && go list ./...`.
- Slice D official runtime fixture: PASS `GSD_OFFICIAL_LOADER=.../@opengsd/gsd-pi/dist/loader.js go test ./internal/gsd -run 'TestPrepareInstalledOfficialHostRuntime|TestLoadInstalledOfficialUnitRegistry' -count=1 -v`; exact GSD Pi 1.11.0 registry and the v3 `darwin/arm64` host snapshot passed.
- Slice D focused gate: PASS `cd agent-runtime/shepherd && go test ./internal/gsd ./internal/validation ./internal/store ./internal/supervisor ./internal/workspace ./cmd/shepherd -count=1`.
- Slice D final nested unit/race/vet/build gates: PASS `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./cmd/shepherd`.
- Slice D nested/root verification: PASS nested `make verify`, root `make verify`, `scripts/tests/shepherd-module-boundary.sh`, `git diff --check`, and root `go list ./...` (145 packages).
- Slice D lint differential: expected nonzero baseline output, 28 findings (`errcheck` 24, `ineffassign` 1, `staticcheck` 2, `unused` 1), below the 29-finding Slice D baseline; the temporary new identity-test finding was removed and there are zero new Slice D production findings.
- Slice D implementation, local review, coherent checkpoint, push, remote-head confirmation, and clean-worktree confirmation are complete.
- Slice E focused GREEN: PASS `cd agent-runtime/shepherd && go test ./internal/recovery ./internal/store ./internal/supervisor ./cmd/shepherd -count=1`.
- Slice E live planner smoke: PASS `POLYMETRICS_SHEPHERD_LIVE_RECOVERY=1 go test ./internal/recovery -run TestLivePiRecoveryPlannerSmoke -count=1 -v`; observed model `openai-codex/gpt-5.6-sol`, thinking `high`, fresh session `019f6721-1e33-7dea-9b20-991a2e004715`, strict bound action/evidence, and no tools.
- Slice E final nested gates: PASS full tests, full race, vet, build, and nested `make verify`.
- Slice E root/repository gates: PASS root `make verify`, module boundary, `git diff --check`, and root `go list ./...` (145 packages).
- Slice E lint differential: expected nonzero baseline output, exactly 28 findings (`errcheck` 24, `ineffassign` 1, `staticcheck` 2, `unused` 1), with no `internal/recovery` finding and zero new Slice E production findings.
- Independent GPT-5.6 Sol/high correctness/security review cycles covered the complete Slice E working diff from Slice D; every actionable finding is dispositioned in `TDD-LEDGER.md` refactor evidence and there is no unresolved finding.
- Slice E checkpoint acceptance: PASS at `9556cb24412f3598b2b8a94a3089b61ef3d1dd91`; local/remote equality and clean worktree were confirmed before Slice F planning edits.
- Slice F RED: PASS as a TDD gate (expected failing capability probe), `cd agent-runtime/shepherd && go test ./internal/outbox -count=1` fails on the intentionally missing typed outbox/controller/store/executor APIs before production edits.
- Slice F GREEN: focused/full/race/vet/build/nested and root `make verify`/module-boundary/diff/root `go list` gates pass on 2026-07-16.
- Slice F lint differential: expected nonzero exit with exactly the accepted 28 findings (`errcheck` 25, `staticcheck` 2, `unused` 1); no Slice F package finding.
- Slice F independent correctness/security/restart reviews: all actionable findings dispositioned through repeated GPT-5.6 Sol/xhigh read-only cycles.
- Slice F live mutation: none; GitHub behavior used fakes and durable local stores only.
- Slice F exact-head correctness/restart closure: no actionable findings after all review fixes.
- Slice F accepted checkpoint: `ea88c92f5f3c0b1c5f3f434fa52efba24624f803`; local/remote equality and cleanliness confirmed before Slice G.
- Slice G programming-loop activation: PASS through the healthy repo-local adapter; requested
  `.pi/skills/go-implementation/SKILL.md` is absent and recorded, without requiring manual fallback.
- Slice G RED: captured before production edits. The required fake-runtime command failed at built-CLI
  runtime admission before any store existed because no compile-tagged executor seam was available.
- Slice G GREEN: PASS actual-CLI success/rejection/registry/recovery/outbox/promotion/final-gate suites,
  including normalized WAL staging, Slice-F post-Git proof compatibility, post-validation/final-gate GSD
  drift rejection, exact proof diff/hash/metadata oracles, strict fake GitHub writes, and two heartbeats.
- Slice G full gates before exact-head review: PASS normal and race integration suites, full nested
  unit/race/vet/build and nested `make verify`, root `make verify`, module boundary, root `go list`,
  JSON/diff/format checks, and exact no-differential 28-finding default/tagged lint baseline.
- Exact-head review-fix focused/full normal integration gates: PASS, including same-process decision reply
  and expiry, real SIGINT, post-claim/post-execution outbox recovery, `agent_settled`, complete ordered
  GSD lifecycle/tool pairing, and inherited-output descendant cleanup. Final normal/race integration,
  nested `make verify`, root `make verify`, vet/build, module-boundary, JSON/diff/hygiene, and exact
  no-differential 28-finding default/tagged lint gates pass before checkpoint amendment.
- First exact-head review at `4592734803f20e1b4893efae2ebd900525a92868`: BLOCKED with actionable
  continuous-decision, SIGINT, pre-send outbox, terminal-event, and descendant-cleanup findings.
- Post-fix working-tree GPT-5.6 Sol/xhigh correctness/security closure: PASS; all first exact-head findings
  dispositioned.
- Replacement exact-head review at `b08c93cc6b1de6a6c89d57c14da6c14d01d7e420`: BLOCKED on incomplete
  validator turn/session provenance. The validator now requires an exclusive fresh session root, ordered
  successful non-retrying Pi lifecycle, and exact stream/final-durable-proof hash identity; focused review
  reports no remaining finding.
- Exact-head review at `c1a34d23585329a9eb7f64a1ef687e0268c17666`: BLOCKED on implementation
  turn/final-stop proof, lifecycle JSON aliases, durable assistant row provenance, and unbounded detached
  output drains. Fixes and adversarial regressions cover missing/error/retrying turns, exact durable stop,
  duplicate/ASCII/Unicode aliases, assistant-shaped non-message rows, detached output, and bounded object
  fields; final working-tree closure reports no finding.
- Exact-head review at `ee8f1fa785a8a44295d839b3bac9c970a81f37cd`: BLOCKED on positive
  workflow-transition and validator evidence-tool provenance. Complete required-tool sets, explicit
  successful outcomes, hashed observed-tool manifests, promotion rechecks, zero/partial workflow negatives,
  and zero/missing/failed validator-tool negatives now pass focused review with no findings.
- Exact-head review at `3542ee007df66648c1f1292e2f0d58d04a8dada5`: implementation and validator
  explicit errored `agent_end` statuses now fail closed with process regressions; RUN-STATE uses the required
  auditable orchestration/verification schema.
- Final Slice G exact-head GPT-5.6 Sol/high correctness/security/restart/verification/test-realism review:
  PASS at `ee474811378edd604e1e86e413f0bcafeced452b` with no findings.
- Slice G checkpoint push/equality/cleanliness: PASS before synchronization.
- Parent synchronization guards: PASS; PR #390 metadata/head, pre-squash ancestry, exact tree identity,
  two-parent `ours` merge, and zero Slice G content diff were verified.
- Fresh post-planning synchronization gates passed, but replacement exact-head GPT-5.6 Sol/high review
  of `e53e9e56b67145419a11f1b577f858922e1a4c50` BLOCKED on typed deletion revalidation and bounded
  Git output/error classification. Both findings are accepted and explicitly authorized for repair.
- During pre-fix verification, one Shepherd query cleanup `EPERM` failed closed then passed 20/20 plus
  full rerun; optional root-wide race timed out without a race report and is not an issue-389 gate.
- Post-Slice-G review-fix RED: PASS as expected before production edits for internal Git/validation
  typed-deletion and output-limit tests; command package control passed; deletion integration RED is
  environment-blocked by the absent packaged official GSD loader.
- Post-Slice-G review-fix GREEN: PASS focused Git/validation/command tests, focused race, full nested
  normal/race, vet/build/nested `make verify`, root `make verify`, module boundary, diff check, root
  package list, and exact default/tagged 28-finding lint baseline. Generated `shepherd` and `pm` binaries
  were removed.
- `verificationPassed` is false during the authorized review-fix cycle because deletion integration and
  exact-head review remain coordinator-owned pending gates.
- Draft stacked PR creation is pending replacement exact-head review. Canaries, credentials,
  cleanup/migration, ready-for-review transition, every PR merge, and `main` mutation remain blocked.
