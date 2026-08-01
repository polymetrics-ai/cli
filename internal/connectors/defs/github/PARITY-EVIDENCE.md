# GitHub connector parity evidence

## GSD plan

- GSD command path: attempted `scripts/gsd prompt programming-loop init --phase issue-2989 --dry-run`; the repo-local adapter reported `unknown GSD command: programming-loop`, so this file records the required manual-GSD fallback before production edits.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-graphql`, `golang-documentation`, and CLI help/docs/website parity guidance.
- Scope: connector-owned GitHub bundle metadata, fixtures, command metadata, operation ledger, API-surface ledger, and this evidence file only.
- Source inventory slice: public, unauthenticated GitHub REST OpenAPI, public GraphQL schema docs, and public webhook event docs; no provider credentials, live data calls, connector certification, merges, or live writes.
- Safety slice: include DELETE/destructive/admin operations in the ledger and implemented command metadata only when they retain the existing reverse ETL plan -> preview -> explicit approval -> execute path plus typed `destructive` confirmation.
- Documentation slice: record counts and limitations without claiming certification or fabricating implemented/fixture-tested counts.

## TDD ledger

| Stage | Red/validation target | Green target |
| --- | --- | --- |
| Source inventory | Current `api_surface.json` covers only the previous repository-scoped slice and does not enumerate the full official REST + GraphQL + webhook inventory. | `api_surface.json` records the full official inventory plus connector conformance coverage rows needed by the current one-target-per-row schema. |
| Destructive safety | Most DELETE-backed write actions omit `confirm: "destructive"`, so implemented destructive writes do not all advertise the typed confirmation gate. | Every implemented DELETE/delete action declares typed destructive confirmation; delete actions carry idempotency notes where the engine can apply them. |
| CLI metadata | gh-style destructive commands are still described as blanket unsafe/disallowed even though the captain policy says they are in scope with typed confirmation. | Command metadata states destructive operations are planned or implemented through typed confirmation, while raw API and token-printing escape hatches remain disallowed. |
| Evidence | No connector-local evidence ties the official counts to the generated ledger and safety policy. | This file records source counts, GSD/manual fallback, safety policy, and verification commands. |

## Verification checklist

- [x] Public source inventory script reports REST 1.1.4 operation count, GraphQL Query+Mutation field count, and webhook event count without credentials.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` returned zero findings and zero warnings across 548 connectors.
- [x] `go test ./cmd/connectorgen -run TestGitHub` passed after API-surface count and destructive metadata test updates.
- [x] `go test ./internal/app -run TestRunReverseETL.*Destructive` passed the typed confirmation safety gate.
- [x] `go test ./internal/cli -run TestGitHubDestructiveCommandRequiresTypedConfirmation` passed the CLI confirmation gate.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts/connectors_inspect_github_json` refreshed the connector-owned inspect golden after metadata changed.
- [x] `go test ./cmd/connectorgen ./internal/app ./internal/cli` passed.
- [x] No connector certification, live provider requests, credentials, pushes, merges, or no-mistakes daemon lifecycle commands were run in this worker slice.

## Source inventory (public, unauthenticated)

Authoritative sources from parent issue #2989:

| Source | Public URL | Inventory method | Count |
| --- | --- | --- | ---: |
| GitHub REST OpenAPI | `https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json` | Count HTTP method objects under `paths` for GET/POST/PUT/PATCH/DELETE. | 1,216 |
| GitHub GraphQL schema | `https://docs.github.com/public/fpt/schema.docs.graphql` | Count top-level `Query` and `Mutation` fields, skipping descriptions and arguments. | 305 |
| GitHub webhook events | `https://docs.github.com/en/webhooks/webhook-events-and-payloads` | Count event `<h2>` IDs after excluding navigation/about headings. | 75 |
| Total official inventory | Parent issue #2989 | REST + GraphQL + webhook events. | 1,596 |

These counts are source inventory counts, not implemented counts. Implementation, fixture, blocked/planned, exclusion, and certification counts must continue to come from the JSON ledgers and tests.

## Safety and dependency notes

- Shared typed confirmation foundation is present in `internal/app`: destructive connector command and reverse ETL plans require the matching `--confirm destructive` challenge before execution.
- This connector slice does not change shared runtime code.
- Destructive/delete operations are not blanket-excluded as unsafe. Missing, unimplemented, unfixture-backed, or multi-segment path-blocked destructive operations remain blocked/planned until they have typed schemas, write fixtures, bounded command metadata, idempotency notes, and the existing plan -> preview -> explicit approval -> execute path.
- `create_or_update_file`, `delete_file`, `update_ref`, and `delete_ref` depend on reviewed shared typed allowlisted multi-segment write-path support before connector dispatch.
- Raw arbitrary GitHub API access and token-printing commands remain disallowed because they would expose a generic API escape hatch or secret values, not because DELETE operations are categorically out of scope.
