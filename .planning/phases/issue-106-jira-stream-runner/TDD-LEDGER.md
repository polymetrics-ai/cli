# TDD Ledger: Issue #106 Jira Stream Runner

## Preflight

```bash
scripts/gsd prompt plan-phase issue-106-jira-stream-runner --skip-research
scripts/gsd prompt programming-loop init --phase issue-106-jira-stream-runner --dry-run
```

Result: plan prompt generated; programming-loop command unavailable with `unknown GSD command: programming-loop`; manual fallback active.

## Planned red tests

```bash
go test ./internal/cli -run 'TestJiraCommandSurfaceRunsStreamBacked' -count=1
```

Expected first failures:

- `pm jira issue list --jql ...` rejects `--jql` because Jira stream command flags are not declared.
- `pm jira project list --query ...` rejects `--query` because Jira stream command flags are not declared.
- `pm jira user list --query ...` rejects `--query` because Jira stream command flags are not declared.

## Green evidence

Pending.

## Refactor evidence

Pending.
