# TDD ledger — source-lock operation import

## Cycle 1 — closed retrieval and integrity

- **Red:** `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport'` initially failed to compile because no `sourceImportLock`, fetch boundary, descriptor, limits, or importer existed.
- **Green:** two checked-in source locks now fetch only their exact fixed fixture URL; mismatched bytes/digest and artifact-size violations return a source-lock-refresh/bound error before command output is created. Connector lock lookup is fixed to `defs/<connector>/sources/<connector>-operation-source-lock.json`.
- **Refactor:** one internal fetch interface keeps fixture retrieval explicit while the production fetcher permits only the lock's HTTPS URL, follows no redirects, sends no credentials, and bounds its read.
- **Lint repair:** the first full `make verify` run reached lint after all tests/smoke checks and reported an unchecked HTTP response-body close. The fetcher now returns read/close failures with context; it does not discard cleanup errors.

## Cycle 2 — canonical provider descriptors

- **Red:** alpha/beta fixture tests required a non-existent descriptor contract and deterministic canonical serializer.
- **Green:** OpenAPI 3 YAML and Swagger 2 JSON fixtures yield byte-stable sorted descriptors with distinct path/query/header/body sections, source URL/digest/location, pagination/auth/byte metadata, empty provider-operation ID plus derived source ID, and JSON/binary/status/text classification. Resolved response declarations retain both `200` and `403` shapes plus `rare_paid_result` and `access_token` fields.
- **Refactor:** response declarations are retained wholesale; classification is a separate additive field so later runtime redaction cannot become an import-time field filter.

## Cycle 3 — bounded reference and request contracts

- **Red:** isolated fixture tests required named failure modes for references, request contracts, identity, callbacks, source drift, and bounds.
- **Green:** external, unresolved, cyclic, sibling-ambiguous, over-depth, and over-count references; duplicate identities; callback routes; unbounded/dynamic request bodies; unsupported XML encoding; oversized artifact/schema; and operation-count excess all fail before descriptor emission.
- **Refactor:** the local JSON Pointer resolver owns bounded ref traversal and canonical map copying, avoiding a second schema walker with different safety rules.

## Cycle 4 — closed adoption command and documentation

- **Red:** `source-import` was an unknown command and migration conventions had no adoption contract.
- **Green:** help documents connector/defs/output/check only; it exposes none of `--url`, `--method`, `--path`, `--header`, `--body`, or `--credential`. Command/check-mode tests prove owned-lock resolution and descriptor drift detection; migration guidance records lock refresh and preservation constraints.
- **Refactor:** standard and connector-qualified `--help` both render the same closed command contract. Inline GSD code review found no remaining correctness, safety, or documentation findings.
