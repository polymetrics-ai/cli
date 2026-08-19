# TDD Ledger — Issue #4302 loader operation-kind registration

## Planning checkpoint

- Manual-GSD fallback: inline execution is required because this firstmate direct-PR lane may not spawn GSD roles. The source issue fixes all product decisions: register only `rest_status` and `text_export`, use their existing `rest` and `binary` blocks respectively, and do not touch Docker Hub or other connector definitions.
- Red: the first test loads a synthetic `operations.json` declaring both kinds through the ordinary `Load` path; it failed on the loader’s unsupported-kind/block-map gap.
- Green: `expectedOperationBlock` now maps `rest_status` to `rest` and `text_export` to `binary`, allowing the already-present kind-specific semantic validation to run.

## Slice 1 — declaration reachability

- Red: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestBundleLoadRegistersStatusAndTextExportOperations|TestBundleLoadRejectsInvalidStatusAndTextExportDeclarations)$'` failed before loader implementation. The happy declaration was rejected as `unsupported kind "rest_status"`; every malformed-case assertion also stopped at the same unreachable-kind error. This proves the real loader did not reach either existing semantic validator.
- Green: the same focused command passed. `Load` returned both parsed `OperationSpec` records with the correct `rest` and `binary` blocks; it includes the status upper metadata-cap boundary (`max_bytes=1024`) and export minimum positive-cap boundary (`max_bytes=1`). This is declaration reachability through the ordinary loader, not executor-only construction.

## Slice 2 — fail-closed declaration semantics

- Red: pending focused engine test. Invalid status response/body behavior, non-HEAD status method, unbounded CSV export, and wrong execution blocks must each be rejected by `Load` before any executor can issue I/O.
- Green: the same focused command passed. Its named cases assert that status rejects JSON output/body declarations and non-HEAD methods, while text export rejects an absent byte bound and a REST block in place of its binary block. Each refusal occurs in `Load`, before any runtime or I/O path exists.
