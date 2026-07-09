# Plan: Zendesk CLI Parity Parent Orchestration

Parent issue: #156
Parent branch: `feat/156-zendesk-cli-parity`
Parent PR: pending creation after this plan seed commit
Default branch: `main`

## GSD Command Path

- Adapter health: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json` passed on 2026-07-09.
- Planning prompt used: `scripts/gsd prompt plan-phase issue-156-zendesk-cli-parity --skip-research`.
- Implementation prompt requested: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-cli-parity --dry-run`.
- Manual GSD fallback: the adapter returned `unknown GSD command: programming-loop`; use the repo-local `execute-phase` prompt plus the manual universal programming loop (plan, red evidence, green implementation, refactor, verification, summary) and record this fallback in issue artifacts.

## Required Skills Loaded

- GSD: `gsd-core`.
- Go orchestrator: `golang-how-to`.
- CLI/connector/runtime: `golang-cli`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`.
- Quality/safety: `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`.
- Documentation/help parity: `golang-documentation` plus `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Scope

Create a parent orchestration lane for the Zendesk connector, then execute sub-issues #157-#163 through stacked branches/PRs targeting `feat/156-zendesk-cli-parity`. The parent branch remains the integration branch and final merge to `main` is human-gated.

## Full-Surface Target

Every operation in the official Zendesk OAS (`https://developer.zendesk.com/zendesk/oas.yaml`) must end in exactly one durable classification:

1. `covered_by.stream` for durable ETL collection reads.
2. `covered_by.direct_read` or `covered_by.direct_reads` for bounded single-resource/read-query commands.
3. `covered_by.write` for named reverse-ETL actions.
4. `covered_by.binary_read` or a bounded direct-read/binary metadata policy for file/binary surfaces.
5. A blocked operation row only for duplicate, deprecated, disallowed, auth-internal, or explicit product-scope blockers.

Sensitive/admin/destructive operations are not blanket exclusions; they require typed reverse-ETL actions, risk text, approval text, redaction policy, and `confirm: destructive` where applicable.

## Sub-Issue Work Queue

| Issue | Lane | Dependencies | Expected write scope | Initial decision |
| ---: | --- | --- | --- | --- |
| #157 | CLI surface metadata | parent PR exists | `internal/connectors/defs/zendesk/`, `.planning/phases/issue-157-*` | local_critical_path after parent PR seed |
| #160 | Operation ledger | #157 OAS inventory | `internal/connectors/defs/zendesk/api_surface.json`, optional generated ledger artifacts | not_spawned_dependency_blocked |
| #159 | Stream runner | #157, #160 stream candidates | `internal/connectors/defs/zendesk/streams.json`, schemas, fixtures | not_spawned_dependency_blocked |
| #161 | Direct reads | #157, #160 direct-read candidates | `internal/connectors/defs/zendesk/cli_surface.json`, operations/direct-read metadata | not_spawned_dependency_blocked |
| #162 | Advanced query/binary engine | #160 binary/query candidates; may feed #161/#159 | connector defs and engine only if a validated gap exists | not_spawned_dependency_blocked |
| #163 | Sensitive/admin policy | #160 write/sensitive candidates | `writes.json`, schemas, policy metadata | not_spawned_dependency_blocked |
| #158 | Help renderer/docs | after command metadata stabilizes | docs/help/website plus generated artifacts | not_spawned_dependency_blocked |

## Orchestration / Parallelism

This Pi harness exposes `read`, `bash`, `edit`, and `write`, but no `subagent` tool. Mutating worker spawn is therefore unavailable in this session. The parent orchestrator will run the critical path locally, one stacked sub-issue branch at a time, and record `not_spawned_runtime_capability_missing` for independent lanes until an isolated worker runtime is available.

If subagent tooling becomes available, spawn only disjoint work scopes in separate worktrees. Shared parent artifacts, parent PR state, sub-PR merge decisions, and automated review routing stay parent-orchestrator owned.

## Slice Boundaries

### Parent seed slice

- Create parent GSD plan, TDD ledger, verification checklist, run state, and orchestration state.
- Commit and push the plan seed.
- Open a draft parent PR from `feat/156-zendesk-cli-parity` to `main`.

### First implementation slice (#157)

- Create branch `feat/157-zendesk-cli-surface-metadata` from the parent branch after the parent PR exists.
- Add red validation/test evidence that Zendesk bundle CLI/API surface metadata is absent.
- Build the minimal safe Zendesk connector bundle scaffold from the official OAS: identity/spec/docs/check metadata plus full OAS operation inventory blocked-by-default as metadata.
- Add `cli_surface.json` command inventory mapped to safe app intents without exposing raw generic HTTP writes.
- Validate with targeted tests and `connectorgen validate`.

## CLI Help / Docs / Website Parity Checklist

Applies to CLI-visible connector surface metadata. For #157, runtime dispatcher/help changes are not implemented, but metadata must be safe for later rendering.

- [ ] `pm help <topic>` checked when help renderer changes (#158) or marked not applicable for metadata-only slice.
- [ ] `pm <namespace>` bare behavior checked when runtime command tree changes or marked not applicable for metadata-only slice.
- [ ] `pm <command> --help` checked when command runtime changes or marked not applicable for metadata-only slice.
- [ ] `docs/cli/**` updated or marked not applicable for metadata-only slice.
- [ ] `website/**` updated or marked not applicable for metadata-only slice.
- [ ] Generated help/manual artifacts updated or marked not applicable for metadata-only slice.
- [ ] Tests cover metadata validation and later renderer behavior as slices land.

## Human Gates

- Parent PR merge to `main`.
- New dependencies or unsafe binary downloads.
- Auth scope changes or `gh auth refresh`.
- Secrets, credential values, or credentialed live connector checks.
- Destructive external actions or production deploys.
- Quality gate reductions.
- Reverse ETL execution without plan → preview → approval → execute.
- Generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tooling.
