# TDD Ledger: Front CLI Parity Parent Orchestration

Parent issue: #188
Connector: `front`

## 2026-07-09 — parent planning and seed slice

- Task type: parent orchestration and planning.
- Production behavior changed: no.
- Red evidence: not applicable for planning-only artifact creation.
- GSD evidence:
  - `scripts/gsd doctor` passed.
  - `scripts/gsd verify-pi` passed.
  - `scripts/gsd list --json` completed.
  - `scripts/gsd prompt plan-phase 188 --skip-research --tdd` generated the planning workflow prompt.
  - `scripts/gsd prompt programming-loop init --phase issue-188-front-cli-parity --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD loop recorded in `PLAN.md`.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-lint`.
- Validation target before commit:
  - `jq empty .planning/phases/issue-188-front-cli-parity/*.json`
  - `git diff --check`

## Planned #189 first red artifact

Before editing `internal/connectors/defs/front/` for #189, create a failing validation/test that
captures the current gap:

- Official Front baseline contains 342 operations.
- Current `internal/connectors/defs/front/api_surface.json` contains 10 endpoint entries.
- Current Front bundle has no `cli_surface.json` or equivalent CLI surface metadata.

The red artifact may be a focused Go test, a connector metadata validation test, or a scripted
validation recorded under `.planning/phases/issue-189-front-cli-surface-metadata/`, but it must fail
against the current state before production connector edits.

## Planned green/refactor evidence

- Green #189 metadata slice: regenerated/curated surface metadata validates with `go run ./cmd/connectorgen validate internal/connectors/defs/front` and relevant focused tests.
- Refactor evidence: run `gofmt -w cmd internal` only if Go files change; otherwise record JSON/docs-only exemption.
- Broader gates before handoff: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`.
