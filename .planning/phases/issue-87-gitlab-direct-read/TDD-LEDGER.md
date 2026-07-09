# TDD Ledger: GitLab Direct Read (#87)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-87-gitlab-direct-read --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

- RED: `TestDirectReadJSONRedactedPolicyRemovesSensitiveFields` initially failed with `direct read output policy "json_redacted" is not supported`.
- GREEN: Added generic `json_redacted` direct-read policy, command-runner/schema/connectorgen validation support, and GitLab implemented direct-read commands for project/group/user-events/issue view.
- GREEN: Added CLI fixture test for `pm gitlab project view --id 123 --json`, verifying bounded direct read, bearer auth, and redaction markers.
