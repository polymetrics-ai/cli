# TDD ledger — GitLab provider-inventory parity G0 + G1

## Setup evidence

- Isolation: `pwd -P` and `git rev-parse --show-toplevel` were verified in an isolated Git worktree.
- Branch: `fm/cli-gitlab-parity-implementation-r1`.
- `no-mistakes doctor`: daemon running and repository initialized.
- No GitLab credentials, live API calls, provider writes, or secret values are used.

## Baseline / red evidence

- Existing GitLab surface: 11 rows, 4 stream-covered reads, 7 legacy exclusions; no `operations.json` and no `cli_surface.json`.
- Provider ledger report: 1,745 documented operations (770 reads, 975 writes).
- Before G1, the built `pm gitlab projects list --help` failed with `error: unknown command "gitlab"` and exit 2 because no definition-owned command surface existed.

## Green evidence

| Slice | Evidence | Result |
| --- | --- | --- |
| G0 provider ledger | `jq` parses 1,745 entries with 4 executable / 1,618 implementable-now / 64 provider-restriction / 45 multipart-foundation / 14 justified-excluded dispositions. | Passed for row and disposition counts; G1 records a provider `source_url` on each matching command in `cli_surface.json`. |
| G1 command reachability | Built `./cmd/pm` resolves `pm gitlab`, `pm gitlab projects list --help`, `groups list --help`, `users list --help`, and `issues list --help`; `TestEveryImplementedCommandPassesRuntimePreflight` covers the real preflight. | Passed; each help path renders, and a non-executing command invocation reaches project initialization (not an unknown command) without credentials or a provider call. |
| Bundle validity | `go run ./cmd/connectorgen validate internal/connectors/defs/gitlab` and `go run ./cmd/connectorgen surface-sync --check internal/connectors/defs/gitlab` pass. | Passed. |
| Documentation parity | GitLab connector manual/website catalog are regenerated or verified through the repository's docs validation command. | Passed; `pm docs generate`, `pnpm --dir website gen:catalog`, and `pm docs validate --connectors-dir docs/connectors` complete. |
| Scope/safety | Connector-boundary and secret-literal validation pass; no shared runtime or live-provider activity. | Passed; no GitLab credential, live request, provider write, shared runtime edit, or output-redaction declaration was added. |
