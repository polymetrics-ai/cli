# TDD Ledger: issue-212 Help Scout all operations

## Red targets

- [ ] `go test ./internal/connectors/engine -run TestDirectReadAllowsGenericJSONOutputPolicy` fails before adding the generic JSON direct-read policy.
- [ ] `go test ./cmd/connectorgen -run TestHelpScoutAllOperationsCovered` fails while Help Scout mutation/direct-read coverage remains blocked/planned.

## Green targets

- [ ] Generic `json` direct-read output policy accepted by schema, commandrunner, and engine.
- [ ] Help Scout `api_surface.json` has 145 official unique endpoints with non-binary GETs covered by implemented direct-read commands or existing streams.
- [ ] Help Scout non-binary mutation endpoints have typed write actions with explicit path fields and record schemas.
- [ ] Binary/file endpoints are represented by bounded `binary_download` operation metadata and remain operation-gated until a safe executor exists.
- [ ] Connector static validation and focused tests pass.

## Evidence log

Pending.
