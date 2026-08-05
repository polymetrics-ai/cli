# Issue #3773 — api_surface v2 per-operation provenance

Refs #3773, #3785, #3787, #3789, #3791.

## GSD setup

- GSD adapter preflight: `scripts/gsd doctor` passed.
- Resolved command sources: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed.
- Manual-GSD fallback: this isolated worker applies the generated lifecycle prompts inline. The
  canonical single-worker contract and this foundation's shared-path ownership make spawned GSD
  roles incompatible.
- Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-cli`, and `golang-documentation`.
- CLI parity reference loaded: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Goal

Add an `operation_ledger_version: 2` contract that provides every API-surface endpoint with
provider-artifact provenance without changing its executable binding. A v2 artifact has a stable
ID, HTTPS artifact URL, ISO-8601 full-date retrieval date, and optional SHA-256. A v2 endpoint has
exactly one `provenance` object containing an artifact ID and HTTPS operation citation URL.

`provenance` is parallel metadata only. `covered_by` continues to bind to an existing stream,
write action, or implemented direct read; no provenance value can advertise a capability, select an
executor, create a command, or bypass a blocked/destructive path.

## Compatibility and collision gates

- V1 ledgers continue to load, validate, and certify as `legacy_unverified`; their nested
  `operation.source_url` is retained only for v1 compatibility.
- V2 validation is complete and non-bypassable only for ledgers that opt into v2. Existing
  definitions are intentionally not migrated in this foundation; the provider-artifact sweep owns
  those 550+ data changes.
- PR #3740 is open and also changes `internal/connectors/engine/bundle.go`. This branch began at
  `56051eada51ac83cc815ff17d39728347609c239`; before final delivery, rebase on current `main`,
  resolve only this issue's owned code, and rerun targeted checks. Do not rebase onto a sibling
  foundation branch.

## Ordered slices

1. **#3785 — structural contract/model (TDD).**
   - Extend `api_surface.schema.json` with a versioned v2 branch and keep v1 structural rules.
   - Model artifacts and endpoint provenance in `engine.APISurface` / `engine.SurfaceEndpoint`.
   - Keep unknown root/artifact/endpoint/provenance keys rejected by the meta-schema and strict
     decode. Keep classifier exclusivity unchanged.
2. **#3787 — shared semantic validation/conformance (TDD).**
   - Add one engine-owned provenance validator for v2 semantic relationships and precise endpoint
     diagnostics.
   - Consume it from conformance's `surface_complete` result, without modifying existing
     `covered_by` target-resolution logic.
3. **#3789 — connectorgen gate (TDD).**
   - Consume the same validator from `connectorgen validate`; do not duplicate provenance rules.
   - Emit connector/file/endpoint-specific findings and retain v1 compatibility.
4. **#3791 — certification evidence and authoring docs (TDD).**
   - Report evidence separately from covered/read/write/blocked: ledger version, artifact count,
     total endpoints, cited endpoints, and `legacy_unverified` / `complete` / `invalid`.
   - Use the shared result for v2 invalidity; never turn a cited endpoint into an executable one.
   - Update authoring/design docs and only the existing CLI/manual/website artifacts that describe
     certification output.

## Scope and exclusions

Owned production paths are the engine API-surface schema/model/validator, conformance static
validation, `cmd/connectorgen/validate.go`, certification surface inventory/reporting, and the
specific authoring/certification docs they require. No `internal/connectors/defs/**` change, live
provider call, credential, new dependency, executor, output redaction/masking, retry behavior,
capability derivation, or approval-flow change is permitted.

## Verification plan

1. Capture the structural red case: v2 root/endpoint fields currently fail closed-schema loading.
2. Add named loader tests, make the structural case green, and prove v1 still loads.
3. Capture semantic and consumer red cases, then make engine/conformance/connectorgen/certify
   tests green against complete, malformed, and legacy fixture bundles.
4. Exercise JSON output, `pm connectors`, `pm connectors certify --help`, relevant `pm help`, and
   docs/website generated-output checks where applicable.
5. Run focused packages plus full vet/build and the individual non-suite repository gates; leave
   aggregate `go test ./...` / `make verify` to CI as required by `AGENTS.md`.
