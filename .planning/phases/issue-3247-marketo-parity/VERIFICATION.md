# VERIFICATION — issue #3247 Marketo parity

## Focused commands

```bash
go test ./cmd/connectorgen -run TestMarketo -count=1
# ok polymetrics.ai/cmd/connectorgen

go run ./cmd/connectorgen validate internal/connectors/defs --json
# findings: 0

go test ./internal/connectors/conformance -run 'TestConformance/marketo' -count=1
# ok polymetrics.ai/internal/connectors/conformance

go test ./internal/cli -run 'Connector|Dynamic|Golden|Marketo' -count=1 -timeout=5m
# ok polymetrics.ai/internal/cli
```

## CLI/help/docs parity checks

```bash
./pm help connectors
./pm connectors
./pm connectors inspect marketo --json
./pm marketo --help
./pm marketo etl get-email-by-id --help
./pm marketo direct get-are-leads-member-of-list --help
./pm marketo reverse update-email --help
```

Observed: contextual help rendered successfully, Marketo inspect reports `read=true`, `write=true`, `query=false`, 117 streams, and `access_token` as the only secret field. ETL path-param help includes `--config id=...`; direct path-param help includes `--list-id`.

## Broader/local gates

```bash
go vet ./...
go build ./cmd/pm
make connector-boundary
git diff --check
make verify
```

Results:

- `go vet ./...`: passed.
- `go build ./cmd/pm`: passed.
- `make connector-boundary`: clean report, 0 findings.
- `git diff --check`: passed.
- First `make verify` attempt hit an unrelated flaky timing assertion in `internal/connectors/certify` (`TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit`); isolated rerun of that test passed.
- Final `make verify` rerun passed through `homebrew release notification assertions passed`; connectorgen validate reported `550 connector(s) checked, 0 findings`, lint reported `0 issues`, smoke passed.

## Review

- Local reviewer pass initially found unsafe Marketo write-query paths, under-required destructive schemas, response-only write fields, and duplicate/missing CLI guidance.
- All blocking findings were fixed.
- Final reviewer check: PASS for count-ledger consistency and no prior blocker resurfaced.
