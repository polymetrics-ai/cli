# TDD Ledger — Issue 599 Connector Boundary Guard

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Shared provider literal | `TestScanDetectsSharedProviderSwitch` synthetic repo contains `case "gong"` in shared `internal/connectors/engine/branch.go`. | Scanner reports `connector_switch`/`gong`; focused tests pass. | Green |
| Helper placement attempt | `TestScanDetectsProviderPolicyInSharedHelper` synthetic shared helper contains `"github_date_range"`. | Scanner reports `provider_policy`/`github`; focused tests pass. | Green |
| Allowed paths | `TestScanAllowsDefinitionsNativeHooksGeneratedTestsAndDocs` places provider literals under defs/native/hooks/generated/testdata/docs. | No blocking findings for definition-owned and non-production/generated paths. | Green |
| Exceptions | `TestExceptionLedgerFailures` covers expired, stale, and broadened rows; fixture includes ignored `approved_by` prose. | Exact bounded exception suppresses only the matching finding; bad exception contracts fail. | Green |
| JSON/schema shape | `TestReportJSONShapeUsesStableArrays` marshals a clean report. | Stable `api_version`, `kind`, and array fields for findings/warnings/exceptions. | Green |
| CLI exits | `TestBoundaryCommand_*` covers clean JSON exit 0, policy exit 1, invalid invocation/config exit 2, and human output. | `cmd/connectorgen` focused tests pass. | Green |
| Base diff mode | `TestScanBaseDiffRestrictsPrimaryScan` scans against `HEAD`. | Changed file is scanned; unchanged baseline policy is ignored for primary diff findings while exception contracts remain whole-tree. | Green |
| Current baseline | `TestCurrentRepositoryBaselinePasses` scans the real repo with fixed time. | Current baseline passes via 23 bounded exceptions; Gong has zero shared-code exceptions. | Green |

## Actual evidence

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
# ok polymetrics.ai/internal/connectors/boundary
# ok polymetrics.ai/cmd/connectorgen

go run ./cmd/connectorgen boundary . --json
# outcome clean; findings 0; warnings 0; exceptions 23; checked_files 83; connectors_loaded 548; gong_exceptions 0

go run ./cmd/connectorgen validate internal/connectors/defs --json
# connectors 548; findings 0; warnings 0

make connector-boundary
# outcome clean; findings 0; exceptions 23; checked_files 83

make verify
# exit 0

git diff --check
# exit 0
```

## Notes

- GSD programming-loop command was unavailable in the adapter registry; manual-GSD fallback recorded in `PLAN.md` and `RUN-STATE.json`.
- No live connector credentials or provider calls are part of this ledger.
