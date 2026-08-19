# TDD Ledger — Issue #4302 loader operation-kind registration

## Planning checkpoint

- Manual-GSD fallback: inline execution is required because this firstmate direct-PR lane may not spawn GSD roles. The source issue fixes all product decisions: register only `rest_status` and `text_export`, use their existing `rest` and `binary` blocks respectively, and do not touch Docker Hub or other connector definitions.
- Red: the first test loads a synthetic `operations.json` declaring both kinds through the ordinary `Load` path; it failed on the loader’s unsupported-kind/block-map gap.
- Green: pending. The smallest implementation is to map `rest_status` to `rest` and `text_export` to `binary`, allowing the already-present kind-specific semantic validation to run.

## Slice 1 — declaration reachability

- Red: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestBundleLoadRegistersStatusAndTextExportOperations|TestBundleLoadRejectsInvalidStatusAndTextExportDeclarations)$'` failed before loader implementation. The happy declaration was rejected as `unsupported kind "rest_status"`; every malformed-case assertion also stopped at the same unreachable-kind error. This proves the real loader did not reach either existing semantic validator.
- Green: pending focused engine test. Observable success: `Load` returns both parsed `OperationSpec` records with the correct kind-specific execution blocks.

## Slice 2 — fail-closed declaration semantics

- Red: pending focused engine test. Invalid status response/body behavior, non-HEAD status method, unbounded CSV export, and wrong execution blocks must each be rejected by `Load` before any executor can issue I/O.
- Green: pending focused engine test. The test will assert each error contains the declared kind and its violated bounded/status contract.
