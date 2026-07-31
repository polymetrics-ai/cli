# Plan: Twenty CRM connector parity wave02 r1 (#277, #278-#284)

Parent issue: #277
Subissues in scope: #278, #279, #280, #281, #282, #283, #284
Nested follow-up observed but not in this connector bundle scope: #323 (agentic control-plane hardening)
Branch: `fm/cli-twenty-parity-wave02-r1`
Worktree: `/Users/karthiksivadas/.treehouse/cli-83d592/25/cli`

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands available.
- `scripts/gsd prompt programming-loop init --phase issue-277-twenty-parity-wave02-r1 --dry-run` — unavailable in this adapter (`unknown GSD command: programming-loop`). Manual GSD fallback recorded in `traces/gsd-programming-loop-unavailable.txt`.
- Manual GSD fallback is active: keep plan, TDD ledger, verification checklist, red/green evidence, and safety gates current before and after implementation.

## Required skills and references loaded

Skills:

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-graphql`
- `golang-documentation`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-lint`

References:

- `AGENTS.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.planning/{config.json,PROJECT.md,ROADMAP.md,STATE.md}`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Mission

Restore the complete documented Twenty CRM connector parity bundle onto `main`-based work: `internal/connectors/defs/twenty/**`, Twenty-owned fixtures, generated/manual connector docs, CLI metadata, conformance evidence, and connector-local planning artifacts. Preserve the 168-operation issue-family count:

- 28 standard objects.
- 56 read operations: 28 list `GET /rest/<object>` rows plus 28 same-shape get-by-id `GET /rest/<object>/{id}` rows covered by the corresponding stream until generic direct-read execution is available.
- 84 non-destructive reverse-ETL actions: create, update, and batch create for each object.
- 28 destructive DELETE actions represented as connector-owned write actions gated by `confirm: "destructive"`, typed schemas, fixtures, and reverse-ETL plan -> preview -> explicit approval -> execute.

## Official-source inventory

- Twenty docs API overview: `https://docs.twenty.com/developers/extend/api`.
- Twenty docs LLM index/full export: `https://docs.twenty.com/llms.txt`, `https://docs.twenty.com/llms-full.txt`.
- Twenty open-source standard-object metadata paths discovered through `gh-axi api '/repos/twentyhq/twenty/git/trees/main?recursive=1'`, especially `packages/twenty-server/src/engine/workspace-manager/twenty-standard-application/utils/{object-metadata,field-metadata,index,...}`.
- Existing parent branch evidence: PR #285 (`origin/feat/277-twenty-connector-parity`) integrated #278-#284 and recorded 168 operations / 28 objects / 546 fields with final parent head `5199c0bb6519155cb0456fb3476e323ba9347d40`.

## Constraints and safety gates

- No live Twenty provider calls, credentials, writes, certification against live provider, VPS/Thaalam work, merges, pushes, or PR creation.
- Do not edit shared runtime/foundation files or other connectors. If execution needs a shared feature not present on `main`, record the dependency and keep the connector-local row planned/blocked rather than claiming support.
- Use `gh-axi` for GitHub operations.
- Use `pm help <topic>` before unfamiliar `pm` commands.
- Never print or store secret values. Use only synthetic fixtures.
- Do not expose generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tools.
- Reverse ETL remains plan -> preview -> explicit approval -> execute. Deletes additionally require typed destructive confirmation.

## Implementation slices

1. **Issue graph and captain-policy addendum**
   - Inventory #277 and #278-#284, plus note nested #323 as a follow-up outside connector bundle scope.
   - Append idempotent captain-policy addendum to #277-#284 using `gh-axi issue edit --body-file`, preserving existing body text and counts.
2. **Red validation baseline**
   - Capture that `internal/connectors/defs/twenty` is absent on this branch.
   - Run targeted connector validation that should fail/miss before the bundle exists.
3. **Connector-local bundle**
   - Restore/create `internal/connectors/defs/twenty/**`: metadata, spec, api surface, streams, writes, CLI surface, schemas, fixtures, and docs.
   - Keep all 168 operations represented and every executable write routed through typed schemas/fixtures/safety gates.
   - Keep same-stream coverage for get-by-id rows; generic direct-read commands stay planned until shared direct-read execution is available.
4. **Docs and tests**
   - Add Twenty generated/manual connector docs under `docs/connectors/twenty/**`.
   - Add Twenty conformance test under `internal/connectors/conformance/twenty_test.go`.
   - Regenerate docs/website data if the project generator changes connector-visible generated artifacts; avoid unrelated churn.
5. **Validation and commit**
   - Run targeted validation and local gates within the allowed no-live scope.
   - Commit the complete connector bundle and planning artifacts.
   - Stop after commit; do not run no-mistakes pipeline, push, or open a PR until firstmate resumes.

## Shared dependencies recorded

- Generic direct-read execution for get-by-id commands is not claimed here. The 28 get-by-id API rows are represented in `api_surface.json` and `cli_surface.json` as planned/direct_read commands, with same-stream coverage for conformance.
- Live provider certification requires credentials and is explicitly out of scope. Credential-free fixture/replay/localhost certification evidence can be recorded only if it does not call Twenty.

## Verification plan and final evidence

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestTwenty|TestConformance/twenty' -count=1
go run ./cmd/pm help docs
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs
go run ./cmd/pm help twenty
go run ./cmd/pm connectors
go run ./cmd/pm connectors inspect twenty --json
cd website && pnpm run gen:website-data
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Final evidence is recorded in `TDD-LEDGER.md` and `VERIFICATION.md`. Targeted Twenty gates and `make verify` passed. The exact default `go test ./...` command was attempted and hit the default 10m package timeout in slow `internal/cli`; the repository-supported `make verify` path uses `go test -timeout 20m ./...` and passed.
