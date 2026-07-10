# Plan: Freshdesk Full Operation Implementation / Ledger Refinement

Parent issue: #172
Primary sub-issue: #176
Related lanes: #175, #177, #178, #179
Branch: `feat/172-freshdesk-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/222 (draft)

## GSD / Safety

- GSD health was verified earlier in this parent session with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- `scripts/gsd prompt programming-loop ...` is not registered in this adapter; continue the manual universal runtime loop through `.pi/prompts/pm-gsd-loop.md` and record the fallback.
- Required skills remain loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.
- No secrets, no credentialed Freshdesk checks, no new dependencies, no reverse ETL execution, no raw generic HTTP write, no generic SQL/shell write.
- Reverse ETL remains plan → preview → approval → execute; destructive/admin writes require `confirm: destructive` where applicable.

## Scope

Implement all Freshdesk operations that can be safely represented by existing connector architecture without creating a raw write escape hatch:

1. Make bounded JSON direct-read output policy generic enough for Freshdesk GET endpoints.
2. Convert Freshdesk GET operation rows to executable stream or direct-read coverage.
3. Add Freshdesk command-surface entries for implemented direct reads with explicit path/query flag mappings.
4. Add named Freshdesk reverse-ETL write actions for the 53 POST/PUT/DELETE baseline operations with path fields, risk text, and destructive confirmation for deletes.
5. Convert mutation operation rows to `covered_by.write` only when the action has a named schema and gated write semantics.
6. Keep truly binary/non-JSON or underspecified operations blocked only if the current engine cannot safely execute them without a binary/file executor or raw payload escape hatch.

## Red Evidence

- Current surface is not fully implemented: `freshdesk implemented coverage=5, blocked=165, total=170`; exit 1.
- Existing direct-read output policy support is GitHub-specific; Freshdesk generic JSON direct reads need a red test before engine changes.

## Implementation Slices

### Slice A — Generic bounded JSON direct reads

- Add a red test showing `output_policy: "json"` is accepted by `commandrunner` and `engine.DirectRead`, preserves JSON body, and still rejects non-JSON/oversized/absolute endpoints.
- Implement the policy in `commandrunner`, `cmd/connectorgen` validation, `engine/direct_read.go`, and `cli_surface.schema.json`.

### Slice B — Freshdesk GET coverage

- Generate command-surface direct-read entries for all non-stream GET rows.
- Mark GET rows with `covered_by.direct_read` or `covered_by.direct_reads` where duplicates share an endpoint key.
- Preserve existing stream coverage for `tickets`, `contacts`, `companies`, `agents`, and `groups`.

### Slice C — Freshdesk mutation coverage

- Generate named write actions for POST/PUT/DELETE rows.
- Use path fields for all endpoint variables.
- Use `body_type: none` for deletes; mark deletes idempotent with `missing_ok_status: [404]` only where product-safe. Otherwise do not claim idempotency.
- Add destructive confirmation for deletes and high-risk destructive/admin actions.
- Use conservative schemas with explicit known path fields plus documented/common body properties; do not expose a raw `payload`/arbitrary HTTP body command.

## Verification Checklist

- [ ] Red tests for generic JSON direct-read policy.
- [ ] `go test ./internal/connectors/commandrunner -run DirectRead`.
- [ ] `go test ./internal/connectors/engine -run DirectRead`.
- [ ] Freshdesk JSON coverage script: blocked rows = 0 or documented safe blockers only.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs --json`.
- [ ] `go test ./cmd/connectorgen -run CLISurface`.
- [ ] `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner`.
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'`.
- [ ] Broader gates as needed: `go vet ./...`, `go test ./... -timeout 20m`, `go build ./cmd/pm`, `make verify`.

## Orchestration Decision

No subagent spawned: this Pi harness lacks the `subagent` tool and all remaining work touches the same Freshdesk definition plus shared direct-read policy code. Proceed inline as `local_critical_path` and keep commits small.
