# Twenty S5 destructive delete actions plan (#282)

Status: GREEN_LOCAL_GSD_EVIDENCE_PASSED. Manual GSD fallback active; no subagents spawned by instruction.
Parent: #277 / base branch `feat/277-twenty-connector-parity`. Worker branch: `feat/282-twenty-deletes`. Start/base head: `bc014ef6b7b138eda1a492672a76ea577f78b695`.

## Issue contract / acceptance

Issue #282: Twenty S5 destructive writes — 28 delete actions, part of #277. Write scope: `writes.json` destructive rows. Deps: #281 (shared writes.json). Acceptance: destructive rows gated; `covered_by` mapping complete.

User correction for S5 gate: S4 batch actions intentionally have `kind:"create"`; preserving existing 84 S4 actions means post-S5 kind counts must be `create=56`, `update=28`, `delete=28`, with 28 `batch_` action names. Existing S4 batch action kinds must not change.

## Required reading / skills loaded

- Read: `AGENTS.md`, `issue-agent-contract.md`, `gsd-universal-runtime-loop.md`, `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, `.planning/{config.json,PROJECT.md,ROADMAP.md,STATE.md}`, `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, issue #282 body, S4 phase artifacts, current Twenty files.
- GSD skill loaded: `.pi/skills/gsd-core/SKILL.md`.
- Repo-local `.pi/skills/go-implementation/SKILL.md`: unavailable (`ENOENT`); fallback skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`; `caveman` for handoff.
- Skill rule anchors: testing rules 1/3/5; security model questions 1-3 + hardcoded secret rule; safety rules 2/4/6/10; error rules 1/2/7/9; design rules 3/5/10/20/21; structs/interfaces small contracts/field tags.

## GSD adapter path

Attempted:

```bash
scripts/gsd doctor && scripts/gsd list && scripts/gsd prompt programming-loop init --phase twenty-s5-deletes --dry-run
```

Result: doctor/list OK; programming-loop command failed with `scripts/gsd: unknown GSD command: programming-loop` (exit 1). Manual GSD loop fallback recorded.

## Red validation evidence before production edits

```text
red writes actions 84
red kind counts {'create': 56, 'update': 28, 'delete': 0}
red name counts {'create_names': 28, 'update_names': 28, 'batch_names': 28, 'delete_names': 0}
red api_surface rows 140
red methods {'GET': 56, 'POST': 56, 'PATCH': 28, 'DELETE': 0}
```

## Green implementation summary

- Appended exactly 28 `delete_<snake_plural>` actions to `internal/connectors/defs/twenty/writes.json`.
- Preserved existing 84 S4 actions, including 28 `batch_` names as `kind:"create"`.
- Delete shape: `kind:"delete"`, `method:"DELETE"`, `/rest/<object>/{{ record.id }}`, `path_fields:["id"]`, `body_type:"none"`, strict id-only schema, no body fields, no GraphQL, no hooks, no `delete.missing_ok_status`, `confirm:"destructive"`, explicit destructive risk.
- Appended exactly 28 `DELETE /rest/<object>/{id}` API surface rows with `covered_by.write` matching delete actions.
- Updated scope text to state S5 covers DELETE rows.

## Verification summary

Passed locally: jq parse, corrected Python S5 gate, `connectorgen validate`, Twenty conformance, focused packages, `go vet ./...`, `go build ./cmd/pm`, `gofmt -l cmd internal`, `go test ./... -count=1`, and post-commit `scripts/verify-gsd-workflow bc014ef6`.

## Safety / non-goals

- No secrets, live credentials, credentialed checks, reverse ETL execution, or `pm reverse run`.
- No generic/raw HTTP write tool; no dependencies; no engine changes.
- No changes to `schemas/**`, `streams.json`, fixtures, `cli_surface.json`, `docs.md`, `docs/cli`, `website`, `go.mod`, or `go.sum`.
- Missing delete idempotency not assumed; no `delete.missing_ok_status` added.
- `make verify` not run because it executes reverse run via smoke target; S5 forbids reverse ETL execution.

## Execution decisions

- plan: `local_critical_path` — user instructed no subagents.
- tdd-gate: `local_critical_path` — red validation captured inline before production edits.
- execute: `local_critical_path` — appended deletes/API rows locally inside allowed scope.
- verify: `local_critical_path` — required local gates passed.
- gsd-evidence: `local_critical_path` — post-commit `scripts/verify-gsd-workflow bc014ef6` passed.
