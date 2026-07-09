# TDD Ledger — #98 Linear help renderer

| Step | Evidence | Result |
|---|---|---|
| Plan | Phase artifact created before production edits; manual-GSD fallback active because `programming-loop` is unavailable. | done |
| Red | `TestLinearConnectorHelpRendersCommandSurface` failed before `pm help linear`, bare `pm linear`, and `pm linear --help` rendered connector command-surface help. | done |
| Green | `internal/cli/cli.go` routes missing built-in help topics to connector command-surface manuals and bare connector namespaces now render help. | done |
| Refactor/verify | `go test ./internal/cli -run TestLinearConnectorHelpRendersCommandSurface -count=1`, `./pm help linear`, `./pm linear --help`, `./pm docs validate --connectors-dir docs/connectors`. | done |

No credentialed Linear checks or live Linear writes were run.
