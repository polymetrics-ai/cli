# Summary — Lever Hiring connector parity (#3223)

## Result

Implemented connector-local Lever Hiring official documentation parity for the current public Lever Developer documentation inventory.

Final operation ledger totals:

- 117 official rows total: `GET=56`, `POST=26`, `PUT=14`, `DELETE=11`, `WEBHOOK=10`.
- 25 fixture-backed ETL streams.
- 23 bounded direct-read operations.
- 14 typed reverse-ETL write actions.
- 55 blocked/planned rows with official-source or shared-runtime dependency evidence.

## Changed surfaces

- Expanded `internal/connectors/defs/lever-hiring/api_surface.json` to cover every audited official row exactly once.
- Expanded Lever streams, schemas, and fixtures under `internal/connectors/defs/lever-hiring/**`.
- Added Lever direct-read operation definitions and typed write definitions/fixtures.
- Added `internal/connectors/defs/lever-hiring/cli_surface.json` for safe dynamic CLI discovery.
- Updated Lever metadata and connector docs/manual/skill/catalog surfaces to truthfully advertise read/write capabilities.
- Updated root CLI golden transcripts for the new Lever dynamic connector command listing.

## Safety notes

- No live Lever credentials, provider reads, provider writes, or certification claims.
- No new dependencies.
- No shared runtime, engine, or CLI Go behavior changes.
- No generic HTTP/path/body/query passthroughs or arbitrary write tools.
- Reverse ETL remains plan -> preview -> explicit approval -> execute; destructive actions have closed schemas and typed confirmation.

## Verification

Passed credential-free local checks:

- Official ledger comparison: `official 117 local 117 missing 0 extra 0 dupes 0`.
- `go run ./cmd/connectorgen validate internal/connectors/defs`.
- `go test ./internal/connectors/conformance -run 'TestConformance/lever-hiring' -count=1`.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`.
- `go vet ./internal/connectors/... ./internal/cli/...`.
- `go build ./cmd/pm`.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`.
- CLI smoke for `lever-hiring`, `lever-hiring opportunities`, `lever-hiring direct`, and `lever-hiring delete-feedback`.
- `make connector-boundary`.
- `git diff --check`.
- `make verify`.

## GSD notes

Repo-local GSD adapter was checked with `scripts/gsd doctor`/`scripts/gsd list`. The requested `scripts/gsd prompt programming-loop ...` command was unavailable, so this slice used the documented manual GSD fallback and recorded plan/TDD/verification artifacts before production edits.
