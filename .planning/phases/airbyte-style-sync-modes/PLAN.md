# PLAN: Airbyte-Style Sync Modes

## Tasks

- [x] task:mode-parse type:behavior - Add typed sync-mode parsing and validation tests.
- [x] task:manifest-defaults type:behavior - Extend connector manifests with all supported sync modes and stream defaults.
- [x] task:warehouse-full-refresh type:behavior - Implement append and overwrite local warehouse semantics with failure-safe temp final files.
- [x] task:incremental-state type:behavior - Implement cursor state, incremental filtering, and checkpoint commit after success.
- [x] task:dedupe type:behavior - Implement raw JSONL history and deterministic deduped final materialization.
- [x] task:cli-docs type:behavior - Update CLI help, generated docs, and skills with sync-mode details.
- [x] task:benchmarks type:behavior - Add synthetic sync-mode benchmark coverage.
- [x] task:verification type:docs - Run verification and update GSD artifacts.

## Constraints

- No new Go dependencies.
- No live GitHub API use.
- No runtime dependency requirement for this phase.
