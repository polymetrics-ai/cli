# Issue 476 Verification

Status: `pass`

No verification result is claimed before the declared commands run. Full verification requires a
successful `make verify`; focused checks alone do not set `verificationPassed`.

| Gate | Result | Evidence |
|---|---|---|
| focused issue tests | pass | 16 tests passed, 0 failed; genuine temporary local repositories |
| full Shepherd tests | pass | 153 tests passed, 0 failed |
| strict TypeScript / Pi 0.80.6 | pass | production adapters passed strict no-emit TypeScript; installed `pi --version` is `0.80.6` |
| Pi extension discovery | pass with documented fallback | Exact `pi --list-extensions` is unsupported by Pi 0.80.6 (`Unknown option`); documented offline RPC `get_commands` returned `true` for `pm-shepherd` from `extension` |
| diff hygiene | pass | `git diff --check` |
| `go vet ./...` | pass, later superseded | Completed before parent policy narrowed sub-worker gates |
| `go test ./...` | `cancelled_by_parent_policy` | First run exceeded `internal/connectors/certify`'s 10-minute timeout under parallel CPU contention; after `make verify` passed with the repository's 20-minute timeout, the exact retry was stopped at parent direction (exit 143). This is not recorded as a functional branch failure. |
| `go build ./cmd/pm` | pass, later superseded | Completed before parent policy narrowed sub-worker gates |
| `make verify` | pass, later superseded | Completed before policy change: tests passed including certify in 522.917s; docs/smoke passed; lint reported 0 issues; 547 connector definitions reported 0 findings |
| exact remote head | pending final evidence push | Local/remote equality will be checked after the final evidence commit |
| stacked ready PR | pending creation | Opens after the final evidence commit |

Runtime-backed services are not part of this issue and will not be started.

## Parent verification-policy update

The parent coordinator narrowed local sub-worker verification after the full `make verify` result:
focused issue tests, complete Shepherd tests, strict TypeScript, offline Pi extension smoke, and diff
hygiene are the authoritative local gates. Full Go/connectors verification is centralized at parent
integration and GitHub CI. Any still-active or planned full-repository retry was therefore stopped
and classified `cancelled_by_parent_policy`.
