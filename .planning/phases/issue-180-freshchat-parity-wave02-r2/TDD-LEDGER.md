# TDD LEDGER — issue #180 Freshchat parity wave02 r2 (fixtures)

## Red / pre-change evidence

- `go test ./internal/connectors/conformance -run 'TestConformance/freshchat' -count=1`
  before this change: the `read_fixture_nonempty` check silently `Skip`ped the 12 streams
  without `fixtures/streams/<name>/page_1.json` (only 6/18 streams exercised real data).
- `verify-gsd-workflow` (CI `gsd-workflow-evidence` gate): failed with
  `cmd/internal changed, but no GSD planning evidence changed` because the fixture change
  touched `internal/` without a corresponding `.planning/` evidence change.

## Green / implementation evidence

- Added 12 replay fixtures under
  `internal/connectors/defs/freshchat/fixtures/streams/<name>/page_1.json`, matching this
  repo's existing fixture/schema conventions (`internal/connectors/conformance/replay.go`):
  request, response, and read_query shape.
- `go run ./cmd/connectorgen validate internal/connectors/defs/freshchat` — 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshchat' -count=1` — passed
  (all 18 streams now pass `read_fixture_nonempty` with real recorded data).

## Refactor / review evidence

- Gofmt / go vet / go build / `make verify` re-verified; docs + website catalog regeneration
  produced zero diff (docs/website already accurate, no commit needed).
- No live Freshchat provider calls, credentials, writes, pushes, merges, or PR actions performed.

## GSD / TDD note

Fixture-only change (no production behavior change). Evidence recorded here (this phase
directory) satisfies the `gsd-workflow-evidence` gate that requires GSD/TDD evidence to
accompany any `internal/` change.
