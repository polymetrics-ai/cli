# PLAN: GitHub Projects And Discussions GraphQL Reads

## Tasks

1. Add red tests for engine GraphQL query-variable interpolation and optional cursor omission.
2. Add red tests for GitHub bundle loading and CLI surface mappings for project/discussion reads.
3. Extend the engine narrowly:
   - add `query.*` template namespace
   - pass `ReadRequest.Query` into GraphQL variable resolution
   - support `omit_when_empty` on GraphQL variable template objects
4. Add GitHub GraphQL read streams, schemas, fixtures, and operation ledger rows:
   - `projects`
   - `project_items`
   - `discussions`
   - `discussion`
5. Update GitHub `cli_surface.json` so read commands are implemented ETL streams.
6. Update GitHub docs and generated website bundle/catalog data.
7. Run validation:
   - focused red/green tests
   - `go run ./cmd/connectorgen validate internal/connectors/defs/github --json`
   - `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen`
   - `go test ./internal/connectors/conformance -run 'TestConformance/github'`
   - `go vet ./...`
   - `go build ./cmd/pm`

## Review-Fix + Pi Orchestration Hardening Slice

CodeRabbit reviewed PR #74 and surfaced actionable comments. In parallel, a pi runtime audit
found the `.pi/` orchestration config has broken prompt-template placeholders and read-only
subagents that lose search tools. This slice addresses both.

### CodeRabbit disposition

- Accept: comments that are still-valid defects or quick wins (contradictory agent guardrails,
  non-portable examples, missing test coverage, generated-file noise, schema gaps, metadata
  auth-labeling).
- Decline: none planned; every comment is either fixable or will be dispositioned with reason.
- Defer: trace-file population that requires actual run evidence from a later phase execution;
  mark as deferred with follow-up issue link.

### Work queue

1. **Pi orchestration runtime** (config/docs, no behavior change):
   - Fix `.pi/prompts/*.md` to use pi `$@`/`$1` placeholders instead of `{{task}}`/`{{target}}`.
   - Document `--tools read,bash,edit,write,grep,find,ls,subagent` launch requirement in `.pi/README.md`.
   - Add `agentScope`, `confirmProjectAgents`, and per-worker `cwd` guidance to `pm-orchestrate.md`.
   - Harden `pm-coderabbit-disposition.md` tools (drop `bash` from read-only disposition planner).
   - Create `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md` adapter.
2. **Agentic-delivery comments**:
   - Fix `gsd-universal-runtime-loop.md` fallback token (`failed_runtime_capability` → `not_spawned_*`).
   - Fix `caveman-token-compression.md` hard-coded home path example.
   - Fix `passb-expander.agent.yaml` contradictory commit guardrail.
   - Align `docs/prompts/universal-programming-loop-prompts.md` decision labels with runtime tokens.
3. **Connector code comments**:
   - Add DraftIssue fragment to `internal/connectors/defs/github/operations.json` ListProjectItems.
   - Validate GraphQL variable `default` against declared `type` in `internal/connectors/engine/bundle.go`.
   - Add regression test for explicitly-empty query variable in `internal/connectors/engine/read_test.go`.
   - Add `read_query` replay conformance gate to `TEST-PLAN.md`.
4. **Generated / metadata noise**:
   - Git-ignore `website/next-env.d.ts` and remove committed version.
   - Fix `website/data/connectors.generated.json` PAT/GitHub App auth labeling.
   - Fix `docs/architecture/repo-profile.json` `.next` artifact noise (regenerate or filter).
5. **Agent role contracts + traces**:
   - Fill TBD placeholders in `agents/*.md` role contracts or mark as draft.
   - Populate trace files with real phase evidence where available; defer placeholder-only traces.
6. **Verification + push**:
   - Re-run local gates (`make verify`, focused tests).
   - Push coherent green slices to `feat/40-github-projects-discussions`.
   - Wait for automatic CodeRabbit incremental review; do not spam manual review commands.

## Spawn Decision

- `read_only_spawned` (intended) / `local_critical_path` (actual): read-only workflow adapter
  audit and GitHub mapping audit sidecars. The phase ran inline (`agentMode: inline`) because the
  audit work was coupled to the coordinator context; inline role passes are recorded as
  `local_critical_path`, not `spawned`, per the gate-integrity rule in
  `gsd-universal-runtime-loop.md`.
- `not_spawned_write_scope_collision`: no mutating subagent for shared engine/GitHub bundle files
  in coordinator checkout. Implementation stays local for this slice.

## Human Gates

- No new dependency.
- No auth-scope refresh.
- No mutation execution.
- Parent PR merge to `main` remains human-gated.
