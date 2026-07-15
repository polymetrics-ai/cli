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

### C. GSD-state promotion — GREEN (candidate diff)

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

### D. Official GSD 1.11 registry loading

- [ ] Structured normalized export from pinned runtime replaces regex parsing.
- [ ] Array spreads such as `RUN_UAT_WORKFLOW_TOOL_NAMES` resolve.
- [ ] Allowed, required, and forbidden tools are preserved.
- [ ] Model routing comes only from official phase metadata.
- [ ] Null/unknown units fail closed unless explicitly governed sidecars.
- [ ] Real pinned GSD 1.11 registry fixture passes.

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
- Lint: FAIL `cd agent-runtime/shepherd && golangci-lint run ./...` with the same 30 pre-existing findings (`errcheck`, one `ineffassign`, two `staticcheck`, two `unused`). A newly introduced validator parser staticcheck finding was fixed; final output returns to exactly the 30-item baseline.
- Canaries, PR creation, final Sol review, and Slice D onward remain blocked.
