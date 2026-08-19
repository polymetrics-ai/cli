# TDD ledger — Zoom direct-read salvage

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Reviewed direct-read cohort | `go test -timeout 20m ./internal/connectors/defs/zoom -run TestReviewedDirectReadSalvageCohort -count=1` fails against the baseline with `reviewed rest_read operations = 0, want 70`. | Salvaged only PR #3951 `rest_read` rows, command rows, ledger dispositions, and direct-read fixtures. The cohort and loopback fixture tests pass; all 70 command paths stop at `error: missing --credential` through the binary. | Regenerated the endpoint ledger, sweep, nine affected CLI manual transcripts, and Zoom's generated MANUAL/SKILL; removed retired `format: date` annotations; record fixture proof only and retain the SCIM2 base-origin caveat. | green |

No credential or live provider request is permitted in this wave. Every direct-read cell is implemented-and-pending-certification, never certified.
