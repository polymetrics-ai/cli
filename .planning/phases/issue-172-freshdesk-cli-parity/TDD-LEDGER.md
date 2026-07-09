# TDD Ledger: Freshdesk CLI Parity Parent

Parent issue: #172

## Red / Baseline Evidence

- `python3` endpoint-count check against `internal/connectors/defs/freshdesk/api_surface.json`
  - Result: failed as expected — `freshdesk api_surface endpoints=10, want 170`; `exit=1`.
- `test -f internal/connectors/defs/freshdesk/cli_surface.json`
  - Result: failed as expected — `cli_surface_file_exit=1`.

## Planned Red Gates By Lane

- #173 metadata: endpoint-count and `cli_surface.json` presence/load checks fail before data edits.
- #176 operation ledger: full official operation classification check fails until every official operation has one primary classification and reason.
- #175 streams: focused conformance/schema checks fail for newly added streams before fixtures/schemas are filled.
- #177 direct reads: command/operation validation fails for bounded safe direct-read operations before metadata/runner support lands.
- #178 query/binary: binary/query policy validation fails before provider-specific policies are declared.
- #179 sensitive/admin: write safety/risk/approval validation fails before typed confirmation and redaction metadata are declared.
- #174 help/docs: CLI help/docs/website parity checks fail until generated/manual surfaces are updated.

## Green Evidence

#173 metadata slice:

- Freshdesk JSON/count validation passed: 170 endpoints (`GET:117`, `POST:10`, `PUT:10`, `DELETE:33`), 5 covered streams, 165 blocked operation rows.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`: passed, 547 connectors checked, 0 findings, 0 warnings.
- `go test ./internal/connectors/engine -run CLISurface`: passed.
- `go test ./cmd/connectorgen -run CLISurface`: passed.
- `go test ./cmd/connectorgen ./internal/connectors/engine`: passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'`: passed.
- Broader gate: `go test ./...` with default timeout failed in `internal/connectors/certify` after 10 minutes; `go test ./internal/connectors/certify -run TestWritePlanPreviewJSONHasNoApprovalToken -count=1 -timeout 20m`, `go test ./... -timeout 20m`, `go build ./cmd/pm`, `make verify`, and final `connectorgen validate` passed.

## Refactor Notes

- Production Go code unchanged in #173 metadata slice.
- `scripts/gsd prompt programming-loop ...` is unavailable in the current adapter; manual universal-loop fallback is recorded in `PLAN.md` and `RUN-STATE.json`.
