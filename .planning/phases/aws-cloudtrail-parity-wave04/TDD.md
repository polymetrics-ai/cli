# AWS CloudTrail parity wave04 TDD ledger

## Red targets before production edits

- `go test ./internal/connectors/native/aws-cloudtrail -run 'TestOperationLedger|TestNativeCloudTrailJSONRPC|TestConnectorContract' -count=1`
  - Expected initially: fail because only 4 legacy streams and no direct/write operation contract exist.
- `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail`
  - Expected initially: legacy command shape treats nested `schemas/` and `fixtures/` as connector roots. Scope-corrected final uses whole-defs validation because the focused connector-dir validator enhancement was reverted with the six shared files.
- `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`
  - Expected initially: skip dynamic due legacy marker; final expected: pass static and dynamic replay.

## Green criteria

- Connector-local contract tests prove:
  - 60 official API operations are enumerated exactly once.
  - 9 stream-backed read operations are implemented; 10 parameterized read operations, 10 direct/provider-query operations, and 31 write/admin actions are blocked/planned.
  - Every executable operation has a sanitized fixture or focused native replay test.
  - Blocked direct/write operations are not listed in executable `operations.json`, `writes.json`, generated catalog write actions, or dynamic CLI help.
  - Implemented reads use fixed AWS CloudTrail JSON-RPC action names and never accept raw AWS action names, raw paths, raw headers, or raw request bodies.
- Required local gates pass per task.

## Evidence log

- Setup and source/audit read complete.
- Red test added: `go test ./internal/connectors/native/aws-cloudtrail -run 'TestOperationLedger|TestNativeCloudTrailJSONRPC|TestNativeCloudTrailWrite' -count=1` failed because `OperationDirectRead`, `cloudTrailTarget`, and the 60-op bundle did not exist yet.
- Initial green implementation added the 60/19/10/31 bundle, native JSON-RPC dispatch, hook write/read/check delegation, promoted-native interface forwarding, bundle registry caching, and focused connector-dir validation.
- Scope correction reverted the shared interface forwarding/validator/cache edits and reclassified the exposed CloudTrail surface to 9 implemented streams plus 51 blocked/planned operations.
- Focused green tests:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json`
  - `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`
  - `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1`
  - `go test ./internal/connectors/commandrunner -count=1`
  - `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestAWSCloudTrailCommandSurfaceHelpScopes|TestRootHelpListsDynamicConnectorCommands' -count=1`

## Scope correction 2026-08-01

Correction target: preserve CloudTrail-owned parity work while proving the six shared files are restored to pre-task contents in a new commit.

Expected red risk after restoring shared files:

- Focused `connectorgen validate internal/connectors/defs/aws-cloudtrail --json` may regress because the removed shared validator enhancement accepted a connector directory directly.
- Runtime command-surface or manifest checks expose a genuine shared-runtime dependency if promoted native connectors cannot surface bundle-owned `Manifest`, `CommandSurface`, `OperationDirectRead`, `ValidateWrite`, `DryRunWrite`, or `InitialState` without edits in `nativeset/promoted.go` and `engine/connector.go`.

Green criteria:

- Focused native/hook/conformance checks and `make verify` pass with only the six shared files restored and no shared edits reintroduced.
- The operation ledger and generated docs/catalog/help show 9 implemented streams, 0 executable direct reads, 0 executable writes, and 51 blocked/planned operations.

Evidence after restoring the six shared files:

- `git diff --exit-code HEAD^ -- <six shared files>` passed, proving the six files match pre-task contents.
- `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json` failed because reverted `connectorgen` no longer treats a connector dir as a single bundle and instead checks nested `fixtures/` and `schemas/` as connector roots. Full `go run ./cmd/connectorgen validate internal/connectors/defs` is the final validator gate and passes.
- `go run ./cmd/pm connectors catalog --json` reports AWS CloudTrail as read-only with 9 streams and 0 write actions.
- `go run ./cmd/pm connectors inspect aws-cloudtrail --json` reports runtime metadata `write=false` and manifest 0/0 because promoted-native manifest forwarding is reverted; connector docs note that ETL streams remain reachable through `pm etl`/catalog surfaces.
- `go run ./cmd/pm aws-cloudtrail --help` fails with `help topic "aws-cloudtrail" not found`, which is now the truthful help surface because `cli_surface.json` was removed.
