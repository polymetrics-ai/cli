# Prompts — Issue 599 Connector Boundary Guard

## Kickoff

Task: implement issue #599 `feat(connectorgen): add connector definition boundary guard` in isolated branch `fm/cli-connector-boundary-guard-r1`.

## GSD commands

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt programming-loop init --phase issue-599 --dry-run # returned unknown command in this checkout
scripts/gsd prompt plan-phase issue-599 --skip-research
```

## Manual-GSD fallback

Because `programming-loop` is missing from the repo-local command registry despite a healthy adapter, the universal loop is executed inline using:

- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.pi/prompts/pm-gsd-loop.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Subagent prompt used

Read-only scout:

> Read-only reconnaissance for issue #599 connector boundary guard. Inspect cmd/connectorgen command style, existing tests, Makefile/workflows, and current provider-specific residue in shared production Go. Do not edit. Return concise implementation notes: files to add/edit, command/output conventions, likely baseline exceptions, and pitfalls for false positives. Use only synthetic/no-secret data.

Result: command style is stdlib `run(args, stdout, stderr) int`; scanner should be new package plus `connectorgen boundary`; false positives require AST-only production Go scanning, path classification, and bounded GitHub/current-residue exceptions.
