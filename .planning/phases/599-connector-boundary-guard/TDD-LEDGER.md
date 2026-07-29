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
| Current baseline | `TestCurrentRepositoryBaselinePasses` scans the real repo with fixed time. | Current baseline passes via 24 bounded exceptions; Gong has zero shared-code exceptions. | Green |
| Branch-name CI repair | GitHub Actions `branch-name` log rejects `HEAD_REF=fm/cli-connector-boundary-guard-r1` because `fm` is missing from `.github/workflows/conventions.yml`. | Extracted `conventions.yml` branch-name run block accepts `fm/cli-connector-boundary-guard-r1`, `fix/stripe-pagination`, and `dependabot/go_modules/example`, while rejecting `invalid/Bad_Name`. | Green |
| Review fix: escape hatch owner | Fixtures place provider-specific Go under `internal/connectors/hooks/shared/` and `internal/connectors/native/common/`; current classifier allows those by path shape alone. | Unknown hook/native first segments scan as shared production Go; real connector hook/native dirs remain allowed. | Green |
| Review fix: weak identifiers | Fixtures use `gongDateRangeFallback` and `boxOutputPolicy` in shared Go; current lexicon does not build identifier prefixes for weak one-word names. | Metadata-derived weak identifier prefixes catch compound provider-policy identifiers while exact weak literals and generic identifiers remain non-blocking. | Green |

## Actual evidence

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
# ok polymetrics.ai/internal/connectors/boundary
# ok polymetrics.ai/cmd/connectorgen

go run ./cmd/connectorgen boundary . --json
# outcome clean; findings 0; warnings 0; exceptions 24; checked_files 129; connectors_loaded 548; gong_exceptions 0

go run ./cmd/connectorgen validate internal/connectors/defs --json
# connectors 548; findings 0; warnings 0

make connector-boundary
# outcome clean; findings 0; exceptions 24; checked_files 129

branch-name validation examples
# extracted conventions.yml run block accepted fm/cli-connector-boundary-guard-r1, fix/stripe-pagination, dependabot/go_modules/example; rejected invalid/Bad_Name

make verify
# exit 0

git diff --check
# exit 0

go test ./internal/connectors/boundary
# ok polymetrics.ai/internal/connectors/boundary
```

## Notes

- GSD programming-loop command was unavailable in the adapter registry; manual-GSD fallback recorded in `PLAN.md` and `RUN-STATE.json`.
- No live connector credentials or provider calls are part of this ledger.
