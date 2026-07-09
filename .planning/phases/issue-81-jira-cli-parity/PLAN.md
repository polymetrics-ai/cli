# Plan: Jira CLI Parity Parent Orchestration

Parent issue: #81
Parent branch: `feat/81-jira-cli-parity`
Default branch: `main`
Parent PR: not yet open at planning start; this plan is the deliberate parent seed artifact for the draft parent PR.

## GSD / Runtime Evidence

- GSD parent plan prompt: `scripts/gsd prompt plan-phase issue-81-jira-cli-parity --skip-research` (generated successfully, 2026-07-09).
- GSD programming loop attempt: `scripts/gsd prompt programming-loop init --phase issue-81-jira-cli-parity --dry-run` failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback: active. Follow `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` manually: plan, red test, green slice, refactor, verify, commit/push, review route.
- Pi subagent status: the current harness exposes `read`, `bash`, `edit`, and `write` only; no `subagent` tool is available. Mutating workers are not spawned from this checkout. Record local critical-path work or `not_spawned_runtime_capability_missing` per cycle.

## Required Skills Loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-graphql`
- `golang-lint`
- `golang-documentation`

## Sources And Constraints

- Parent issue #81 and sub-issues #104-#110 read through `gh issue view --json`; all are open.
- Official Jira source: Atlassian Jira Cloud OpenAPI `swagger-v3.v3.json` (620 operations: GET 276, POST 135, PUT 119, DELETE 90) per mission prompt.
- Current baseline: `internal/connectors/defs/jira/` has 3 implemented streams (`issues`, `projects`, `users`) and 12 excluded API-surface aggregate rows; no `cli_surface.json`, no `operations.json`, no writes.
- Safety: no secrets, no credentialed checks, no new dependencies, no generic raw HTTP write, no shell/SQL write tools, no destructive/sensitive writes without plan → preview → approval → typed confirmation.

## Parent Workflow

1. Seed parent orchestration artifacts under `.planning/phases/issue-81-jira-cli-parity/`.
2. Commit and push the parent seed slice to `feat/81-jira-cli-parity`.
3. Open a draft parent PR from `feat/81-jira-cli-parity` to `main` using `Refs #81` until all sub-issues are integrated.
4. Execute sub-issues with separate evidence under `.planning/phases/issue-<N>-.../`.
5. Prefer stacked sub-issue branches/PRs when the slice is large. In this harness, begin #104 locally as `local_critical_path`; split before push if the diff grows beyond the CLI-surface lane.
6. Keep parent PR draft while sub-issues land. The final parent PR uses `Closes #81` only when ready for human approval.
7. Do not merge the parent PR to `main`.

## Sub-Issue Queue

| Issue | Lane | Dependencies | Expected write scope | Initial state |
| ---: | --- | --- | --- | --- |
| #104 | CLI surface metadata | none | `internal/connectors/defs/jira/`, tests, `.planning/phases/issue-104-jira-cli-surface/` | first local critical path |
| #105 | Help renderer/docs | #104 | help/docs/website/rendering paths plus Jira docs | blocked on #104 metadata |
| #106 | Stream runner | #104/#105 | command runner/CLI + Jira metadata | blocked on metadata/help |
| #107 | Operation ledger | #104 | `internal/connectors/defs/jira/api_surface.json`, possible `operations.json` | can follow #104 |
| #108 | Direct reads | #107 | command runner/engine/direct-read + Jira metadata | blocked on ledger |
| #109 | GraphQL/advanced body support | #107 | engine/body/GraphQL metadata if Jira needs it | blocked on ledger; may become N/A |
| #110 | Sensitive/admin policy | #107 | policy metadata/validator + Jira writes blocked by default | blocked on ledger |

## Slice 1: Issue #104 CLI Surface Metadata

Goal: add validated Jira command-surface metadata without enabling new dispatcher behavior.

Planned red test:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Expected first failure: Jira `CLISurface` is nil because `internal/connectors/defs/jira/cli_surface.json` is absent.

Green implementation:

- Add `internal/connectors/defs/jira/cli_surface.json` with provider-like Jira commands mapped to existing streams and safe unsupported/planned classifications.
- Only `availability=implemented` commands target existing streams (`issues`, `projects`, `users`) and covered API-surface endpoints.
- Writes remain unimplemented, blocked/planned/unsafe with plan-preview-approval notes, not executable write refs.
- Validate with `go run ./cmd/connectorgen validate internal/connectors/defs --json`.

## Verification Plan

Targeted first:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

CLI parity checks when runtime help/docs are touched:

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors
go run ./cmd/pm connectors inspect jira --json
rg -n "jira|Jira" docs/cli website internal/connectors/defs/jira
```

## Human Gates

- Parent PR merge to `main`.
- Auth scope changes or `gh auth refresh`.
- Secrets or credentialed connector checks.
- New dependencies.
- Destructive external actions or production deploys.
- Quality-gate reductions.
- Generic shell/HTTP/SQL write tools.
- Reverse ETL execution outside plan → preview → approval → execute.
