# Mailchimp parity wave03-r1 plan

## Scope

Parent issue: #3078. Capability subissues: #3079-#3085. Worker branch: `fm/cli-mailchimp-parity-wave03-r1`.

Implement connector-local Mailchimp Marketing API parity evidence for the current wave without live provider calls, credentials, pushes, PR updates, or `/no-mistakes`.

Allowed write scope:

- `internal/connectors/defs/mailchimp/**`
- connector-owned generated docs/catalog outputs under `docs/connectors/**` and website generated connector data/examples when produced by repo tooling
- connector-owned CLI/help/golden metadata when generated from the Mailchimp definition
- `.planning/phases/mailchimp-parity-wave03-r1/**`

No shared runtime/engine behavior changes were planned. During verification, the expanded bundle exposed two tooling/runtime scalability gaps that were fixed generically without connector-specific policy: `connectorgen validate` now accepts a single bundle path (required by this task's exact gate), and `bundleregistry.New` reuses the immutable embedded bundle parse across repeated registry construction while still returning a fresh registry instance. If a Mailchimp operation requires generic raw batch execution, live provider certification, new dependencies, or shared runtime support outside existing definition-owned surfaces, keep it blocked with exact evidence.

## GSD / skills

- GSD adapter checked: `scripts/gsd doctor` passed.
- Requested command per repo guidance: `scripts/gsd prompt programming-loop init --phase mailchimp-parity-wave03-r1 --dry-run`.
- Adapter result: command was unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so this phase records a manual GSD universal programming-loop fallback per `.agents/agentic-delivery/references/gsd-pi-adapter.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- Supporting prompt generated: `scripts/gsd prompt execute-phase mailchimp-parity-wave03-r1 --dry-run`.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-error-handling`, `golang-safety`, `golang-documentation`, `golang-structs-interfaces`, `golang-design-patterns`, `golang-context`, `golang-concurrency`.
- CLI/docs parity reference loaded: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Official audit plan

1. Fetch the official Mailchimp Marketing Swagger root `https://api.mailchimp.com/schema/3.0/Swagger.json` and every provider-owned path `$ref` referenced by the root.
2. Persist a sanitized trace with version, ref count, operation count by method/top-level family, and every operation identity (`METHOD path`, operationId, summary, tags). Do not store credentials or live response data.
3. Reconcile the official operation inventory against `internal/connectors/defs/mailchimp/api_surface.json`.
4. Inventory truthful post-change counts from the final `api_surface.json`: executable stream/write/direct-read rows, blocked operation rows, and excluded/N/A rows.

## Implementation slices

1. **Ledger repair (red first):** prove current `api_surface.json` is incomplete against the official 298-operation audit. Replace the legacy 9-row surface with a complete operation-ledger-mode manifest.
2. **Executable coverage:** add definition-owned streams, direct-read/search/query metadata, write actions, operations metadata, schemas, and sanitized fixtures for every Mailchimp operation that is supportable by current declarative engine/commandrunner surfaces and within the task safety constraints.
3. **Blocked evidence:** keep only operations that need a precise shared-runtime dependency, expose disallowed generic batch/raw execution, require live-only certification, or are N/A/deprecated as blocked/excluded with source evidence.
4. **Docs/catalog parity:** update `docs.md`, generated connector `MANUAL.md`/`SKILL.md`, all-connectors catalog rows, website generated connector data, and CLI golden transcripts.
5. **Issue addendum:** append the established captain-policy addendum idempotently to #3078-#3085 via `gh-axi`, using actual final counts and no certification claims.

## TDD / validation strategy

- Red validation: custom official-audit check shows current Mailchimp surface covers only 9 of 298 official operations.
- Green validation: `go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp`, `go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1`.
- CLI/docs validation: `pm help connectors`, `pm connectors inspect mailchimp --json`, `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`.
- Full local gates from task: `go build ./cmd/pm`, `make connector-boundary`, `make verify`, `git diff --check`.

## Safety notes

- No live provider data calls, writes, credentials, secret values, certification, VPS, Herdr, Thaalam, or daemon lifecycle actions.
- Reverse ETL remains plan -> preview -> explicit approval -> execute; destructive actions use `confirm: "destructive"` and typed schemas.
- No generic shell, generic HTTP write, generic SQL write, or raw batch operation execution surfaces.

## Orchestration decision log

- Cycle `plan`: `local_critical_path` — firstmate supplied an isolated worker worktree and asked this crewmate to implement locally; mutating subagents not spawned into this checkout.
- Cycle `red`: official Swagger audit showed 298 current operations with only 9 legacy rows represented.
- Cycle `green`: generated fixture-backed Mailchimp definition now represents all 298 operations: 79 streams, 68 typed direct reads, 148 approval-gated reverse-ETL actions, and 3 blocked/local-workflow rows.
- Cycle `verification-fix`: made `connectorgen validate <single-bundle>` generic and cached embedded bundle loading so expanded connector metadata does not push repeated CLI/certify tests past the 20m make-verify timeout.
