# TDD LEDGER — issue #3721 Codex project-local workers

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | Generated workers cannot delegate to an ambient agent | `go test ./internal/agentcontract -run '^TestCodexWorkersCannotDelegateToAmbientAgents$' -count=1` failed: `pm-delivery-worker can delegate to ambient agent "worker" because agents.enabled is not false` | `TestCodexWorkersCannotDelegateToAmbientAgents` parses each generated TOML and proves `agents.enabled` is explicitly false before modeling built-in `worker` and ambient custom-agent reachability. | Full focused package test passed; the test deliberately does not invoke a live model. |
| R2 | Every generated role is valid standalone TOML with required fields | Added parser assertions before renderer support. | The same isolation test parses both standalone TOML files with the repository's Viper TOML parser and checks `name`, `description`, and `developer_instructions`. | `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1` passed. |
| R3 | Removing delegation isolation causes drift failure | Added `TestCodexProjectionDriftRejectsDelegationRegression`. | The test changes `agents.enabled = false` to `true`, observes `CheckProjections` fail, then observes sync restore the file. | `go run ./cmd/agentcontractgen check` passed after generation. |
| R4 | Required Codex projections are generated only from canonical source | A blank projection root was accepted when all targets were optional. | `TestOptionalNonCodexProjectionsMayBeAbsent` proves required Codex projections fail absent, then sync creates exactly the two Codex files while optional adapters remain absent. | Full-file drift comparison protects the standalone TOML files; `TestProjectionIORejectsSymlinkEscape` covers the new required Codex parent path. |
| R5 | Trust/collision claims remain constrained to official evidence | Canonical source had no Codex evidence surface. | Contract validation requires the two official URLs plus nonempty trust, precedence, discovery, and collision guidance; generated files and README carry that limited language. | The collision rule remains explicitly undocumented; no test or prose invents an answer. |

## Safety boundary

The implementation must prove generated `agents.enabled = false` configuration is present and
drift-protected. It must not claim that Codex loads project configuration before a project is
trusted, or that user/project standalone filename collision precedence is documented.
