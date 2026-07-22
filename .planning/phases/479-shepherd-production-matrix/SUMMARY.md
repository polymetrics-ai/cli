# Summary

Status: **17/17 functional matrix complete; PR #489 CI repaired locally, remote GREEN pending.**

Issue #479 delivers the production in-process Shepherd path: canonical issue intake and plan
bootstrap, dependency/collision-aware child scheduling, isolated worktrees, scoped Pi AgentSessions,
host-owned bounded verification, typed commit/push/stacked-PR publication, exact-head
review/correction, non-default-parent integration, crash-safe reconciliation, and an authenticated
human wait. Shepherd has no authority to merge the parent PR into the default branch.

Production behavior is frozen at `78708cbef64b33e54ed32078bf2a107d81126236`; the CI/evidence
implementation checkpoint is `307ea409648e2f293c8a48cc957ffc312cc44542`, and deterministic
Pi-family verification is `a594be98`. Its complete local
inventory reported 1,712 tests: 1,647 passed, 64 failed before their assertions because the managed
sandbox denied `/bin/ps` process-identity discovery with `spawn EPERM`, and one was skipped.
Strict production TypeScript, the pinned offline Pi RPC, GSD/TDD workflow validation, and diff
hygiene pass. Earlier parent-proportional Go/static probes are not used as child release evidence;
the parent policy runs those gates once on the exact integrated parent head.

The merge-readiness closure adds a least-privilege GitHub Actions gate for the complete sequential
Shepherd inventory on ordinary infrastructure, repairs committed diff hygiene, and records this
missing phase summary. Workflow structure, strict TypeScript, offline Pi registration, GSD/TDD
evidence, and diff hygiene pass locally. A fresh CI result and one bounded Codex 5.6 Sol xhigh
exact-head review follow-up remain required before the child may integrate into
`feat/471-pi-agent-session-shepherd`.

The first bounded review at `ca3f6c6f` found that CI must prove Pi's complete package family and
that the local parent must be published before the child PR. The runtime now asserts Pi's published
shrinkwrap and all four installed family packages at exactly 0.80.6. The parent ledger was
reconciled locally at `45c27b9d` and merged into the child at `766709b3` without rewriting
history.

PR #489's first ordinary-host run then exposed an unpinned Go fixture toolchain and inherited
`golang.org/x/text v0.36.0`. Current main `873cd7b2` is synchronized into the parent at
`383fcf93`, yielding `x/text v0.39.0`; parent repair state is `387d41fd`. The child workflow pins
Go `1.25.12` at `52866972` and incorporates the repaired parent at `57ca31ae`. The focused real-Go
fixture, exact Pi runtime family, strict TypeScript, workflow YAML, and `go mod verify` all pass.
The complete inventory and security scan require their fresh ordinary-host reruns because this
managed sandbox blocks `/bin/ps` and vulnerability-database DNS.

The parent PR (#472) remains human-gated. No push or merge to `main` is authorized by this phase.
GitHub authentication is ambient host authority; no secret value is requested, printed, persisted,
summarized, or placed in an agent prompt.
