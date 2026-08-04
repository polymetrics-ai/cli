# Verification checklist: AWS CloudTrail parity resume r1

- [x] Verify the official AWS Actions page's 60-operation total and record its URL: <https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html> lists actions from `AddTags` through `UpdateTrail` (60 action links).
- [x] Capture red native-provider and runtime-help evidence before the delegate: `TestCommandSurfaceExposesDocumentedOperations` initially failed because `New()` did not implement `connectors.CommandSurfaceProvider`; the wrapped factory then separately failed the same assertion. `pm aws-cloudtrail --help` was absent before the delegate.
- [x] Add a native command-surface delegate and coverage test: `engine_delegate.go` delegates to the CloudTrail bundle; its contract test checks 60 command rows (57 implemented, 3 policy-disallowed).
- [x] Run `go run ./cmd/connectorgen surface-sync` and `go run ./cmd/connectorgen surface-sync --check`: the former repaired nine direct-read endpoint references from `operations.json`; the check subsequently scanned 550 connectors with zero drift.
- [x] Run `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail`: zero findings.
- [x] Run `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1`: passed.
- [x] Run `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1`: passed.
- [x] Run affected native tests: `go test ./internal/connectors/native/nativeset ./internal/connectors/native/aws-cloudtrail -count=1` passed. The new wrapper test has an individual red/green subtest for each enumerated optional runtime interface, plus explicit absence behavior.
- [x] Re-run the complete scoped `go test ./internal/cli/... -count=1` after regenerating root-help golden transcripts: passed in 388.154 seconds.
- [x] Run `go vet` for changed packages and `go build ./cmd/pm`: passed.
- [x] Build and run runtime commands without credentials: `./pm aws-cloudtrail --help` and bare `./pm aws-cloudtrail` both list 17 groups; `./pm help aws-cloudtrail` lists 57 implemented and 3 `unsafe_or_disallowed` commands; `query cancel --help` shows required typed `--query-id`; `tags add --help` retains plan -> preview -> explicit approval -> execute; `query start --help` explains the SQL-text policy refusal. Invocation without an initialized project stops at the local project guard, before any provider call.
- [x] Run `cd website && pnpm run gen:website-data`: it updated only the CloudTrail-driven connector catalog capability count/data.
- [x] Run generated help/docs gates: `make docs-check-no-build` and `make smoke-no-build` both passed. The smoke project's exact temporary directory was moved to Trash after verification.
- [x] Run `git diff --check`: passed. `connector-boundary` returned `outcome: clean`; all non-connector changes are generated help/catalog data, phase evidence, or the captain-authorized generic wrapper and its tests.

## Shared runtime finding

`definitionConnector` previously stored the native as the narrow `connectors.Connector` interface,
silently erasing `CommandSurfaceProvider` and all eleven optional runtime interfaces. This made a
definition look valid while the registry lost its executable capability. The authorized generic
forwarder preserves `DirectReader`, `OperationDirectReader`, `OperationBinaryDownloader`,
`WriteValidator`, `DryRunWriter`, `Querier`, `CDCReader`, `StatefulReader`, `SchemaMapper`,
`LiveConformanceProvider`, and `LocalWarehouseMaterializer`; the companion command-surface
provider is forwarded too. A uniform Go wrapper cannot conditionally preserve method sets for
future unknown interfaces, so known absent error-returning capabilities now produce typed
`connectors.ErrUnsupportedOperation` rather than disappearing silently; the availability/boolean
interfaces return their explicit false absence values.
