# TDD Ledger — #99 Linear stream runner

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | `TestLinearCommandSurfaceRunsGraphQLIssueList` failed before Linear streams had fixed GraphQL request bodies runnable from `pm linear issue list`. | done |
| Green | Linear list streams now use fixed `POST /graphql` GraphQL documents and commandrunner executes implemented stream-backed commands. | done |
| Refactor/verify | `go test ./internal/cli -run TestLinearCommandSurfaceRunsGraphQLIssueList -count=1`, `go test ./internal/connectors/conformance -run TestConformance/linear -count=1`. | done |

No credentialed Linear checks or live Linear writes were run.
