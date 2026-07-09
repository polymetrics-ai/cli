# Plan: Help Scout CLI Parity Parent Orchestration

Parent issue: #212
Parent branch: `feat/212-helpscout-cli-parity`
Default branch: `main`
Parent PR: https://github.com/polymetrics-ai/cli/pull/230 (draft)
Connector slug: `help-scout`

> Note: the kickoff prompt names `internal/connectors/defs/helpscout/`, but this checkout already has the canonical bundle at `internal/connectors/defs/help-scout/` and public docs/catalog entries use `help-scout`. This plan uses the existing canonical slug to avoid a duplicate connector.

## GSD / Skills Evidence

- GSD adapter health: `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json` passed on 2026-07-09.
- GSD planning prompt: `scripts/gsd prompt plan-phase 212 --skip-research --tdd` captured in `PROMPTS.md`.
- Mandatory programming-loop command attempted: `scripts/gsd prompt programming-loop init --phase issue-212-helpscout-cli-parity --dry-run`.
- Manual fallback: the repo-local command registry has no `programming-loop` command and `scripts/gsd` returned `unknown GSD command: programming-loop`; use the manual GSD loop from `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` with TDD ledger and verification artifacts kept current.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.
- Required references loaded: issue/parent contracts, parent orchestration loop, stacked PR workflow, CodeRabbit and automated review routing loops, GSD Pi adapter, CLI help/docs/website parity, connector migration handoff/conventions/design.

## Objectives

1. Create/confirm the parent branch and draft parent PR to `main` before sub-issue execution.
2. Maintain durable parent state in `ORCHESTRATION-STATE.json`.
3. Execute #213 first: refresh Help Scout API/CLI surface metadata from the official Inbox API docs.
4. Keep all Help Scout operations in one of the allowed classifications: stream, direct read, write, binary read, or blocked operation row with exact reason.
5. Preserve safety gates: no secrets, no credentialed checks, no generic raw HTTP write, no generic shell/SQL write, and reverse ETL only plan → preview → approval → execute.

## Sub-Issue Queue

| Issue | Status | Dependencies | Initial write scope | Notes |
| ---: | --- | --- | --- | --- |
| #213 | worker_ready | none | `.planning/phases/issue-213-*`, `internal/connectors/defs/help-scout/{metadata.json,api_surface.json,cli_surface.json,docs.md}` | First lane; refresh official docs inventory and command metadata. |
| #214 | planned | #213 | connector docs/help renderer/docs/website surfaces | Help/docs parity after metadata shape exists. |
| #215 | planned | #213, #216 | `internal/connectors/defs/help-scout/{streams.json,schemas,fixtures}` | Stream-backed reads and pagination/cursors. |
| #216 | planned | #213 | `internal/connectors/defs/help-scout/api_surface.json`, optional `operations.json` | Exact operation ledger classification. |
| #217 | planned | #213, #216 | direct-read command/runtime metadata | Bounded safe direct reads only. |
| #218 | planned | #213, #216 | provider-specific query/body/binary policies | No unsafe binary downloads; bounded policies only. |
| #219 | planned | #216 | writes/policy/redaction/approval metadata | Sensitive/admin/destructive operations gated. |

## Slice Boundaries

### Parent planning checkpoint

- Write parent plan/TDD/verification/orchestration artifacts.
- Commit and push to `feat/212-helpscout-cli-parity`.
- Open a draft parent PR to `main` with `Refs #212`.

### #213 local critical path

- Create branch `feat/213-helpscout-cli-surface-metadata` from the parent branch after the parent PR exists.
- Capture official Inbox API endpoint inventory (~146 endpoint pages; current crawl observed 146 endpoint pages, 145 unique method/path pairs, duplicate thread-source JSON/RFC822 path).
- Add/refresh `cli_surface.json` for safe Help Scout command metadata.
- Refresh `api_surface.json`/metadata without overclaiming implementation.
- Run focused validation before broader gates.

## TDD Strategy

- Use validation as the red gate for metadata changes: intentionally add `cli_surface.json`/`api_surface.json` references before they pass, then fix metadata until `connectorgen validate` is green.
- Add or update tests only if the existing validators cannot express a needed Help Scout safety rule.
- Do not weaken current validator rules; direct-read/binary output policy gaps become #217/#218 blockers if needed.

## Verification Checklist

Targeted before sub-PR handoff:

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json
# If cli_surface.json is added:
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

Parent/full local gate before final handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

CLI/help/docs parity checks apply once #214+ change runtime help or website/manual surfaces.

## Human Gates

- Parent PR merge to `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Credentialed Help Scout checks.
- Reading, printing, storing, or inventing secrets.
- Destructive external actions.
- Production deployment.
- Quality gate reductions.
- Generic shell, generic HTTP write, generic SQL write, raw GraphQL mutation, or raw API write escape hatches.
- Reverse ETL execution outside plan → preview → approval → execute.
