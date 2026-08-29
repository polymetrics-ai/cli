# Verification: Issue #4283 Batch-1 source-rigidity R2

## Completed planning checks

- [x] Fetched `origin/fm/cli-top100-declaration-batch-r1` and checked out SHA `6ec34964fda7a78a1736a5cd8933a46418346261` detached in the assigned task worktree.
- [x] Verified the required ancestor with `git merge-base --is-ancestor`.
- [x] Verified PR #4294 API base/head/SHA: `main`, `fm/cli-top100-declaration-batch-r1`, `6ec34964fda7a78a1736a5cd8933a46418346261`.
- [x] Ran `scripts/gsd doctor`, resolved all required lifecycle command sources, generated `discuss-phase` and `plan-phase --tdd` prompts, and ran `go run ./cmd/agentcontractgen check`.
- [x] Read the required routing, GSD adapter, issue contract, CLI parity, connector lane, and source-rigidity audit materials.
- [x] Processed Firstmate inbox messages `001.msg` and `002.msg` as historical scope context.
- [x] Captured RED policy evidence: the three authoritative direct clauses in required-skills routing, task-skill matrix, and issue-agent contract had the obsolete mandatory split.
- [x] Captured GREEN policy evidence: `focused-policy-regression` passes against the three authoritative clauses; `AGENTS.md` has no direct obsolete clause; all 17 non-authoritative copies were restored exactly; YAML/JSON parsing, `git diff --check`, and `go run ./cmd/agentcontractgen check` pass. The repository has no dedicated delivery-policy regression harness.
- [x] Processed Firstmate inbox message `006.msg`: commit `924197b49adc516ec7bdb3b6d928c9701cfab946` was normal-fast-forward pushed and remote verified, then `006.msg` was moved to `handled/`.
- [ ] Processed current Firstmate inbox message `007.msg`: narrowed correction, focused validation, commit/push, then move to `handled/`.

## Not run by design

- [ ] Red/green source or runtime tests — policy correction and the exhaustive cohort matrix precede production changes.
- [ ] Production implementation/generator/docs edits — pending the cohort matrix and Foundation Atlas classification.
- [ ] Provider-live requests, credentials, reverse-ETL dispatch, or destructive operations — excluded by the task.
- [ ] Exact-SHA Codex review, PR checks remediation, commit, or push — require a valid scoped production candidate.

## Current constraint

The captain has changed the relevant repository policy, not reduced a quality gate. The immediate prerequisites are the committed policy correction, a green active-policy scan, an exhaustive machine-checkable cohort matrix, Atlas classification, and red tests before any production behavior change.
