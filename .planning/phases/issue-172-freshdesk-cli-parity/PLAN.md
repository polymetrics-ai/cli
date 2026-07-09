# Plan: Freshdesk CLI Parity Parent Orchestration

Parent issue: #172
Parent branch: `feat/172-freshdesk-cli-parity`
Parent PR: pending draft PR creation
Default branch: `main`
Connector: `freshdesk`

## GSD / Skill Evidence

- GSD adapter health: `scripts/gsd doctor` passed; `scripts/gsd verify-pi` passed; `scripts/gsd list --json` returned 69 commands.
- GSD planning prompt: `scripts/gsd prompt plan-phase 1 --skip-research` generated a prompt successfully.
- Programming-loop command path: `scripts/gsd prompt programming-loop init --phase issue-172-freshdesk-cli-parity --dry-run` failed with `unknown GSD command: programming-loop`; use the repo-local Pi prompt `.pi/prompts/pm-gsd-loop.md` plus the manual universal runtime loop and record this fallback in all issue artifacts.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.
- Required references loaded: issue-agent contract, parent-orchestrator contract, parent orchestration loop, stacked PR workflow, GSD universal runtime loop, CodeRabbit loop, automated review routing loop, GSD Pi adapter, CLI help/docs/website parity, connector migration handoff, connector conventions, connector architecture v2 design.

## Scope

Bring `internal/connectors/defs/freshdesk/` and related CLI/docs surfaces to full connector CLI parity against the official Freshdesk API docs. Every official operation must be exactly one of stream, direct read, typed reverse-ETL write, bounded binary/file policy, duplicate/deprecated/disallowed/auth-internal/product-scope block, or other typed operation-ledger block. Sensitive/admin/destructive operations are not blanket exclusions; implement or ledger them behind reverse-ETL approval/typed confirmation as applicable.

## Sub-Issue Plan

| Issue | Lane | Dependencies | Primary write scope | Initial status |
|---:|---|---|---|---|
| #173 | CLI surface metadata | none | `internal/connectors/defs/freshdesk/**`, `.planning/phases/issue-173-*` | local critical path |
| #176 | Operation ledger | #173 source inventory | `internal/connectors/defs/freshdesk/api_surface.json`, possible `operations.json` | planned |
| #175 | Stream runner | #173/#176 classifications | `internal/connectors/defs/freshdesk/streams.json`, schemas, fixtures | planned |
| #177 | Direct reads | #176 classifications | direct-read metadata/operations and CLI docs | planned |
| #178 | Advanced query/binary engine | #176 classifications | provider query/body/binary metadata/policies | planned |
| #179 | Sensitive/admin policy | #176 classifications, write inventory | writes/policy/risk/approval metadata | planned |
| #174 | Help renderer/docs | #173 surface metadata; updates as lanes land | docs/help/website/generated artifacts | planned |

## First Slice (#173)

1. Refresh Freshdesk official operation inventory from https://developers.freshdesk.com/api/ without credentials.
2. Create/refresh `api_surface.json` with 170 official operations and honest initial classifications.
3. Add `cli_surface.json` metadata mapping safe provider-style commands to existing streams or planned/blocked intents.
4. Update metadata/docs only as far as #173 can prove; do not overclaim implemented streams/writes.
5. Validate JSON, connector bundle loading, and static validation.

## TDD / Validation Strategy

- Red baseline: prove current `api_surface.json` has 10 entries, not the 170-operation official baseline; prove `cli_surface.json` is absent.
- Green #173: JSON parses; `engine.Load(defs.FS, "freshdesk")` loads non-nil `CLISurface`; `connectorgen validate` accepts the Freshdesk bundle and full defs tree.
- Later behavior-changing lanes must add focused red Go tests before code changes.

## Verification Checklist

- [ ] `python3`/`jq` JSON parse checks for Freshdesk metadata/surface files.
- [ ] `go test ./internal/connectors/engine -run CLISurface`.
- [ ] `go test ./cmd/connectorgen -run CLISurface`.
- [ ] `go test ./cmd/connectorgen ./internal/connectors/engine`.
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] `pm help <topic>`, `pm <namespace>`, and docs/website parity checks when CLI-visible behavior changes land.
- [ ] Parent handoff gates before final readiness: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`.

## Orchestration Decision

No mutating subagent was spawned in this Pi harness because the available tool surface does not include the Pi `subagent` tool and the first lanes share the same Freshdesk definition directory. The orchestrator will execute #173 inline as `local_critical_path`, then re-evaluate the queue. Record this as a runtime-capability/isolation constraint, not as a completed worker spawn.

## Human Gates

- Parent PR merge to `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Secrets or credentialed connector checks.
- Destructive external actions or reverse ETL execution.
- Production deploys.
- Quality-gate reductions.
- Generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tools.
