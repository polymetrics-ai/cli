# Plan — source-lock operation import

## Scope and boundaries

Implement shared `cmd/connectorgen` source-import infrastructure and checked-in synthetic fixtures only. The lane does not edit any production `internal/connectors/defs/<connector>` bundle, cache a production OpenAPI artifact, add a dependency, make provider calls, add `pm` execution behavior, accept credentials, or provide caller-controlled HTTP transport fields.

## TDD slices

1. **Red — closed lock and artifact integrity.** Add tests for absent/out-of-tree/mismatched-connector locks, exact source URL use, byte-count mismatch, SHA-256 mismatch, oversized artifacts, and no output on failure. The existing importer must not exist yet.
2. **Green — validated retrieval and source normalization.** Add a bounded fetch seam, strict lock parser, byte/digest comparison, JSON/YAML parsing, and canonical JSON encoder. Keep remote fetch fixed to the lock metadata and bounded by configured size/time limits.
3. **Red — descriptor extraction.** Add fixtures/tests that require path/query/header/body separation, request media type, response class/output policy, pagination, byte limits, auth scopes, missing `operationId`, and a second connector in sorted canonical output.
4. **Green — fixed descriptor projection.** Extract one descriptor for each ordinary provider operation; preserve provider IDs (including empty), derive only missing source IDs, lowercase fixed method/path, and stable source locations. Preserve every provider-declared response status shape and ordinary field exactly; classify binary, status, and text output kinds without making them executable or filtering fields by risk/tier/familiarity.
5. **Red — rejection matrix.** Add isolated tests for external, cyclic, unresolved, ambiguous, and over-depth references; duplicate identities; callback-only routes; free-form/dynamic or unbounded request contracts; unsupported encodings; artifact/schema byte limits; and configured operation/ref/count limits.
6. **Green — fail-closed resolver and schema bounds.** Resolve only local bounded JSON Pointers with cycle/depth/count guards; reject every unsafe/unrepresentable form before serialization.
7. **Adoption surface.** Register `connectorgen source-import <connector> [--defs <dir>] [--out <path>] [--check]`, document its lock verification and descriptor handoff in migration guidance, and add help/docs tests. The command owns no URL/method/path/header/body/credential flags.
8. **Verification and review.** Run focused tests then package and repository gates. Generate/execute lifecycle prompts for execute, verify, and code review inline. Record outcomes and resolve any gap plan before final commit.

## Expected red/green/refactor evidence

- **Red:** tests compile against planned `sourceimport` APIs but fail because no closed importer/descriptor exists; rejection tests prove unsafe source forms cannot produce descriptors.
- **Green:** every fixture produces exact canonical descriptor bytes or a named failure, fixture transport observes only the source-lock URL, and the output retains every declared response status/schema field including unusual and sensitive-looking names.
- **Refactor:** share parsing/ordering helpers without weakening a rejection, retain separate structs for lock, descriptor, resolved reference, and output classification, and maintain byte-stable JSON.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport|TestRunSourceImport|TestSourceImportHelp'`
- `go test -timeout 20m ./cmd/connectorgen`
- `go run ./cmd/connectorgen source-import --help`; source-import fixture golden/check mode; `go run ./cmd/connectorgen validate internal/connectors/defs`; `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
- `go vet ./...`; `go build ./cmd/pm`; `git diff --check`; completion-tracked `make connector-boundary`; `make verify`
- No `pm help` check is applicable: no `pm` surface changes. `docs/cli/**` and `website/**` are checked for unintended references rather than edited.

## Commit checkpoints

1. Planning artifacts and task header (`Refs #4306`).
2. Red tests/fixtures (`Refs #4306`).
3. Green importer, command, documentation, and tests (`Refs #4306`).
4. Verification/review evidence and any focused fixes (`Refs #4306`).
