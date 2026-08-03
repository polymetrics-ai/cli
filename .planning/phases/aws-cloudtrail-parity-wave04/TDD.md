# AWS CloudTrail parity wave04 TDD ledger

> **2026-08-04 correction**: the 19/0/0/41 counts and "shared promoted-native forwarding" blocker
> reasoning throughout this file are superseded. See VERIFICATION.md's "Wave04-r1 correction
> 2026-08-04" section for the corrected 19/9/29/3 surface, the re-audited genuine blockers
> (query-text policy, not missing infrastructure), and updated red/green targets
> (`TestNativeCloudTrailDirectReadDispatchesOperationTarget`,
> `TestNativeCloudTrailWriteDispatchesActionTarget`,
> `TestNativeCloudTrailQueryTextOperationsStayBlocked`, `write_request_shape:*` conformance checks).

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
  - 19 stream-backed read operations are implemented; 10 direct/provider-query operations and 31 write/admin actions are blocked/planned.
  - Every executable operation has a sanitized fixture or focused native replay test.
  - Blocked direct/write operations are not listed in executable `operations.json`, `writes.json`, generated catalog write actions, or dynamic CLI help.
  - Implemented reads use fixed AWS CloudTrail JSON-RPC action names and never accept raw AWS action names, raw paths, raw headers, or raw request bodies.
- Required local gates pass per task.

## Evidence log

- Setup and source/audit read complete.
- Red test added: `go test ./internal/connectors/native/aws-cloudtrail -run 'TestOperationLedger|TestNativeCloudTrailJSONRPC|TestNativeCloudTrailWrite' -count=1` failed because `OperationDirectRead`, `cloudTrailTarget`, and the 60-op bundle did not exist yet.
- Initial green implementation added the 60/19/10/31 bundle, native JSON-RPC dispatch, hook write/read/check delegation, promoted-native interface forwarding, bundle registry caching, and focused connector-dir validation.
- Scope correction reverted the shared interface forwarding/validator/cache edits and reclassified the exposed CloudTrail surface to 19 implemented streams plus 41 blocked/planned operations.
- Focused green tests:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json`
  - `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`
  - `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1`
  - `go test ./internal/connectors/commandrunner -count=1`
  - `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestAWSCloudTrailCommandSurfaceHelpScopes|TestRootHelpListsDynamicConnectorCommands' -count=1`

## Scope correction 2026-08-01

Correction target: preserve CloudTrail-owned parity work while keeping shared command/direct/write forwarding out of this branch.

Expected red risk after restoring shared command/direct/write files:

- Focused `connectorgen validate internal/connectors/defs/aws-cloudtrail --json` may regress because the removed shared validator enhancement accepted a connector directory directly.
- Runtime command-surface, operation-direct-read, write-validation, or dry-run checks expose a genuine shared-runtime dependency without promoted-native `CommandSurface`, `OperationDirectRead`, `ValidateWrite`, `DryRunWrite`, or `InitialState` forwarding in the shared runtime.

Green criteria:

- Focused native/hook/conformance checks and `make verify` pass without reintroducing shared command/direct/write forwarding.
- The operation ledger and generated docs/catalog/help show 19 implemented streams, 0 executable direct reads, 0 executable writes, and 41 blocked/planned operations.

Evidence after restoring shared command/direct/write files and reverting the manifest wrapper:

- The final head keeps no shared-runtime change. The bundle-backed promoted-native `Manifest()` override was extracted into standalone foundation PR #3676 because it affects every bundle-based connector (~30 today) rather than aws-cloudtrail-owned surface, so this connector must not claim catalog/inspect manifest truthfulness it does not deliver. The resulting metadata-only manifest for promoted natives is a repo-wide gap that already exists on `main` and is not a regression from this connector or this revert. CloudTrail command-surface, operation-direct-read, write-validation, and dry-run forwarding remain blocked/planned.
- `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json` failed because reverted `connectorgen` no longer treats a connector dir as a single bundle and instead checks nested `fixtures/` and `schemas/` as connector roots. Full `go run ./cmd/connectorgen validate internal/connectors/defs` is the final validator gate and passes.
- `go run ./cmd/pm connectors catalog --json` reports AWS CloudTrail as read-only with 19 streams and 0 write actions.
- `go run ./cmd/pm connectors inspect aws-cloudtrail --json` reports runtime metadata `write=false` and 0 write actions; it does not enumerate the 19 streams, because bundle-backed manifest forwarding is deferred to #3676. `pm connectors catalog --json` is the bundle-backed check for the 19-stream surface.
- `go run ./cmd/pm aws-cloudtrail --help` fails with `help topic "aws-cloudtrail" not found`, which is now the truthful help surface because `cli_surface.json` was removed.
