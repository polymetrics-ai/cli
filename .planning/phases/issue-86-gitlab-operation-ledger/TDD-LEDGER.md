# TDD Ledger: GitLab Operation Ledger (#86)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-86-gitlab-operation-ledger --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

- GREEN: Rebuilt `internal/connectors/defs/gitlab/api_surface.json` from official GitLab OpenAPI v2 source (1,144 operations: GET 503, POST 241, PUT 263, DELETE 125, PATCH 9, HEAD 3).
- GREEN: Added `operations.json` with 1,144 official operation specs plus one `/users` compatibility stream row required by the existing connector.
- GREEN: `TestBundleLoadGitLabOperationLedgerFromDisk` asserts 1,145 surface rows/operation specs, stream/direct-read coverage, blocked-by-default operation rows, and required model coverage.
