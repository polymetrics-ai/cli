# Twenty S3 read streams — TDD ledger

Issue: #280
PR: #290

| Step | Red / validation target | Green evidence | Status |
| --- | --- | --- | --- |
| 1 | Current `api_surface.json` failed full S3 invariant: `endpoints 56`, `covered_stream 28`, `excluded 28`, `direct_read 0`, then `AssertionError`. | Rewrote get-by-id rows to same-stream coverage; invariant now prints `endpoints 56`, `covered_stream 56`, `excluded 0`, `direct_read 0`. | Green |
| 2 | `test ! -f internal/connectors/engine/twenty_bundle_test.go` exited 1, proving the unowned engine test was present before cleanup. | Deleted `internal/connectors/engine/twenty_bundle_test.go`; focused package tests pass without it. | Green |
| 3 | Minimal fixtures must remain because conformance binds as soon as streams are declared. | Kept `fixtures/streams/attachments/page_1.json` and `page_2.json`; `go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1` passed. | Green |
| 4 | Full defs validation must stay clean with 56 stream-covered rows and existing streams. | `go run ./cmd/connectorgen validate internal/connectors/defs --json` returned `findings: []`, `warnings: []`, `connectors_checked: 548`. | Green |
| 5 | Broad repo gates must stay green without the unowned engine test. | `go vet ./...`, `go test ./... -count=1`, `go build ./cmd/pm`, and `gofmt -l cmd internal` passed. | Green |

## Notes

- Turn36 operator decision resolved the previous fixture blocker: S3 keeps minimal fixtures; S7 #284 refines/expands later.
- Red validation captured before production edits in this correction pass.
- `make verify` not run because this task forbids reverse ETL execution and the Makefile `verify` target depends on `smoke`, which runs `./pm reverse run`.
