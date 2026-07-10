# GSD Plan — Issue #204 Crisp all-ops executable coverage

Issue: https://github.com/polymetrics-ai/cli/issues/204
Parent branch: `feat/204-crisp-cli-parity`
Working branch: `feat/204-crisp-all-ops`
Base dependency: #205 metadata scaffold (`feat/205-crisp-cli-surface-metadata` / PR #235)
GSD command path: `scripts/gsd prompt plan-phase 204 --skip-research`; manual programming-loop fallback remains active because `scripts/gsd prompt programming-loop ...` is not registered.

## Scope requested

Implement executable coverage for the official Crisp REST API v1 operation inventory while preserving repository safety gates.

## Safety interpretation

- All GET operations become explicit, bounded direct-read commands with fixed provider endpoints, explicit path/query flags, and a generic JSON response output policy capped by the existing direct-read byte limit.
- POST/PUT/PATCH/DELETE operations become explicit named reverse-ETL write actions with fixed method/path, path fields, record schema validation, dry-run preview, approval token, and destructive confirmation where applicable.
- No raw generic HTTP command is introduced: every executable command maps to one fixed Crisp operation from the official ledger.
- No credentialed Crisp checks or live API calls are run.
- Body shapes are transport-safe scaffolds for official operations: path fields are typed and required; mutation body fields remain per-record JSON fields until later issue-specific semantic schemas are refined from provider body docs.
- Binary/export/import-like operations remain explicit fixed actions/operations with high-risk approval text and output size policy; no unbounded file download or shell/write escape hatch is introduced.

## Planned implementation slices

1. Extend direct-read output policy support with a provider-neutral bounded JSON policy (`json_response`) in schema, command runner validation, and engine direct-read application.
2. Generate Crisp operation mappings from the #205 ledger:
   - Add `writes.json` for all non-GET operations.
   - Mark all GET CLI commands `implemented` with one `api_surface` endpoint, path/query flags, and `output_policy=json_response`.
   - Mark all mutation CLI commands `implemented` with a `write` action, path/body flags where safe, risk and approval text.
   - Replace all `operation` classifiers in `api_surface.json` with `covered_by.direct_read` or `covered_by.write`.
   - Set metadata capabilities for read/write true.
3. Update docs/catalog/manual artifacts and count summaries.
4. Add/adjust tests for generic JSON direct-read policy, URI-template query expression handling, and Crisp operation coverage counts.

## Red / green strategy

Red evidence before production edits:

```bash
go test ./internal/connectors/commandrunner -run TestRunDirectReadSupportsGenericJSONResponsePolicy -count=1
```

Expected: fail because generic JSON output policy is unsupported.

Green evidence:

```bash
go test ./internal/connectors/commandrunner ./internal/connectors/engine -count=1
go test ./cmd/connectorgen -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/crisp' -count=1
./pm docs validate --connectors-dir docs/connectors
make verify
```

## Required skills used

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-lint`

## Done criteria

- All 220 Crisp API surface endpoints have executable coverage (`covered_by.direct_read` for GET or `covered_by.write` for mutations) or an explicit safety exception if execution would violate policy.
- `connectorgen validate internal/connectors/defs` passes.
- Direct-read policy tests pass.
- Write actions are explicit and gated by reverse-ETL plan/preview/approval/confirmation rules.
- Docs/manual/catalog artifacts reflect the implemented surface.
