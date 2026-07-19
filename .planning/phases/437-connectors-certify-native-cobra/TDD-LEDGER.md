# Phase 437 TDD Ledger

Issue #437; invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

GSD doctor/list and plan-phase prompt passed. Adapter has no `programming-loop`; manual universal-runtime-loop fallback was used. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

| Step | Kind | Evidence | Status |
|---:|---|---|---|
| 0 | Planning | Six issue-local artifacts written before tests/production | Complete |
| 1 | Initial RED | `go test ./internal/cli -run 'TestConnectorsCommandIsNativeCobraSubtree|TestNativeConnectors|TestNativeCertify' -count=1` | Failed before production: `newConnectorsCobraCommand` and `newConnectorsCobraCommandWithRuntime` undefined |
| 2 | GREEN | Native subtree, typed flags/runtime seam, compatibility normalization, namespace-only parser removal | Focused `3.876s`; router/golden/certify/telemetry `111.507s`; repeated ×10 `34.915s`; race `40.844s`; certify package `336.422s` |
| 3 | Help RED | Trailing list/catalog/inspect/certify help contract | Failed before final help edit: operations ran, certify positional help returned usage 2, JSON trailing help returned certification output |
| 4 | Help GREEN | Contextual direct/trailing help before effects | Focused `3.884s` |
| 5 | Operand-help RED | Direct `connectors inspect --help|help` | Failed before correction: help token captured as connector name and returned internal connector-not-found |
| 6 | Operand-help GREEN | Help-aware private operand capture | Focused `3.989s` |
| 7 | Refactor | Repeated/race/router/golden/differential/parity | Final native ×10 `34.833s`; native race `40.842s`; router/golden/certify/telemetry `111.919s`; exact-start operations 21/21 |
| 8 | Full gate | certify smoke, connector validation, docs/website, gofmt/vet/test/build/make verify | Complete: smoke exit 0/pass; validation 547/0; final `make verify` exit 0 |
| 9 | Delivery | Finalize six artifacts, commit/push, no PR/review | Complete |

## Contract proven

- `connectors` is native Cobra and absent from legacy wrappers; native list/catalog/inspect/manual aliases/help/certify are declared.
- Certify supports single connector, `--all`, and `--sweep`, with the current command-contract flags represented as repeatable `StringArray` flags with `NoOptDefVal=true`.
- Repeated values, bare/assigned/space forms, positional connector selection, ignored tails, unknowns, literal `--`, malformed action heads, no later discovery, and globals retain compatibility.
- Bare/topic/direct/long/short/positional/trailing help is canonical text/JSON and side-effect free; invalid actions remain usage errors.
- Certification reports retain exits 0 pass, 1 internal, 2 certification failure, 3 leak dominance and one-envelope output.
- Fresh command trees, context cancellation, bounded batch concurrency, event sequence, telemetry spans, and credential-value exclusion remain covered.
- Dynamic connector dispatch remains the only CLI `parseFlags` consumer pair; `parse.go` is unchanged.

## Final evidence

- Full `make verify`: CLI `431.305s`, certify `337.280s`, docs, ordered local smoke, lint 0, connector validation 547/0.
- Required explicit local certify smoke: exit 0, sample `ConnectorCertification`, passed, empty stderr.
- No live credential check/write/sweep, service, dependency, connector definition, or private material used.

## Accepted review correction ledger

Session `issue-437-review-correction-20260719T113319Z`; exact correction start `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`; all five findings in `/tmp/pm-397-review-437.log` accepted.

| Step | Kind | Required evidence | Status |
|---:|---|---|---|
| C0 | Planning | Six artifacts reopened before tests/production; GSD manual fallback, skills, safety scope, and verification commands recorded | Complete |
| C1 | RED | `go test ./internal/cli -run 'TestNativeCertifyRejectsUnsupportedSafetyAndModeControls|TestNativeConnectorsOnlyExactHelpFormsRenderManual|TestConnectorsManualSeparatesCLIAndCompletedCertificationExits|TestCertifyCLIBatchLoadsCredentialsBeforeValidatingParallelAndPreservesErrorBytes|TestTelemetryCertifyInvalidOptionsPreserveSingleSpanAndConnectorValidationPrecedence' -count=1` | Complete: unsupported controls invoked runtime and were visible; false/malformed/cluster help rendered manuals; docs phrases absent; missing creds gained batch wrapper and invalid parallel won; invalid connector lost to option usage and emitted no certify span |
| C2 | GREEN | Smallest production correction preserving all safety gates and rejecting/unadvertising unimplemented controls | Pending |
| C3 | Refactor | Focused differential/repeated/race, certify exit/redaction/replay-no-live, local sample smoke, docs/golden/website parity | Pending |
| C4 | Full gate | gofmt, vet, full tests, build, `make verify`, connector validation | Pending |
| C5 | Delivery | Finalize artifacts, commit/push, no services/credentials/PR/review | Pending |

Correction RED must be captured and committed before any production edit. Fixture/temp data only; no live credentials, HTTP writes, reverse ETL execution, or external sweep.
