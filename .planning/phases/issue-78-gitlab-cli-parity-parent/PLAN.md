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
| #83 | CLI surface metadata | completed local / pushed | Added 73 glab-inspired command entries and 4 implemented stream commands. |
| #84 | Help renderer/docs | completed local | `pm help gitlab`, `pm gitlab`, and `pm gitlab --help` render connector manual/help from metadata. |
| #85 | Stream runner | completed local | Fixture test proves `pm gitlab issue list` dispatches through generic stream runner. |
| #86 | Operation ledger | completed local | Inventoried 1,144 official GitLab OpenAPI operations plus `/users` compatibility row; non-enabled operations blocked by default. |
| #87 | Direct reads | completed local | Added `json_redacted` policy and four bounded direct-read commands. |
| #88 | GraphQL/advanced engine | completed local / not-enabled | Recorded GraphQL as not required for this REST slice; no generic GraphQL/raw body executor added. |
| #89 | Sensitive/admin policy | completed local | Operation ledger risk tiers, approval policy, typed confirmation markers, and redaction coverage added while writes remain disabled. |

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

Focused per-lane gates are listed in `.planning/phases/issue-83-*` through `.planning/phases/issue-89-*` verification files.
