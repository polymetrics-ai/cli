# TDD Ledger — Generic Certification Evidence Importer

## Planned red/green slices

| Slice | Red: observable failing check | Green: implementation evidence |
| --- | --- | --- |
| Generic report importer | A completed HTTP-proof report cannot be imported to accepted evidence without the PostgreSQL-only command path. | A definition-derived source binding imports the report and the record validates/matches its matrix cell. |
| Proof redaction | A planted header/query/body secret can be detected in the emitted record or a raw body survives serialization. | The generated record contains only allowed proof metadata/fingerprints and no planted value. |
| Honest matrix accounting | A valid report with no accepted evidence can read as live-tested, or a corrupted record leaves its cell green. | Missing evidence stays red; corrupted evidence makes the named cell red; restoring the original record returns green. |
| Connector agnosticism | A differently shaped source definition requires a new command branch. | The same generic importer handles a second definition-owned binding without shared connector identifiers. |

## Red / green execution record

_Append exact commands and results while executing; do not include secrets or provider payloads._

- Red: `go test ./cmd/connectorgen -run TestCertificationEvidenceReportImport -count=1` — importer test not yet present; the existing command only accepts database-shaped `transport` and `change-capture` inputs.
- Green: `go test -timeout 20m ./cmd/connectorgen -run 'TestCertificationEvidence(Report|Postgres)' -count=1` — pass; a definition-owned GitHub binding and a second Xero binding import the same validated external-proof shape, while the PostgreSQL database records still pass their existing contract tests.
- Green: `go test -timeout 20m ./internal/connectors/certify -run 'Test(ReadExternalProof|WriteExternalProof)' -count=1` — pass; a deliberately corrupted raw response is rejected at the import boundary.
- Green: `TestCertificationEvidenceReportRefusesUnverifiedFullParityScope` makes a completed but non-full-parity report fail before an accepted record can be created. This is a guard on the new generic path, not a change to the existing record contract; #4211 remains open for the legacy writer's unconditional scope stamp.
- Green: `TestCertificationEvidenceReportRefusesProofFromDifferentRun` makes a valid redacted proof fail when its fresh-child run ID does not derive from the completed report's start time; a report and its observed exchanges cannot be spliced across runs.
- Red: `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationEvidenceReportRefusesIncompleteBindingsWithoutPartialPublication$' -count=1` — failed as expected: a later incomplete binding left the earlier `operation:rest_read` record on disk.
- Green: the importer now validates every binding and output-path identity before opening a record. The same test leaves zero accepted records when its GraphQL binding is resumed.
- Red: a synthetic secret-shaped string temporarily added to a valid `github/certification.json` `live_unavailable.contains` entry made `go run ./cmd/connectorgen validate internal/connectors/defs/github` exit 1 with `[secret_literal]`. An earlier placement in `source.default_stream` was rejected by meta-schema first and was immediately moved; it was not counted as the scanner proof.
- Green: the synthetic value was removed with `apply_patch`; the same validation command returned `0 findings`. No credential value or fixture was committed.
- Red: after temporarily changing the published PostgreSQL CDC record's `function_kind` from `capability:cdc` to `capability:read`, `go run ./cmd/connectorgen certification-matrix --connector postgres` regenerated the CDC cell as `live_tested=false, records=0`.
- Green: restoring that one field and regenerating restored the same cell to `live_tested=true, records=1`; `git diff --exit-code db494bc8f -- internal/connectors/certifications/evidence` passed afterward.
- Green: `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/certify ./internal/connectors/engine ./internal/cli` — pass after rebasing to `db494bc8f`; the refreshed GitHub ledger supplies 25 eligible direct reads (23 REST and 2 GraphQL), and the test suite keeps an explicitly empty accepted-evidence input non-live-tested.
