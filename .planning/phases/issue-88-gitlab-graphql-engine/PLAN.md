# Plan: GitLab Graphql Engine (#88)

Parent issue: #78
Parent PR: #127
Branch: `feat/78-gitlab-cli-parity`

## GSD Evidence

- Lane prompt: `scripts/gsd prompt execute-phase issue-88-gitlab-graphql-engine --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry; continue manual universal GSD loop with red/green/refactor evidence.

## Required Skills Loaded

gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation

## Goal

Record GitLab GraphQL as not required for this REST-backed slice; keep fixed-document-only policy.

## Scope

planning/docs only unless needed

## TDD / Execution Plan

1. Add or update failing tests/validation evidence for this lane before production changes.
2. Implement the smallest safe slice.
3. Keep GitLab command metadata honest: do not enable raw generic API writes, generic shell/SQL writes, unsafe binary downloads, or reverse ETL execution outside plan → preview → approval → execute.
4. Run focused tests and connector validation.
5. Update ledger and verification artifacts.

## Safety Gates

- No secrets, credentialed GitLab checks, new dependencies, or external writes.
- Sensitive/admin/destructive operations remain blocked by default unless an explicit approved policy exists.
- Binary/file transfer requires bounded executor and output policy before enabling.

## Verification Plan

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```
