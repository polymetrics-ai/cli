# #4093 — definition-owned production transports

## Task delivery header

- **Issue:** `Closes #4093 — feat(synctransport): load definition-owned transports and register production transports`
- **Base branch:** `integration/4015-mvp-flat-r1`
- **Working branch:** `fm/cli-4093-transport-registration-r1`
- **Delivery:** direct PR to `integration/4015-mvp-flat-r1`; never merge or push `main`.
- **Isolation:** verified with `pwd -P` and `git rev-parse --show-toplevel` at the disposable Treehouse worktree.
- **Base admission:** the prescribed rebase encountered only a CI/toolchain conflict, was aborted, and the branch was recreated directly from `origin/integration/4015-mvp-flat-r1`.

The task-delivery-header template named in `AGENTS.md` is not present on this
base branch. This header retains its required facts without inventing a
replacement contract. Likewise, the brief-named
`data/cli-mvp-verify-certification-r1/report.md` is absent; the canonical
#4093 issue body and the landed #4081/#4090 phase evidence are the design
inputs used here.

## Locked implementation decisions

1. `sync_transport.json` is optional, versioned at schema version 1, and
   rejected by JSON Schema, strict decoding, and the existing closed Go
   descriptor validation. It is a declaration only: loading it never registers
   an executor.
2. The engine adds the descriptor to `Bundle`, deep-clones it into
   `Definition`, and does not let callers mutate the loaded definition or later
   projections.
3. A new transport definition composer first enumerates and validates every
   declared role, connector-family pairing, known executor factory, and
   external evidence admission. Only after that complete plan succeeds can it
   call `RegisterSource` or `RegisterDestination`; an unknown or malformed
   declaration therefore has zero registration side effects.
4. Factories are indexed only by exact `family` plus `id`; the App composition
   root contains no connector-name branch. A factory may be a named hook for a
   typed provider adapter, but no generic HTTP, SQL, or shell capability is
   created.
5. GitHub declares both production roles in its own definition directory;
   PostgreSQL declares its native snapshot source in its own definition
   directory. The old wrapper descriptor and PostgreSQL Go-authored descriptor
   are removed after parity is proven.
6. `change_capture` is a source-only transport mode. A destination declaration
   containing it fails structurally before composition or I/O.
7. Existing closed GitHub stage/approval/reconciliation behavior is reused.
   This issue does not alter #4154 sync-apply work, #4136 certification sweep
   behavior, #4125, or #4090's bounded snapshot implementation.

## Manual GSD fallback

`scripts/gsd doctor`, all required `sources` lookups, generated prompts for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review`, and `go run ./cmd/agentcontractgen check` have passed. #4093 is
an issue phase rather than a numeric roadmap phase and this single-worker
environment cannot run compatible isolated Pi roles; the lifecycle therefore
runs inline with durable evidence in this directory.

## Required skills loaded

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-naming`, `golang-error-handling`, `golang-safety`, `golang-security`,
`golang-database`, `golang-testing`, and `golang-documentation`.

CLI help/manual/website parity is not applicable: no command, flag, help text,
manual, or website surface changes.
