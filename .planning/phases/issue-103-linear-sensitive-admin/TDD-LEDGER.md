# TDD Ledger — #103 Linear sensitive/admin policy

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | Raw GraphQL/admin/sensitive Linear operations lacked a complete blocked-by-default inventory and help-visible safety posture. | done |
| Green | `cli_surface.json` marks raw GraphQL, webhook, invite, auth, config, and admin/destructive surfaces unsafe/unsupported; `api_surface.json` blocks non-approved SDK operations by default. | done |
| Refactor/verify | `go run ./cmd/connectorgen validate internal/connectors/defs --json`, `./pm help linear`, `grep` evidence for `api graphql` and `Raw arbitrary GraphQL is disallowed`. | done |

No credentialed Linear checks or live Linear writes were run.
