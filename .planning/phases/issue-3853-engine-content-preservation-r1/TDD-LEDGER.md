# TDD LEDGER — issue #3853 engine content preservation

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | A write preview preserves declared `redact_fields` values, including nested values | `go test ./internal/connectors/engine -run '^(TestDryRunWritePreviewResolvedPathPreservesConfiguredRecordFields|TestDryRunWritePreviewResolvedPathPreservesNestedRecordFields)$' -count=1 -v` failed on the unchanged engine: both resolved paths were `/patients/redacted` | The full six-test focused run passed after preview interpolation used the unmodified record; both top-level and nested values appear in the resolved warning | Keep declarations and load compatibility; no bundle rewrite |
| R2 | A write preview preserves resolved config-secret substitutions | The same red run failed: `fixture-preview-secret` became `***` in the resolved URL warning | The focused run passed after preview interpolation used the unmodified runtime secrets; the synthetic value remains in the resolved warning | Existing preview digest/gate coverage remains unchanged |
| R3 | Direct-read and operation-direct-read errors preserve bounded HTTP URL/query/body diagnostics | The named `httptest` cases failed: the query was removed and the JSON diagnostic became `[redacted]` on both paths | The focused run passed after the engine formatted bounded `connsdk.HTTPError` URL/body fields directly; both query and JSON diagnostics remain visible | Existing error-map class/hint and non-HTTP error behavior remains unchanged |
| R4 | Binary-download errors preserve bounded HTTP URL/query/body diagnostics | The named `httptest` case failed: the query was removed and the JSON diagnostic became `[redacted]`; the existing no-file assertion remains | The focused run passed after binary download shared the same engine error formatter; diagnostics are complete and no destination file is left | Destination, bound, redirect, and cleanup guards remain covered by package tests |
| R5 | CLI/manual/golden/website wording describes the true connector-engine boundary | The tracked help/manual/golden/website wording asserted masking before the source change | `go test ./internal/cli -run '^(TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -count=1` passed after regeneration; approval-token JSON behavior now says “omits”, not “redacts” | Token omission and the separate source-table sample boundary remain explicit |

## Red-test rule

Every behavior assertion runs against the unchanged engine and fails before its production fix.
Existing tests asserting the old behavior are deliberately reversed and retained as permanent
content-preservation coverage; no test is weakened, skipped, or deleted.
