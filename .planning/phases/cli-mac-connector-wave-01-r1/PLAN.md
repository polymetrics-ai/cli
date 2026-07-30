# Connector Wave 01 Parent Orchestration

This document is the tracked parent scaffold for the `feat/connector-wave-01` integration branch. It coordinates child work only; it does not implement connector gaps or certify provider behavior.

## GSD / skills evidence

- GSD adapter: `scripts/gsd doctor` passed in this worktree.
- Required programming loop: `scripts/gsd prompt programming-loop init --phase cli-mac-connector-wave-01-r1 --dry-run` was attempted; this adapter returned `unknown GSD command: programming-loop`, so this docs-only scaffold uses the documented manual fallback via `scripts/gsd prompt quick "connector wave 01 parent orchestration scaffold"`.
- Skills loaded for this parent scaffold: `gsd-core`, `no-mistakes`, `golang-documentation`, and connector/Go safety routing (`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`). Child implementation agents must reload the task-specific skills required by their issue.

## Branch and PR model

- Integration branch: `feat/connector-wave-01`, created from fresh `origin/main`.
- Parent PR base: `main`; keep the parent PR draft until all required children land and full integrated validation is green.
- Child PR base: `feat/connector-wave-01`; child PR bodies use `Refs`, not closing keywords, for their sub-issues and connector parents.
- Firstmate may merge validated child PRs into `feat/connector-wave-01` only after no-mistakes, required local checks, GitHub checks, and automated-review disposition are satisfied.
- Only the captain may merge the parent PR to `main`. No force-pushes, destructive resets, discarded unique work, provider credentials, live writes, false certification, or no-mistakes bypasses are authorized.

## Serialized foundations

No connector starts until these land into the integration branch in order:

1. [#2985 — typed bounded provider search/query capabilities](https://github.com/polymetrics-ai/cli/issues/2985)
2. [#2986 — truthful CDC capability discovery](https://github.com/polymetrics-ai/cli/issues/2986)
3. [#2987 — typed bounded source query/search surfaces](https://github.com/polymetrics-ai/cli/issues/2987)
4. [#2988 — database CDC state model and synthetic certification lab harness](https://github.com/polymetrics-ai/cli/issues/2988)

## Connector cohorts

Phase 1 correction cohort, after all foundations land:

- [GitHub parent #2989](https://github.com/polymetrics-ai/cli/issues/2989)
- [Gong parent #2997](https://github.com/polymetrics-ai/cli/issues/2997)
- [Stripe parent #3005](https://github.com/polymetrics-ai/cli/issues/3005)
- [Zendesk Support parent #156](https://github.com/polymetrics-ai/cli/issues/156)
- [Intercom parent #164](https://github.com/polymetrics-ai/cli/issues/164)

Phase 2 starts only after the phase 1 PRs are integrated. Within phase 2, finish the correction PRs before beginning new zero-operation connectors:

1. [Jira parent #81](https://github.com/polymetrics-ai/cli/issues/81)
2. [Linear parent #80](https://github.com/polymetrics-ai/cli/issues/80)
3. [HubSpot parent #132](https://github.com/polymetrics-ai/cli/issues/132)
4. [Shopify parent #3013](https://github.com/polymetrics-ai/cli/issues/3013)
5. [Bitbucket parent #79](https://github.com/polymetrics-ai/cli/issues/79)

Before a child worker starts, re-read that parent issue, its native GitHub sub-issues, and the current dispatch manifest; GitHub remains the canonical source for current labels, state, and child relationships.

## Child-base mechanism record

Current help inspection found this exact support:

- `gh-axi pr create` supports `--base`, so GitHub can create a child PR targeting `feat/connector-wave-01`.
- `no-mistakes axi run --help` exposes only `--intent`, `--skip`, and `--yes`; it does not advertise a non-default PR base flag.

Therefore the first foundation child must prove the supported no-mistakes path for a non-default-base PR before stacked connector work proceeds. Acceptable proof is a no-mistakes-green child PR whose base is `feat/connector-wave-01` without a duplicate/wrong-base main PR. If that cannot be proven, stop and report a blocker instead of faking stacked PRs.

## Fixture, certification, and safety truth

- Fixture replay and synthetic labs do not imply live provider certification.
- Certification can be claimed only by a connector-owned certification record and matching no-mistakes/GitHub evidence.
- No child may use provider credentials or live provider writes unless explicitly authorized by a separate issue/human gate.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.
- Do not add generic shell, generic HTTP write, generic SQL write, arbitrary GraphQL, or raw passthrough escape hatches.

## TDD ledger for this scaffold

- Red evidence: not applicable; this commit is a docs-only orchestration scaffold with no production behavior change.
- Green evidence before commit: fresh `origin/main` matched local `HEAD`; GitHub issues, native sub-issue links, duplicate branches/PRs, repository default branch, and branch protection were revalidated with `gh-axi` reads.
- Refactor evidence: not applicable.

## Verification checklist

Initial parent setup:

- `pwd -P` and `git rev-parse --show-toplevel` resolve to the disposable worktree.
- `git fetch origin main --prune`; local branch starts from current `origin/main`.
- `gh-axi repo view polymetrics-ai/cli` confirms default branch `main`.
- `gh-axi api /repos/polymetrics-ai/cli/branches/main` confirms the current protected main SHA.
- `gh-axi issue view` / `gh-axi issue subissue list` revalidate the foundations, connector parents, and native children.
- `gh-axi pr list` / `gh-axi search prs` / branch API reads find no conflicting `feat/connector-wave-01` branch or parent PR.
- `no-mistakes doctor` passes; `no-mistakes axi run --intent ...` validates, pushes, opens the parent PR, and reaches CI-ready green.
- If no-mistakes opens the parent PR ready, convert it to draft with `gh-axi` before releasing child workers.

Final parent readiness after Firstmate integrates every required child:

- Update `feat/connector-wave-01` from current `origin/main` through no-mistakes-supported sync/rebase paths only.
- Re-run complete integrated validation through no-mistakes.
- Ensure automated review coverage and dispositions exist for each integrated child range.
- Mark the parent PR ready for captain review only after checks are green.
- Do not merge the parent PR to `main`.

Wave 02 must not begin until the captain explicitly merges the wave 01 main PR and local work is refreshed from `origin/main`.
