# TDD Ledger: GitLab Stream Runner (#85)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-85-gitlab-stream-runner --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

- GREEN: Added `TestGitLabCommandSurfaceRunsStreamBackedIssueList`; it creates a local GitLab fixture credential from env and verifies `pm gitlab issue list` dispatches through the generic stream runner.
- SAFETY: Fixture token is supplied via environment variable; no credentialed external GitLab request is made.
