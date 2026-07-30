# Verification — Issue #1950 Lucid ELD Operation Ledger

## Required commands

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/lucid-eld
go test ./internal/connectors/conformance -run 'TestConformance/lucid-eld' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

## Additional planning validator commands

```bash
python3 .planning/issue-775/1950/tools/validate_surface.py --surface .planning/issue-775/1950/fixtures/red/missing-endpoints.api_surface.json --openapi .planning/issue-775/1950/evidence/openapi-doc.json
python3 .planning/issue-775/1950/tools/validate_surface.py --surface internal/connectors/defs/lucid-eld/api_surface.json --openapi .planning/issue-775/1950/evidence/openapi-doc.json
```

## Results

| Command | Result | Exact output / note |
|---|---|---|
| `scripts/gsd doctor` | pass | captured in session; all adapter resources ok |
| `scripts/gsd prompt programming-loop init --phase issue-775-1950-lucid-eld-operation-ledger --dry-run` | fallback | `scripts/gsd: unknown GSD command: programming-loop` |
| red fixture validator | pass-red | `FAIL .planning/issue-775/1950/fixtures/red/missing-endpoints.api_surface.json: 1 error(s)` / `- missing official endpoint GET /v2/company-info` / `exit=1` |
| negative fixtures validator | pass-negative | duplicate, invalid category, stale review, unknown target, and wildcard fixtures each exited 1 with intended failure text; wildcard also reported missing official endpoint and extra wildcard endpoint |
| final planning validator | pass | `PASS internal/connectors/defs/lucid-eld/api_surface.json: 8 endpoint(s) match official OpenAPI` |
| `go run ./cmd/connectorgen validate internal/connectors/defs/lucid-eld` | pass | `connectorgen validate: 0 connector(s) checked, 0 findings` / `exit=0` |
| `go test ./internal/connectors/conformance -run 'TestConformance/lucid-eld' -count=1` | fail-expected | bundle load failed because #1950 scope permits only `api_surface.json`; exact output below |
| `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` | pass | `ok  	polymetrics.ai/internal/cli	138.429s` / `exit=0` |
| `go vet ./internal/connectors/... ./internal/cli/...` | pass | no stdout/stderr / `exit=0` |
| `go build ./cmd/pm` | pass | no stdout/stderr / `exit=0` |
| `make connector-boundary` | fail-expected | boundary scanner loads connector metadata and fails until #1951 adds `metadata.json`; exact output below |
| `git diff --check` | pass | no output / `exit=0` |

## Exact expected failure output

### `go test ./internal/connectors/conformance -run 'TestConformance/lucid-eld' -count=1`

```text
--- FAIL: TestConformance (1.77s)
    --- FAIL: TestConformance/lucid-eld (0.00s)
        conformance_test.go:95: bundle lucid-eld failed to load: load bundle lucid-eld: missing required file metadata.json
FAIL
FAIL	polymetrics.ai/internal/connectors/conformance	2.628s
FAIL
exit=1
```

### `make connector-boundary`

```text
go run ./cmd/connectorgen boundary . --json
connectorgen boundary: load connector metadata lucid-eld: open /Users/karthiksivadas/Development/polymetrics-cli-agents/add_terminal-connector-issues/worktrees/lucid-eld-children/1950-operation-ledger/internal/connectors/defs/lucid-eld/metadata.json: no such file or directory
exit status 2
make: *** [connector-boundary] Error 1
exit=2
```

## Incomplete-bundle expectation

This issue is operation-ledger only. The Lucid ELD bundle does not yet include `metadata.json`, `spec.json`, `streams.json`, schemas, fixtures, `cli_surface.json`, or docs. Conformance and boundary failures above are expected incomplete-bundle failures for #1951-#1955, not #1950 ledger defects.
