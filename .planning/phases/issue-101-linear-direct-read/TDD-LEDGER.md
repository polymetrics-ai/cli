# TDD Ledger — #101 Linear direct read

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | `TestLinearCommandSurfaceRunsStreamBackedDirectRead` failed before direct-read commands could run single-object Linear streams. | done |
| Green | Added `issue`, `team`, `project`, and `user` fixed GraphQL view streams plus stream-backed direct-read runner support. | done |
| Refactor/verify | `go test ./internal/cli -run TestLinearCommandSurfaceRunsStreamBackedDirectRead -count=1`, `go test ./internal/connectors/engine -run TestLinearStreamsUseFixedGraphQLDocuments -count=1`. | done |

No credentialed Linear checks or live Linear writes were run.
