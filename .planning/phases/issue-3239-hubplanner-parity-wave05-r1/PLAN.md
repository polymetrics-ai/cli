# Issue #3239 Hubplanner parity wave05-r1 plan

## GSD workflow

- GSD adapter checked with `scripts/gsd doctor` (green).
- Required command `scripts/gsd prompt programming-loop init --phase issue-3239 --dry-run` is unavailable in this checkout (`unknown GSD command: programming-loop`), so this phase uses the documented manual-GSD fallback from `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Scope

Implement Hubplanner official API parity from parent #3239 and child issues #3240-#3246, staying connector-local to `internal/connectors/defs/hubplanner/**` plus phase artifacts and issue status evidence. No shared runtime behavior, dependencies, infrastructure, live provider calls, credentials, pushes, PRs, or no-mistakes pipeline.

## Official source inventory

Authoritative sources are the provider-owned `hubplanner/API` repository and `Sections/*.md` files at tree SHA `91217d34486e43fa590e9f9e3e477aee20da157a` from #3239. The issue-provided audit counts are the starting oracle: total 107 operations = 33 ETL/read, 59 reverse-ETL write, 2 direct/provider search, 13 CDC/webhook.

## Implementation plan

1. Re-audit the official Markdown and create a complete `api_surface.json` operation ledger with `operation_ledger_version: 1`, source URLs, duplicate-free rows, and post-change counts.
2. Expand `streams.json`, schemas, and stream fixtures for all list-style Hubplanner resources expressible as declarative streams.
3. Add `operations.json` and `cli_surface.json` for bounded typed direct reads where the existing connector contract supports GET/POST provider reads with fixed endpoint, closed body schema, max bytes, and explicit flags.
4. Add `writes.json` and one sanitized write fixture per implemented write action. Use closed `record_schema`, no generic request body, no raw method/path/query, fixed provider paths only, explicit destructive confirmation for deletes, and documented idempotency notes where provider status behavior is represented.
5. Keep webhook event delivery/changefeed rows blocked with exact official evidence and the documented shared CDC/runtime dependency where no existing declarative connector contract can receive provider callbacks as CDC.
6. Update Hubplanner docs and certification metadata truthfully: fixture-only, not certified.
7. Run focused validation, conformance, CLI/golden checks, boundary, diff check, build, and `make verify`; update issues once through `gh-axi`; commit cleanly.

## Safety constraints

- No secret values in fixtures/docs/issue updates.
- No live Hubplanner API calls and no credentialed checks.
- No generic HTTP method/path/body/query, shell, file, SQL, or passthrough escape hatches.
- Reverse ETL remains plan -> preview -> explicit approval -> execute; destructive write commands require typed confirmation.
- Provider query/search remains connector-owned metadata and does not change warehouse `pm query` semantics.
