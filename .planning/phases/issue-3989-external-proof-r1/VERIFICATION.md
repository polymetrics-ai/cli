# #3989 verification checklist

## Residual verification — pending

- [x] The live GitHub smoke runs, not skips, with its disposable identity and records an observable sanitized proof result (`TestExternalProofGitHubSmoke` parses a passing GitHub report and an observed 2xx proof response).
- [x] Complete opaque request and response bodies carry credential canaries that are substituted before proof serialization (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] A held real external child is present during an OS command-list scan and scoped temporary-artifact scan, neither of which contains the raw credential; after release, its response, handler return, natural exit, complete parseable report, and temporary-build cleanup all settle under bounded condition polling (`TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts`).
- [x] One proof demonstrates same-A equality, distinct-B separation, and absence of the repository salt from output (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] Focused suites, consumer package, repository gates, generator byte stability, and CLI unchanged-surface checks are recorded with their exact results.

## Residual validation record

| Command | Result |
| --- | --- |
| `go test -count=1 -run '^TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials$' ./internal/connectors/certify` | Passed after the planned red compile failure. |
| `go test -count=1 -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed after the planned red compile failure; rerun after the process-list child-presence assertion. |
| `go test -count=3 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed after the PR #4198 second CI race. The one held provider request, its handler return, child exit, valid persisted report, and removal of scoped external-build artifacts are independently observed before the proof returns. |
| `go test -timeout 20m -count=1 -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` | Passed with the designated disposable identity. The run stays credential-free in this record. |
| `go test -timeout 20m ./internal/connectors/certify` | Passed. |
| `go test -timeout 20m ./internal/cli` | Passed. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed (required consumer package). |
| `make tidy-check`, `make fmt`, `git diff --check`, `go vet ./...`, `go build ./cmd/pm` | Passed. |
| `make docs-check`; `./pm connectors`; `./pm help connectors`; `./pm connectors certify --help` | Passed; no CLI/docs source change was applicable. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website/lib/docs.generated.ts` | Passed; generated website docs were byte-stable. |
| `make smoke-no-build` | Passed. |
| `make lint`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync` | Passed. |
| `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make connector-canon-check`, `make release-workflow-check` | Passed. |

The full `go test -timeout 20m ./...` / aggregate `make verify` commands were intentionally not run as a single per-command-timeout process: repository guidance requires their constituent gates and changed packages plus consumers to run separately because the 550+ connector suite routinely exceeds the execution wrapper window. No requested component gate was skipped.

After rebasing the delivery commit on refreshed base `4967fa2a0`, the focused certify/CLI/consumer tests, `go vet ./...`, `go build ./cmd/pm`, tidy/docs/smoke/lint, website byte-stability, and every remaining individual generator/boundary/canon/release gate above were rerun successfully. The base delta has no change under either external-proof test file, so the authenticated disposable-identity smoke is retained as the already-run live evidence rather than needlessly repeating provider traffic.

## CI complete-state settlement validation

After PR #4198 exposed the second, distinct report-persistence race, the
fresh-child OS proof was redesigned around its complete observable state rather
than another one-point handoff. The following commands passed without widening
any timeout:

| Command | Result |
| --- | --- |
| `go test -count=3 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed. Three fresh-child runs under the test runner's parallel setting each reached the complete settlement state. |
| `go test -timeout 20m ./internal/cli` | Passed. No third distinct OS-proof failure occurred. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed (required consumer package). |
| `go vet ./...`; `go build ./cmd/pm`; `make docs-check`; `git diff --check` | Passed. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website` | Passed; the second generator pass left all website files byte-stable. |
| `go run ./cmd/agentcontractgen check` | Passed. |

- [x] Every acceptance row has an observable state-change assertion: evidence writer, observer, ephemeral-session, relay, and fresh-child tests each assert a positive write/request/fingerprint or an explicit zero-write refusal. The OS fresh-child test additionally waits for provider delivery, handler completion, child exit, report parseability, and temporary-build removal as independent conditions.
- [x] External binary is freshly built; evidence records its SHA, exact safe argv, and successful `flow_plan`/`flow_preview`/`flow_run`/`flow_status` references (`TestExternalProofFreshChildCapturesCompleteHTTPSProviderTranscript`).
- [x] No raw credential appears in captured parent streams, project tree, vault/key, or artifact in the fresh TLS child test; parent relay refuses both streams before writing on a canary match. `--value-stdin` is rewritten to a child-only environment reference with no value in argv.
- [x] Error/refusal paths write zero accepted-evidence artifacts (`TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites` and `TestExternalProofFreshChildRefusesNoHTTPSWithoutArtifact`).
- [x] HTTPS transcript covers exact request/response observation and explicit byte bounds; transport tests cover bounded error bodies, complete redirect source/final exchanges, and a zero-write refusal beyond the redirect/retry cap while preserving the child-visible body.
- [x] `go test -timeout 20m ./internal/connectors/certify/...` passes after the final TLS/relay changes.
- [x] Required scoped local gates and inline standard code review complete: `go test -timeout 20m ./internal/{app,cli}`, `go test -timeout 20m ./cmd/connectorgen`, scoped `go vet`, `go build ./cmd/pm`, and the individual `make verify` component gates pass. No #3989 review finding remains.
- [x] CLI help/manual/website parity: `pm connectors`, `pm help connectors`, `pm connectors certify --help`, golden transcript, and `make docs-check` pass after final documentation changes.
- [x] Opt-in PostgreSQL container lane was run with the supplied Docker endpoint. `TestPostgresManagedTargetDriverLiveControlAssertions` still fails with the pre-existing #4158 route-mismatch refusal where a durable acknowledgement is expected; it is out of #3989 scope and no PostgreSQL source was changed.
