# Asana source-operation strict-import compatibility R1

## Task Delivery Header

- Issue: Refs Batch 1 Asana source-lock compatibility repair; no independent issue was assigned.
- Base branch: `fm/cli-batch1-authoring-rgr` at `14aa19c76c327617216a891f394c9a658819208f`.
- Merges into: Batch 1 authoring branch, then its human-selected integration route.
- Delivery: local committed repair only; no push, PR creation, or merge.
- Working branch: `codex/asana-source-operation-compat-r1`.
- Task: Permit the existing Asana `source_operation` enrichment through strict source-import decoding while retaining closed lock decoding and source-retention validation.
- Verification: focused source-import tests; the three retained Asana source-projection regressions; package scope test; `gofmt`; `git diff --check`.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The immutable enriched Asana lock parses through strict source import | live | A test reads the retained lock and asserts that all 249 REST operations are admitted; without the compatibility field it fails with `unknown field "source_operation"`. |
| Enrichment does not make source-lock decoding permissive | live | Table cases inject unknown source-operation and parameter members and assert strict decoding rejects them. |
| Enrichment cannot weaken retained artifact identity | live | A fetched artifact with one extra byte is rejected against the lock's SHA-256 and byte count. |
| Existing Asana projection use remains source-import compatible | live | The three retained Asana source-projection tests generate and validate source-backed projections from the immutable lock. |

## Manual GSD/TDD fallback

`scripts/gsd doctor` and the registered lifecycle sources were verified. This bounded isolated worker task cannot invoke the repository's Pi-role workflow without violating the active single-worker/parent ownership, so the lifecycle is executed inline and recorded here. Required skills loaded: `go-engineering`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-code-style`, `golang-naming`, `connector-lane-build-order`, and `firstmate-exhaustive-review`.

## Red

- `go test ./cmd/connectorgen -run 'TestRetainedAsanaSourceImportRejectsReadProjectionDrift|TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams|TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation' -count=1`
- Result before the repair: all three tests failed because strict source import rejected the immutable lock with `json: unknown field "source_operation"`.
- Added `sourceimport_source_operation_test.go` before implementation. Its valid enrichment case reproduced the same failure; its negative cases define the closed envelope boundary.
- Once the closed `source_operation` envelope was admitted, the immutable lock advanced to its second documented enrichment, `source_contract`. The same enrichment commit adds its fixed legacy shape (`openapi`, `servers`, `security`, and `components`), so a second red step correctly failed with `json: unknown field "source_contract"`.

## Green

- `gofmt -w cmd/connectorgen/sourceimport.go cmd/connectorgen/sourceimport_source_operation_test.go`
- `go test ./cmd/connectorgen -run 'TestParseSourceImportLockAcceptsAsanaSourceOperationEnrichment|TestParseSourceImportLockSourceOperationEnrichmentStaysClosed|TestParseSourceImportLockSourceContractEnrichmentStaysClosed|TestSourceOperationEnrichmentCannotRelaxRetainedArtifactHash|TestSourceImportLegacyByteBackedLocksRejectReferenceOnlyFields|TestSourceImport_RejectsUnknownSectionAndIndependentIndexOverflow' -count=1` — pass.
- `go test ./cmd/connectorgen -run '^TestSourceImport|^TestParseSourceImportLock|^TestSourceOperationEnrichment' -count=1` — pass.
- `go run ./cmd/connectorgen source-import --help` — confirms `--check` is non-writing.
- `git diff --check` — pass.

## Scope-bound baseline failures

The non-writing real command reaches source import after this repair, proving that the prior unknown-field blocker is gone:

```text
go run ./cmd/connectorgen source-import asana --defs internal/connectors/defs --check
connectorgen source-import: apply source-cited partial mutation coverage dispositions: partial mutation coverage disposition cites unknown source operation "asana.rest.createCustomField"
```

The full package command was run after the repair:

```text
go test ./cmd/connectorgen -count=1
```

It still has five out-of-scope Asana failures: two fixed-100 ETL-classification failures for `asana.rest.getCustomFieldsForWorkspace`; two partial-mutation-disposition failures for the unknown imported `asana.rest.createCustomField`; and one retained source-projection failure because `getSectionsForProject` is omitted. These failures are downstream importer/projection/evidence defects, not strict source-lock decoding, and this repair does not change their owning code.
