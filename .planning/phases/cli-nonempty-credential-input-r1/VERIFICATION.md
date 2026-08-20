# Verification — provider-neutral non-empty credentials

## Outcome

The implementation satisfies the automated acceptance criteria. No real
credentials were used: tests use synthetic canaries and report only typed
classification, byte length, and SHA-256 fingerprints.

## Acceptance evidence

| Criterion | Evidence | Result |
| --- | --- | --- |
| Stdin preserves all valid bytes except one final LF/CRLF | `TestNormalizeStdinRemovesOnlyOneDocumentedTerminalDelimiter`; `TestCredentialsAddStdinPreservesSingleTerminalDelimiterAndRoundTrips` | pass |
| Empty, LF-only, and CRLF-only input cannot persist | `TestCredentialsAddStdinRejectsEmptyNormalizedSecretBeforePersistence`; `TestAddCredentialRejectsEmptySecretBeforeVaultPersistence` | pass |
| App/vault writes cannot introduce an empty secret and existing data survives | `TestRuntimeConfigSecretStoreRejectsEmptyWriteAndPreservesExistingSecret` | pass |
| Required bearer, basic, API-key, and OAuth forms cannot emit empty authentication | `TestRequiredAuthenticatorsRejectEmptyCredentialBeforeRequestMutation`; `TestOAuth2ClientCredentialsRejectsEmptyRequiredMaterialBeforeTokenRequest`; `TestOAuth2RefreshTokenRejectsEmptyRequiredTokenBeforeExchange`; `TestSelectAuthRejectsEmptyRequiredCredential` | pass |
| Optional authentication remains available | `TestSelectAuthOptionalMissingCredentialSelectsNone`; refresh-token tests retain the documented optional public-client secret | pass |
| Help/manual/website generated artifacts are in sync | golden credentials transcript, `TestGoldenDocsGenerateMatchesTrackedCLIManuals`, `pm help credentials`, `pm credentials`, `pm credentials --help`, `pm docs validate --connectors-dir docs/connectors` | pass |
| Twenty lane remains isolated | changed-path check found no `internal/connectors/defs/twenty/**` or Twenty docs path | pass |

## Commands passed

```text
go test -timeout 20m ./internal/credential ./internal/vault ./internal/connectors/connsdk ./internal/connectors/engine -count=1
go test -timeout 20m ./internal/app -run '^(TestAddCredentialRejectsEmptySecretBeforeVaultPersistence|TestRuntimeConfigSecretStoreRejectsEmptyWriteAndPreservesExistingSecret)$' -count=1
POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=help_credentials go test -timeout 20m ./internal/cli -run '^(TestCredentialsAddStdin|TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -count=1
go build ./cmd/pm
make fmt
go vet ./...
make tidy-check
make docs-check-no-build
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make lint
make github-parity-artifacts-check
make connectorgen-certification-matrix
make connectorgen-certification-candidates
make connectorgen-certification-sweep
make connector-canon-check
make connector-boundary
make release-workflow-check
make build
```

`connector-boundary` reported `outcome: clean`, loading all 552 bundles.

## CI timeout repair

- **Red:** PR #4310 `verify` exhausted the 20-minute `internal/cli` test
  timeout while `TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts`
  was running. In isolation, its one-route fixture spent about 170 seconds on
  the unrelated Recurly `--full-parity` sweep.
- **Green:** The test now runs the bounded external-proof path that it actually
  asserts: a fresh child makes an authenticated provider request while its
  argv/process observation and temporary artifacts remain free of the synthetic
  credential. Its direct live check passed in 32.502 seconds; the focused
  credential-and-child regression command below passed in 95.244 seconds:

  ```text
  go test -count=1 -timeout=2m -v -run '^TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts$' ./internal/cli
  go test -count=1 -timeout=5m -run '^(TestExternalProofFreshChildHidesCredentialFromProcessListAndTemporaryArtifacts|TestExternalProofFreshChildPublishesBoundedProofWithoutFullParity|TestPrepareExternalCertifyStdinCredential(UsesChildMemoryOnly|NormalizesOneDelimiter|RejectsEmptyNormalizedValue)|TestCredentialsAdd(StdinPreservesSingleTerminalDelimiterAndRoundTrips|StdinRejectsEmptyNormalizedSecretBeforePersistence|FromEnvRejectsTransportOnlySecretBeforePersistence))$' ./internal/cli
  ```

- The active no-mistakes run owns the remaining aggregate validation and PR/CI
  phases; no pipeline command, push, or PR action was performed by this repair.

## Controlled limitations

- A whole-package `go test -timeout 20m ./internal/cli -count=1` was run.
  Two pre-existing external-certification tests failed because their local
  external provider/Redis fixtures were unavailable. The only task-caused
  failure was the credentials help golden, which was regenerated and then
  passed in the focused transcript and generated-doc checks above.
- `npm --prefix website run typecheck` could not run because `tsc` is not
  installed in this worktree. This change updates prose plus the checked-in
  generated docs file; `npm --prefix website run gen:docs` completed before
  the tracked artifact was reviewed.
- Repository guidance prohibits running aggregate `go test ./...` or
  `make verify` under this per-command execution limit. The changed packages,
  `internal/cli`, and every other `make verify` component gate were run
  separately. The full aggregate suite and CI are deferred to the required
  Firstmate no-mistakes/PR gate; no no-mistakes run has started yet.
