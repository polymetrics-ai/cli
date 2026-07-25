# TDD ledger — Bahmni rename/parity follow-up

## Red / baseline

Baseline before production edits on branch `feat/bahmni-docker-connector` at head `e18d58adf`:

- Connector identity was still `bahmni-docker` in bundle path/name, generated docs/catalogs, website generated data, PR title, and issue #516–#526 titles/bodies.
- The previous no-mistakes review found four authorized real defects:
  1. `document_path` was not matched by reverse-plan redaction markers.
  2. Four root-array streams inherited `offset_limit` pagination even though the Bahmni endpoints return bare arrays and do not honor `limit`/`startIndex`.
  3. The connector check endpoint used `/ws/rest/v1/session`, which can return 200 with `authenticated:false` for bad credentials.
  4. CLI group metadata used a hyphenated drug-order command name while the implemented stream command is `drug_orders`.
- The official-operation ledger covered the originally implemented operations but did not yet provide an evidence-backed parity matrix from official Bahmni/OpenMRS sources.
- The two captain-owned findings remained unresolved and were not production-edit decisions in this follow-up: global PHI field-redaction semantics and nullable `diagnoses.existingObs` primary-key semantics.

## Green / implementation

- Renamed connector id/path and generated user-facing identity to `bahmni` / `Bahmni`.
- Preserved `Bahmni/bahmni-docker` only as a deployment/setup repository reference.
- Renamed generated docs from `docs/connectors/bahmni-docker` to `docs/connectors/bahmni` and regenerated catalog/website entries.
- Fixed the four authorized defects:
  - `upload_patient_document` now uses `document_file_path`, which matches the existing `file_path` plan-redaction marker.
  - `drug_orders`, `lab_results`, `appointments`, and `diagnoses` set per-stream `pagination: {type: none}`; the appointment fixture no longer expects `limit/startIndex`.
  - `base.check` and `fixtures/check.json` now use authenticated `GET /ws/rest/v1/provider?v=default&limit=1` instead of `/session`.
  - CLI group and generated manual entries consistently show `drug_orders` for the drug-order command family.
- Softened false PHI-redaction prose without implementing or deciding an engine-level PHI policy. The connector now states that output is bounded and secret-shaped/file-path fields are redacted by current runtime policies, while broad clinical PHI field redaction remains an unresolved engine policy decision.
- Produced `docs/migration/issue-516-bahmni-operation-parity.md`, mapping official operation families to subissues and implementation/blocked/exclusion rows.
- Expanded `api_surface.json` with evidence-backed blocked/exclusion rows for official Bahmni operation families discovered during the audit: notes, ADT/end-visit/discharge, UI/admin config/global properties, appointment workflow mutations, appointment-service reference data, patient image/binary reads, and display/reporting helpers.
- Updated GitHub #516–#526 titles/bodies via `gh-axi issue edit`, preserving the parent/subissue hierarchy.

## Validation results

- `go run ./cmd/connectorgen validate internal/connectors/defs` → 548 connector(s), 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/bahmni$' -count=1` → PASS.
- `./pm docs validate --connectors-dir docs/connectors` → PASS.
- CLI discovery:
  - `./pm connectors inspect bahmni --json` → exit 0, connector name `bahmni`, display `Bahmni`.
  - `./pm bahmni` → exit 0, renders connector manual.
  - `./pm connectors inspect bahmni-docker --json` → exit 1, old connector id no longer resolves.
- Focused packages: `go test ./cmd/connectorgen ./internal/connectors/bundleregistry ./internal/connectors/conformance ./internal/cli -count=1` → PASS.
- `go vet ./...` → PASS.
- `go test -timeout 20m ./...` → PASS.

## Remaining gaps / non-decisions

- Appointment service/reference-data reads, notes, appointment workflow mutations, discharge/ADT/end-visit mutations, and patient image/binary reads are recorded as blocked/excluded gaps in the parity matrix/API surface rather than silently omitted.
- Broad clinical PHI field redaction remains captain-owned and unresolved; no engine change was made.
- `diagnoses.existingObs` nullable primary-key semantics remain captain-owned and unresolved.

## PM/no-mistakes

Pending after commit on the exact candidate head.
