# TDD Ledger

## Planned Red Evidence

- `go test ./internal/connectors -run CLISurface`
  - Adjusted to `go test ./internal/connectors/bundleregistry -run CLISurface` because GitHub is
    registered through the bundle registry, not the built-in registry used by base connector tests.
  - Initial failure: GitHub manual missing `COMMAND SURFACE`.
- `pnpm --filter cli-polymetrics-ai test -- connector-data`
  - Initial failure: generated GitHub connector data does not expose `cliSurface`.

## Green Evidence

- `go test ./internal/connectors/bundleregistry -run CLISurface`
  - Result: passed.
  - Evidence: GitHub connector manual now includes `COMMAND SURFACE`, usage, grouped commands,
    ETL stream mappings, reverse ETL write mappings, local workflow exclusions, JSON flag, and
    approval notes.
- `pnpm --filter cli-polymetrics-ai test -- connector-data`
  - Result: passed.
  - Evidence: generated connector data route returns GitHub `cliSurface`, and the existing docs
    smoke suite still passes.

## Refactor Evidence

Broader targeted verification passed:

- `go test ./internal/connectors/engine -run CLISurface`
- `go test ./cmd/connectorgen -run CLISurface`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill`
- `pnpm --filter cli-polymetrics-ai build`
- `git diff --check`
