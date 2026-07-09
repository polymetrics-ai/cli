# Issue #97 Plan — Linear CLI surface metadata

Date: 2026-07-09
Parent issue: #80
Issue: https://github.com/polymetrics-ai/cli/issues/97
Branch: `feat/80-linear-cli-parity` (parent integration branch for this local slice)
Scope: `internal/connectors/defs/linear/cli_surface.json`, focused tests, issue evidence.

## Objective

Produce and validate `internal/connectors/defs/linear/cli_surface.json` so Linear has a safe, provider-inspired command metadata surface. The metadata must map implemented stream-backed commands to existing streams and classify other commands as planned, unsupported, or unsafe without exposing raw generic GraphQL or write escape hatches.

## GSD command path

- Parent planning prompt: `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-97-linear-cli-surface --dry-run`.
- Result: unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Manual-GSD fallback active for this lane.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-graphql`, `golang-documentation`, `web-design-guidelines`, `vercel-react-best-practices`, `frontend-design` (loaded for generated website/docs data parity; no hand-written UI changes in this slice).

## Implementation boundaries

Allowed:

- Add `internal/connectors/defs/linear/cli_surface.json`.
- Add focused test coverage proving the embedded Linear bundle exposes implemented stream commands.
- Update issue phase artifacts.

Not allowed in this slice:

- New dependencies.
- Credentialed Linear checks.
- Reverse ETL execution.
- New write actions.
- Raw/generic GraphQL command execution.
- Direct-read execution beyond descriptive planned metadata.
- Destructive/admin external actions.

## TDD plan

1. RED: add `TestLinearCLISurfaceMapsImplementedStreams` in `internal/connectors/engine/bundle_test.go`.
   - It should fail because Linear currently lacks `cli_surface.json`.
   - Required mappings: `issue list` → `issues`, `team list` → `teams`, `project list` → `projects`, `user list` → `users`.
   - It should assert all four are `intent=etl`, `availability=implemented`, and have no `operation` executor.
2. GREEN: add `cli_surface.json` with:
   - source CLI reference to `linear` CLI docs/repo where available;
   - global `json` and `connection` flags;
   - implemented ETL list commands for the four existing streams;
   - planned direct-read candidates such as `issue view`, `project view`, `team view`, `user view` without marking them implemented;
   - planned or unsafe reverse-ETL/admin commands with explicit plan/preview/approval notes and blocked status;
   - docs-only/auth/config/local workflow entries marked non-executable.
3. REFACTOR: ensure schema validity and no unsupported `output_policy` for Linear.
4. VERIFY: focused test, Linear bundle validation, and commandrunner/engine packages.

## CLI help/docs/website parity position

This slice adds metadata consumed by future help rendering but does not add a new executable `pm linear` command path by itself. Runtime help/docs/website checks will be marked metadata-only unless generators consume the file in this slice. Issue #98 owns rendering/docs expansion.

## Safety rationale

Only existing read streams are marked implemented. Mutations and admin/destructive/sensitive operations are not executable. No generic GraphQL document/body flag is exposed.
