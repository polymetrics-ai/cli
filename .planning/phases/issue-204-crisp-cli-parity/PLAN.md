# GSD Plan — Issue #204 Crisp CLI parity parent roadmap

Parent issue: https://github.com/polymetrics-ai/cli/issues/204
Branch: `feat/204-crisp-cli-parity`
Default branch: `main`
Connector: `crisp`
Definition scope: `internal/connectors/defs/crisp/`
GSD command path: `scripts/gsd prompt plan-phase 204 --skip-research`
Programming loop fallback: `scripts/gsd prompt programming-loop init --phase issue-204-crisp-cli-parity --dry-run` returned `unknown GSD command: programming-loop`; manual GSD/TDD loop recorded here per adapter fallback rules.

## Required context loaded

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-lint`

## Parent acceptance target

Every official Crisp REST API operation from https://docs.crisp.chat/references/rest-api/v1/ must land in exactly one final class before the parent PR is human-ready:

1. `covered_by.stream` for durable ETL reads.
2. `covered_by.direct_read` / `covered_by.direct_reads` for bounded safe read commands.
3. `covered_by.write` for typed reverse-ETL writes.
4. `covered_by.binary_read` or bounded binary metadata/download policy.
5. Blocked only when duplicate, deprecated, disallowed, auth-internal, or explicitly out of product scope.

Sensitive/admin/destructive operations are not blanket exclusions. They must become typed reverse-ETL actions with risk text, explicit schemas, redaction, approvals, and destructive typed confirmation where required.

## Sub-issue queue

| Issue | Lane | Dependency | Expected write scope | Initial state |
|---:|---|---|---|---|
| #205 | CLI surface metadata | parent PR exists | `internal/connectors/defs/crisp/{metadata.json,spec.json,streams.json,api_surface.json,cli_surface.json,docs.md}` plus issue #205 planning | worker_ready |
| #208 | Operation ledger | #205 baseline | `internal/connectors/defs/crisp/api_surface.json`, `.planning/phases/issue-208-*` | planned |
| #207 | Stream runner | #205 + #208 stream classifications | `internal/connectors/defs/crisp/streams.json`, `schemas/**`, fixtures | planned |
| #209 | Direct reads | #205 + #208 direct-read classifications | `internal/connectors/defs/crisp/operations.json`, `cli_surface.json`, direct-read docs/tests if needed | planned |
| #210 | Advanced query/binary engine | #208 binary/query gaps | provider-specific operation metadata, bounded binary policy docs/tests | planned |
| #211 | Sensitive/admin policy | #208 write/admin classifications | writes/operations risk policy, schemas, redaction docs/tests | planned |
| #206 | Help renderer/docs | after executable metadata exists | connector docs/help/manual/website generated artifacts | planned |

## Orchestration decision

Cycle `plan`: `local_critical_path`.
Reason: Parent PR is missing. Current Pi tool surface in this harness exposes no `subagent` tool, so no isolated mutating workers can be spawned from this coordinator. The orchestrator will create parent planning artifacts, commit/push a parent seed, open a draft parent PR, then run #205 locally or on a stacked branch. Record `not_spawned_runtime_capability_missing` for ready workers until a runtime with subagent support is available.

## TDD / validation strategy

- Parent planning slice: no production behavior change; no red test required. Validate GSD adapter and issue context.
- #205 first production slice: red validation is `go run ./cmd/connectorgen validate internal/connectors/defs/crisp` before the bundle exists (expected failure). Green is the same command after adding a non-executable metadata/ledger scaffold.
- Later behavior slices (#207/#209/#211) must add targeted failing tests or fixtures before implementing stream/direct-read/write behavior.

## Verification checklist

Parent planning checkpoint:

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase 204 --skip-research
```

Issue #205 targeted checkpoint:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
go run ./cmd/connectorgen validate internal/connectors/defs
```

Parent handoff checkpoint after integrated implementation slices:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

CLI parity checks become mandatory once executable command/help surfaces land:

```bash
pm help connectors
pm connectors
pm connectors inspect crisp --json
rg -n "crisp|Crisp" docs/cli website
```

## Safety gates

- No credentialed Crisp checks unless explicitly requested.
- No secrets in prompts, fixtures, docs, logs, or PR body.
- No new dependencies without human approval.
- No generic HTTP write, shell write, SQL write, or raw API write command.
- Reverse ETL remains plan → preview → approval → execute.
- Parent PR merge to `main` remains human-gated.

## Commit / PR checkpoints

1. Parent planning seed commit on `feat/204-crisp-cli-parity`; push and open draft parent PR to `main` with `Refs #204`.
2. #205 stacked branch from parent branch; commit red validation evidence when useful, then metadata/ledger scaffold after green targeted validation.
3. Later sub-PRs target `feat/204-crisp-cli-parity` and use `Refs #<subissue>` + `Refs #204`.
4. Parent PR body changes from `Refs #204` to `Closes #204` only when all accepted subissues are integrated and final verification/review gates are clean.
