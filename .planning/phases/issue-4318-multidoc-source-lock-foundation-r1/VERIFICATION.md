# Verification — additive v3 multi-document source locks

## Planned checks

- Focused `cmd/connectorgen` importer, projection, and surface-sync behavioral tests.
- Checked-in GitHub source-lock, descriptor, and combined-ledger byte/hash assertions.
- `go test -timeout 20m ./cmd/connectorgen`
- `go vet ./...`
- `go build ./cmd/pm`
- `go run ./cmd/connectorgen source-import github --check`
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
- `go run ./cmd/connectorgen certification-subject --check`
- Every `make verify` gate, then full `make verify` as required by the task.
- Generated `verify-work` and `code-review` prompts, executed inline; automated-review routing after PR creation.

## Result

Passed.

### Behavioral and parity evidence

- Red before implementation:
  - `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3ImportsDocumentOwnedProvenanceAndDuplicateProviderIDs$'` failed at strict legacy decoding because `rest.openapi` was an array.
  - `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportNormalizesScalarYAMLMappingKeys$'` failed because an unquoted `200:` mapping key was rejected as non-string.
- Green after implementation:
  - `go test -count=1 -timeout 20m ./cmd/connectorgen` — pass.
  - `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3SynchronizesDuplicateArtifactDigests$'` — pass.
  - `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportPreservesFrozenGitHubArtifacts$'` — pass.
  - `go run ./cmd/connectorgen source-import github --check` — pass (1,525 operations).
  - `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` — pass (552 connectors, zero drift).
  - `go run ./cmd/connectorgen certification-subject --check` — pass.
  - `node scripts/github-combined-operation-ledger.mjs --check` — pass.
- The GitHub artifacts remain byte-identical to the mandated baseline:
  - source lock: 3,420,025 bytes, SHA-256 `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`;
  - descriptor: 43,354,021 bytes, SHA-256 `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`;
  - combined ledger: 2,553,169 bytes, SHA-256 `c4a904a919f30065fcc8453c6689e1a3dcc7be5ac8e11a7154d310b334972de3`.

### Repository gates

- `go vet ./...` — pass.
- `go build ./cmd/pm` and `go build ./cmd/connectorgen` — pass.
- `make verify` — pass in full, including `go test -timeout 20m ./...`, docs validation,
  smoke, `golangci-lint` (0 issues), connector validation, surface sync, GitHub source/ledger
  tests, certification checks, connector boundary, and release/install checks.
- `git diff --check` — pass.

### Scope note

`go run ./cmd/connectorgen source-import dockerhub --check` cannot run on this base because
`internal/connectors/defs/dockerhub/sources/` does not exist. The correction is therefore proven
through the real source-lock/fetch/import path using hermetic YAML fixtures; it does not add a
Docker Hub capture or any connector-owned artifact outside this foundation lane.

### GSD/manual review fallback

The `verify-work` and `code-review` prompts were generated with `scripts/gsd prompt` and executed
inline. The project contract prohibits lifecycle role spawning in this environment, so the manual
fallback is recorded in `UAT.md` and `REVIEW.md`; no compatible isolated GSD worker was available.
