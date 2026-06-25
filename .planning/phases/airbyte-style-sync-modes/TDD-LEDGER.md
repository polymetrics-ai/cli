# TDD Ledger

Phase: airbyte-style-sync-modes

Record failing test evidence before production code for every behavior-adding task.

## 2026-06-25 Red Evidence: Sync-Mode Implementation

Status: red-confirmed

Command:

```bash
go test ./internal/app ./internal/connectors ./internal/cli ./internal/perf
```

Tasks:

- mode-parse
- manifest-defaults
- warehouse-full-refresh
- incremental-state
- dedupe
- cli-docs
- benchmarks

Expected failures:

- `internal/app/sync_modes_test.go` cannot compile because `SourceSyncMode`, `DestinationSyncMode`, `ParseSyncMode`, stream state, and related sync-mode helpers do not exist yet.
- `internal/connectors/manifest_test.go` cannot compile because `Manifest.SourceSyncModes` and `Manifest.DestinationSyncModes` do not exist yet.
- `TestPerfSyncModesJSON` returns usage error because `pm perf sync-modes` is not implemented.
- `TestETLHelpListsAllSyncModes` fails because ETL help does not list the five Airbyte-style sync modes.
- `TestSkillsGenerateWritesAgentSkills` fails because generated ETL skill guidance does not mention deduped sync modes.

Expected behavior:

- All five user-facing sync modes are parsed, validated, documented, and benchmarked.
- Local warehouse ETL supports append, overwrite, incremental, and deduped final materialization with failure-safe overwrite and checkpoint commit after success.

## 2026-06-25 Green Evidence: Focused Packages

Status: green-confirmed

Command:

```bash
gofmt -w internal/app internal/connectors internal/cli internal/perf && go test ./internal/app ./internal/connectors ./internal/cli ./internal/perf
```

Result:

- `polymetrics/internal/app` passed.
- `polymetrics/internal/connectors` passed.
- `polymetrics/internal/cli` passed.
- `polymetrics/internal/perf` passed.

## 2026-06-25 Green Evidence: Full Test Suite

Status: green-confirmed

Command:

```bash
go test ./...
```

Result:

- All packages passed.

## 2026-06-25 Green Evidence: GitHub Connector Sync-Mode Matrix

Status: green-confirmed

Command:

```bash
gofmt -w internal/app/github_sync_modes_test.go && go test ./internal/app -run 'TestGithubPullRequestsETLSupportsAllSyncModes|Test.*Sync' && go test ./...
```

Result:

- Added and passed `TestGithubPullRequestsETLSupportsAllSyncModes`.
- Verified `github` pull request ETL through the real connector path and local warehouse destination for:
  - `full_refresh_append`
  - `full_refresh_overwrite`
  - `full_refresh_overwrite_deduped`
  - `incremental_append`
  - `incremental_append_deduped`
- Full package suite passed after the GitHub-specific matrix was added.
