# Reddit full-parity completion verification

## Ledger outcome

- 230 concrete documented URI-variant rows in `api_surface.json`.
- 225 rows are covered and five are named provider exclusions.
- 50 ETL streams, 122 write actions, and 224 implemented CLI commands.
- `unsafe_or_disallowed` rows: 0.

The five exclusions are machine-checkable in `api_surface.json`:

1. `POST /api/store_visits`: Reddit Premium subscription required.
2. `POST /api/block_user`: approved OAuth application required.
3. `GET /api/recommend/sr/{srnames}`: deprecated by Reddit.
4. `GET /api/needs_captcha`: legacy captcha-flow, non-data endpoint.
5. `GET [/r/subreddit]/comments/{article}`: superseded by the existing disclosed sub-wide comments stream.

## TDD evidence

### Red:

- Before adding the four residual actions, `go test -timeout 20m ./internal/connectors/engine -run TestRedditResidualWritesAreExecutable -count=1` failed because `emoji_asset_upload_s3`, `upload_sr_img`, `widget_image_upload_s3`, and `vote` were absent.
- Before declaring the 100 RPM policy, `TestRedditEnforcesItsDocumentedRateLimit` failed because both stream and inspected metadata rate limits were absent.
- After declaring `batchable: false` but before listing its intentional adopters, `TestEveryShippedWriteActionHasExpectedBatchability` failed for Reddit's three bounded image commands and `vote`.

### Green:

- `go test -timeout 20m ./internal/connectors/engine -run '^(TestEveryShippedWriteActionHasExpectedBatchability|TestRedditResidualWritesAreExecutable|TestRedditEnforcesItsDocumentedRateLimit)$' -count=1` passed.
- `go test -timeout 20m ./internal/connectors/hooks/reddit -count=1` passed, including lease acquisition, bounded multipart transfer, OAuth-header isolation, and hostile-lease rejection.
- `go test -timeout 300s ./internal/connectors/conformance/... -run 'TestConformance$/reddit' -v -count=1` passed.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` passed across every implemented command.

## Rebase and runtime checks

- Rebased onto `f96a47e80` (`fix(cli): reject unresolved connector help paths (#3964)`) before creating a PR; regenerated docs/catalog artifacts afterward.
- A rebuilt `pm` exercised representative help for all 19 command groups and the four residual commands. The unresolved `pm reddit definitely-not-a-reddit-command --help` path exited 2, proving it no longer succeeds by rendering the connector manual.
- A temporary fixture project exercised `pm reddit vote --credential reddit-fixture --id t3_fixture --dir 1 --json`; it created a one-record command plan with the `destructive` confirmation challenge and a `batchable: false` plan seal, without a provider call.

## Local gates

- `gofmt` completed for all changed Go files.
- `go vet ./internal/connectors/engine ./internal/connectors/hooks/reddit ./internal/connectors/commandrunner ./internal/cli` passed.
- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/hooks/reddit ./internal/connectors/commandrunner` and `go test -timeout 20m ./internal/cli -count=1` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs/reddit --json`, `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`, and `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors` passed.
- `make tidy-check`, `make docs-check`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, `make smoke-no-build`, and `make lint` passed. The connector-boundary report was clean.

The repository guidance keeps the full `go test ./...` and `make verify` aggregate runs for CI because they exceed the per-command agent timeout; their component local gates above were run individually.
