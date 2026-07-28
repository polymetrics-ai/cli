# Prompt snapshot — CLI PM Broker profile/context foundation

## Kickoff

- Task: add CLI-side PM Broker profile/context/domain foundation for Organization, Workspace, Environment, BrokerProfile, runtime mode selection, and incompatible contract-version handling without live provider operations.
- Runtime: Pi in disposable worktree.
- Base: `origin/integration/pm-broker-production-program`.
- Branch: `fm/cli-pm-broker-profile-context-r1`.
- Sub-issue PR base: `integration/pm-broker-production-program`.
- Primary issue: CLI #566; parent PR #593.
- PM Broker contract references: PM Broker PR #33 and merged PR #35.
- GSD: `scripts/gsd doctor` and `scripts/gsd list` passed; `scripts/gsd prompt plan-phase cli-pm-broker-profile-context-r1 --skip-research` generated the planning prompt; `scripts/gsd prompt programming-loop init --phase cli-pm-broker-profile-context-r1 --dry-run` is unavailable (`unknown GSD command: programming-loop`), so manual-GSD fallback is recorded.
- Execution decision: `local_critical_path` because the scope is one small dependency-ordered CLI/domain slice in an isolated worktree.
- Downstream artifact: pending.
- Verification result: pending.
