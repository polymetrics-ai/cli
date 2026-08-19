# VERIFICATION — issue 3809 curated icon-collapse authority

## Required evidence

| Stage | Command | Expected result |
| --- | --- | --- |
| GSD contract | `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` | Adapter and canonical delivery projections are valid. |
| Red regression | `go test ./cmd/iconregistrygen -run TestBuildIconEntriesAllowsCuratedRowToResolveConflictingSourceURLs -count=1` | Fails before production code because current upstream collapse errors. |
| Focused green | `go test ./cmd/iconregistrygen -count=1` | All generator regressions, including uncurated refusal, pass. |
| Runtime coverage | `go test ./internal/connectors -count=1` | `MustValidateIconCoverage` remains strict and passes with regenerated data. |
| Generation | `PM_ICON_REGISTRY_SOURCE='https://connectors.airbyte.com/files/registries/v0/oss_registry.json' make icons-generate` | Generator completes; `internal/connectors/icon_data.json` is generator-produced. |
| Website invariant | `pnpm run test:scripts` from `website/` | Registry-to-lockfile and canonical icon invariants pass. |
| Scoped hygiene | `gofmt -w cmd/iconregistrygen`; `go vet ./cmd/iconregistrygen ./internal/connectors`; `go build ./cmd/pm` | Formatting, static checks, and CLI build pass. |
| Verify gates (individual) | `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | All applicable non-full-suite verification gates pass. |

## Review checklist

- Confirm no changes under `internal/connectors/commandrunner/runner.go`, shared engine paths, connector defs, Shopify data, Simple Icons fetcher, or lockfile.
- Confirm the only `icon_data.json` modification was produced by `make icons-generate`.
- Confirm no missing coverage fallback is added at runtime and `MustValidateIconCoverage` is unchanged.
- Confirm uncurated conflicts still reject instead of selecting an upstream source by order.
- Before code review, load `golang-lint`; run the inline/manual `code-review` fallback because issue 3809 is not a roadmap phase, and record its findings/dispositions here.

## Actual evidence

- 2026-08-06: `scripts/gsd doctor`, all five `scripts/gsd sources` resolutions, and `go run ./cmd/agentcontractgen check` passed. Generated `discuss-phase`, `plan-phase --tdd`, `execute-phase --interactive`, `verify-work`, and explicit-file `code-review` prompts; `gsd-sdk query init.phase-op 3809` reports `phase_found: false`, so the documented worker-artifact manual fallback is used.
- RED: `go test ./cmd/iconregistrygen -run TestBuildIconEntriesAllowsCuratedRowToResolveConflictingSourceURLs -count=1` failed before the implementation with the old conflicting-source-URL abort.
- GREEN: `go test ./cmd/iconregistrygen -count=1`; `go test ./internal/connectors -count=1`; `go test ./internal/cli -count=1`; `go vet ./cmd/iconregistrygen ./internal/connectors`; and `go build ./cmd/pm` all passed.
- GREEN: `PM_ICON_REGISTRY_SOURCE='https://connectors.airbyte.com/files/registries/v0/oss_registry.json' make icons-generate` completed twice, each time producing 554 connector entries and 5 SVG assets. A key-sorted comparison with the pre-generation registry found no semantic JSON changes.
- GREEN: `pnpm run test:scripts` in `website/` passed 22/22 tests, including exact registry-to-lockfile coverage.
- GREEN: `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` passed.
- Baseline note: an additional `golangci-lint run ./cmd/iconregistrygen/...` reports four pre-existing findings at unchanged lines `cmd/iconregistrygen/main.go:122`, `:559`, `:622`, and `:630` (two unchecked pre-existing close errors and two unused pre-existing helpers). The repository's configured `make lint` passed with 0 issues; no unrelated cleanup was added to #3809.
- Inline/manual code review of `cmd/iconregistrygen/main.go` and `main_test.go`: no new correctness, security, or code-quality findings. Confirmed that final shared-path validation is unchanged and the no-curated test asserts that no upstream URL is chosen silently.
