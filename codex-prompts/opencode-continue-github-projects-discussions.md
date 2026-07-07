# OpenCode Continuation Prompt: GitHub Projects/Discussions CLI Parity Slice

You are continuing in repository:

`/Users/karthiksivadas/Development/polymetrics-cli-agents/connector-cli-parity-research`

Branch:

`feat/40-github-projects-discussions`

Use the repo-local OpenCode/GSD system. Do not restart research from scratch. Continue from the
current dirty worktree.

## Required Runtime Contract

Read before acting:

- `AGENTS.md`
- `/Users/karthiksivadas/.codex/skills/gsd-programming-loop/SKILL.md` if available, otherwise use
  repo policy files below
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `docs/prompts/universal-programming-loop-prompts.md`
- `.agents/skills/caveman/SKILL.md`

Run as active orchestrator, not passive planner:

- After each helper/preflight command, record one decision:
  `spawned`, `read_only_spawned`, `local_critical_path`, or exact `not_spawned_*`.
- Use `.opencode/commands/gsd-worker.md` for independent worker scopes.
- Mutating workers need isolated worktrees or working directories.
- Read-only verifier workers may share checkout.
- If no worker can run, record the blocker and do local critical-path work.
- Keep exact commands/test output uncompressed.

## Current Implementation State

Already implemented:

- Fixed-document GraphQL stream support reads GraphQL variables from `config.*`, `query.*`,
  `cursor`, and optional `default`/`omit_when_empty`.
- Static validation now checks GraphQL variable templates in `connectorgen validate` and
  conformance.
- Conformance fixture harness supports stream fixture `read_query` for parameterized reads.
- GitHub connector now has four GraphQL read streams:
  - `projects`
  - `project_items`
  - `discussions`
  - `discussion`
- GitHub `cli_surface.json` maps:
  - `project list` -> stream `projects`
  - `project item-list --project-id` -> stream `project_items`
  - `discussion list` -> stream `discussions`
  - `discussion view --number` -> stream `discussion`
- GitHub `api_surface.json` now has four `GRAPHQL` covered rows.
- GitHub `operations.json` has fixed GraphQL operation ledger rows for the same read operations.
- Website connector data was regenerated with bundled Node.
- OpenCode/Codex agent docs were patched so helper scripts are preflight only and worker spawning is
  explicit.
- New OpenCode worker adapter files were added:
  - `.opencode/agents/gsd-worker.md`
  - `.opencode/commands/gsd-worker.md`

## Verification Already Run

Passed:

```bash
go vet ./...
go build ./cmd/pm
PM_CRONTAB_FILE=$(mktemp) go test ./...
go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/commandrunner -count=1
go test ./internal/connectors/conformance -run 'TestConformance/github$|TestConformance/github/|TestReadRawRecordsWithReplayUsesFixtureReadQuery' -count=1 -v
go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '.findings[] | select(.connector=="github")'
/Users/karthiksivadas/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node node_modules/typescript/bin/tsc --noEmit
/Users/karthiksivadas/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node node_modules/vitest/vitest.mjs run
```

`make verify` was interrupted after `go test ./...` had reached and passed `internal/cli`; it had
not completed `docs-check`, `smoke`, `lint`, or `connectorgen-validate`.

## Remaining Critical Path

1. Inspect current diff and ensure no unrelated files are accidentally included.
2. Run final gates:

```bash
PM_CRONTAB_FILE=$(mktemp) make verify
```

If `make verify` is too slow, run and record equivalent sub-gates:

```bash
gofmt -w cmd internal
go mod tidy
git diff --exit-code -- go.mod go.sum
go vet ./...
PM_CRONTAB_FILE=$(mktemp) go test ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
PM_CRONTAB_FILE=$(mktemp) make smoke-no-build
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/...
go run ./cmd/connectorgen validate internal/connectors/defs
```

3. Update `.planning/phases/github-projects-discussions/VERIFICATION.md`, `SUMMARY.md`,
   `PRD-COVERAGE.md`, and `RUN-STATE.json` with actual evidence.
4. If PRD coverage still says missing `Design direction`, reference the generated design prompt and
   critique under `docs/design/generated/` or mark not applicable with reason if no UI design change
   exists.
5. If PRD coverage still says missing `Data model`, reference `API-CONTRACT.md`, schemas, streams,
   and fixture `read_query`, or add a short `DATA-MODEL.md`.
6. Review generated website files. `website/next-env.d.ts` may have changed from `.next/types` to
   `.next/dev/types`; decide whether to keep only if generated by current Next workflow.
7. Commit a coherent green slice and push the branch if local gates pass.
8. Create/update PR to the parent branch/parent PR according to AGENTS policy. Do not merge to
   `main`.

## Expected Final Summary

Report:

- GitHub now has 37 streams, 67 writes, 507 API surface rows, 105 covered rows.
- Four GraphQL read commands are implemented.
- Runtime command dispatch remains limited to stream-backed ETL and existing reverse ETL writes.
- Project/discussion mutations remain planned until write action policy is implemented.
- Verification commands and results.
- Any blockers, especially if `make verify` cannot finish.
