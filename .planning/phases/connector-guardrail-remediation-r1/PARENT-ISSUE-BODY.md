## Objective

Deliver a forward-only connector path-ownership guardrail remediation program before any additional connector campaign PR merges. The program must add a target-aware changed-path ownership guard, wire it into local and remote merge gates, update PM/no-mistakes connector-lane guidance, and record/remediate every first-eight connector campaign finding without rewriting `main` history.

## Background

Captain authorization requires this parent issue, linked sub-issues, stacked sub-PRs into a parent branch, a draft parent PR to `main`, and final no-mistakes validation before Firstmate performs any authorized final merge. The first-eight connector guardrail audit showed the existing `connector-boundary` check scanned connector-specific lexemes in shared Go but did not enforce target connector path ownership.

Audit scope:

- HubSpot: https://github.com/polymetrics-ai/cli/pull/3529
- Stripe: https://github.com/polymetrics-ai/cli/pull/3530
- Bitbucket: https://github.com/polymetrics-ai/cli/pull/3531
- Zendesk Support: https://github.com/polymetrics-ai/cli/pull/3532
- Google Ads: https://github.com/polymetrics-ai/cli/pull/3535
- Freshchat: https://github.com/polymetrics-ai/cli/pull/3536
- Xero: https://github.com/polymetrics-ai/cli/pull/3537
- Asana: https://github.com/polymetrics-ai/cli/pull/3538

## Roadmap shape

- Parent issue: this issue
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580 (draft)
- Final target branch: `main`
- Milestones: connector path ownership guardrail remediation
- Project: not added unless project scope is already available
- Orchestrator: active PM parent issue orchestrator via `/pm-orchestrate` / Pi project agents; fallback must record exact reason
- Orchestration workflow: `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- Stacked PR workflow: `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- Automated review routing: `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`

## Sub-issues

| Issue | Milestone | Branch | PR base | Intent | Dependencies | Status |
| --- | --- | --- | --- | --- | --- | --- |
| #3595 | icon registry foundation | `fix/3595-icon-registry-single-source` | `fix/3579-connector-path-ownership-guardrails` | Canonical bare connector icon registry and docs-icon asset authority | none | Draft sub-PR #3596 open; next critical path |
| #3581 | guardrail core | `fix/3581-target-scope-core-validator` | `fix/3579-connector-path-ownership-guardrails` | Target-scope contract and core changed-path validator | #3595 before R5/R6 reconciliation | Sub-PR #3590 open; integration blocked on #3595 |
| #3582 | guardrail gate | `fix/3582-connector-ownership-ci-gate` | `fix/3579-connector-path-ownership-guardrails` | GitHub Actions, label/tag, local hook, and required remote gate | #3581 | Dependency-blocked |
| #3583 | PM/no-mistakes | `fix/3583-pm-no-mistakes-connector-lane` | `fix/3579-connector-path-ownership-guardrails` | PM orchestrator, issue/PR templates, connector instructions, no-mistakes guidance | none | Provisionally integrated via #3588; parent review/final gate pending |
| #3584 | forward remediation A | `fix/3584-hubspot-bitbucket-forward-remediation` | `fix/3579-connector-path-ownership-guardrails` | HubSpot and Bitbucket shared path dispositions | none | Sub-PR #3591 open; fresh no-mistakes recovery pending |
| #3585 | forward remediation B | `fix/3585-shared-engine-runner-remediation` | `fix/3579-connector-path-ownership-guardrails` | Stripe, Freshchat, and Google Ads shared engine/runner/connectorgen dispositions | none for ledger-only fallback | Sub-PR #3593 open |
| #3586 | generated remediation | `fix/3586-generated-unrelated-connector-remediation` | `fix/3579-connector-path-ownership-guardrails` | Zendesk Support and Google Ads unrelated-connector/generated remediation | none | Sub-PR #3589 open |
| #3587 | audit ledger/proof | `fix/3587-first-eight-audit-ledger-proof` | `fix/3579-connector-path-ownership-guardrails` | Historical audit ledger and end-to-end enforcement proof for all eight first merges | #3581, #3582, #3584, #3585, #3586 | Dependency-blocked |

## Orchestration state

State ledger lives in the parent branch under `.planning/phases/connector-guardrail-remediation-r1/RUN-STATE.json` and must follow `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`.

| Issue | Worker | Branch | PR | Latest SHA | Verification | Automated review coverage | Merge state | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| #3595 | pm-gsd-worker | `fix/3595-icon-registry-single-source` | #3596 | `b814e85a6` | Planning scaffold pushed; implementation and 5.6 SOL validation pending | pending | Draft sub-PR open; next critical path | None |
| #3581 | pm-gsd-worker | `fix/3581-target-scope-core-validator` | #3590 | pending after #3595 | Prior-config no-mistakes run parked; fresh 5.6 SOL validation required after #3595 reconciliation | blocked | Sub-PR open; blocked | #3595 canonical icon registry foundation |
| #3582 | deferred | `fix/3582-connector-ownership-ci-gate` | pending | pending | Pending | pending | Planned | Blocks on #3581 |
| #3583 | pm-gsd-worker | `fix/3583-pm-no-mistakes-connector-lane` | #3588 | `0c321595d7ae4852550a5012a895c3e11f7e8298` / parent `86b91fc40f46b8653538531fc40c183913676f05` | no-mistakes `01KZ0SEAKBB9TG7N3SMG97XKJS` passed | parent PR fallback/human final gate pending | Provisionally integrated | Parent PR final review/readiness pending |
| #3584 | pm-gsd-worker | `fix/3584-hubspot-bitbucket-forward-remediation` | #3591 | pending current validation | Checks green; fresh no-mistakes recovery pending | pending | Sub-PR open | Recovery pending |
| #3585 | pm-gsd-worker | `fix/3585-shared-engine-runner-remediation` | #3593 | pending current validation | Checks green; merge arbitration pending | pending | Sub-PR open | None |
| #3586 | pm-gsd-worker | `fix/3586-generated-unrelated-connector-remediation` | #3589 | pending current validation | Checks green; merge arbitration pending | pending | Sub-PR open | None |
| #3587 | deferred | `fix/3587-first-eight-audit-ledger-proof` | pending | pending | Pending | pending | Planned | Blocks on #3581/#3582 and remediation slices |

## Branch and PR policy

- Parent branch starts from current `main` and is named `fix/<parent-issue-number>-connector-path-ownership-guardrails`.
- Parent PR targets `main`, is opened as draft before sub-issue implementation, and remains draft until all required sub-issues are integrated and final verification is ready.
- Parent PR uses `Refs #<parent-issue>` while incomplete. Firstmate owns final readiness/no-mistakes and final merge.
- Sub-issue branches start from the parent branch and use `fix/<sub-issue-number>-<slug>`.
- Sub-PRs target the parent branch and use `Refs #<sub-issue>` and `Refs #<parent-issue>`.
- Sub-PRs do not use closing keywords because they do not target the default branch.
- Sub-PRs may be merged into the parent branch only after scoped verification, no-mistakes evidence, required CI, changed-path compliance, and automated review coverage are satisfied.
- The parent PR into `main` is human-gated and must not be merged by an agent.

## Human gates

- Parent PR merge to `main`
- Auth scope changes or `gh auth refresh`
- Secret handling changes or credentialed connector checks
- New dependencies
- Destructive external actions or production deploys
- Quality gate reductions
- Generic shell, unrestricted HTTP write, unrestricted SQL write, or unrestricted raw API tooling
- Reverse ETL execution outside plan → preview → approval → execute
- Branch protection/ruleset permission blocker if GitHub denies required-check configuration

## Acceptance criteria

- One public parent issue, linked public sub-issues, one parent branch, and one draft parent PR to `main` exist before sub-issue coding.
- Parent PR and issue expose dependency graph, disjoint write scopes, worker ownership, integrated sub-PRs, review coverage, and every first-eight disposition.
- Connector implementation diffs are automatically detected even when label/tag/scope presentation is omitted or changed.
- Exactly one target connector is declared and validated for connector implementation PRs.
- Target definitions, connector-owned fixtures/tests, target generated docs, and necessary shared indexes/goldens are allowed only under narrow tested rules.
- Shared runtime/tooling, unrelated connectors, unrelated docs, and unrelated generated churn fail unless moved to a separately reviewed foundation PR.
- Connector-lane exception editing cannot silently weaken the gate.
- Local hook and GitHub workflow call the same validator. Local bypass does not permit remote merge because required GitHub check remains red.
- GitHub ruleset or branch protection requires the guard check for `main` merges, with read-back evidence from GitHub.
- Every audit finding for all eight prior merges has a recorded forward disposition and evidence; no history rewrite, force-push, or blanket revert.
- Every implementation sub-issue uses GSD/TDD, required skills, scoped no-mistakes validation, CI, review disposition, and worker handoff.
- Final parent branch is synchronized with current `main`, full repository verification passes, final no-mistakes validation is current-code-matched, required CI is green, and the draft parent PR is ready for Firstmate.

## Verification

Required local gates before parent readiness:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Focused guardrail gates:

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership --help
go run ./cmd/connectorgen ownership . --base origin/main --scope-file <fixture>
make connector-boundary
```

GitHub/read-back evidence:

```bash
gh-axi pr checks <parent-pr>
gh-axi api /repos/polymetrics-ai/cli/rulesets
gh-axi api /repos/polymetrics-ai/cli/branches/main/protection
```

## Sources

- Captain authorization: local decision record supplied to orchestrator
- First-eight audit evidence: local report supplied to orchestrator
- `AGENTS.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`
