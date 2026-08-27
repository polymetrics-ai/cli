# Verification checklist — issue 4371

## Required local gates

- [ ] Focused cited-only non-executable/partial red then green tests.
- [ ] Ordinary OpenAPI/Swagger disposition stability and invalid citation matrix.
- [ ] Salesloft/Copper source import, source projection, operation evidence,
  validate, and surface-sync checks; or exact record that clean-worktree source
  locks are unavailable and no cohort output was regenerated.
- [ ] Real registry/commandrunner missing-foundation-before-credential/I/O
  proof for any generated unavailable command under test.
- [ ] `go test -timeout 20m ./cmd/connectorgen -count=1`.
- [ ] Changed dependent engine/commandrunner suites with `-timeout 20m`.
- [ ] `go vet ./...`, `go build ./cmd/connectorgen`, `go build ./cmd/pm`, and
  the individually applicable tidy/lint/docs/smoke/agent-contract/
  connectorgen/connector-boundary/release-workflow gates.
- [ ] `git diff --check` before commit and after current-main rebase.
- [ ] Inline GSD `verify-work`, `code-review`, and a separate exact-head Codex
  audit request/result recorded against the final code SHA.
- [ ] PR base API read-back equals `main`; CI/review are requested but no merge
  is performed.

## Results

## Current-main integration

- Initial and final `origin/main`: `cf29d302c13f7fcd340d31ad6dc27872880ccf42`.
- `git fetch origin main && git rebase origin/main` completed normally with
  `Current branch … is up to date`; no history rewrite, conflict resolution,
  reset, stash, force push, or generated-file merge happened.

## Captured green evidence

- [x] Red then green cited-only disposition suite:

  ```text
  go test -timeout 20m ./cmd/connectorgen -run 'TestSourceProjection(CitedOnlyMutationDispositionsKeepReferenceClosed|WriteDisabledMutationArtifactsKeepCitedOnlyReferenceClosed|MutationDispositionInputRemainsFailClosed|SourceCitedNonExecutableMutationDispositionRejectsCompleteAction|SourceCitedPartialMutationCoveragePreservesImplementedIncompleteAction|SourceCitedMutationDispositionLeavesExistingProjectionByteIdentical|WriteDisabledMutationArtifactsRetainGraphQLMutations|WriteDisabledMutationArtifactsRequireProviderCitation)$' -count=1 -v
  PASS
  ```

- [x] Full changed package: `go test -timeout 20m ./cmd/connectorgen -count=1`
  — PASS in `227.505s`.
- [x] Engine: `go test -timeout 20m ./internal/connectors/engine -count=1`
  — PASS in `17.505s`; rebase coverage PASS in `14.005s`.
- [x] Commandrunner: `go test -timeout 20m
  ./internal/connectors/commandrunner -count=1` — PASS in `22.432s`, covering
  `TestRunSourceBoundReadMissingFoundationRefusesBeforeDispatch`; real registry
  preflight `make connector-runtime-preflight` — PASS in `8.648s`.
- [x] Static/build: `go vet ./...`; `go build ./cmd/connectorgen`; and
  `go build ./cmd/pm` — PASS. The post-rebase focused regression and engine
  rerun are recorded above; `make docs-check` independently rebuilt `pm`.
- [x] Formatting/diff: `gofmt` changed only the edited Go files; `git diff
  --check origin/main...HEAD` — PASS.
- [x] Repository gates: `make tidy-check`, `make lint` (`0 issues`), `make
  docs-check`, `make smoke-no-build` (`smoke ok`), `make connector-canon-check`,
  `make connector-boundary` (exit `0`), and `make release-workflow-check`
  (exit `0`) — PASS.
- [x] Canonical/generator: `go run ./cmd/agentcontractgen check` — current;
  `connectorgen validate` — `553 connector(s) checked, 0 findings`;
  `surface-sync --check` — 553 scanned with 0 changes; `operation-evidence
  --check` — `1774 rows; 5 rollups; fixed-100 passed`.

## Cohort and command-boundary truth

- Salesloft/Copper source locks and retained artifacts are absent from this
  clean current-main worktree, so no cohort descriptor was regenerated and no
  source-import command was fabricated. Read-only historical lock inspection
  records 211 Salesloft and 89 Copper operations in `CONTEXT.md`; it does not
  replace a preserved retained-artifact input or make a current-main claim.
- The shared source-reference fixture proves the only allowed cited-only state
  is byte-identical `source_contract_unavailable`; each operation remains
  visible and merge-blocked without a new action, command, credential lookup,
  transport, record emission, or mutation.
- Usable-surface delta is **0**. There is no generated unavailable command in
  this change to invoke. The existing real commandrunner missing-foundation
  test proves structured refusal before dispatch; the registry preflight sweep
  proves every command still marked `implemented` reaches the runtime-owned
  preflight path. No false missing-credential witness or provider I/O was
  added.

## Still external

- [ ] CI and GitHub's configured automated reviewer run after normal branch
  publication.
- [ ] A separate fresh-context Codex exact-head audit is requested before
  merge consideration. It must name the final code SHA; this worker will not
  merge.
