# GitHub CLI feature parity roadmap

Parent issue: https://github.com/polymetrics-ai/cli/issues/44

## Objective

Use the GitHub pilot research to deliver GitHub connector CLI feature parity in small, reviewable,
issue-backed slices. The work should start with docs and metadata foundations, then proceed through
safe execution layers, GraphQL, sensitive/admin policy, and cross-connector rollout learnings.

## Planning model

- Parent issue: #44, the GitHub CLI feature parity epic.
- Parent branch: `feat/44-github-cli-parity`.
- Parent PR: draft PR from `feat/44-github-cli-parity` into `main`.
- Sub-PRs: one PR per sub-issue, each targeting `feat/44-github-cli-parity`.
- Sub-PR body: `Refs #<sub-issue>` and `Refs #44`.
- Parent PR body: `Closes #44` and closes integrated sub-issues when the parent branch lands.
- Human approval: required for the parent PR into `main`; not required for sub-PRs into the parent
  branch when all automated gates pass and no human gate is triggered.

## Milestone roadmap

| Milestone | Purpose | Issues |
| --- | --- | --- |
| Connector CLI Foundation & Guardrails | Agent workflow, issue hierarchy, parent branch rules | #43, #44 |
| GitHub CLI Surface Docs | Validated `cli_surface` metadata and gh-like help output | #34, #35 |
| GitHub Operation Ledger & Direct Reads | Safe execution classification and read execution | #36, #37, #38 |
| GitHub GraphQL Projects & Discussions | Fixed GraphQL operations and command groups not covered by REST | #39, #40 |
| GitHub Sensitive/Admin Actions | Secret, variable, elevated-scope, and admin policy | #41 |
| Cross-Connector Rollout | Convert GitHub learnings into reusable connector rollout templates | #42 |

## Sub-issue plan

| Issue | Branch | PR base | Primary agent | Intent |
| --- | --- | --- | --- | --- |
| #34 | `feat/34-cli-surface-metadata` | `feat/44-github-cli-parity` | provider-cli-surface-agent | Promote GitHub `cli_surface` metadata into validated production form. |
| #35 | `feat/35-github-help-renderer` | `feat/44-github-cli-parity` | help-renderer-agent | Render gh-like GitHub connector help from metadata. |
| #36 | `feat/36-github-stream-runner` | `feat/44-github-cli-parity` | issue-first-implementation-agent | Execute stream-backed commands through generic connector runner. |
| #37 | `feat/37-github-operation-ledger` | `feat/44-github-cli-parity` | operation-ledger-architect | Reclassify GitHub REST rows into execution models. |
| #38 | `feat/38-github-direct-read` | `feat/44-github-cli-parity` | direct-read-executor-agent | Add constrained direct-read execution for safe GitHub operations. |
| #39 | `feat/39-github-graphql-engine` | `feat/44-github-cli-parity` | graphql-engine-agent | Add declarative GraphQL support for fixed operations. |
| #40 | `feat/40-github-projects-discussions` | `feat/44-github-cli-parity` | provider-cli-surface-agent | Model GitHub Projects and Discussions command groups. |
| #41 | `feat/41-github-sensitive-admin` | `feat/44-github-cli-parity` | sensitive-admin-policy-agent | Design sensitive/admin reverse ETL policy. |
| #42 | `docs/42-cross-connector-rollout` | `feat/44-github-cli-parity` | cross-connector-rollout-agent | Generalize GitHub learnings to other connectors. |

## Dependency order

1. #43 must be merged before using the automated issue-first workflow for sub-PRs.
2. #34 precedes #35 because the renderer needs validated metadata.
3. #37 precedes #38 because direct reads need the operation-ledger classification.
4. #39 precedes the executable parts of #40 when Projects/Discussions require GraphQL.
5. #41 is gated by security policy review before any sensitive/admin execution is exposed.
6. #42 starts after at least one GitHub implementation slice has landed in the parent branch.

## Sub-PR merge policy

Agents may merge sub-PRs into `feat/44-github-cli-parity` without human approval only when:

- the PR targets the parent branch, not `main`
- the PR references exactly one sub-issue and #44
- all local and remote checks pass
- CodeRabbit review has no unresolved actionable findings
- no hard stop is triggered
- the diff stays inside the sub-issue scope

The parent PR into `main` must remain human-approved.

## Source anchors

See `.agents/agentic-delivery/references/issue-roadmap-best-practices.md`.
