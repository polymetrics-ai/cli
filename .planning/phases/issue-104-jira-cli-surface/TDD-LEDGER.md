# TDD Ledger: Issue #104 Jira CLI Surface Metadata

## Preflight

- `scripts/gsd prompt plan-phase issue-104-jira-cli-surface --skip-research`: generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-104-jira-cli-surface --dry-run`: failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active.

## Planned Red

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Expected first failure: `Jira CLISurface is nil; defs.FS must embed cli_surface.json`.

## Red Evidence

Pending.

## Green Evidence

Pending.

## Refactor Evidence

Pending.

## Notes

- The first production edit must be the failing test, not `cli_surface.json`.
- Implementation remains metadata-only; no Jira credentials or live API checks are required.
