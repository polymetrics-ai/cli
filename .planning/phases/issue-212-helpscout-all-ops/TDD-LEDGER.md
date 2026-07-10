# TDD Ledger: issue-212 Help Scout all operations

## Red targets

- [ ] `go test ./internal/connectors/engine -run TestDirectReadAllowsGenericJSONOutputPolicy` fails before adding the generic JSON direct-read policy.
- [ ] `go test ./cmd/connectorgen -run TestHelpScoutAllOperationsCovered` fails while Help Scout mutation/direct-read coverage remains blocked/planned.

## Green targets

- [x] Generic `json` direct-read output policy accepted by schema, commandrunner, and engine.
- [x] Help Scout `api_surface.json` has 145 official unique endpoints with non-binary GETs covered by implemented direct-read commands or existing streams.
- [x] Help Scout non-binary mutation endpoints have typed write actions with explicit path fields and record schemas.
- [x] Binary/file endpoints are represented by bounded `binary_download` operation metadata and remain operation-gated until a safe executor exists.
- [x] Connector static validation and focused tests pass.

## Evidence log

### Red — 2026-07-10

```bash
go test ./internal/connectors/engine -run TestDirectReadAllowsGenericJSONOutputPolicy
go test ./cmd/connectorgen -run TestHelpScoutAllOperationsCovered
```

Results:

- `TestDirectReadAllowsGenericJSONOutputPolicy` failed with `direct read output policy "json" is not supported`.
- `TestHelpScoutAllOperationsCovered` failed because `internal/connectors/defs/help-scout/writes.json` is missing.

### Green — 2026-07-10

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json
go test ./internal/connectors/engine -run 'DirectRead|Operations|CLISurface'
go test ./cmd/connectorgen -run 'HelpScout|CLISurface|Operation'
go test ./internal/connectors/commandrunner -run 'DirectRead|Operation|WriteCommand'
go test ./internal/cli -run 'HelpScoutCommandSurface'
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: all passed. Current Help Scout operation coverage: 145 endpoint rows, 73 JSON direct-read commands, 66 typed reverse-ETL writes, 4 stream commands, 2 bounded binary operation rows.

### Red/green — Help renderer

Red:

```bash
go test ./internal/cli -run TestHelpScoutConnectorNamespaceRendersCommandSurfaceHelp
```

Result: failed because `pm help help-scout`, `pm help-scout`, and `pm help-scout --help` returned missing-help/missing-path errors.

Green:

```bash
go test ./internal/cli -run 'HelpScoutConnectorNamespaceRendersCommandSurfaceHelp|HelpScoutCommandSurface|Manual'
go build ./cmd/pm
./pm help help-scout
./pm help-scout
./pm help-scout --help
```

Result: passed; connector namespace/help now renders the connector manual and command surface without credentials.
