# Twenty S3 read streams — TDD ledger

Issue: #280

| Step | Red / validation target | Green evidence | Status |
| --- | --- | --- | --- |
| 1 | Focused bundle test should fail against S1 stub (`streams: []`) because Twenty has 0 streams. | `internal/connectors/engine/twenty_bundle_test.go` asserts 28 streams and passes after `streams.json` fill. | Green |
| 2 | `api_surface.json` empty skeleton should not cover new streams. | 28 list rows with `covered_by.stream` match the 28 snake_case stream names. | Green |
| 3 | Get-by-id endpoints cannot honestly use `covered_by.direct_read` because the engine has no generic direct-read output policy. | 28 get-by-id rows recorded as `excluded.category=out_of_scope` with engine-gap reason. | Green |
| 4 | Cursor pagination and schema refs must be static-checked. | Focused test asserts cursor `starting_after`, token `pageInfo.endCursor`, stop `pageInfo.hasNextPage`, page size 60, and schema ref path for every stream. | Green |
| 5 | Full connector validation must stay clean. | Worker reported `go run ./cmd/connectorgen validate` -> `548 connector(s) checked, 0 findings`; VERIFY stage will independently rerun. | Pending independent VERIFY |

## Red/green note

The first Codex attempt self-blocked before a green commit because the initial task prompt incorrectly required camelCase stream names and generic direct-read coverage. The corrected prompt made the red condition explicit: generic direct-read is an engine gap, so get-by-id rows are excluded in this slice instead of implemented.
