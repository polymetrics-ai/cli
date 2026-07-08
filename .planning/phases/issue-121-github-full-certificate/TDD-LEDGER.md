# Issue #121 TDD Ledger

## GSD Mode

- Mode: manual-GSD fallback.
- Reason: the stacked base branch `feat/44-github-cli-parity` does not include the repo-local `scripts/gsd` adapter.
- Enforcement added in this PR: `scripts/verify-gsd-workflow` and `.github/workflows/gsd-workflow.yml` require GSD/TDD planning evidence when `cmd/` or `internal/` changes.

## Red / Green Evidence

### Enforcement gate

- Red expectation: implementation changes under `cmd/` or `internal/` without `.planning` GSD/TDD evidence should fail.
- Green implementation: added `scripts/verify-gsd-workflow` and PR workflow `GSD workflow evidence`.

### Full read sweep

- Red: `TestCatalogStreamSpecsFromStreams`, `TestFullSweepNamesAreStreamScoped`, and `TestFullSweepStreamSpecsFallbackToSelectedStream` failed before stream-spec/full-sweep helpers existed.
- Green: all focused full-sweep helper tests passed.

### Direct read stage

- Red: `TestDirectReadCandidateForGitHub` and `TestDirectReadCandidateForUnknownConnector` failed before direct-read candidate helper existed.
- Red follow-up: `TestDirectReadCandidatesForGitHub` failed while only `repo read-file` was swept.
- Green: direct-read sweep now accounts for both implemented GitHub direct-read commands: `repo read-file` and `repo read-dir`.

### 507 endpoint surface accounting

- Red: `TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints` failed before the surface inventory stage existed.
- Green: GitHub API surface accounting verifies 507 endpoints: 105 covered and 402 blocked with typed reasons.

### 67 write action accounting

- Red: `TestGithubWriteActionInventoryAccountsForAllDeclaredActions` failed before write action inventory accounting existed.
- Green: GitHub write action inventory accounts all 67 declared write actions, marking safe curated/inferred pairings separately from unpaired/blocked actions.

### Binary safety gate

- Red: `TestBinaryDownloadCandidateForGitHub` and `TestBinaryDownloadCandidateForUnknownConnector` failed before binary candidate helper existed.
- Green: focused binary tests passed.

### Flow/schedule per stream

- Red: `TestFullSweepFlowAndScheduleNamesAreStreamScoped` failed before stream-scoped glue helpers existed.
- Green: focused glue naming and full sample tests passed.

### GitHub bootstrap stream

- Red: `TestDefaultStreamName` failed before `defaultStreamName` existed.
- Green: `TestDefaultStreamName` and full sample tests passed.

## Live Tests

No live GitHub credentialed test has been run in this PR. Live tests require a rotated token via environment variable only.
