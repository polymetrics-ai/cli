# VERIFICATION — issue 3809 curated icon-collapse authority

## Required evidence

| Stage | Command | Expected result |
| --- | --- | --- |
| GSD contract | `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` | Adapter and canonical delivery projections are valid. |
| Red regression | `go test ./cmd/iconregistrygen -run TestBuildIconEntriesAllowsCuratedRowToResolveConflictingSourceURLs -count=1` | Fails before production code because current upstream collapse errors. |
| Focused green | `go test ./cmd/iconregistrygen -count=1` | All generator regressions, including uncurated refusal, pass. |
| Runtime coverage | `go test ./internal/connectors -count=1` | `MustValidateIconCoverage` remains strict and passes with regenerated data. |
| Generation | `PM_ICON_REGISTRY_SOURCE='<public OSS registry JSON>' make icons-generate` | Generator completes; `internal/connectors/icon_data.json` is generator-produced. |
| Website invariant | `pnpm run test:scripts` from `website/` | Registry-to-lockfile and canonical icon invariants pass. |
| Scoped hygiene | `gofmt -w cmd/iconregistrygen`; `go vet ./cmd/iconregistrygen ./internal/connectors`; `go build ./cmd/pm` | Formatting, static checks, and CLI build pass. |
| Verify gates (individual) | `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | All applicable non-full-suite verification gates pass. |

## Review checklist

- Confirm no changes under `internal/connectors/commandrunner/runner.go`, shared engine paths, connector defs, Shopify data, Simple Icons fetcher, or lockfile.
- Confirm the only `icon_data.json` modification was produced by `make icons-generate`.
- Confirm no missing coverage fallback is added at runtime and `MustValidateIconCoverage` is unchanged.
- Confirm uncurated conflicts still reject instead of selecting an upstream source by order.
- Before code review, load `golang-lint`; run the inline/manual `code-review` fallback because issue 3809 is not a roadmap phase, and record its findings/dispositions here.
