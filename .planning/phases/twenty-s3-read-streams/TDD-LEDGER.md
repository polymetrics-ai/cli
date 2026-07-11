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
| 6 | Review F1 accepted: pre-production invariant failed on current head because cursor `limit_param` is still present before static `query.limit` exists. Evidence: review-fix Python command exited 1 with `AssertionError` at line 9. | Removed no-op `pagination.limit_param`, added static `query: {"limit":"60"}` to all 28 streams, updated attachments fixture queries. Required gates passed: `jq`, review-fix Python invariant, `connectorgen validate`, Twenty conformance, focused package tests, `go build ./cmd/pm`, `gofmt -l cmd internal`; additional local gates passed: `gofmt -w cmd internal`, `go vet ./...`, `go test ./... -count=1`, `scripts/verify-gsd-workflow b4895064`. | Green |

## Notes

- Turn36 operator decision resolved the previous fixture blocker: S3 keeps minimal fixtures; S7 #284 refines/expands later.
- Red validation captured before production edits in this correction pass.
- `make verify` not run because this task forbids reverse ETL execution and the Makefile `verify` target depends on `smoke`, which runs `./pm reverse run`.
- Review F1 disposition: Accepted; action = removed no-op cursor `limit_param` and added static stream query `limit=60`; red and green evidence recorded in `VERIFICATION.md`.
