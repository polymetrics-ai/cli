# Plan — Zoom provider-owned inventory parity, R1

## Delivery record

- Phase: `cli-zoom-parity-implementation-r1`
- Scope owner: `internal/connectors/defs/zoom/**`, Zoom-local fixtures/tests, generated connector documentation/catalog output, and this phase trace only.
- Locked source of truth: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-zoom-provider-inventory-rebuild-r1/report.md`, including its complete replacement ledger at lines 162–24751.
- GSD evidence: `scripts/gsd doctor`, `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`, and all five generated prompts passed. `gsd-sdk query init.phase-op cli-zoom-parity-implementation-r1` reports `phase_found: false`, so the official workflow exits before it may create artifacts. This directory is the required inline/manual fallback; no subagents are authorized for this single-connector task.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `no-mistakes`.

## Locked decisions and scope fences

- Land Wave 0 first as a commit containing only `internal/connectors/defs/zoom/api_surface.json`. The report's 54 legacy `excluded.category=deprecated` rows require the repository's operation-ledger-v1 representation, so they are normalized mechanically to `operation.model=deprecated`, `status=blocked`, retained reason/source URL, and `classification=justified_excluded`; no provider classification changes.
- The ledger must retain exactly 1,913 provider operations: 3 existing stream-backed executable reads, 1,839 connector-locally implementable operations that remain ledger-blocked pending their own typed contracts, 17 provider-restricted operations, and 54 provider-deprecated exclusions.
- Wave 1 ends after hardening the existing `users`, `meetings`, and `webinars` read streams and exposing each as an individually reachable `pm zoom <command>` ETL command. Do not start a Wave 2 resource family or any write implementation.
- No live provider calls, credentials, new dependencies, raw HTTP/SQL tools, shared runtime/engine/command-runner changes, or changes outside this connector lane.
- Writes remain unavailable. The 1,032 provider writes stay honestly ledger-disposed for later bounded family tickets; no redaction policy is introduced.

## Execution slices

1. **Wave 0 inventory landing**
   - Mechanically extract the report's fenced JSON body, validate it as JSON, verify 1,913 unique method/path rows and the cited disposition totals, then replace only Zoom's `api_surface.json`.
   - Run focused `connectorgen validate` and conformance/static checks permitted before a command surface exists.
   - Commit the ledger file alone before any command/test/documentation implementation.
   - Completed in `46eff2585` (`fix(connectors): rebuild Zoom provider operation inventory`): focused `connectorgen validate` and Zoom conformance both passed.

2. **Wave 1 existing-core hardening**
   - Add a Zoom-local failing test proving all three covered streams have an implemented command-surface route and pass the real command-runner preflight.
   - Add `cli_surface.json` routes for `users list`, `meetings list`, and `webinars list`; expose an optional `--user-id` config override only for the two user-scoped stream routes. Each command must map its exact provider method/path from the ledger.
   - Update Zoom's `docs.md` to describe the full provider inventory, the three executable routes, and the remaining exact dispositions. Regenerate the connector website catalog/manual output from the bundle rather than hand-editing derived data.
   - Run `surface-sync --check`; it must not invent metadata or alter the inventory because this wave contains stream-backed ETL commands rather than operation-backed direct commands.

3. **Verification and handoff**
   - Execute the red/green ledger below, focused connector/conformance/CLI tests, `go build ./cmd/pm`, and actual built-binary help invocations for `pm zoom`, `pm zoom users list --help`, `pm zoom meetings list --help`, and `pm zoom webinars list --help`.
   - Check changed paths are Zoom-local plus the explicit phase and generated catalog artifacts. Commit the Wave 1 slice, append the required status, and stop for firstmate's no-mistakes instruction.

## CLI parity checklist

- [x] `pm zoom` renders contextual command help successfully.
- [x] `pm help zoom`, `pm zoom users list --help`, `pm zoom meetings list --help`, and `pm zoom webinars list --help` resolve without credentials.
- [x] The three exact executable ledger rows map to one command each; no non-executable row is promoted.
- [x] `docs.md` and generated website catalog reflect the new command surface and honest 1,913-row ledger counts.
- [x] No command accepts a generic provider path, arbitrary query/body, or secret in a flag.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-zoom-provider-inventory-rebuild-r1/report.md`
