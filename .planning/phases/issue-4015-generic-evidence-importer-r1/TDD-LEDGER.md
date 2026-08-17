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
