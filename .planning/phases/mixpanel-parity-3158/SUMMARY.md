# Summary — Mixpanel connector parity (#3158)

## What changed

- Rebuilt Mixpanel from the official OpenAPI YAML inventory with 105 operation rows in `api_surface.json` and matching `operations.json` metadata.
- Added 25 fixture-backed read/changefeed-like streams, 15 bounded JSON direct-read operation commands, and 60 typed reverse-ETL write actions.
- Kept 5 official operations truthfully blocked/planned: JQL/raw script, two form/query POST reads, raw event export, and CSV lookup-table replacement.
- Added closed write schemas, destructive confirmation metadata, idempotent 404 handling for delete/destructive actions, redaction fields, stream/write fixtures, and a direct-read replay test for Mixpanel.
- Regenerated Mixpanel connector docs, catalog data, website connector data, and CLI golden transcripts.
- Appended the captain-policy destructive-operation addendum idempotently to #3158-#3165 with actual local counts.

## Counts

| Metric | Count |
| --- | ---: |
| Official operations | 105 |
| Implemented operations | 100 |
| Stream replay fixtures | 25 |
| Write fixtures | 60 |
| Direct-read replay subtests | 15 |
| Blocked/planned operations | 5 |
| Excluded/not-applicable | 0 |
| Certified live operations | 0 |

## Safety

- No live Mixpanel credentials, provider calls, provider writes, certification, VPS/Thaalam, pushes, PRs, or merges.
- No generic HTTP/path/body, JQL/script, raw CSV, shell, file, or binary passthrough was exposed.
- Reverse ETL remains plan → preview → explicit approval → execute.

## Verification

Passed:

- Official inventory comparison: 105/105 rows, missing 0, extra 0, duplicate IDs 0.
- `go run ./cmd/connectorgen validate internal/connectors/defs`.
- Temp defs-root Mixpanel-only `connectorgen validate`.
- `go test ./internal/connectors/conformance -run 'TestConformance/mixpanel|TestMixpanelOperationDirectReadsReplay' -count=1`.
- Focused CLI/golden tests.
- `go build ./cmd/pm`.
- `./pm docs validate --connectors-dir docs/connectors`.
- `make connector-boundary`.
- `git diff --check`.
- Mixpanel help smoke commands.

Attempted but not green:

- `make verify` reached the repository-wide `go test -timeout 20m ./...` phase and timed out in long-running `internal/cli` and `internal/connectors/certify` packages on this machine. Focused Mixpanel gates passed, and `internal/cli` passed separately with `-timeout 35m`.
