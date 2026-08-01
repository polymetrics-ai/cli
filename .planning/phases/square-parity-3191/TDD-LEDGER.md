# Square parity #3191 — TDD ledger

## Red / baseline

- `go run ./cmd/connectorgen validate internal/connectors/defs/square` was the wrong single-connector invocation for this tool shape and treated `fixtures/`/`schemas/` as connector roots, producing two `missing_file` findings. Full-root validation was used for the real gate.
- Baseline Square conformance before expansion: `go test ./internal/connectors/conformance -run 'TestConformance/square' -count=1` passed with the prior 4-stream bundle, proving the implementation needed parity expansion rather than bug reproduction.
- Official Square OpenAPI latest inventory from developer.squareup.com contained 346 operations.

## Green / implementation evidence

- Generated complete Square ledger: 346 total official operations.
- Implemented surfaces: 110 stream-backed GET operations, 31 bounded direct-read/search operations, 153 typed reverse-ETL write actions.
- Fixture-backed surfaces: 110 stream fixture page sets and 153 write fixtures.
- Blocked/planned rows: 52, with closed operation models and source evidence.
- Certification: 0 live certified; `certification.json` intentionally has no live write pairings/direct/binary candidates.

## Refactor / review fixes

- Regenerated CLI golden transcripts and website connector generated data after adding Square `cli_surface.json` and write-capable metadata.
- Added two-page `payments` fixture coverage so cursor pagination termination remains exercised.
- `make verify` initially timed out in `internal/connectors/certify` during the full concurrent suite at the package 20m timeout. Evidence showed no connector failure: `go test ./internal/connectors/certify -timeout 30m -count=1` passed, and the final full `make verify` passed.
