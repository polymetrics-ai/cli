# AWS CloudTrail parity resume r1 plan

## GSD workflow

- `scripts/gsd doctor` passed.
- The required non-interactive command `scripts/gsd prompt programming-loop init --phase aws-cloudtrail-parity-resume-r1 --dry-run` is unavailable in this checkout (`unknown GSD command: programming-loop`). This phase therefore uses the documented manual-GSD fallback from `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-documentation`.

## Scope and ownership

This lane owns only AWS CloudTrail connector files under `internal/connectors/defs/aws-cloudtrail/`, `internal/connectors/native/aws-cloudtrail/`, `internal/connectors/hooks/aws-cloudtrail/`, connector-generated documentation/data when the scoped generator changes it, and this phase's artifacts. The target connector is exactly `aws-cloudtrail`.

The captain approved one narrow exception after runtime reproduction: `internal/connectors/native/nativeset/promoted.go` may receive only connector-agnostic optional-runtime forwarders. `definitionConnector` otherwise erases native optional capabilities before the CLI registry sees them. No other shared path is authorized.

The enumerated runtime interfaces declared by `internal/connectors/connectors.go` are `DirectReader`, `OperationDirectReader`, `OperationBinaryDownloader`, `WriteValidator`, `DryRunWriter`, `Querier`, `CDCReader`, `StatefulReader`, `SchemaMapper`, `LiveConformanceProvider`, and `LocalWarehouseMaterializer`. `CommandSurfaceProvider` is the companion CLI provider in `command_surface.go`. `DefinitionProvider` remains intentionally supplied by the wrapper's bundle-backed implementation; `ManifestProvider` and `GuideProvider` are documentation metadata providers rather than runtime dispatch interfaces and are not changed by this authorization.

Do not alter shared engine/runtime/schema files, including `internal/connectors/engine/write.go`, `internal/connectors/engine/schema/writes.schema.json`, `internal/connectors/engine/schema/operations.schema.json`, the connector definition validator, or `docs/migration/conventions.md`.

## Rehydration and source inventory

- Preserved head `3159b798` was rehydrated and rebased onto current `origin/main` (`36b431cf`) in this isolated worktree. The rebase kept current generated AWS CloudTrail docs/catalog output where the old merge carried stale generated content.
- AWS's official [CloudTrail Actions](https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html) page lists 60 actions (action links 4 through 63). This is the documented-operation total for this lane.
- The rehydrated bundle already inventories 60 command rows: 57 `implemented` (19 ETL streams, 9 typed direct reads, 29 typed reverse-ETL writes) and 3 `unsafe_or_disallowed` (not planned).
- `StartQuery`, `CreateDashboard`, and `UpdateDashboard` remain unavailable because their request model admits unrestricted CloudTrail Lake SQL `QueryStatement` text. This is a standing provider-operation policy boundary: no raw SQL passthrough, generic HTTP write, or raw AWS action escape hatch.

## Implementation plan

1. Capture red evidence that the native connector does not implement `connectors.CommandSurfaceProvider`; consequently `pm aws-cloudtrail --help` is rejected as an unknown command despite a valid `cli_surface.json`.
2. Add the established connector-local native-to-engine command-surface delegate, loading only the AWS CloudTrail bundle. Add a native contract test that checks all 60 command rows are exposed with the 57/3 implemented/disallowed split.
3. Add a generic `definitionConnector` forwarding test and the single authorized forwarder so every enumerated native optional runtime interface survives registry wrapping, or returns a typed `ErrUnsupportedOperation` when the wrapped connector does not implement it.
4. Build `pm` and run runtime help/bare/representative direct-read/reverse-ETL commands without credentials or provider writes. The results must prove command registration and preserve the plan -> preview -> approval -> execute write gate.
5. Complete the provider-field research matrix in `traces/` without changing shared citation schema: every request field must reference its operation's AWS API Reference Request Parameters section, evidence type, confidence, and requiredness rationale. Rebase the eventual shared citation convention before final validation if it lands.
6. Run the connector-local and current-main gates specified by the parity resume contract, regenerate only scoped generated data if it changes, then commit the green slice.

## Safety constraints

- No credentialed CloudTrail calls, secret values, live writes, or reverse-ETL execution.
- All command inputs remain fixed typed mappings from the bundle; no raw method/path/header/body/SQL escape hatch.
- Destructive writes retain typed confirmation and the normal reverse-ETL plan/preview/approval/execute lifecycle.
