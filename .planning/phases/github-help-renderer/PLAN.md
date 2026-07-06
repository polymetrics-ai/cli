# GitHub Help Renderer Plan

Issue: #35
Parent issue: #44
Parent PR: https://github.com/polymetrics-ai/cli/pull/49
Branch: `feat/35-github-help-renderer`
Base: `feat/44-github-cli-parity`

## Goal

Render gh-like GitHub connector help from `cli_surface.json` metadata without enabling runtime
command dispatch.

## Scope

- Add a generic renderer for connector CLI surface help.
- Wire engine-backed connector guides to include CLI surface sections when `cli_surface.json` is
  present.
- Add golden tests for GitHub help output and a small synthetic surface.
- Expose CLI surface data to the website generated connector catalog.
- Render a compact command-surface section on connector pages.

## Non-goals

- Do not execute `pm github ...` command aliases.
- Do not add stream-backed or write-backed command dispatch.
- Do not add raw API, GraphQL, generic HTTP, shell, or SQL execution.
- Do not request or use secrets.

## TDD Plan

1. Add a red Go test proving GitHub rendered help contains usage, command groups, ETL, reverse ETL,
   local workflow, JSON flag, and approval notes.
2. Add a red website data test proving GitHub connector JSON includes a CLI surface summary.
3. Implement the smallest renderer and website data path.
4. Refactor only after targeted tests pass.

## Verification

- `go test ./internal/connectors -run CLISurface`
- `go test ./internal/connectors/engine -run CLISurface`
- `go test ./cmd/connectorgen -run CLISurface`
- `pnpm --filter cli-polymetrics-ai test -- connector-data`
- `pnpm --filter cli-polymetrics-ai test`
- `pnpm --filter cli-polymetrics-ai build`
- `go run ./cmd/connectorgen validate internal/connectors/defs/github`
- `git diff --check`

## Human Gates

- New frontend dependencies.
- Runtime command dispatch.
- Auth or token scope changes.
- Generic unrestricted HTTP/write tooling.
- Parent PR merge to `main`.

## Manual GSD Fallback

The local `scripts/programming-loop.mjs` and `scripts/tdd-gate.mjs` helpers are absent in this
worktree. The phase will use the manual GSD loop: plan, red test, green implementation, refactor,
verification, summary, and run-state artifacts.
