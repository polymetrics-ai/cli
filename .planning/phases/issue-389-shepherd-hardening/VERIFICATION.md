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

### E. Sol/high recovery planning

- [ ] Static recovery text removed.
- [ ] GPT-5.6 Sol/high recovery-planning unit dispatch is observed and persisted.
- [ ] Evidence hash, typed action, bounded plan, and model/thinking are stored.
- [ ] Action allowlist enforced.
- [ ] Per-class durable recovery budgets enforced.
- [ ] Exhaustion enters durable `awaiting_decision` and survives restart.

### F. Authority-gated external effects

- [ ] Decision summaries/questions/statuses/future GitHub mutations route through fenced outbox.
- [ ] Direct `SyncDecisionComment` production paths removed.
- [ ] Pending/claim/send/failure/restart/idempotent replay covered.
- [ ] Workers cannot receive direct GitHub mutation paths.

### G. Integration coverage

- [ ] Successful implementation -> independent Sol/high validation -> ratification -> promotion ->
      `final_human_gate`.
- [ ] Missing or GPT-5.5 validator evidence.
- [ ] Stale candidate head.
- [ ] Validator `RETRY`/`HALT`.
- [ ] Crash/restart at every promotion boundary.
- [ ] Retained failed attempt followed by fresh attempt.
- [ ] Recovery planning and `awaiting_decision` restart.
- [ ] Outbox restart and duplicate suppression.
- [ ] Official registry spread metadata.
- [ ] Canonical worktree remains unchanged on every failed path.

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

## Current verification status

- GSD adapter health: PASS (`scripts/gsd doctor`, `scripts/gsd list`).
- Programming-loop prompt generation: PASS (`scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`).
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
- Slice E onward, canaries, PR creation, final Sol review, GitHub mutation, and parent PR #390 merge remain blocked.
