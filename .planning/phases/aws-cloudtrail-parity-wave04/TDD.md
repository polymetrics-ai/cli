# AWS CloudTrail parity wave04 TDD ledger

## Red targets before production edits

- `go test ./internal/connectors/native/aws-cloudtrail -run 'TestOperationLedger|TestNativeCloudTrailJSONRPC|TestConnectorContract' -count=1`
  - Expected initially: fail because only 4 legacy streams and no direct/write operation contract exist.
- `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail`
  - Expected initially: legacy command shape treated nested `schemas/` and `fixtures/` as connector roots; final expected after small connectorgen fix: one connector checked and zero findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`
  - Expected initially: skip dynamic due legacy marker; final expected: pass static and dynamic replay.

## Green criteria

- Connector-local contract tests prove:
  - 60 official API operations are enumerated exactly once.
  - 19 stream-backed read operations, 10 bounded direct/provider-query operations, 31 typed write actions.
  - Every executable operation has a sanitized fixture or focused native replay test.
  - Direct/write operations use fixed AWS CloudTrail JSON-RPC action names and never accept raw AWS action names, raw paths, raw headers, or raw request bodies.
  - Write previews and errors redact configured fields; destructive delete/admin actions are typed and approval-gated by metadata.
- Required local gates pass per task.

## Evidence log

- Setup and source/audit read complete.
- Red test added: `go test ./internal/connectors/native/aws-cloudtrail -run 'TestOperationLedger|TestNativeCloudTrailJSONRPC|TestNativeCloudTrailWrite' -count=1` failed because `OperationDirectRead`, `cloudTrailTarget`, and the 60-op bundle did not exist yet.
- Green implementation added the 60/19/10/31 bundle, native JSON-RPC dispatch, hook write/read/check delegation, promoted-native interface forwarding, bundle registry caching, and focused connector-dir validation.
- Focused green tests:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json`
  - `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`
  - `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1`
  - `go test ./internal/connectors/commandrunner -count=1`
  - `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestAWSCloudTrailCommandSurfaceHelpScopes|TestRootHelpListsDynamicConnectorCommands' -count=1`
