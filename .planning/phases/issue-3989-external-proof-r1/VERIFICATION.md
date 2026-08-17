# #3989 verification checklist

## Residual verification — pending

- [x] The live GitHub smoke runs, not skips, with its disposable identity and records an observable sanitized proof result (`TestExternalProofGitHubSmoke` parses a passing GitHub report and an observed 2xx proof response).
- [x] Complete opaque request and response bodies carry credential canaries that are substituted before proof serialization (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] A real fresh child snapshots its own process-list entry, argv, project root, runner workdir, and fresh-binary directory at the credential-live boundary. The parent verifies the finished secret-free artifact records no raw credential, no external timing window, and no CI opt-out (`TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts`).
- [x] One proof demonstrates same-A equality, distinct-B separation, and absence of the repository salt from output (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] Focused suites, consumer package, repository gates, generator byte stability, and CLI unchanged-surface checks are recorded with their exact results.

## Residual validation record

| Command | Result |
| --- | --- |
| `go test -count=1 -run '^TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials$' ./internal/connectors/certify` | Passed after the planned red compile failure. |
| `go test -count=1 -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed after the planned red compile failure; rerun after the process-list child-presence assertion. |
| `go test -count=3 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed after the child-side observation inversion. Each completed fresh process emits a secret-safe snapshot at the credential-live boundary; the parent verifies the snapshot after ordinary exit. |
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

## CI child-side observation validation

After PR #4198 exposed a third distinct failure with every parent-side
settlement observable false, the fresh-child OS proof was inverted. The child
now captures its own evidence at the credential-live boundary; the parent does
not hold, release, poll, or otherwise race that process. The following commands
passed without adding any timeout or CI opt-out:

| Command | Result |
| --- | --- |
| `go test -count=3 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed. Three fresh-child runs under the test runner's parallel setting each wrote and verified their own observation artifact. |
| `go test -timeout 20m ./internal/cli` | Passed. The proof no longer relies on a parent-side scheduling window. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed (required consumer package). |
| `go vet ./...`; `go build ./cmd/pm`; `make docs-check`; `git diff --check` | Passed. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website` | Passed; the second generator pass left all website files byte-stable. |
| `go run ./cmd/agentcontractgen check` | Passed. |

- [x] Every acceptance row has an observable state-change assertion: evidence writer, observer, ephemeral-session, relay, and fresh-child tests each assert a positive write/request/fingerprint or an explicit zero-write refusal. The OS fresh-child test additionally requires child-captured process-list/argv/project/temporary evidence at the credential-live boundary.
- [x] External binary is freshly built; evidence records its SHA, exact safe argv, and successful `flow_plan`/`flow_preview`/`flow_run`/`flow_status` references (`TestExternalProofFreshChildCapturesCompleteHTTPSProviderTranscript`).
- [x] No raw credential appears in captured parent streams, project tree, vault/key, or artifact in the fresh TLS child test; parent relay refuses both streams before writing on a canary match. `--value-stdin` is rewritten to a child-only environment reference with no value in argv.
- [x] Error/refusal paths write zero accepted-evidence artifacts (`TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites` and `TestExternalProofFreshChildRefusesNoHTTPSWithoutArtifact`).
- [x] HTTPS transcript covers exact request/response observation and explicit byte bounds; transport tests cover bounded error bodies, complete redirect source/final exchanges, and a zero-write refusal beyond the redirect/retry cap while preserving the child-visible body.
- [x] `go test -timeout 20m ./internal/connectors/certify/...` passes after the final TLS/relay changes.
- [x] Required scoped local gates and inline standard code review complete: `go test -timeout 20m ./internal/{app,cli}`, `go test -timeout 20m ./cmd/connectorgen`, scoped `go vet`, `go build ./cmd/pm`, and the individual `make verify` component gates pass. No #3989 review finding remains.
- [x] CLI help/manual/website parity: `pm connectors`, `pm help connectors`, `pm connectors certify --help`, golden transcript, and `make docs-check` pass after final documentation changes.
- [x] Opt-in PostgreSQL container lane was run with the supplied Docker endpoint. `TestPostgresManagedTargetDriverLiveControlAssertions` still fails with the pre-existing #4158 route-mismatch refusal where a durable acknowledgement is expected; it is out of #3989 scope and no PostgreSQL source was changed.
