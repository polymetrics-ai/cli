# Verification: Issue #4283 Batch-1 source-rigidity R2

## Completed planning checks

- [x] Fetched `origin/fm/cli-top100-declaration-batch-r1` and checked out SHA `6ec34964fda7a78a1736a5cd8933a46418346261` detached in the assigned task worktree.
- [x] Verified the required ancestor with `git merge-base --is-ancestor`.
- [x] Verified PR #4294 API base/head/SHA: `main`, `fm/cli-top100-declaration-batch-r1`, `6ec34964fda7a78a1736a5cd8933a46418346261`.
- [x] Ran `scripts/gsd doctor`, resolved all required lifecycle command sources, generated `discuss-phase` and `plan-phase --tdd` prompts, and ran `go run ./cmd/agentcontractgen check`.
- [x] Read the required routing, GSD adapter, issue contract, CLI parity, connector lane, and source-rigidity audit materials.
- [x] Processed Firstmate inbox messages `001.msg` and `002.msg` as historical scope context.
- [x] Captured RED policy evidence: the three authoritative direct clauses in required-skills routing, task-skill matrix, and issue-agent contract had the obsolete mandatory split.
- [x] Captured GREEN policy evidence: the correct `focused-policy-regression` passes with the three authoritative paths as separate positional arguments (`exit_status=0`); `AGENTS.md` has no direct obsolete clause; all 17 non-authoritative copies were restored exactly; YAML/JSON parsing and `go run ./cmd/agentcontractgen check` pass (`exit_status=0`). The earlier space-joined-path invocation is excluded from evidence. The repository has no dedicated delivery-policy regression harness.
- [x] Processed Firstmate inbox message `006.msg`: commit `924197b49adc516ec7bdb3b6d928c9701cfab946` was normal-fast-forward pushed and remote verified, then `006.msg` was moved to `handled/`.
- [x] Processed Firstmate inbox messages `007.msg` through `010.msg`: narrowed policy scope to the three named direct clauses, completed focused validation, and moved each message to `handled/`. Message `010.msg` additionally prohibits installing ad hoc tooling or scripts.
- [x] Committed immutable Batch-1 discovery denominator: `data/connector-canon/batch1-source-rigidity-r2-cohort-ledger.json` pins all ten source locks, schema forms, identity locations, source-lock SHA-256 values, and the 4,341-identity total. It expressly has zero projected cells and makes no reachability claim.

## Not run by design

- [ ] Red/green source or runtime tests — policy correction and the exhaustive cohort matrix precede production changes.
- [ ] Production implementation/generator/docs edits — pending the cohort matrix and Foundation Atlas classification.
- [ ] Provider-live requests, credentials, reverse-ETL dispatch, or destructive operations — excluded by the task.
- [ ] Exact-SHA Codex review, PR checks remediation, commit, or push — require a valid scoped production candidate.

## Current constraint

The captain has changed the relevant repository policy, not reduced a quality gate. The immediate prerequisites are the committed policy correction, a green active-policy scan, an exhaustive machine-checkable cohort matrix, Atlas classification, and red tests before any production behavior change.
