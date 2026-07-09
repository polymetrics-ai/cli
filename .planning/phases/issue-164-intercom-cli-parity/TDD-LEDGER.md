# TDD Ledger: Intercom CLI Parity Parent Orchestration

## Red / Planning Evidence

- Parent PR lookup:
  - Command: `gh pr list --head feat/164-intercom-cli-parity --base main --json number,title,state,isDraft,url,headRefName,baseRefName,mergeable,reviewDecision,statusCheckRollup`
  - Result: `[]`; parent PR is missing and must be created before stacked sub-issue completion.
- GSD programming-loop attempt:
  - Command: `scripts/gsd prompt programming-loop init --phase issue-164-intercom-cli-parity --dry-run`
  - Result: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback recorded.

## Green / Execution Evidence

- Plan checkpoint commit:
  - Commit: `3afecf62 chore(intercom): plan CLI parity orchestration`
  - Push: `git push -u origin feat/164-intercom-cli-parity`
  - Draft parent PR opened: https://github.com/polymetrics-ai/cli/pull/220

Each sub-issue must add its own red/green/refactor entries before production edits.

## Refactor Evidence

Pending.
