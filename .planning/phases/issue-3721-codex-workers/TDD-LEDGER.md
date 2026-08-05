# TDD LEDGER — issue #3721 Codex project-local workers

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | Generated workers cannot delegate to an ambient agent | `go test ./internal/agentcontract -run '^TestCodexWorkersCannotDelegateToAmbientAgents$' -count=1` failed: `pm-delivery-worker can delegate to ambient agent "worker" because agents.enabled is not false` | Pending | Pending |
| R2 | Every generated role is valid standalone TOML with required fields | Pending | Pending | Pending |
| R3 | Removing delegation isolation causes drift failure | Pending | Pending | Pending |
| R4 | Required Codex projections are generated only from canonical source | Pending | Pending | Pending |
| R5 | Trust/collision claims remain constrained to official evidence | Pending | Pending | Pending |

## Safety boundary

The implementation must prove generated `agents.enabled = false` configuration is present and
drift-protected. It must not claim that Codex loads project configuration before a project is
trusted, or that user/project standalone filename collision precedence is documented.
