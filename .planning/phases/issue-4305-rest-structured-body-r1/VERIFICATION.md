# Verification: Issue #4305

## Required checks

- [x] Focused configured-base-URL GraphQL application-error persistence regression: state excludes the configured secret while returned provider receipt remains verbatim.

- [x] Focused credential-bearing `base_url` validation red/green regression: URI user-info and a query component reject through configuration validation; an ordinary endpoint URI remains accepted, including for source-locked fields that omit `format: uri`.

- [x] Targeted engine, commandrunner, and installed CLI tests with -timeout 20m.
- [x] Existing scalar, form, SCIM, binary, and specialized GitHub focused regressions.
- [x] go vet ./...
- [x] go build ./cmd/pm
- [x] go run ./cmd/connectorgen validate internal/connectors/defs
- [x] go run ./cmd/connectorgen surface-sync --check
- [x] Generated help/manual/schema checks and applicable website parity check.
- [x] Completion-tracked make connector-boundary.
- [x] Fresh `make verify` after the credential-bearing `base_url` admission slice.
- [x] git diff --check.
- [x] Code-review evidence with actionable findings resolved.

## Results

### Focused TDD and behavior checks

### 2026-08-24 — credential-bearing `base_url` input hardening

- Red tests reproduced acceptance of URL user-info/query components for a schema-declared `base_url`, then reproduced that 29 format-omitted source declarations bypassed a format-only repair.
- Green: `go test -timeout 20m ./internal/connectors/engine -count=1` passed. `go test -timeout 20m ./internal/app -count=1` passed in 393.089s under a completion-tracked session.
- The app-level regression proves `AddCredential` rejects either unsafe URI shape before a credential appears in state and without echoing the rejected URL. The existing configured-base-URL GraphQL application-error regression remains unchanged and green; it uses an ordinary endpoint and confirms that input hardening does not supersede receipt persistence safety.

### 2026-08-25 — completion-tracked whole-tree verification attempt

- `make verify` progressed through formatting, tidy, vet, the complete `internal/app` package (1089.715s), and `internal/connectors/boundary` (1024.093s). It then failed only because the unrelated `internal/cli` package reached its repository-enforced 20-minute test timeout (1203.558s) while another lane's whole-tree suite was actively running on the shared host.
- The changed packages themselves had already passed in full (`./internal/connectors/engine` and `./internal/app`). Once the competing run cleared, the clean retry `go test -timeout 20m ./internal/cli -count=1` passed in 519.864s.
- This first whole-tree attempt remains incomplete rather than green; the clean CLI result removed the localized timeout concern. A second completion-tracked `make verify` rerun on the idle host passed with exit 0: complete tests (`internal/cli` 498.426s and connector-boundary 424.818s), build/docs/smoke, lint (0 issues), contract and source/generator checks, a clean 553-connector boundary report, release layout, and installed GitHub certification.

### 2026-08-24 — plaintext configured-base-URL persistence follow-up

- Red: `go test -timeout 20m ./internal/app -run '^TestGraphQLBaseURL(Secret|Config)ApplicationErrorDoesNotPersistEcho$' -count=1` failed at the plaintext state-file assertion before production edits.
- Green: the same command passed after the persistence-only projection. It observes both required effects: `state.json` lacks the configured secret and the returned direct-write result plus typed provider error retain the exact provider receipt. No output/header/body redaction, truncation, or state-encryption change was made.

- `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO|WriteActionOptionalRecordQueryIsOmittedOrPreservedAtProvider|OperationDirectWriteStructuredRESTBody)' -count=1` — pass.
- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen ./internal/cli -run 'Test(OperationDirectWriteStructuredRESTBody|OperationStructuredJSONBodyPreflight|BuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody|GitHubUserDraftCommandBuildsFixedGraphQLMutation|Validate_CLISurfaceDirectWriteStructuredRESTBody|StructuredRESTBodyCommandHelpAndManual|ConnectorsManualDocumentsConnectorArchitectureAndGithubExamples)' -count=1` — pass.
- Full packages passed with `-timeout 20m`: `./internal/connectors/engine`, `./internal/connectors/commandrunner`, `./cmd/connectorgen`, and `./internal/cli`.
- The structured REST fixture proves exact provider request materialization, path/query/body isolation, action isolation, nested approval mutation rejection, and zero transport calls for malformed, missing, unknown, wrong-type, over-limit, and metadata-override inputs. The write-query fixture proves only the exact object-form `record.*` declaration can omit an absent value; supplied values reach the provider and required/undeclared/wrong-source/malformed inputs reject pre-I/O.

### Generated command surface and docs

- Built the actual binary: `go build -o ./pm ./cmd/pm`.
- Runtime help/manual checks passed: `./pm help connectors`, `./pm connectors`, and `./pm docs generate --dir docs/cli --connectors-dir docs/connectors`. A second generation preserved the `docs/cli/connectors.md` content hash.
- `POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=help_connectors,bare_connectors_manual,json_connectors_manual,connectors_help_github_known_legacy_namespace_help_intercept go test -timeout 20m ./internal/cli -run TestGoldenTranscripts -count=1` passed after regenerating the complete affected manual fixture set. `go test -timeout 20m ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1` passed.

### Repository gates

- `go vet ./...` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — pass: 552 connectors, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` — pass: 552 scanned, 0 changes.
- `make connector-boundary` — pass: clean whole-tree report, 294 files checked and 552 connectors loaded; only pre-existing dated exceptions.
- Final `make verify` — pass. It includes `gofmt`, tidy check, `go vet ./...`, `go test -timeout 20m ./...`, `pm` build/docs/smoke, lint (0 issues), agent contract, generator validation/sync, GitHub artifacts, connector boundary, connector canon, pinned dependencies, and release target parity.
- `git diff --check` — pass.

### Final local validation — 2026-08-24

- Built the real `pm` binary and confirmed `pm help connectors` plus bare `pm connectors` behavior.
- Focused structured-body engine and commandrunner tests, the configured-base-URL error-persistence regression, and the complete CLI golden/manual suite passed with `-timeout 20m`.
- `go vet ./...`, `go run ./cmd/connectorgen validate internal/connectors/defs` (553 connectors, 0 findings), and `go run ./cmd/connectorgen surface-sync --check` (553 scanned, 0 changes) passed.
- `make connector-boundary` completed with exit 0 in a completion-tracked session.
- `make verify` completed with exit 0 in a completion-tracked session. Its full repository suite passed, including formatting/tidy, full tests, binary/docs/smoke, lint, generated artifacts, connector boundary, release checks, and installed GitHub certification.

### Delivery note

The explicit Firstmate correction defers the no-mistakes, push, and PR stages. They were not invoked in this worktree.

### CI remediation — 2026-08-21

- Red reproduced locally: `go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectWriteStructuredRESTBody' -count=1` failed to compile because `interpolateDeclaredHeader` called the hardened four-argument `resolveExpr` with its obsolete three-argument form.
- Green: restored the explicit `allowControlCharacters=false` policy at that declaration-bound header call. `go test -timeout 20m ./internal/connectors/engine`, `go test -timeout 20m ./internal/connectors/commandrunner`, `go test -timeout 20m ./internal/app`, `go test -timeout 20m ./internal/cli`, `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m -run '^$' ./internal/cli/ ./internal/connectors/certify/`, `go build -trimpath -ldflags='-s -w' ./cmd/pm`, `go vet ./...`, `go run ./cmd/connectorgen boundary . --json`, `go run ./cmd/connectorgen validate internal/connectors/defs`, `go run ./cmd/connectorgen surface-sync --check`, and `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` passed.
- Website CI/data green evidence: regenerated `website/lib/docs.generated.ts` with `npm run gen:website-data`; repeated generation was stable. `pnpm run typecheck` passed, and `pnpm run lint` completed with only the repository's existing 13 warnings.
- This was an assigned CI repair only. No no-mistakes pipeline command, push, PR mutation, or CI re-run was invoked.

### CI Verify follow-up — 2026-08-21

- Hosted Verify run `32430447912` exposed two stale test artifacts, not a real-invocation-budget failure: the `83 (budget 92)` line was a successful final accounting report. The actual failures were the stale secret-bearing HTTP-error expectation in `internal/app` and four stale generated connector-manual transcripts in `internal/cli`.
- `go test -timeout 20m ./internal/app -count=1` — pass (256.922s).
- `go test -timeout 20m ./internal/cli -count=1` — pass (561.504s).
- The targeted generator-backed transcript refresh and validation passed; `go vet ./internal/app ./internal/cli` and `git diff --check` also passed.

### CI Verify follow-up — 2026-08-24

- Hosted Verify run `32695514116` failed only because `json_connectors_manual` and `connectors_help_github_known_legacy_namespace_help_intercept` had stale generated copies of the connector manual.
- Refreshed those two records only. `go test -timeout 20m ./internal/cli -run '^(TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -count=1` passed without update mode, and `git diff --check` passed.
- This assigned CI phase did not invoke no-mistakes controls, a CI re-run, a push, or a PR mutation.
