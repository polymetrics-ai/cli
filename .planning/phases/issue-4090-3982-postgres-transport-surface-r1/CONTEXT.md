# Context — PostgreSQL transport surface truthfulness (#4090, #3982)

## Task Delivery Header

- Issues: Refs #4090, Refs #3982
- Base branch: `integration/4015-mvp-flat-r1` at `13fcc9d5dec09b4f631b049ceb4f21b685d9c486`
- Merge target: `integration/4015-mvp-flat-r1` → `main`
- Working branch: `fm/cli-pg-declared-surface-truthfulness-r1`
- Delivery: direct PR; push only the working branch, API-read the observed PR base, and never run `no-mistakes`.
- Target connector: `postgres` only. Changed production paths are limited to its bundle; registry proof lives in `internal/app` and lifecycle evidence lives in this directory.

The task-delivery-header template named by `AGENTS.md` is absent on this integration base. This records the repository's documented manual fallback fields.

## Evidence-led decisions

- The `full_overwrite` reduction in PR #4167 was not intentional: #4090 requires full modes, and the PR's own plan calls for full-mode PostgreSQL source behavior without documenting a scope reduction. PR #4175, included in the required base, has already restored `full_overwrite`; `pm connectors inspect postgres --json` and the regenerated PostgreSQL certification shard agree.
- Do not revert #4175's reachable `incremental_append` and `incremental_dedupe` source/destination modes. `app.Open` constructs the real registry and `Registry.Preflight` admits these exact matched declarations without source or destination I/O.
- Remove only `incremental_dedupe_history` from the PostgreSQL **source** declaration. No shipped destination advertises that mode, so current preflight necessarily refuses it after a misleading source declaration. This is the smallest honest reduction and does not flip `write:true`.
- The certification shard is derived only through `go run ./cmd/connectorgen certification-matrix --connector postgres`; it is never hand-edited.

## Lifecycle and skills

`scripts/gsd doctor`, sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`, and `go run ./cmd/agentcontractgen check` passed. The generated prompts are executed inline: the delivery contract is a single-worker direct PR and this runtime may not spawn lifecycle roles.

Loaded for this Go/PostgreSQL/CLI surface work: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-database`; the CLI parity and runtime/PostgreSQL integration references were read.

CLI parity is limited to existing `pm connectors inspect postgres --json` metadata. No command, flag, help text, manual, or website page changes; the focused inspection test and live command provide the applicable evidence.
