# TDD ledger — #3771 command-runner runtime content

| Slice | Red evidence to capture before production edit | Green evidence | Status |
| --- | --- | --- | --- |
| #3782 | Captured 2026-08-06 with `go test ./internal/connectors/commandrunner -run 'TestBuildWriteCommandPreviewDryRunsAndPreservesDeclaredFields|…' -count=1`: previews and ETL/error content became `***`; direct-read, operation-direct-read, and binary-download requests received declared fields. | The same focused command passed after removing runner mutation/forwarding and deleting its masking helpers. | Green captured |
| #3784 | The shared #3782 red checkpoint demonstrated the same legacy `reverse_etl` preview mutation before this public-path regression test was added. | `TestPlanConnectorCommandPersistsCompleteDeclaredContent` passed on 2026-08-06: a destructive, preview-only fake preserves nested/token/content fields through plan state reload while the token is omitted from stored state. | Green captured |
| #3786 | Captured 2026-08-06 with `TestReverseManualExplainsConnectorCommandContentPolicy`: old help lacked the complete-content/source-table/no-retry policy and retained masking claims. | Focused help test and regenerated `TestGoldenTranscripts` passed after source/manual/website parity updates. | Green captured |
| #3790 | The #3782 red checkpoint demonstrated the behavior-level mutation and forwarding that this new matrix covers. | `content_preservation_test.go` passed on 2026-08-06: exact ETL values, equivalent reverse preview, original executor errors, and empty executor `RedactFields`. | Green captured |

Existing tests asserting masking are intentionally revised rather than deleted because the captain
overruled the old policy. The final PR must call out this reversal.
