# Verification — source-lock operation import

## Status

Planning complete; execution pending.

## Required gates

| Command or evidence | Status |
| --- | --- |
| GSD command resolution and `agentcontractgen check` | pass before production edits |
| Focused source-import red/green tests | pending |
| `go test -timeout 20m ./cmd/connectorgen` | pending |
| Generator validation, source-import golden/check mode, and `surface-sync --check` | pending |
| `go vet ./...`; `go build ./cmd/pm`; `git diff --check` | pending |
| completion-tracked `make connector-boundary`; `make verify` | pending |
| Execute/verify/code-review prompt evidence | pending |

## CLI/docs parity disposition

- `pm help <topic>`, `pm <namespace>`, `pm <command> --help`: not applicable; no `pm` runtime surface changes.
- `connectorgen source-import --help`: required and pending.
- `docs/cli/**`, `website/**`, generated `pm` manual/completions: not applicable; check that no unintended changes are introduced.
- Migration/adoption documentation: required and pending.
