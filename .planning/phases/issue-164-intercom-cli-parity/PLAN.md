# Plan: Intercom CLI Parity Parent Orchestration

Parent issue: #164
Parent branch: `feat/164-intercom-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/220 (draft; opened after the plan seed commit)
Default branch: `main`

## GSD Command Path

- Adapter health: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json` all passed on 2026-07-09.
- Requested programming-loop path attempted: `scripts/gsd prompt programming-loop init --phase issue-164-intercom-cli-parity --dry-run`.
- Result: unavailable in this checkout (`scripts/gsd: unknown GSD command: programming-loop`).
- Fallback path used: `scripts/gsd prompt quick --full "issue #164 Intercom CLI parity parent planning and issue #165 CLI surface metadata"`, plus the manual GSD loop required by `AGENTS.md`: plan, TDD ledger, verification checklist, red/green/refactor evidence, coherent commits, and review routing.

## Required Skills Loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-spf13-viper`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`

Required references loaded: `AGENTS.md`, issue/parent contracts, parent orchestration loop, stacked PR workflow, GSD universal runtime loop, CodeRabbit and automated-review routing loops, GSD Pi adapter reference, CLI help/docs/website parity reference, connector migration handoff, conventions, and architecture design.

## Objective

Coordinate Intercom connector CLI parity across sub-issues #165-#171 while preserving safety gates and stacked-PR review coverage. Every official Intercom REST API 2.14 operation must become one of: stream, direct read, typed reverse-ETL write, bounded binary/file policy, or explicit blocked-by-default operation metadata with a reason.

## Initial State

- Branch `feat/164-intercom-cli-parity` exists locally and starts at `origin/main`.
- No parent PR existed at planning start; draft parent PR #220 now tracks `feat/164-intercom-cli-parity` → `main`.
- Current Intercom bundle has 5 streams, 10 `api_surface.json` entries, and no write actions.
- Official baseline: 149 operations across 105 paths; method split GET 67, PUT 16, POST 47, DELETE 19.
- No credentials are needed or allowed for this phase.

## Parent Orchestration Steps

1. Create parent and sub-issue planning artifacts before production edits.
2. Seed the parent branch with the planning checkpoint and open a draft parent PR to `main` before sub-PR execution.
3. Execute #165 locally as the first critical-path slice because the current Pi tool surface has no subagent tool and the parent PR is missing.
4. After #165 is green and pushed, either open a stacked #165 PR targeting `feat/164-intercom-cli-parity` or keep it in the parent branch only if the coordinator explicitly chooses single-branch integration.
5. Continue sub-issues in dependency order and only spawn workers when isolated working directories plus subagent runtime are available.
6. Run automated review routing before any CodeRabbit/Copilot review command. Do not manually request CodeRabbit on a non-draft `main`-target PR unless fallback conditions apply.

## Sub-Issue Queue

| Issue | Goal | Dependencies | Expected write scope | Initial decision |
|---:|---|---|---|---|
| #165 | CLI/API surface metadata | parent plan | `internal/connectors/defs/intercom`, `cmd/connectorgen/*test.go`, `.planning/phases/issue-165-*` | `local_critical_path` |
| #168 | Operation ledger refinement | #165 metadata baseline | `internal/connectors/defs/intercom/api_surface.json`, tests/docs | `not_spawned_dependency_blocked` |
| #166 | Help renderer/docs | #165, dispatcher metadata | help/docs/website and command metadata | `not_spawned_dependency_blocked` |
| #167 | Stream runner | #165/#168 stream classification | Intercom streams/schemas/fixtures/runner tests | `not_spawned_dependency_blocked` |
| #169 | Direct reads | #165/#168 | direct-read operation metadata/engine/docs/tests | `not_spawned_dependency_blocked` |
| #170 | Advanced query/binary engine | #168/#169 | provider-specific query/body/binary policy | `not_spawned_dependency_blocked` |
| #171 | Sensitive/admin policy | #168/#170 | writes/risk/approval/typed confirmation | `not_spawned_dependency_blocked` |

## TDD Strategy

- Use data-contract tests for the official Intercom API surface count and method split.
- Add/red-run validation before modifying Intercom production metadata.
- Keep every behavior-changing slice constrained by targeted tests and `connectorgen validate`.
- Do not execute reverse ETL or live connector checks.

## Verification Gates

Parent-level final handoff requires:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Issue-level slices may run focused gates first, then escalate to the parent gates before handoff.

## CLI Help / Docs / Website Parity

Applies to this parent roadmap. Each CLI-visible sub-issue must record:

- `pm help <topic>` or explicit not-applicable status.
- `pm <namespace>` bare help behavior or explicit not-applicable status.
- `pm <command> --help` or explicit not-applicable status.
- `docs/cli/**`, `website/**`, generated help/manual artifacts, and tests updated or explicitly exempted.

#165 is metadata-only and does not add runtime command dispatch; runtime help/docs/website changes are deferred to #166 unless metadata changes surface through generated docs.

## Human Gates

- Parent PR merge into `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Credentialed connector checks.
- Destructive/admin external actions.
- Reverse ETL execution beyond plan/preview/approval/execute.
- Production deploys.
- Quality-gate reductions.
- Generic shell, raw generic HTTP write, generic SQL write, or unrestricted raw API tools.
