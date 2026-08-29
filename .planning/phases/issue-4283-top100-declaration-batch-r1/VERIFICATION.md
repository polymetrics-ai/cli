# Verification: Issue #4283 Batch-1 source-rigidity R2

## Completed planning checks

- [x] Fetched `origin/fm/cli-top100-declaration-batch-r1` and checked out SHA `6ec34964fda7a78a1736a5cd8933a46418346261` detached in the assigned task worktree.
- [x] Verified the required ancestor with `git merge-base --is-ancestor`.
- [x] Verified PR #4294 API base/head/SHA: `main`, `fm/cli-top100-declaration-batch-r1`, `6ec34964fda7a78a1736a5cd8933a46418346261`.
- [x] Ran `scripts/gsd doctor`, resolved all required lifecycle command sources, generated `discuss-phase` and `plan-phase --tdd` prompts, and ran `go run ./cmd/agentcontractgen check`.
- [x] Read the required routing, GSD adapter, issue contract, CLI parity, connector lane, and source-rigidity audit materials.
- [x] Processed Firstmate inbox messages `001.msg` and `002.msg` as historical scope context.
- [x] Captured RED policy evidence: an active-policy scan found 38 obsolete mandatory-split references before the correction.
- [x] Captured GREEN policy evidence: the active-policy scan finds no obsolete mandatory-split requirement and 37 bounded-cohort/ledger/Atlas declarations; YAML/JSON parsing, `git diff --check`, and `go run ./cmd/agentcontractgen check` pass.
- [ ] Processed current Firstmate inbox message `006.msg`: its committed correction and normal fast-forward push are pending; it will then be moved to `handled/`.

## Not run by design

- [ ] Red/green source or runtime tests — policy correction and the exhaustive cohort matrix precede production changes.
- [ ] Production implementation/generator/docs edits — pending the cohort matrix and Foundation Atlas classification.
- [ ] Provider-live requests, credentials, reverse-ETL dispatch, or destructive operations — excluded by the task.
- [ ] Exact-SHA Codex review, PR checks remediation, commit, or push — require a valid scoped production candidate.

## Current constraint

The captain has changed the relevant repository policy, not reduced a quality gate. The immediate prerequisites are the committed policy correction, a green active-policy scan, an exhaustive machine-checkable cohort matrix, Atlas classification, and red tests before any production behavior change.
