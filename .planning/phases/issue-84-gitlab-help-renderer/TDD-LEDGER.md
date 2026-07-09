# TDD Ledger: GitLab Help Renderer (#84)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-84-gitlab-help-renderer --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

- RED: `go test ./internal/cli -run 'TestGitLabCommandSurfaceHelp' -count=1` failed because bare `pm gitlab` returned `missing connector command path` and `pm gitlab --help`/`pm help gitlab` returned `help topic "gitlab" not found`.
- GREEN: `internal/cli` now resolves connector topics through `connectors.RenderConnectorManual`; bare connector invocations render contextual manual help and exit 0.
- GREEN: JSON help emits `CommandManual` for `pm --json gitlab --help`.
- DOCS: `docs/connectors/gitlab/MANUAL.md`, website `gitlab-cli-surface.mdx`, `cli-reference.mdx`, and generated `website/lib/docs.generated.ts` refreshed.
