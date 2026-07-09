# TDD Ledger: Issue #105 Jira Help Renderer / Docs

## Preflight

- `scripts/gsd prompt plan-phase issue-105-jira-help-renderer --skip-research`: generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-105-jira-help-renderer --dry-run`: failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active.

## Planned Red

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
cd website && pnpm test:unit -- connector-data
```

Expected first failures:

- Runtime CLI does not render connector help for `pm jira --help` or bare `pm jira`.
- Website data route does not yet assert Jira `cliSurface` metadata.

## Red Evidence

Pending.

## Green Evidence

Pending.

## Refactor Evidence

Pending.
