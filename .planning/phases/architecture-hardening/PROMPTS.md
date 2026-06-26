# Prompts

## User Request

Implement all seven architecture-hardening changes using the GSD programming loop, multiple subagents, tests, and push the changes.

## Subagent Workstreams

- Runtime Module: implement `internal/runtime` only, with Dragonfly/Postgres adapters over existing dependencies and targeted tests.
- State Module: implement `internal/state` only, with locked atomic JSON store and targeted tests.
- HTTP Source Template: implement `internal/connectors/httpsource` only, with fixture/base URL/auth/pagination/read-only behavior and targeted tests.

## Main Orchestration

- Coordinate shared app, CLI, registry, connector, and docs changes.
- Use TDD vertical slices for reverse ETL, validation, path policy, materialization, and read limits.
- Run local gates before commit and push.
