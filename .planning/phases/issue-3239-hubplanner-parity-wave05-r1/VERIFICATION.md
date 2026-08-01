# Verification checklist: Hubplanner parity wave05-r1

- [x] Official inventory script/check records 107 Hubplanner operations with truthful implemented/blocked counts (`traces/official-inventory-summary.md`).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 550 connector(s), 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/hubplanner' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `make connector-boundary` — clean.
- [x] `git diff --check` — passed.
- [x] `go run ./cmd/pm help hubplanner`, `go run ./cmd/pm hubplanner`, `go run ./cmd/pm hubplanner resource-custom-fields search --help`, and `go run ./cmd/pm connectors inspect hubplanner --json` — passed.
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed.
- [x] `make verify` — passed.
- [x] `gh-axi` parent/subissue final update posted once with counts and verification evidence — comments posted to #3239 and #3240-#3246.
