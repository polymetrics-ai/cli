# TDD ledger — Lever Hiring connector parity (#3223)

## Red / pre-edit validation

### 2026-08-01 — official operation inventory parity is incomplete

Credential-free source audit against current public Lever documentation found 117 official operations (107 HTTP operations plus 10 webhook trigger/event names). The pre-edit local ledger had 11 rows.

Expected green evidence:

- `api_surface.json` contains every official operation exactly once.
- `api_surface.json` row count equals 117 unless the current official documentation changes during the slice.
- Existing implemented streams remain covered and fixture-backed.
- Any newly implemented direct reads, streams, or writes have typed, bounded, connector-local surfaces.
- Unsupported operations are blocked/planned with exact shared-runtime or official-source evidence, not silently omitted.

### 2026-08-01 — connector capabilities are stale

Pre-edit `metadata.json` advertised `write=false` and no command surface while official Lever docs include documented mutations and fixed-target direct/binary/webhook surfaces. Green required capabilities and CLI/docs to reflect only the implemented subset truthfully.

## Green / verification log

### 2026-08-01 — connector-local implementation green

- Official ledger comparison: `official 117 local 117 missing 0 extra 0 dupes 0`.
- Method totals: `DELETE=11`, `GET=56`, `POST=26`, `PUT=14`, `WEBHOOK=10`.
- Disposition totals: `stream=25`, `direct_read=23`, `write=14`, `blocked=55`.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: passed (`550 connector(s) checked, 0 findings`).
- `go test ./internal/connectors/conformance -run 'TestConformance/lever-hiring' -count=1`: passed.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`: passed after updating the tracked golden transcripts for the new Lever command listing.
- `go vet ./internal/connectors/... ./internal/cli/...`: passed.
- `go build ./cmd/pm`: passed.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`: passed.
- CLI smoke: `go run ./cmd/pm lever-hiring`, `go run ./cmd/pm lever-hiring opportunities`, `go run ./cmd/pm lever-hiring direct`, and `go run ./cmd/pm lever-hiring delete-feedback` all rendered credential-free help successfully.
- `make connector-boundary`: passed (`outcome clean`, `findings=0`, `warnings=0`).
- `git diff --check`: passed.
- `make verify`: passed.

No live Lever credentials, provider calls, provider writes, new dependencies, or shared runtime/engine/CLI Go edits were used.
