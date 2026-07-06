# TDD Ledger

## Planned Red Evidence

- `go test ./internal/connectors -run CLISurface`
  - Adjusted to `go test ./internal/connectors/bundleregistry -run CLISurface` because GitHub is
    registered through the bundle registry, not the built-in registry used by base connector tests.
  - Initial failure: GitHub manual missing `COMMAND SURFACE`.
- `pnpm --filter cli-polymetrics-ai test -- connector-data`
  - Initial failure: generated GitHub connector data does not expose `cliSurface`.

## Green Evidence

Pending.

## Refactor Evidence

Pending.
