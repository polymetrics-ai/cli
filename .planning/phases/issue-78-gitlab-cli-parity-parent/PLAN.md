# Plan: GitLab CLI Parity Parent Orchestration

Parent issue: #78
Parent PR: #127 (draft)
Parent branch: `feat/78-gitlab-cli-parity`
Default branch: `main`
Connector: `gitlab`
Definition scope: `internal/connectors/defs/gitlab/`

## GSD Command Evidence

- Adapter health: `scripts/gsd doctor` passed on 2026-07-09.
- Pi adapter verification: `scripts/gsd verify-pi` passed on 2026-07-09.
- Command registry: `scripts/gsd list --json` completed on 2026-07-09.
- Parent planning prompt: `scripts/gsd prompt plan-phase issue-78-gitlab-cli-parity --skip-research` generated the Pi `/gsd-plan-phase` prompt.
- Programming-loop command fallback: `scripts/gsd prompt programming-loop init --phase issue-78-gitlab-cli-parity --dry-run` returned `unknown GSD command: programming-loop`. Use the manual universal programming loop from `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` while preserving TDD evidence.

## Required Skills Loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-spf13-cobra` for CLI command-tree awareness if runtime help code is touched.

## Scope And Sub-Issue Queue

| Issue | Lane | Status | Dependency / collision note |
| ---: | --- | --- | --- |
| #83 | CLI surface metadata | local critical path | First slice; owns `internal/connectors/defs/gitlab/cli_surface.json` and focused metadata tests. |
| #84 | Help renderer/docs | planned | Depends on #83 metadata. |
| #85 | Stream runner | planned | Depends on #83 metadata and existing generic command runner behavior. |
| #86 | Operation ledger | planned | Write-scope collision with #83 inside `defs/gitlab`; run after metadata slice. |
| #87 | Direct reads | planned | Depends on #86 ledger classifications. |
| #88 | GraphQL/advanced engine | planned | Depends on #86 and direct/read operation needs. |
| #89 | Sensitive/admin policy | planned | Depends on #86 operation classifications and write inventory. |

## Parent Deliverables

- Draft parent PR from `feat/78-gitlab-cli-parity` to `main`; final merge remains human-gated.
- Issue-scoped planning artifacts under `.planning/phases/issue-<N>-.../`.
- Status table for official operations inventoried, mapped by app type, implemented, blocked/deferred, and verified.
- Connector artifacts under `internal/connectors/defs/gitlab/` plus docs/website/help updates when a lane makes CLI-visible behavior available.

## First Slice: #83 CLI Surface Metadata

1. Create #83 planning artifacts before production edits.
2. Add a failing test that `engine.Load(defs.FS, "gitlab")` exposes non-nil CLI surface metadata with implemented ETL commands for the four existing GitLab streams.
3. Add `internal/connectors/defs/gitlab/cli_surface.json` with glab-inspired safe command metadata:
   - implemented ETL commands only for existing streams (`projects`, `groups`, `users`, `issues`);
   - planned/partial commands for safe future direct-read and reverse-ETL candidates;
   - local workflow and raw API command families marked unsupported/unsafe, not implemented;
   - no generic raw HTTP write, shell, SQL, or arbitrary GraphQL mutation escape hatches.
4. Validate with focused tests and `connectorgen validate`.
5. Update #83 ledger and verification evidence.

## CLI Help / Docs / Website Parity Plan

#83 adds metadata used by future help/runner lanes. Runtime help/docs/website generated pages are not overclaimed in this slice. #84 owns rendered help/docs parity. Any #83 doc changes must explicitly state metadata-only status and avoid claiming new executable commands beyond existing stream-backed ETL mappings.

## Safety Gates

- No secrets in prompts, logs, examples, fixtures, or docs.
- No credentialed GitLab checks.
- No new dependencies.
- No broad raw HTTP write, generic shell write, generic SQL write, or arbitrary GraphQL mutation tooling.
- No binary downloads until #87/#88 define explicit bounded output policy.
- Reverse ETL remains plan → preview → approval → execute; #83 metadata may describe planned writes but must not execute them.
- Parent PR merge to `main` remains human-gated.

## Verification Checklist

Minimum parent gates before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Focused #83 gates are listed in `.planning/phases/issue-83-gitlab-cli-surface/VERIFICATION.md`.
