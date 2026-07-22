## Summary

Deliver the production in-process Pi AgentSession Shepherd matrix and close the bounded release
blockers around issue-driven plan bootstrap, AgentSession-selected host verification, trusted-local
execution, same-second parent readiness, and complete-suite CI.

Refs #479
Refs #471

## Stacked PR

- Base: `feat/471-pi-agent-session-shepherd`
- Parent PR: #472
- Parent-to-`main` merge remains human-gated.

## Scope

- Production intake, scheduler, controller, child pipeline, isolated worktree/Git lifecycle,
  GitHub/review/human-gate adapters, effect recovery, and Pi composition.
- Complete 17-row production behavior matrix plus focused release-blocker correction.
- Complete sequential Shepherd GitHub Actions gate.
- Exact post-install assertion for Pi 0.80.6's published shrinkwrap and installed package family.
- GSD/TDD/verification/handoff artifacts.

## GSD / TDD / Skills

- Manual-GSD fallback recorded because `scripts/gsd doctor` passes but the adapter does not expose
  `programming-loop`.
- Behavior-level correction RED/GREEN evidence is in `TDD-LEDGER.md`; the historical original
  all-row RED-evidence gap remains explicit.
- Skills: `gsd-programming-loop`, `github-issue-first-delivery`,
  `architecture-patterns`, `javascript-testing-patterns`, `golang-how-to`,
  `golang-testing`, `golang-continuous-integration`, `golang-documentation`,
  `golang-lint`, and `golang-security`.

## Verification

- Focused release-blocker suite: 767/767 pass.
- Strict production TypeScript: pass across 49 non-test modules.
- Pinned offline Pi 0.80.6 RPC: pass.
- Exact Pi shrinkwrap/runtime-family assertion: pass.
- Complete local inventory: 1,712 total; 1,647 pass; 64 blocked before assertions by managed-sandbox
  `/bin/ps` `spawn EPERM`; 1 skip. The new GitHub Actions job is the required ordinary-host gate.
- Branch-range and worktree `git diff --check`: pass.
- GSD/TDD workflow evidence check: pass.

## CLI Help / Docs / Website Parity

This closure adds no Go `pm` command or command behavior. Existing `pm-shepherd` Pi registration
passes offline RPC. `docs/cli/**`, website docs, and generated `pm` manuals are not applicable.

## Safety Boundaries

- No generic shell/Git/GitHub/default-branch-merge capability is exposed to child AgentSessions.
- No secret is requested, printed, persisted, summarized, or placed in prompts.
- Child integration targets only the non-default parent branch.
- Shepherd observes, but never performs, the human-owned parent merge to `main`.

## Review / Deferred Gates

- Internal Codex 5.6 Sol xhigh review at `ca3f6c6f` found two blockers. The deterministic Pi-family
  assertion is corrected at `a594be98`; the parent ledger is reconciled at `45c27b9d` and merged
  into the child at `766709b3`. Authoritative parent-first publication/range verification remains
  external.
- Repository policy still requires `claude_auto` coverage or an allowed recorded fallback. Codex
  review is retained as an additional internal exact-head gate and is not mislabeled as Claude.
- Before opening this PR: publish/fetch parent `45c27b9d`, verify its authoritative head, then
  push this child and confirm the GitHub range.
- Before parent integration: remote complete-suite CI and policy review coverage must be clean.
