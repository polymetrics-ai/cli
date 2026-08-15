# Legacy PR Triage R1

## Task Delivery Header

- Issue: Refs PR #4062, PR #4061, and PR #4071 — legacy MVP pull-request triage.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Each existing pull request is either proven already landed by per-file content evidence and reported for a firstmate close decision, or reconciled onto the stated base with required checks green and an API-observed clean merge state.
- Working branch: Existing heads `feat/3995-certification-shepherd-nm5`, `test/3993-github-live-roundtrip-nm5`, and `fix/4069-flow-case-equivalent-unique-tables-r1`, sequentially; a compliant `fm/<head>-r2` replacement is allowed only if a normal base merge is unworkable.
- Task: Classify every changed file in PRs #4062, #4061, and #4071 by content against the base, preserve all undelivered behavior while reconciling conflicts, repair only each PR's named required-check failure, regenerate invalidated generated artifacts, verify touched packages, and observe the final GitHub API state.
- Verification: Prove every ref with `git rev-parse --verify`; compare every PR file with `git diff --quiet BASE:path PR:path`; run affected package tests with `-timeout 20m` plus repository gates appropriate to each changed set; inspect required checks; read `.base.ref`, `.mergeable`, and `.mergeable_state` from the pulls API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every changed file is content-classified against BASE | live | The trace records `IDENTICAL-ON-BASE`, `DIFFERENT-ON-BASE`, or `ABSENT-ON-BASE` for every API-reported PR file after both refs resolve. |
| Undelivered PR content is non-conflicting and required checks are green | live | GitHub reports the exact base, a clean merge state, and successful required checks after the reconciled branch is pushed. |
| Already-landed substance is not merged or closed by this worker | live | The evidence table shows every non-planning file identical and the status record requests a firstmate decision instead of mutating the PR. |
| No history or branch is destructively rewritten | live | Pushes are normal fast-forward updates of the named PR heads or a newly created replacement branch; no force push, deletion, or protected-branch push is performed. |

## Manual GSD Fallback

- Reason: The canonical phase commands assume one phase/issue branch, while this supervisor task explicitly operates three pre-existing PR branches and forbids creating the standard task branch. The single worker therefore executes the generated lifecycle intent inline and records its evidence here.
- Command resolution: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`; `scripts/gsd sources plan-phase`; `scripts/gsd sources execute-phase`; `scripts/gsd sources verify-work`; `scripts/gsd sources code-review`; `go run ./cmd/agentcontractgen check`.
- Discuss: Scope is locked to PRs #4062, #4061, and #4071 in that order; decisions and conflict rules come from the supervisor brief.
- Plan: Per-PR content classification precedes any mutation. Genuine undelivered content is reconciled by a normal merge, then tested, reviewed, committed, pushed, and API-observed.
- Red: Capture the pre-change content table, merge conflict/check failure, and any failing targeted regression before modifying the branch.
- Green: Preserve both sides and callers, run targeted tests and applicable generators/gates, then observe the required checks and clean merge state after push.
- Refactor: No opportunistic changes; only conflict reconciliation, named check repair, and invalidated generated artifacts are allowed.
- Verify: Re-run scoped tests/gates and compare the final branch to BASE; inspect live pull-request API state rather than assuming it.
- Code review: Review the complete BASE-to-head diff for correctness, security, error handling, test quality, discarded callers, and scope drift before handoff.
- Skills loaded: `github-issue-first-delivery`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, `golang-how-to`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, and `golang-testing`.

## PR Evidence

Per-file classification and terminal-state observations are appended below in PR order.

### PR #4062 — `feat/3995-certification-shepherd-nm5`

- Compared refs: BASE `origin/integration/4015-mvp-flat-r1` at `85ff21487c9a29ac4663cc67f655867630d682f4`; PR head `origin/feat/3995-certification-shepherd-nm5` at `159348ee6b8d9d1a6a802dcbf5c4d50332f4b602`. Both resolved before comparison.
- Classification method: complete API-reported file list from `/repos/polymetrics-ai/cli/pulls/4062/files`, followed by per-file object existence checks and `git diff --quiet "BASE:path" "PRHEAD:path"`.
- Decision: genuine undelivered content. Every non-planning file is different from BASE or absent there, so the branch was reconciled by a normal merge of BASE. No force push, reset, branch deletion, or protected-branch mutation was used.
- Named required check: `.github/workflows/conventions.yml` accepts `^(feat|fix|docs|chore|ci|test|refactor|perf|build|release|revert|deps)/[a-z0-9][a-z0-9._-]*$`; the existing branch name satisfies it. A normal push retriggers the stale failed check.

| Classification | File |
| --- | --- |
| DIFFERENT-ON-BASE | `.agents/agentic-delivery/README.md` |
| DIFFERENT-ON-BASE | `.agents/agentic-delivery/canonical/delivery-contract.json` |
| DIFFERENT-ON-BASE | `.claude/agents/pm-connector-worker.md` |
| DIFFERENT-ON-BASE | `.claude/agents/pm-delivery-worker.md` |
| DIFFERENT-ON-BASE | `.codex/agents/pm-connector-worker.toml` |
| DIFFERENT-ON-BASE | `.codex/agents/pm-delivery-worker.toml` |
| ABSENT-ON-BASE | `.opencode/agents/pm-connector-worker.md` |
| ABSENT-ON-BASE | `.opencode/agents/pm-delivery-worker.md` |
| DIFFERENT-ON-BASE | `.pi/agents/pm-connector-worker.md` |
| DIFFERENT-ON-BASE | `.pi/agents/pm-delivery-worker.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CONTEXT.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CORRECTION-ROUND-1.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CORRECTION-ROUND-2.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CORRECTION-ROUND-3.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CORRECTION-ROUND-4.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/CORRECTION-ROUND-5.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/DISCUSSION-LOG.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/PLAN.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/REVIEW.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/RUN-STATE.json` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/SUMMARY.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/TDD-LEDGER.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/UAT.md` |
| ABSENT-ON-BASE | `.planning/phases/issue-3995-certification-shepherd-gate/VERIFICATION.md` |
| DIFFERENT-ON-BASE | `cmd/agentcontractgen/main.go` |
| DIFFERENT-ON-BASE | `cmd/agentcontractgen/main_test.go` |
| ABSENT-ON-BASE | `internal/agentcontract/certification.go` |
| ABSENT-ON-BASE | `internal/agentcontract/certification_test.go` |
| ABSENT-ON-BASE | `internal/agentcontract/certificationcatalog.go` |
| DIFFERENT-ON-BASE | `internal/agentcontract/check.go` |
| DIFFERENT-ON-BASE | `internal/agentcontract/check_test.go` |
| DIFFERENT-ON-BASE | `internal/agentcontract/contract.go` |
| DIFFERENT-ON-BASE | `internal/agentcontract/contract_test.go` |
| ABSENT-ON-BASE | `internal/agentcontract/opencode.go` |
| ABSENT-ON-BASE | `internal/agentcontract/opencode_test.go` |
| DIFFERENT-ON-BASE | `internal/agentcontract/render.go` |
| ABSENT-ON-BASE | `internal/certificationcatalog/flow.go` |
| ABSENT-ON-BASE | `internal/certificationcatalog/flow_gen.go` |
| ABSENT-ON-BASE | `internal/certificationcatalog/flow_test.go` |

Reconciliation preserves the Shepherd gate while adapting its generated inputs to BASE's definition-owned certification shards. The gate evaluator and certification-catalog producer callers remain present. Verification before push: `go test -timeout 20m ./internal/agentcontract ./internal/certificationcatalog ./cmd/agentcontractgen`; `go vet ./internal/agentcontract ./internal/certificationcatalog ./cmd/agentcontractgen`; `go run ./cmd/agentcontractgen check`; `make agent-contract-check`; `bash scripts/tests/pi-clean-project-agents.sh`; and `go run ./cmd/connectorgen certification-matrix --check` all passed.
