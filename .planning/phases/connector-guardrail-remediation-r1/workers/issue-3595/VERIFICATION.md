# VERIFICATION — issue 3595 icon registry single-source foundation

## Required focused gates

```bash
gofmt -w cmd internal
go test ./internal/connectors ./internal/connectors/boundary ./cmd/iconregistrygen ./cmd/connectorgen
go test ./cmd/pm ./internal/cli
node --check website/scripts/gen-connector-bundles.mjs
node --check website/scripts/fetch-simple-icons.mjs
```

Add or adjust package-specific commands if implementation places tests elsewhere. Use existing package managers/tools only; do not add dependencies without approval.

Review repair focused gate:

```bash
go test ./internal/connectors ./internal/connectors/bundleregistry ./cmd/iconregistrygen && node --test website/scripts/icon-registry.test.mjs
```

Review repair round 2 focused gate:

```bash
node --test website/scripts/icon-registry.test.mjs && go test ./internal/cli -run TestValidateConnectorDocsRejectsStaleIconMetadata && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 3 focused gate:

```bash
go test ./internal/cli -run '^TestValidateConnectorDocsRejectsStaleIconMetadata$' && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 4 focused gate:

```bash
go test ./internal/connectors ./internal/cli ./cmd/iconregistrygen -run 'Test(ConnectorIconRegistryProjectsCompleteMetadata|ConnectorIconMetadataOmitsAbsentOptionalFields|ValidateConnectorDocsRejectsStaleIconMetadata|BuildIconEntriesPreservesCuratedAttribution)$' && go run ./cmd/pm docs validate --connectors-dir docs/connectors
```

Review repair round 5 focused gate:

```bash
node website/scripts/gen-connector-bundles.mjs
node website/scripts/gen-connector-catalog.mjs
node website/scripts/gen-connectors.mjs
# Assert the first run changes only derived icon values/deletes, rerun all three generators, and require byte-stable outputs.
go test ./internal/cli -run 'Test(GeneratedConnectorIconBlockRequiresExactUniqueHeading|ValidateConnectorDocsRejectsStaleIconMetadata)$'
node --test website/scripts/icon-registry.test.mjs
```

Review repair round 6 focused gate:

```bash
gofmt -w internal/cli/connector_docs.go internal/cli/connector_docs_test.go cmd/iconregistrygen/main.go cmd/iconregistrygen/main_test.go
go test ./internal/cli ./cmd/iconregistrygen -run 'Test(GeneratedConnectorIconBlockRequiresExactUniqueHeading|BuildIconEntriesRejectsDuplicateCuratedKeys|BuildIconEntriesRejectsSharedAssetPathSourceURLConflict|BuildIconEntriesAllowsSharedAssetPathWithIdenticalSourceURL)$'
node --test website/scripts/icon-registry.test.mjs
node website/scripts/gen-connector-bundles.mjs
# Hash bounded generated outputs, run the generator again, and require identical hashes plus a clean output diff.
```

## Repository gates before integration

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make connector-boundary
make verify
```

If a gate is not applicable or blocked by environment, record the exact reason and do not claim it passed.

## GitHub / no-mistakes gates

- PR targets `fix/3579-connector-path-ownership-guardrails` and uses `Refs #3595` and `Refs #3579`.
- Required/current GitHub checks green before parent integration.
- Comprehensive native-Codex `gpt-5.6-sol` no-mistakes validation at `xhigh`, including full-diff comprehensive review/rereview of all material substantiated issues.
- Do not integrate PR #3590 from the prior no-mistakes run; #3590 needs fresh 5.6 SOL validation after this foundation lands and is reconciled.

## Current evidence

- `scripts/gsd doctor`: pass in `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3595-icon-registry`.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run`: failed with `unknown GSD command: programming-loop`; manual GSD fallback uses `.pi/prompts/pm-gsd-loop.md` and must be recorded in PR evidence.
- Pre-edit audit/proof commands completed without credentialed checks.
- RED `go test ./internal/connectors ./cmd/iconregistrygen`: failed before implementation on missing exact bare lookup/ownership/generator-collision support.
- RED `node --test website/scripts/icon-registry.test.mjs`: failed before implementation on prefixed registry keys, website override authority, and website script prefix handling.
- GREEN `go test ./internal/connectors ./cmd/iconregistrygen`: pass.
- GREEN `go test ./internal/connectors ./internal/connectors/boundary ./cmd/iconregistrygen ./cmd/connectorgen`: pass.
- GREEN `node --test website/scripts/icon-registry.test.mjs`; `node --check website/scripts/gen-connector-bundles.mjs`; `node --check website/scripts/fetch-simple-icons.mjs`: pass.
- GREEN `go test ./internal/cli ./cmd/pm`: pass (`internal/cli` took 365.247s).
- GREEN `go vet ./...`: pass.
- GREEN `go test ./...`: pass.
- GREEN `go build ./cmd/pm`: pass.
- GREEN `make connector-boundary`: clean boundary report.
- GREEN `make verify`: pass, including docs validation, smoke, lint, connectorgen validate, connector boundary, and Homebrew notification checks.
- GREEN review round 6 focused Go gate: `internal/cli` and `cmd/iconregistrygen` targeted F11/F13/F14 regressions passed.
- GREEN review round 6 Node gate: all 6 icon-registry tests passed, including F12 invalid unimplemented-row coverage.
- GREEN review round 6 deterministic generation: two consecutive bundle generations emitted 550 connectors and 334 icons with identical data/public-icon hashes and no checked-in derived-output diff.
