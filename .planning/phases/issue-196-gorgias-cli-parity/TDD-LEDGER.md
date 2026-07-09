# TDD Ledger: Gorgias CLI Parity Parent Orchestration

## Mode

Manual GSD universal runtime fallback. `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry (`unknown GSD command: programming-loop`), so the live loop is: plan → red validation/test → minimal implementation → green verification → refactor → commit/push.

## Parent planning red/validation evidence

- Parent PR did not exist yet for `feat/196-gorgias-cli-parity` during initial planning.
  - Command: `gh pr list --head feat/196-gorgias-cli-parity --json number,title,state,isDraft,baseRefName,headRefName,url,body`
  - Initial result: `[]`.
  - Green follow-up: draft parent PR opened at https://github.com/polymetrics-ai/cli/pull/229.
- Parallel worker spawn is unavailable in this Pi harness.
  - Evidence: available tools are `read`, `bash`, `edit`, `write`; no `subagent` tool is available.
  - Decision: run #197 as `local_critical_path` after parent seed.

## Red tests planned for #197

Before production metadata edits on #197, add or run validation that fails against the current Gorgias baseline:

```bash
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

Expected issue-specific red condition: current `internal/connectors/defs/gorgias/` does not account for the 114-operation official baseline in API/CLI surface metadata.

## Green evidence

Pending.

## Refactor evidence

Pending.
