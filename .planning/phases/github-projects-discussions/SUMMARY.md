# Phase Summary

Phase: github-projects-discussions

## What Was Done

Implemented four GraphQL-backed GitHub CLI read commands:

- `pm github project list` -> stream `projects`
- `pm github project item-list --project-id <id>` -> stream `project_items`
- `pm github discussion list` -> stream `discussions`
- `pm github discussion view --number <n>` -> stream `discussion`

## Engine Changes

- Fixed-document GraphQL stream support reads GraphQL variables from `config.*`, `query.*`, `cursor`, and optional `default`/`omit_when_empty`.
- Static validation checks GraphQL variable templates in `connectorgen validate` and conformance.
- Conformance fixture harness supports stream fixture `read_query` for parameterized reads.
- Omitable query interpolation: `optional: true` and `omit_when_empty: true` template object flags.

## GitHub Connector State

- Streams: 37 (4 new: projects, project_items, discussions, discussion)
- Writes: 67 (unchanged)
- API surface rows: 507 (105 covered, 4 new GRAPHQL covered rows)
- Operations.json: 4 new GraphQL read operation ledger rows
- CLI surface: 4 new commands mapped

## Runtime Command Dispatch

Runtime command dispatch remains limited to stream-backed ETL and existing reverse ETL writes. Project/discussion mutations remain planned until write action policy.

## Verification

All local gates pass:
- gofmt, go mod tidy, go vet: clean
- golangci-lint: 0 issues
- go build ./cmd/pm: builds
- Engine, connectorgen, commandrunner, conformance tests: all pass
- connectorgen validate: 547 connectors, 0 findings
- github conformance: PASS
- docs validate: clean
- smoke: ok
- PRD coverage: all artifacts present or marked not-applicable

## Unrelated File Review

Agent doc hardening patches across `.agents/`, `.opencode/`, `.codex/` are expected OpenCode/Codex agent docs patching. `website/next-env.d.ts` auto-generated dev mode path. `passb-expander.agent.yaml` GSD compliance update. `cli_test.go` test adjusted for changed command dispatch. No unrelated files to revert.

## File Inventory

### New files
- `.opencode/agents/gsd-worker.md`
- `.opencode/commands/gsd-worker.md`
- `.planning/phases/github-projects-discussions/*` (phase artifacts)
- `codex-prompts/connector-cli-parity-research.txt`
- `codex-prompts/opencode-continue-github-projects-discussions.md`
- `docs/design/generated/github-projects-discussions-*`
- `internal/connectors/defs/github/fixtures/streams/projects/*`
- `internal/connectors/defs/github/fixtures/streams/project_items/*`
- `internal/connectors/defs/github/fixtures/streams/discussions/*`
- `internal/connectors/defs/github/fixtures/streams/discussion/*`
- `internal/connectors/defs/github/schemas/projects.json`
- `internal/connectors/defs/github/schemas/project_items.json`
- `internal/connectors/defs/github/schemas/discussions.json`
- `internal/connectors/defs/github/schemas/discussion.json`

### Modified files (37 total)
Core engine: `engine/bundle.go`, `engine/bundle_test.go`, `engine/graphql.go`, `engine/interpolate.go`, `engine/read.go`, `engine/read_test.go`
Validation: `cmd/connectorgen/validate.go`, `cmd/connectorgen/github_api_surface_test.go`
Conformance: `conformance/dynamic.go`, `conformance/dynamic_test.go`, `conformance/replay.go`, `conformance/static.go`
GitHub bundle: `streams.json`, `api_surface.json`, `cli_surface.json`, `operations.json`, `docs.md`
Agent docs: `.agents/`, `.opencode/`, `.codex/`, `AGENTS.md`
Website: `connectors.generated.json`, `connectors.catalog.data.generated.json`, `next-env.d.ts`
Docs: `docs/architecture/repo-profile.json`, `docs/migration/conventions.md`, `docs/prompts/`
CLI test: `internal/cli/cli_test.go`

## Blockers

None. Phase complete.
