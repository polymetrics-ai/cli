# TDD Ledger: GitLab Sensitive/Admin Policy (#89)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-89-gitlab-sensitive-admin --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

- GREEN: Non-enabled GitLab operations are blocked by default with risk labels across direct_read, binary_read, sensitive_reverse_etl, admin_reverse_etl, destructive_action, and deprecated models.
- GREEN: `operations.json` rest writes require approval; secret/destructive write candidates include typed-confirmation sensitive policy with env/stdin input and redaction fields.
- GREEN: Direct-read output uses recursive `json_redacted`; raw API, unsafe binary downloads, generic shell/SQL writes, and GitLab writes remain disabled.
