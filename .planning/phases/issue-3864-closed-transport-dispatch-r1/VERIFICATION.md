# #3864 verification checklist

## Status: local gates, manual verify-work, and manual code review recorded; child delivery pipeline pending

- [x] TDD RED outputs are recorded before production code.
- [x] Focused `internal/connectors`, `internal/synctransport`, `internal/app`, and
  `internal/cli` tests pass with `-timeout 20m`.
- [x] Transport package race test and cancellation regression pass.
- [x] `go vet` and build pass.
- [x] Required non-suite `make verify` components pass individually.
- [x] `make connector-runtime-preflight`, connector canon, connectorgen validation, and
  surface-sync checks pass.
- [x] `pm connectors`, `pm help connectors`, `pm connectors --help`, and
  `pm connectors inspect sample --json` are checked in an initialized project without
  credentials; docs/website parity is checked.
- [x] Manual `verify-work` outcome (zero automated gaps), code-review findings/dispositions, and
  supervisor-compatible local evidence are recorded in `UAT.md`, `REVIEW.md`, and
  `SUPERVISOR-EVIDENCE.md` using the documented manual-GSD fallback.
- [ ] Complete child no-mistakes push/PR/CI result (without `--yes`), automated-review coverage,
  and child-local check state are recorded. Its PR base must be
  `feat/3862-any-to-any-transport`; it must not merge or create another parent/default PR.

## Local evidence

- `go test -timeout 20m ./internal/connectors ./internal/synctransport`,
  `go test -timeout 20m ./internal/app`, and `go test -timeout 20m ./internal/cli` passed.
- `go test -race -timeout 20m ./internal/synctransport -run
  '^(TestRegistryPreflightIsRaceSafeDuringRegistration|TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches|TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply)$'`
  passed.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`, and
  `make connector-runtime-preflight` passed.
- `golangci-lint run --new-from-rev=origin/feat/3862-any-to-any-transport
  ./internal/app/... ./internal/cli/... ./internal/connectors/... ./internal/synctransport/...`
  returned `0 issues`. A broader non-diff sweep reports pre-existing findings in unrelated
  app/connsdk files; they are not changed or waived by this child.
- A freshly built binary in an initialized temporary project rendered the three connector
  help forms with `sync_transport` wording and `connectors inspect sample --json` with
  both roles `unsupported`. `docs/cli/connectors.md`, its generated transcript fixture,
  and `website/content/docs/agent-guide.mdx` carry the same non-certification boundary.
- `scripts/gsd prompt verify-work issue-3864-closed-transport-dispatch-r1` and
  `scripts/gsd prompt code-review issue-3864-closed-transport-dispatch-r1` were executed.
  The Pi adapter cannot create the mandated worker roles for this non-numbered issue phase, so
  `UAT.md` and `REVIEW.md` record the inline/manual fallback and its results.

## Explicit limits

Correction loop 1/5 is tracked in [#4021](https://github.com/polymetrics-ai/cli/issues/4021):
an authored invalid descriptor must reach closed preflight rather than legacy routing. Its RED and
GREEN command evidence is in T11 of `TDD-LEDGER.md`.

Correction loop 2/5 is tracked independently in [#4023](https://github.com/polymetrics-ai/cli/issues/4023):
the closed descriptor must reject `generic-http` just as it rejects `generic_http`. Its RED and
GREEN command evidence is in T13 of `TDD-LEDGER.md`. Shared correction commit `9775f420c`
references #4021, #4023, #3864, and #3862 while retaining each issue's bounded scope.

This verification can prove only fake-backed dispatch and metadata surfaces. It cannot
truthfully assert executable #3810 conformance, a real API/database transport, a live
provider flow, automatic Shepherd certification, or a green GitHub CI/review state until
those gates actually run.
