# Issue 389 Verification — proof-recovery repair

## Evidence policy

All prior PASS/checkmark claims for independent validation, ratification, recovery planning, canaries,
and final readiness are superseded for this repair run. A gate is checked only after it is rerun against
the current branch and exact candidate head.

## Focused gates to add/run by slice

### A. Independent validation and ratification

- [ ] Missing validator evidence fails closed.
- [ ] GPT-5.5 validator evidence fails closed.
- [ ] Stale candidate head fails closed.
- [ ] Validator `RETRY`/`HALT` fails closed.
- [ ] `authority.Ratify` is called with real validation result and stored evidence.
- [ ] Canonical branch remains unchanged on every failed validation/ratification path.

### B. Attempt lifecycle and crash recovery

- [ ] Attempt identity, branch, path, base/candidate/validated heads, and all lifecycle states persist
      in SQLite.
- [ ] Restart reconciles database-owned orphan worktrees/branches.
- [ ] Unknown or live worktrees are never deleted automatically.
- [ ] Early preparation/query/runtime failures transition explicitly.
- [ ] Retry always receives a fresh attempt worktree.

### C. GSD-state promotion

- [ ] No `RemoveAll` and in-place copy into canonical `.gsd` tree.
- [ ] Staged snapshot is validated before promotion.
- [ ] SQLite database/WAL state is handled consistently.
- [ ] Promotion journal supports failures before Git promotion, after Git promotion, before state swap,
      and after state swap.
- [ ] Restart converges idempotently to one consistent Git/GSD state.

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
- Important remaining blocker: production `persistSuccessProof` still manufactures validator/ratification evidence and must be replaced before Slice A can be considered complete.
