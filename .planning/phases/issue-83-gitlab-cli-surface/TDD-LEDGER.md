# TDD Ledger: GitLab CLI Surface Metadata (#83)

## 2026-07-09 — planned red test

### GSD / Skill Evidence

- GSD lane prompt: `scripts/gsd prompt execute-phase issue-83-gitlab-cli-surface --tdd`.
- Manual programming-loop fallback is recorded because `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-spf13-cobra`.

### Red Target

Add `TestBundleLoadEmbeddedGitLabCLISurface` before creating `internal/connectors/defs/gitlab/cli_surface.json`.

Expected initial failure:

```text
GitLab CLISurface is nil; defs.FS must embed cli_surface.json
```

### Green Target

Create schema-valid `internal/connectors/defs/gitlab/cli_surface.json` where:

- `project list`, `group list`, `user list`, and `issue list` are `intent=etl`, `availability=implemented`, and point to the existing streams.
- Future direct-read, reverse-ETL, local workflow, raw API, binary, and admin/destructive commands are planned/unsupported/unsafe with explicit notes, not executable.
- Examples and notes contain no secret-shaped literals.

### Verification To Record

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
