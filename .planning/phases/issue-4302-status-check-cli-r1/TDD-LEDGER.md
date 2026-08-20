# TDD Ledger — PR #4308 status-check result preservation

## Planned red → green evidence

### Slice 1 — status JSON and human boundary

- Red: `go test -count=1 -timeout 20m ./internal/cli -run '^TestWriteConnectorCommandResult'` first failed to compile because `writeConnectorCommandResult` did not exist. The fixture requires the exact output boundary to accept a typed `commandrunner.Result.StatusCheck`; before the production branch it cannot express the expected dedicated JSON or human representation.
- Behavioral red: with only the new generic `StatusCheck` branch temporarily removed, the same focused test failed exactly as the qualification observed: JSON was `ConnectorCommandRead` with zero-valued status fields, and human output was empty for both `200` and `503`. The branch was restored immediately after the captured failure.
- Green: extracted the existing binary/direct-read/fallback output branches unchanged into `writeConnectorCommandResult` and added one generic `StatusCheck` branch. `go test -count=1 -timeout 20m ./internal/cli -run '^TestWriteConnectorCommandResult'` passed. JSON asserts exact `api_version`, `ConnectorCommandStatusCheck`, connector, command, operation, method, path, status, and `body_bytes` fields.

### Slice 2 — edge classification and retention

- Red: the first targeted run also could not reach the test assertions because the shared output function did not exist; the absent branch is therefore the causal failure, not a test-only helper mismatch.
- Green: the status fixture intentionally carries a non-empty stream, count, and ETL row. The JSON assertion rejects all three fallback fields, proving `StatusCheck` wins its result branch. Human subtests exactly assert success (`200`) and non-200 (`503`) lines, each retaining the zero-byte HEAD field. The binary-download fixture retains its kind, operation, filename, size, SHA-256, and truncation fields unchanged.

### Slice 3 — closed loader and live proof

- Red/green source closure remains in the existing #4302 loader ledger: malformed `rest_status` declarations are refused in `engine.Load` before I/O. This remediation reran `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestBundleLoadRegistersStatusAndTextExportOperations|TestBundleLoadRejectsInvalidStatusAndTextExportDeclarations)$'`, which passed, and added no runtime override or connector-specific route.

## Resource note

- A newer standalone `internal/cli` package run (`73382` → `75699`) was revalidated against the qualified head and found already exited before its explicitly authorized cancellation could be sent. It is superseded and is not used as verification evidence. The older chained run (`57896` → `57973` → `69549`) was preserved and had also exited; its uncaptured terminal result is likewise not claimed as evidence. Subsequent checks are serial and use captured output only.
