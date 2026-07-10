# Summary: GitHub CLI Parity Parent Orchestration

## Current State

- Parent branch `feat/44-github-cli-parity` is locally rebased onto `origin/main`.
- #51 parent orchestrator rules are retained from `main`; older duplicate workflow edits from #34
  were not reintroduced.
- GitHub CLI surface metadata, validation, website docs, and #34 phase artifacts remain in the
  parent branch.
- GSD model policy now records Codex `gpt-5.5` with `xhigh` reasoning effort.
- Claude review fixes are applied locally for PR #49: stale parent-review fallback wording,
  agent-task grouping, raw CLI-surface bytes, CLI command stream/write ambiguity, agent matrix
  review-disposition coverage, issue-first parent PR blocker wording, and catalog JSON
  serialization.
- #69 was green and merged into the parent branch as the #57 reverse-ETL command runner slice.
- Parent orchestration is now documented as active ownership: the orchestrator must build a ready
  queue, spawn every independent ready worker when subagent tooling exists, or record a blocker.
- Added cross-runtime GSD adapters for Codex and OpenCode plus a repo-local `caveman` skill for
  compact status, prompts, and handoffs.

## Next

Push the parent branch, wait for automatic Claude review on PR #49, resolve any actionable
comments with dispositions, then continue #44 subissues through the active orchestrator.
