# PRD — Wave 0: GitHub Native Package + Data-Driven Registry

## Problem
The connector platform is moving to **one Go package per system, named by the bare slug**
(`github`, not `source-github`), each exposing the unified ruby-style contract
(check/catalog/read/write/query, +cdc). Today every connector lives flat in `package connectors`,
and `NewRegistry()` (internal/connectors/connectors.go) hand-registers built-ins in a switch.
GitHub is the only real, live connector and must become the **reference** per-system package that
every future connector (Wave 1–5) is modeled on.

## Goal
Migrate the existing GitHub connector (github.go, github_auth.go, github_streams.go) into
`internal/connectors/github/` as a self-contained package implementing the `connectors.Connector`
contract (and optional WriteValidator/DryRunWriter as today), and replace the hand-switch in
`NewRegistry()` with a **data-driven self-registration** mechanism (`registry_gen.go`) so adding a
connector package is a one-line import, not a switch edit.

## Non-Goals
- DuckDB / query-engine work (separate, human-gated phase — dependency + CGO).
- Migrating other connectors or rewriting `catalog_data.json` pair-merges (later waves).
- Changing the `connectors.Connector` interface (it is correct).

## Users
- The `pm` CLI and `internal/app` orchestration (consume the registry).
- LLM agents (consume `pm connectors inspect github --json`).
- Future connector authors (copy the github package as the template).

## Success Metrics
- `internal/connectors/github/` builds and exposes `github.New() connectors.Connector`.
- Registry resolves both `github` and `source-github` to the live connector (identities preserved).
- All existing tests + `make verify` stay green; GitHub conformance unchanged.
- `NewRegistry()` no longer hard-codes GitHub; it is registered via the new mechanism.

## Constraints
- No new third-party dependencies.
- No secret values requested/printed; reverse-ETL stays plan→preview→approve→execute.
- Keep `make verify` green at every gate.
