# TDD Ledger — PR 3712 Connector Validation Honesty

Manual-GSD fallback: `scripts/gsd prompt programming-loop` is absent from the repo-local command
registry (`scripts/gsd: unknown GSD command: programming-loop`, with `scripts/gsd doctor` otherwise
green at 69 commands), so this ledger records the manual GSD/TDD loop per `PLAN.md`.

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Validator/runtime drift closed | Red: 174 commands carried `availability: implemented` while `commandrunner.Preflight` blocked them, because `cmd/connectorgen/validate.go` exempted operation-backed `direct_read` commands from the `api_surface` assertion and skipped an empty `output_policy` instead of reporting it. | Green: both checks assert unconditionally; `go run ./cmd/connectorgen validate internal/connectors/defs` reports `550 connector(s) checked, 0 findings`. | Green (commit `e56a5950b`) |
| Runtime-truth sweep | Red: the drift above was invisible to every existing test, because each one described the runtime instead of calling it. | Green: `TestEveryImplementedCommandPassesRuntimePreflight` walks every bundle in `defs.FS` through the real `commandrunner.Preflight` — the same entry point `internal/cli` uses — so a new executor kind is covered the day it lands. | Green (commit `e56a5950b`) |
| Binary download routing | Red: the last 8 dead commands could not be made honest without an executor; `binary_download` had no runtime path. | Green: `engine.OperationBinaryDownload` + `commandrunner` `binary_download` routing, covered by `TestRunBinaryDownloadCommandPassesDestinationThrough`, `TestRunBinaryDownloadRequiresDestinationRoot`, `TestPreflightBlocksBinaryDownloadWithUnsafeMetadata`, `TestRunBinaryDownloadReachesOperationDeclaredCap`, `TestBinaryDownloadAppliesRedactFields`. | Green (commits `f846ad296`, `7a732e955`) |
| Derivable metadata drift gate | Red: `api_surface`, flag `maps_to`, `output_policy`, and `rest.max_bytes` were hand-editable and could silently diverge from `operations.json`. | Green: `connectorgen surface-sync --check` (Makefile gate `connectorgen-surface-sync`) plus `TestSyncBundleReportsDivergentAPISurface`, `TestSyncBundleReportsDivergentFlagMapsTo`, `TestSyncBundleOutputPolicyCorrectsOnlyUnsupportedValues`, `TestSyncBundleRemovesOutputPolicyFromBinaryDownload`, `TestSyncBundleMaxBytesFillsAndCorrects`, `TestSyncBundleConsistentBundleIsClean`. | Green (commits `7a732e955`, `ce76276e6`) |
| **Conformance copy of the coverage rule (this CI slice)** | Red: `go test ./internal/connectors/conformance/ -run TestConformance/github` fails with `surface_complete ... endpoint 7 (GET /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format}) covered_by.direct_read "artifact download" is not an implemented direct_read command`. New focused test `TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage` reproduces the same string on a 1-endpoint bundle. | Green: `checkSurfaceComplete` builds its `directReads` set from `direct_read \|\| binary_download` implemented commands, matching `cmd/connectorgen/validate.go:576-587` verbatim; `TestConformance/github` and the full conformance package pass. | Green |
| Availability still enforced (mutant) | Red: dropping `cmd.Availability == "implemented"` from the widened condition makes `TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage` fail with `a planned binary_download command must not satisfy covered_by.direct_reads`. | Green: with the availability guard in place, a `planned` `binary_download` command covers nothing and the endpoint is reported uncovered. | Green |

## Actual evidence — this CI slice

```bash
# RED (pre-fix checker, current tests)
go test ./internal/connectors/conformance/ -run 'TestConformance/github'
# --- FAIL: TestConformance/github (0.06s)
#     surface_complete Passed:false
#     Error:endpoint 7 (GET /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format})
#           covered_by.direct_read "artifact download" is not an implemented direct_read command

go test ./internal/connectors/conformance/ -run 'TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage' -v
# --- FAIL: TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage (0.00s)
#     static_test.go:161: checkSurfaceComplete: implemented binary_download must cover its
#     covered_by.direct_reads endpoint: endpoint 0 (GET /files/{file_id}/download)
#     covered_by.direct_read "file download" is not an implemented direct_read command

# GREEN (after widening the directReads set)
go test ./internal/connectors/conformance/
# ok  polymetrics.ai/internal/connectors/conformance  10.639s

# MUTANT: availability guard removed from the widened condition
go test ./internal/connectors/conformance/ -run 'TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage'
# --- FAIL: static_test.go:173: a planned binary_download command must not satisfy
#           covered_by.direct_reads
# (mutation reverted immediately; the guard is present on the branch)
```

## Notes

- The `github` bundle was deliberately not touched to make this pass. Endpoint 7 keeps
  `covered_by.direct_reads: ["artifact download"]`; the checker was stale, not the bundle. Rewriting
  the bundle to satisfy a drifted checker is exactly the failure mode this PR exists to close.
- No `api_surface` endpoint was invented. The only production change in this slice is the coverage
  set in `internal/connectors/conformance/static.go`.
- No secrets requested, printed, summarized, or stored. No live provider or credentialed calls.
