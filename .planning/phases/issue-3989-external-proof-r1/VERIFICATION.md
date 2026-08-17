# #3989 verification checklist

## Residual verification — bounded credential scope

The prior full-parity blocker is resolved by #4215’s schema-v2 bounded scope,
not by weakening full parity. The remaining implementation proof must emit a
truthful `observed_operations` / `protocol_exchanges` artifact from the real
fresh child, preserve any certification failure exit, and record the two
historical stage failures separately.

- [x] Bounded external proof omits the obsolete `--full-parity` gate, emits
  schema-v2 scope/proof fields from complete observed HTTPS exchanges, and
  requires an observable successful provider response without claiming parity.
- [x] The opt-in GitHub smoke ran with the designated disposable identity and
  retained one secret-free bounded proof plus exact-transcript verification.
- [x] `schedule_create` (typed CLI error, exit 3) and `resume` (typed CLI
  error, exit 1) are named separately with redacted evidence; neither is
  treated as environmental success or patched around.

- [x] The live GitHub smoke ran, not skipped, with the designated disposable identity and the required non-secret `rate_limit_account` subject. With no `--full-parity`, it passed and retained a schema-v2 `observed_operations` / `protocol_exchanges` proof from a real GitHub exchange. The report does not claim full parity; the earlier `schedule_create` and `resume` results remain separate redacted findings.
- [x] Complete opaque request and response bodies carry credential canaries that are substituted before proof serialization (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] A real fresh child snapshots its own process-list entry, argv, project root, runner workdir, and fresh-binary directory at the credential-live boundary. The parent verifies the finished secret-free artifact records no raw credential, no external timing window, and no CI opt-out; it also requires Recurly's legitimate incomplete-full-parity exit because the one-route TLS fixture does not certify its declared write surface (`TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts`).
- [x] One proof demonstrates same-A equality, distinct-B separation, and absence of the repository salt from output (`TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials`).
- [x] Final changed-package and consumer verification passed: full
  `internal/connectors/certify`, full `internal/cli` under its unchanged
  20-minute ceiling, required `cmd/connectorgen`, `go vet ./...`,
  `go build ./cmd/pm`, `git diff --check`, all individual repository gates,
  CLI manual/golden checks, and four generated-surface byte-stability checks.

## Final bounded-scope validation

| Command | Result |
| --- | --- |
| `go test -count=1 -v -run '^(TestWriteExternalProofFingerprintsObservedExternalTranscript|TestWriteExternalProofPublishesBoundedObservedOperations|TestWriteExternalProofRefusesBoundedClaimWithoutSuccessfulProviderResponse|TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials|TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites|TestWriteExternalProofRefusesMissingFlowReferencesWithoutArtifactWrites)$' ./internal/connectors/certify` | Passed. Covers v2 bounded claim, no-provider-success refusal, opaque request/response substitution, same-run A/B semantics, truncation, and full-parity reference refusal. |
| `go test -count=1 -v -run '^TestCertifyCLIExternalProofRunsWithoutFullParityBeforeNoHTTPSRefusal$' ./internal/cli` | Passed. A fresh child starts without `--full-parity`, saves its report, and still writes no proof when it observed no HTTPS exchange. |
| `go test -count=1 -v -run '^TestExternalProofFreshChildPublishesBoundedProofWithoutFullParity$' ./internal/cli` | Passed in 37.327s. Fresh TLS child creates one v2 bounded proof with a provider 2xx. |
| `go test -count=1 -v -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed in 157.150s. Real child self-audits command list, argv, project, runner workdir, and fresh-binary temporary locations without raw credential presence. |
| `go test -count=1 -v -run '^TestExternalProofFailureDiagnosticFingerprintsPlantedCredential$' ./internal/cli` | Passed. Parsed-report and pre-report diagnostics both substitute the planted credential with a fingerprint marker. |
| Credentialed `go test -count=1 -timeout 20m -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` with only the designated token, owner, and repository environment variables | Passed in 30.239s. The real GitHub fresh child retained and verified one secret-free schema-v2 `observed_operations` / `protocol_exchanges` proof. |
| `go test -count=1 -timeout 20m ./internal/connectors/certify`; `go test -count=1 -timeout 20m ./internal/cli`; `go test -count=1 -timeout 20m ./cmd/connectorgen` | Passed in 10.476s, 666.954s, and 187.484s respectively. |
| `go vet ./...`; `go build ./cmd/pm`; `go run ./cmd/agentcontractgen check`; `git diff --check` | Passed. |
| `make tidy-check fmt docs-check smoke-no-build lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check` | Passed. |
| `./pm connectors`; `./pm help connectors`; `./pm connectors certify --help`; `go test -count=1 -run '^TestGoldenTranscripts$' ./internal/cli` | Passed. |
| Docs, website, skills, and certification generators run twice with `git hash-object` aggregate comparison | Second run byte-stable for all four generated surfaces. |

### Separate historical full-parity findings

Before #4215 enabled the bounded claim, the authorized disposable-identity
full-parity sequence produced two distinct non-passing facts: `schedule_create`
returned a typed CLI error with exit 3; a later diagnostic run reached `resume`
and returned a typed CLI error with exit 1. Both were emitted only through the
fingerprint-redacted diagnostic path. They are retained as product findings,
not reclassified as environmental success, and the bounded proof above makes no
claim about either stage.

## Residual validation record

| Command | Result |
| --- | --- |
| `go test -count=1 -run '^TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials$' ./internal/connectors/certify` | Passed after the planned red compile failure. |
| `go test -count=1 -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed after the planned red compile failure; rerun after the process-list child-presence assertion. |
| `go test -count=2 -parallel=4 -timeout 20m -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' -v ./internal/cli` | Passed after rebase (115.27s and 117.20s). Each child emits its secret-safe snapshot at the credential-live boundary, then Recurly's intentionally one-route fixture exits 1 for incomplete full parity. |
| `go test -timeout 20m -count=1 -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` | Ran repeatedly with the designated disposable identity and did not skip. The first run exposed missing declared non-secret rate-limit coordination; after adding `rate_limit_account`, full parity remained non-passing. Final safe diagnostics named `schedule_create` and then `resume` with typed CLI errors. No accepted proof or raw credential was retained. |
| `go test -count=1 -timeout 20m -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` with the designated disposable token, owner, and repository environment variables | Passed after the schema-v2 bounded-scope change. The fresh child retained one secret-free `observed_operations` / `protocol_exchanges` proof, asserted an observed GitHub 2xx, and verified the exact child transcript. No full-parity claim was made. |
| `go test -count=1 -v -run '^TestExternalProofFreshChildPublishesBoundedProofWithoutFullParity$' ./internal/cli`; `go test -count=1 -v -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli` | Passed. The first proves the gate-free v2 TLS-child route; the second proves child-side process-list/argv/project/temporary scans remain clear of the prepared credential while the intentionally incomplete full-parity fixture exits with the current certification-failure code 2. |
| `go test -count=1 -v -run '^TestExternalProofFailureDiagnosticFingerprintsPlantedCredential$' ./internal/cli` | Passed. A planted secret is absent from both parsed-report and pre-report diagnostics, which instead carry `{{pmcertfp:v1:...}}`; the rendered finding stays concise. |
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

## Refreshed integration-base revalidation (2026-08-18)

The branch was rebased onto `integration/4015-mvp-flat-r1` at `c2dedecbc`.
The post-rebase reader fixture initially failed because a `FullParity` test
fixture had no flow references. It was corrected by providing the existing
four required references; no production admission rule was relaxed.

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m -run '^TestReadExternalProofRefusesAnUnfingerprintedResponseRegression$' ./internal/connectors/certify` | Passed after the fixture correction. |
| Credentialed `go test -count=1 -timeout 20m -v -run '^TestExternalProofGitHubSmoke$' ./internal/cli` with only the designated disposable token, owner, and repository environment variables | Passed in 24.932s. The real GitHub child observed a provider success, verified its exact transcript, wrote a secret-free schema-v2 bounded proof, and did not skip. No credential value is retained here. |
| `go test -count=1 -timeout 20m ./internal/connectors/certify` | Passed in 9.359s. |
| `go test -count=1 -timeout 20m ./internal/cli` | Passed in 532.221s, below the unchanged 20-minute ceiling. |
| `go test -count=1 -timeout 20m ./cmd/connectorgen` | Passed in 80.792s (required consumer package). |
| `go vet ./...`; `go build ./cmd/pm`; `make fmt`; `git diff --check`; `go run ./cmd/agentcontractgen check` | Passed. |
| `./pm connectors`; `./pm help connectors`; `./pm connectors certify --help`; `go test -count=1 -timeout 20m -run '^TestGoldenTranscripts$' ./internal/cli` | Passed; the golden test completed in 9.610s. |
| `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check` | Passed. `connector-boundary` reported `outcome: clean`. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -count=1 -timeout 20m -run '^TestGoldenTranscripts$' ./internal/cli`; `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` twice; `pnpm --dir website run gen:docs` twice; `pnpm --dir website run gen:website-data` twice | Passed. Each of the four generated surfaces matched the committed bytes after its second pass. |

As required for this repository's per-command execution environment, the
aggregate `go test -timeout 20m ./...` and `make verify` were not run as one
wrapper process. Their changed packages, consumer package, build/vet checks,
and every individual `make verify` component target above were run separately.

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
