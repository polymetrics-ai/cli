# TDD ledger — Segment parity (#3150-#3157)

## Red checks planned before/with implementation

- [x] Add/maintain a focused Segment parity test asserting:
  - official total = 197 operations;
  - API surface endpoint count = 197;
  - operations ledger count = 197;
  - lane coverage = stream 80, direct_read 19, write 96, operation binary 1, operation excluded/disallowed 1;
  - no stale `/workspaces` endpoint remains.
- [x] Existing conformance should fail on stale Segment ledger before regeneration and pass after definitions/fixtures are updated.
- [x] Existing CLI golden transcript tests should show dynamic connector command root drift once Segment `cli_surface.json` is added; update only generated/golden surfaces after implementation.
- [x] Add bundleregistry isolation regression coverage so cached definition loads return fresh registry maps and do not leak test mutations.

## Green checks planned

- [x] `go test ./cmd/connectorgen -run Segment -count=1`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/segment' -count=1`
- [x] `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestConnector' -count=1`
- [x] `go test ./internal/connectors/bundleregistry -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify`
- [x] `git diff --check`

## Notes

- No live Segment credentials or provider calls are permitted; fixture replay and local CLIs are the proof route.
- Reverse ETL execution is covered by existing shared plan/preview/approval gates; this phase only registers Segment fixed actions.
