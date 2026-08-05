# TDD LEDGER — issue #3853 engine content preservation

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | A write preview preserves declared `redact_fields` values, including nested values | `go test ./internal/connectors/engine -run '^(TestDryRunWritePreviewResolvedPathPreservesConfiguredRecordFields|TestDryRunWritePreviewResolvedPathPreservesNestedRecordFields)$' -count=1 -v` failed on the unchanged engine: both resolved paths were `/patients/redacted` | Pending | Keep declarations and load compatibility; no bundle rewrite |
| R2 | A write preview preserves resolved config-secret substitutions | The same red run failed: `fixture-preview-secret` became `***` in the resolved URL warning | Pending | Verify existing preview digest/gate behavior remains covered |
| R3 | Direct-read and operation-direct-read errors preserve bounded HTTP URL/query/body diagnostics | The named `httptest` cases failed: the query was removed and the JSON diagnostic became `[redacted]` on both paths | Pending | Preserve existing error-map class/hint and non-HTTP error behavior |
| R4 | Binary-download errors preserve bounded HTTP URL/query/body diagnostics | The named `httptest` case failed: the query was removed and the JSON diagnostic became `[redacted]`; the existing no-file assertion remains | Pending | Preserve destination, bound, redirect, and cleanup guards |
| R5 | CLI/manual/golden/website wording describes the true connector-engine boundary | Pending: source/golden mismatch before regenerated artifacts | Pending | Verify token omission and generic source-table caveat remain accurate |

## Red-test rule

Every behavior assertion runs against the unchanged engine and fails before its production fix.
Existing tests asserting the old behavior are deliberately reversed and retained as permanent
content-preservation coverage; no test is weakened, skipped, or deleted.
