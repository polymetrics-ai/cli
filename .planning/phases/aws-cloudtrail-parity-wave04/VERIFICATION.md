# AWS CloudTrail parity wave04 verification checklist

No live provider calls or credentialed checks are part of this verification.

## Initial 3dd65b20d verification

The initial commit verified a 60/19/10/31 implementation with shared promoted-native support. That shared support was later ruled out of scope by the user, so the current corrective head preserves 60/19/0/0 with 41 blocked/planned operations by keeping resource-detail reads inside connector-local discovery/fan-out while shared direct/write forwarding remains blocked.

## Scope-corrected final surface

Final implemented/blocked counts:

- 60 official CloudTrail API actions inventoried exactly once.
- 19 implemented ETL/read streams.
- 0 implemented direct/provider query commands.
- 0 implemented reverse-ETL write/admin actions.
- 41 blocked/planned operations: 10 direct/provider query + 31 write/admin.
- 0 binary, 0 CDC, 0 excluded.

## Scope correction 2026-08-01

Verification after restoring shared command/direct/write files, reverting the promoted-native manifest wrapper, and reclassifying dependent CloudTrail commands:

- [x] The final head carries no shared-runtime change at all. The bundle-backed `Manifest()` override on the shared `definitionConnector` wrapper was extracted into standalone foundation PR #3676 (`fix(connectors): derive nativeset manifest from bundle definition`) because it changes every `withBundleDefinition`-wrapped connector (~30 today), not aws-cloudtrail-owned surface. CloudTrail command-surface, operation-direct-read, write-validation, and dry-run forwarding remain blocked/planned.
- [x] Consequence, recorded honestly: without #3676, `ManifestOf` falls back to a metadata-only manifest for every bundle-backed promoted native, so `pm connectors inspect aws-cloudtrail` renders no ETL STREAMS or authored CONFIGURATION/SECURITY detail. This manifest-fidelity gap is repo-wide and already present on `main` for all ~30 such connectors; it is not a regression introduced by this connector or by this revert, and it resolves for all of them when #3676 merges. `pm connectors catalog`, the generated catalog docs, and the website data read bundles directly and are unaffected.
- [x] `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` -> pass.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` -> pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` -> pass, 549 connectors checked, 0 findings.
- [x] `go build ./cmd/pm` -> pass.
- [x] `go run ./cmd/pm connectors catalog --json` -> AWS CloudTrail reports read-only, 19 streams, 0 write actions.
- [x] `go run ./cmd/pm connectors inspect aws-cloudtrail --json` -> runtime metadata reports `write=false` and 0 write actions. The manifest block does not enumerate the 19 streams, because the bundle-backed manifest wrapper is deferred to #3676; the authoritative 19-stream/0-write surface is verified from the bundle via `pm connectors catalog --json` and `pm etl catalog` instead.
- [x] `go run ./cmd/pm aws-cloudtrail --help` -> fails with `help topic "aws-cloudtrail" not found`, which is the truthful reduced help surface because `cli_surface.json` is removed.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` -> pass and refreshes root help to remove `pm aws-cloudtrail`.
- [x] `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` -> CloudTrail generated catalog/docs no longer claim 31 writes. `docs/connectors/aws-cloudtrail/MANUAL.md` and `SKILL.md` are byte-identical to generator output at this head, so a later `pm docs generate` is a no-op rather than silently overwriting hand edits. Both are correspondingly sparse (no ETL STREAMS/SYNC MODES, `No connector-specific config fields.`, generic `connector-specific` risk, and an incorrect `No secret authentication is required` line) because bundle-backed `Manifest()` forwarding is deferred to #3676; the `Generated manual fidelity` section of the bundle's `docs.md` records that and names the authoritative surfaces in the meantime.
- [x] `cd website && node scripts/gen-connector-bundles.mjs && node scripts/gen-connector-catalog.mjs` -> AWS CloudTrail website data reports 19 streams, 0 write actions, `cli_surface: null`.
- [x] Posted scope-correction addendum with marker `pm-aws-cloudtrail-wave04-r1-scope-correction` to issues #3142-#3149.
- [x] `make connector-boundary` -> clean.
- [x] `make verify` -> pass.
- [x] `git diff --check` -> pass.
- [x] `git diff --cached --check` -> pass.

Blocked shared-runtime dependencies documented by this scope correction:

1. Focused connector-dir validation depends on the reverted `cmd/connectorgen` shared enhancement; the final supported local gate is whole-defs validation.
2. Runtime-visible CloudTrail dynamic commands, direct/provider queries, and typed reverse-ETL write/admin actions depend on separate promoted-native command-surface, operation-direct-read, write-validation, and dry-run forwarding. They are now blocked/planned in the ledger and generated surfaces rather than claimed executable.

## Wave04-r1 correction 2026-08-04: cli_surface.json and executable direct-read/write parity

**The 60/19/0/0/41 surface above is superseded.** Captain review found that "60 operations dispositioned in `api_surface.json`, 0 reachable as `pm aws-cloudtrail <command>`" is a sync-only connector, which does not meet the parity bar — CI passing was never evidence the bar was met. The "shared promoted-native forwarding" blocker cited throughout this file for the 41 blocked/planned operations was re-audited and found to be inaccurate: `internal/connectors/native/aws-cloudtrail/aws_cloudtrail.go` already had fully generic `OperationDirectRead`/`ValidateWrite`/`DryRunWrite`/`Write` implementations (dispatching through two maps that were simply left empty), and `internal/connectors/hooks/aws-cloudtrail/hooks.go` already registered an `ExecuteWrite` hook delegating to them — no shared-runtime dependency existed. Activating them required only: populating `cloudTrailDirectOperations`/`cloudTrailWriteActions`/`cloudTrailDeleteActions` in `api_contract.go` (field schemas for all 60 actions already existed there, "DO NOT EDIT BY HAND" / AWS-doc-verified), authoring `cli_surface.json` (previously absent entirely) plus `operations.json`/`writes.json` entries, and reclassifying the corresponding `api_surface.json` endpoints from `operation` (blocked) to `covered_by`.

Per-operation re-audit against the official AWS CloudTrail API Reference (not the prior blanket "requires shared runtime" reason) found only 3 of 60 actions are genuinely blocked, each for the same specific, non-infrastructure reason: `StartQuery` accepts an unrestricted CloudTrail Lake SQL `QueryStatement` directly, and `CreateDashboard`/`UpdateDashboard` require one embedded in every `Widgets[].QueryStatement`. This project disables generic/unrestricted query-text execution for every connector by policy (`capabilities.query` fixed `false` repo-wide; no raw SQL passthrough tool is ever exposed, per `AGENTS.md`) — this is a standing project constraint, not a missing-infrastructure placeholder, and `api_surface.json` now records each with `operation.model: "disallowed"` and a `source_url` citation. The other 9 query/Insights-family actions (`CancelQuery`, `DescribeQuery`, `GenerateQuery`, `GetQueryResults`, `ListInsightsData`, `ListInsightsMetricData`, `ListQueries`, `LookupEvents`, `SearchSampleQueries`) take only typed identifiers/enums/bounded strings as input — none accept raw SQL — so they are genuinely safe and are now executable direct-read commands.

Corrected final surface:

- 60 official CloudTrail API actions inventoried exactly once.
- 19 implemented ETL/read streams (unchanged).
- 9 implemented direct-read commands (`cancel_query`, `describe_query`, `generate_query`, `get_query_results`, `list_insights_data`, `list_insights_metric_data`, `list_queries`, `lookup_events`, `search_sample_queries`).
- 29 implemented reverse-ETL write/admin actions (10 destructive with `confirm: destructive` and provider-supported idempotency; 19 non-destructive).
- 3 blocked, each with a specific `disallowed`-model, source_url-backed reason (`StartQuery`, `CreateDashboard`, `UpdateDashboard`).
- 0 binary, 0 CDC, 0 excluded.

Verification:

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail` -> pass, 1 connector checked, 0 findings.
- [x] `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` -> pass, including new `TestNativeCloudTrailDirectReadDispatchesOperationTarget`, `TestNativeCloudTrailWriteDispatchesActionTarget`, `TestNativeCloudTrailQueryTextOperationsStayBlocked`, and updated `TestOperationLedgerCounts` (9/29/3).
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` -> pass, including `write_request_shape:*` for all 29 writes against the real `ExecuteWrite` hook (real native dispatch, not a declarative fiction — confirmed writes POST to `/` with the `CloudTrail_20131101.<Action>` `X-Amz-Target` header, matching `aws_cloudtrail.go`'s actual protocol; `writes.json`'s documentation-only `path` field is never dispatched to literally).
- [x] `go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/certify ./internal/connectors/boundary -count=1` -> pass.
- [x] `go test ./internal/cli -count=1` -> pass (no golden-transcript change; `pm aws-cloudtrail --help` still correctly reports no dynamic command surface, since `cli_surface.json` is docs/validation metadata only and never drives live command dispatch).
- [x] `go build ./cmd/pm` -> pass.
- [x] `./pm docs generate --dir docs/cli` (scoped restore of all other connectors' docs afterward; this is known pre-existing repo-wide generator drift unrelated to this change, see prior no-mistakes finding `docs-generate-repo-wide-drift`) -> `docs/connectors/aws-cloudtrail/MANUAL.md`/`SKILL.md` regenerated with `write=true`; still sparse (no ETL STREAMS/SYNC MODES section) because bundle-backed `Manifest()` forwarding remains deferred to #3676, which is still open/unmerged as of this correction — unchanged, pre-existing, out-of-scope gap, not reintroduced or worsened by this change.
- [x] `cd website && node scripts/gen-connector-bundles.mjs && node scripts/gen-connector-catalog.mjs` -> AWS CloudTrail website data updated to 19 streams / 9 direct reads / 29 writes; diffed against origin to confirm only the AWS CloudTrail entry changed (`Connectors with write actions: 229 -> 230`).
- [x] `make connector-boundary` -> clean.
- [x] `make verify` -> pass.
- [x] `git diff --check` -> pass.
