# TDD Ledger — issue #253 Gong top-level JSON array request bodies

## Red (planned before production code)

- Engine write test proves top-level array body is sent unwrapped.
- Validation fails schema mismatch before network send.
- Preview/redaction does not leak content-like fields.
- Validator rejects unsafe implemented array-body commands.

## Green

- Added `body_type: json_array` write support with `body_field` selection and `body_schema` validation before network send.
- Added schema/meta-schema validation for `json_array` write actions.
- Added Gong `upload_crm_entity_schema` write metadata using a top-level selected-fields array without exposing raw JSON CLI body flags.

Evidence:
- `go test ./internal/connectors/engine -run 'OperationDirectRead|WriteJSONArray|WriteMultipart|DirectRead|Write' -count=1`
- `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|DirectRead|RedactRecord' -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

## Refactor

- Kept the Gong CRM entity-schema CLI command `partial` because provider-style raw JSON flags remain intentionally unavailable; executable support is available through schema-gated write records/table mappings.
- Reused existing reverse-ETL plan/preview/approval flow rather than adding a direct write shortcut.

## Skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.
