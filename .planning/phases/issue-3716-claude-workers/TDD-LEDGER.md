# TDD LEDGER — issue #3716 clean project-local Claude workers

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | Both Claude workers have generated, parseable frontmatter with a minimal allowlist and no `Agent` | `go test ./internal/agentcontract -run TestClaudeProjectWorkersBlockAmbientAgentDelegation -count=1` failed because both project projections are absent | Pending | Pending |
| R2 | A generated worker cannot delegate to an ambient user/plugin agent | The same executed RED test reported: `pm-delivery-worker can resolve to an ambient same-name agent: missing project projection .claude/agents/pm-delivery-worker.md leaves no project-local tools allowlist that omits Agent` (and the equivalent connector failure) | Pending | Pending |
| R3 | Missing required Claude projections are generated and any frontmatter/body drift fails `check` | Pending | Pending | Pending |
| R4 | Installed CLI discovers project workers by name while ambient fixtures remain non-delegable | Pending | Pending | Pending |
