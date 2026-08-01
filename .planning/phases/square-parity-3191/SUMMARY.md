# Square parity #3191 — summary

Implemented final-wave Square parity from the official Square OpenAPI latest document served by developer.squareup.com.

## Counts

| total official operations | implemented | fixture-backed | blocked/planned | excluded | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 346 | 294 | 263 | 52 | 0 | 0 |

Implemented breakdown: 110 streams, 31 bounded direct reads/searches, 153 typed reverse-ETL write actions. Fixture-backed breakdown: 110 stream fixture page sets and 153 write fixtures. Blocked breakdown: 25 deprecated, 15 CDC/webhook/event, 6 OAuth/mobile authorization lifecycle, 5 multipart/file/attachment, 1 generic V1 batch passthrough.

## Files

- `internal/connectors/defs/square/api_surface.json` — complete 346-operation ledger.
- `internal/connectors/defs/square/streams.json`, `schemas/*.json`, `fixtures/streams/**` — stream parity.
- `internal/connectors/defs/square/writes.json`, `fixtures/writes/**` — typed reverse ETL parity.
- `internal/connectors/defs/square/operations.json`, `cli_surface.json` — bounded direct read/search surfaces and CLI metadata.
- `internal/connectors/defs/square/docs.md`, `certification.json` — docs and truthful fixture-only certification evidence.
- `internal/cli/testdata/golden_transcripts.json`, `website/data/connectors.generated.json`, `website/lib/connectors.catalog*.generated.*` — generated Square discovery/docs surfaces.

## Safety

No credentials, no live Square calls, no provider writes, no new dependencies, no shared runtime edits. Generic batch passthrough remains blocked. Binary/multipart and CDC/webhook/event lifecycle operations remain blocked behind shared executor/foundation dependencies. Reverse ETL remains plan -> preview -> approval -> execute.

## Verification

Final `make verify` passed. See `VERIFICATION.md` for focused gates and timeout evidence.
