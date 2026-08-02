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
- Parent PR: `pending; draft PR to main required before sub-issue implementation`
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
| pending | guardrail core | pending | `fix/3579-connector-path-ownership-guardrails` | Target-scope contract and core changed-path validator | none | Planned |
| pending | guardrail gate | pending | `fix/3579-connector-path-ownership-guardrails` | GitHub Actions, label/tag, local hook, and required remote gate | core validator | Planned |
| pending | PM/no-mistakes | pending | `fix/3579-connector-path-ownership-guardrails` | PM orchestrator, issue/PR templates, connector instructions, no-mistakes guidance | core validator | Planned |
| pending | forward remediation A | pending | `fix/3579-connector-path-ownership-guardrails` | HubSpot and Bitbucket shared path dispositions | core validator | Planned |
| pending | forward remediation B | pending | `fix/3579-connector-path-ownership-guardrails` | Stripe, Freshchat, and Google Ads shared engine/runner/connectorgen dispositions | core validator | Planned |
| pending | generated remediation | pending | `fix/3579-connector-path-ownership-guardrails` | Zendesk Support and Google Ads unrelated-connector/generated remediation | core validator | Planned |
| pending | audit ledger/proof | pending | `fix/3579-connector-path-ownership-guardrails` | Historical audit ledger and end-to-end enforcement proof for all eight first merges | core + gate | Planned |

## Orchestration state

State ledger lives in the parent branch under `.planning/phases/connector-guardrail-remediation-r1/RUN-STATE.json` and must follow `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`.

| Issue | Worker | Branch | PR | Latest SHA | Verification | Automated review coverage | Merge state | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| pending | pending | pending | pending | pending | Pending | pending | Planned | None |

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
