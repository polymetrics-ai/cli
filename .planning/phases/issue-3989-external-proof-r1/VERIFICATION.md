# #3989 verification checklist

## Residual verification — blocked on live full parity

- [ ] The live GitHub smoke ran, not skipped, with the designated disposable identity, but did not produce an accepted proof. After supplying its required non-secret `rate_limit_account` coordination subject, full parity reached distinct non-passing stages (`schedule_create`, then `resume`) on the final authorized runs. The raw credential was not rendered; the remaining live full-parity failure is the blocker.
- [x] Complete opaque request and response bodies carry credential canaries that are substituted before proof serialization (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] A real fresh child snapshots its own process-list entry, argv, project root, runner workdir, and fresh-binary directory at the credential-live boundary. The parent verifies the finished secret-free artifact records no raw credential, no external timing window, and no CI opt-out; it also requires Recurly's legitimate incomplete-full-parity exit because the one-route TLS fixture does not certify its declared write surface (`TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts`).
- [x] One proof demonstrates same-A equality, distinct-B separation, and absence of the repository salt from output (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] Changed-package and consumer verification passed after the safe diagnostic change: full `internal/connectors/certify`, full `internal/cli` under its unchanged 20-minute ceiling, required `cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, and `git diff --check`. The broader historical gate record remains valid for unchanged generated and documentation surfaces.

## Residual validation record

| Command | Result |
| --- | --- |
| `go test -count=1 -run '^TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials$' ./internal/connectors/certify` | Passed after the planned red compile failure. |
| `go test -count=1 -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed after the planned red compile failure; rerun after the process-list child-presence assertion. |
| `go test -count=2 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed after rebase (115.27s and 117.20s). Each child emits its secret-safe snapshot at the credential-live boundary, then Recurly's intentionally one-route fixture exits 1 for incomplete full parity. |
| `go test -timeout 20m -count=1 -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` | Ran repeatedly with the designated disposable identity and did not skip. The first run exposed missing declared non-secret rate-limit coordination; after adding `rate_limit_account`, full parity remained non-passing. Final safe diagnostics named `schedule_create` and then `resume` with typed CLI errors. No accepted proof or raw credential was retained. |
| `go test -timeout 20m ./internal/connectors/certify` | Passed. |
| `go test -timeout 20m ./internal/cli` | Passed. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed (required consumer package). |
| `go test -timeout 20m ./internal/connectors/certify` | Passed after the universal diagnostic change in 9.238s. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed after the universal diagnostic change in 91.170s. |
| `go test -timeout 20m ./internal/cli` | Passed after the universal diagnostic change in 648.115s, below the unchanged 20-minute ceiling. |
| `go vet ./...`; `go build ./cmd/pm`; `git diff --check` | Passed after the universal diagnostic change. |
| `make tidy-check`, `make fmt`, `git diff --check`, `go vet ./...`, `go build ./cmd/pm` | Passed. |
| `make docs-check`; `./pm connectors`; `./pm help connectors`; `./pm connectors certify --help` | Passed; no CLI/docs source change was applicable. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website/lib/docs.generated.ts` | Passed; generated website docs were byte-stable. |
| `make smoke-no-build` | Passed. |
| `make lint`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync` | Passed. |
| `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make connector-canon-check`, `make release-workflow-check` | Passed. |

The full `go test -timeout 20m ./...` / aggregate `make verify` commands were intentionally not run as a single per-command-timeout process: repository guidance requires their constituent gates and changed packages plus consumers to run separately because the 550+ connector suite routinely exceeds the execution wrapper window. No requested component gate was skipped.

After rebasing the delivery commit on refreshed base `4a0289bcc`, the focused certify/CLI/consumer tests, `go vet ./...`, `go build ./cmd/pm`, tidy/docs/smoke/lint, website byte-stability, and every remaining individual generator/boundary/canon/release gate above were rerun successfully. The base's stricter full-parity roll-up correctly makes the local one-route Recurly fixture exit 1; the OS proof now asserts that honest exit after the child snapshot instead of claiming Recurly is certified. The authenticated disposable-identity smoke is retained as the already-run live evidence rather than needlessly repeating provider traffic.

## CI child-side observation validation

After PR #4198 exposed a third distinct failure with every parent-side
settlement observable false, the fresh-child OS proof was inverted. The child
now captures its own evidence at the credential-live boundary; the parent does
not hold, release, poll, or otherwise race that process. Rebase then showed
that Recurly's one-route fixture must honestly exit 1 under the stricter
full-parity roll-up, so the OS test requires that exit after obtaining the
snapshot. The following commands passed without adding any timeout or CI
opt-out:

| Command | Result |
| --- | --- |
| `go test -count=2 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed. Two fresh-child runs (115.27s and 117.20s) each wrote and verified their own observation artifact, then required Recurly's honest incomplete-parity exit. |
| `go test -timeout 20m ./internal/cli` | Passed in 725.538s. The proof no longer relies on a parent-side scheduling window. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed (required consumer package). |
| `go vet ./...`; `go build ./cmd/pm`; `make docs-check`; `git diff --check` | Passed. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website` | Passed; the second generator pass left all website files byte-stable. |
| `go run ./cmd/agentcontractgen check` | Passed. |

## Rebased gate results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./internal/cli` | Passed in 725.538s. |
| `go test -timeout 20m ./internal/connectors/certify` | Passed in 9.294s. |
| `go test -timeout 20m ./cmd/connectorgen` | Passed in 100.643s. |
| `go vet ./...`; `go build ./cmd/pm`; `make docs-check`; `git diff --check`; `go run ./cmd/agentcontractgen check` | Passed. |
| `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website` | Passed; the second pass was byte-stable. |
| `make tidy-check`; `make lint`; `make smoke-no-build`; `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make github-parity-artifacts-check`; `make connectorgen-certification-matrix`; `make connector-boundary`; `make connector-canon-check`; `make release-workflow-check` | Passed. |

## CLI package-capacity validation

CI's timeout displayed `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked`, but a `go test -count=1 -timeout 20m -v ./internal/cli` baseline passed in 706.417s and showed that test at 39.140s. The actual slowest tests are the two external-child proof cases (118.770s and 118.210s); the new dynamic leaf-help sweep runs 17,800 independent manual-render cases in 22.500s locally. The failure is therefore aggregate capacity, not a Bahmni hang.

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m -v ./internal/cli` | Passed in 706.417s before the scheduling change. It retained all verbose per-test timings; the ranked costs are recorded above. |
| `go test -count=1 -timeout 20m -v -run '^TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch$' ./internal/cli` | Passed; the parent logs exactly `checked 17800 dynamic connector command help variants` and the package completes in 6.137s. |
| `go test -race -count=1 -timeout 20m -run '^TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch$' ./internal/cli` | Passed in 61.753s; no race is reported while the independent leaf-help cases run in parallel. |
| `go test -count=1 -timeout 20m ./internal/cli` | Passed in 694.432s after the scheduling change, under the unchanged 20-minute package limit. |

The parallel change neither filters the generated case set nor reduces assertions: each discovered command still runs with both `--help` and `-h`, validates help interception before dispatch, validates the exact rendered command, and requires a `NAME` section. It merely assigns independent read-only cases to the test runner's bounded parallel scheduler.

- [x] Every acceptance row has an observable state-change assertion: evidence writer, observer, ephemeral-session, relay, and fresh-child tests each assert a positive write/request/fingerprint or an explicit zero-write refusal. The OS fresh-child test additionally requires child-captured process-list/argv/project/temporary evidence at the credential-live boundary.
- [x] External binary is freshly built; evidence records its SHA, exact safe argv, and successful `flow_plan`/`flow_preview`/`flow_run`/`flow_status` references (`TestExternalProofFreshChildCapturesCompleteHTTPSProviderTranscript`).
- [x] No raw credential appears in captured parent streams, project tree, vault/key, or artifact in the fresh TLS child test; parent relay refuses both streams before writing on a canary match. `--value-stdin` is rewritten to a child-only environment reference with no value in argv.
- [x] Error/refusal paths write zero accepted-evidence artifacts (`TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites` and `TestExternalProofFreshChildRefusesNoHTTPSWithoutArtifact`).
- [x] HTTPS transcript covers exact request/response observation and explicit byte bounds; transport tests cover bounded error bodies, complete redirect source/final exchanges, and a zero-write refusal beyond the redirect/retry cap while preserving the child-visible body.
- [x] `go test -timeout 20m ./internal/connectors/certify/...` passes after the final TLS/relay changes.
- [x] Required scoped local gates and inline standard code review complete: `go test -timeout 20m ./internal/{app,cli}`, `go test -timeout 20m ./cmd/connectorgen`, scoped `go vet`, `go build ./cmd/pm`, and the individual `make verify` component gates pass. No #3989 review finding remains.
- [x] CLI help/manual/website parity: `pm connectors`, `pm help connectors`, `pm connectors certify --help`, golden transcript, and `make docs-check` pass after final documentation changes.
- [x] Opt-in PostgreSQL container lane was run with the supplied Docker endpoint. `TestPostgresManagedTargetDriverLiveControlAssertions` still fails with the pre-existing #4158 route-mismatch refusal where a durable acknowledgement is expected; it is out of #3989 scope and no PostgreSQL source was changed.
