# TDD LEDGER — Twilio parity wave04 r1

## Red target

Add `internal/connectors/conformance/twilio_full_coverage_test.go` before production edits. The test must fail against baseline because:

- `api_surface.json` lacks `operation_ledger_version: 1`.
- Only 7 stream fixture files exist for 103 executable streams.
- 0 write fixture files exist for 94 executable writes.
- Official lane arithmetic is not asserted in code.

## Expected green behavior

- Official-source lane arithmetic is asserted: total 197 = 96 ETL read + 93 reverse ETL write + 0 direct/provider query + 3 media/binary + 5 event/notification changefeed + 0 excluded.
- `api_surface.json` contains 197 method/path rows matching official Twilio OpenAPI v2010 exactly.
- Every declared stream has `fixtures/streams/<stream>/page_1.json`.
- Every declared write action has `fixtures/writes/<action>.json`.
- Media/binary official operations are present and covered by `medias`, `media`, and `delete_media`; event/notification official changefeed operations are present and covered by streams.
- `go test ./internal/connectors/conformance -run 'TestTwilioOfficialParityLedgerAndFixtureCoverage|TestConformance/twilio' -count=1` passes.

## Evidence log

- RED: `go test ./internal/connectors/conformance -run 'TestTwilioOfficialParityLedgerAndFixtureCoverage' -count=1` failed as expected with `operation_ledger_version = 0, want 1` before production edits.
- GREEN: `go run ./cmd/connectorgen validate internal/connectors/defs --json` reported 0 findings/0 warnings for Twilio, and `go test ./internal/connectors/conformance -run 'TestTwilioOfficialParityLedgerAndFixtureCoverage|TestConformance/twilio' -count=1` passed after adding operation-ledger version, fixture coverage, and docs updates.
- REFACTOR: kept the coverage test connector-local; docs/generated surfaces were narrowed back to Twilio-only output after full docs generation initially touched unrelated connector manuals.
- VERIFICATION CAVEAT: focused Twilio validation/conformance/CLI gates passed, but aggregate `make verify` did not finish green locally because `go test -timeout 20m ./...` hit slow pre-existing package timeouts; see `VERIFICATION.md`.
