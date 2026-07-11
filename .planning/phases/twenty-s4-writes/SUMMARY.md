# Twenty S4 writes summary (#281)

Status: GREEN for PR #304 F1 review fix; fix committed/pushed at `ccc3047f0202562de49c7c85d1e19d1c9a660c73` and re-reviewed PASS.

Review fix F1 accepted:

- Existing batch actions used `body_fields:["records"]`, which sends `{ "records": [...] }`.
- Twenty server source evidence says `POST /rest/batch/{objects}` consumes `request.body` directly as `Partial<ObjectRecord>[]`; body must be a bare JSON array.
- Fixed by adding schema-gated, action-specific `body_field:"records"` JSON payload support; updated 28 `batch_` actions; kept `record_schema.required == ["records"]`; no generic/raw HTTP tool and no reverse ETL execution.

Original delivered:

- Created `internal/connectors/defs/twenty/writes.json` with 84 non-delete actions: 28 create, 28 update, 28 batch.
- Updated `internal/connectors/defs/twenty/api_surface.json` to 140 rows: existing 56 GET rows + 56 POST + 28 PATCH; 84 `covered_by.write`; 0 DELETE.
- Derived create/update/batch record schemas from S2 schemas and `FIELD-MANIFEST.json` after immutable/system pruning; `additionalProperties:false`; update schemas require `id` and `minProperties:2`; batch uses an outer `records` array for validation/preview and `body_field:"records"` so execution sends a bare top-level JSON array.
- Added narrow engine/schema support for declarative JSON `body_field`; no write fixtures needed; conformance passed without S4 fixture edits.

Evidence:

- Manual GSD fallback active: `scripts/gsd prompt programming-loop init --phase twenty-s4-writes --dry-run` unavailable (`unknown GSD command: programming-loop`); review-fix used `scripts/gsd prompt gsd-quick "S4 #281 PR #304 review-fix correction for Twenty batch body shape"` as adapter prompt evidence.
- Red evidence before production edits: `writes.json` missing; API surface 56 total / 56 GET / 0 write rows.
- Original gates passed: jq parse/count/shape, `connectorgen validate` (`findings: []`, `warnings: []`), twenty conformance, focused packages, `gofmt -l`, `go vet ./...`, `go build ./cmd/pm`, `go test ./... -count=1`, `scripts/verify-gsd-workflow 1a86cc1a`.
- Review-fix red: focused engine test failed before production support with `unknown field BodyField in struct literal of type WriteAction`.
- Review-fix green gates passed: jq/schema parse, Python batch shape (`batch body_field ok 28`), focused body-field tests, `connectorgen validate`, twenty conformance, focused packages, `go vet ./...`, `go build ./cmd/pm`, `gofmt -l cmd internal`, `go test ./... -count=1`, `scripts/verify-gsd-workflow 1a86cc1a`.
- `make verify` not run because it executes `pm reverse run` via smoke; S4 forbids reverse-ETL execution.

Safety:

- No live credentials, no generic raw HTTP tool, no reverse-ETL execution, no DELETE/destructive rows, no dependencies, and no CLI/help/website edits. The F1 fix intentionally changed the connector engine/write schema plus migration docs only to add action-scoped JSON `body_field` support.
