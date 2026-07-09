# TDD Ledger — #103 Linear sensitive/admin policy

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | Raw GraphQL/admin/sensitive Linear operations lacked a complete executable-or-exact-block inventory and help-visible safety posture. | done |
| Green | SDK mutation rows, including admin/destructive/sensitive-shaped operations, are typed fixed-document reverse-ETL actions with risk text and destructive confirmation where applicable; raw arbitrary GraphQL remains blocked by default. | done |
| Refactor/verify | `go run ./cmd/connectorgen validate internal/connectors/defs --json`, `./pm help linear`, `grep` evidence for `api graphql` and `Raw arbitrary GraphQL is disallowed`. | done |

No credentialed Linear checks or live Linear writes were run.
