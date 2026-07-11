# Twenty S4 writes plan (#281)

Status: REVIEW_FIX_GREEN_LOCAL (F1 accepted/fixed; manual GSD fallback; worker inline, no subagents spawned by instruction).
Parent: #277 / draft PR #285. Worker branch: `feat/281-twenty-writes`. Base branch for sub-PR: `feat/277-twenty-connector-parity`.

## Required reading / skills loaded

- Read: `AGENTS.md`, `issue-agent-contract.md`, `gsd-universal-runtime-loop.md`, `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, connector validation gates, migration handoff/conventions/design, S4 task PLAN/TDD/VERIFICATION, issue #281 body, research files, current `twenty` stream/API surface files, engine write/schema files.
- GSD skill: `.pi/skills/gsd-core/SKILL.md` loaded.
- Repo-local `.pi/skills/go-implementation/SKILL.md`: unavailable in this worktree (`ENOENT`; `find . -name '*go-implementation*'` found none). Fallback loaded Go skills per `required-skills-routing.md`: `golang-how-to`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-naming`, `golang-code-style`.
- Skill rule anchors for handoff: testing rules 1/3/5 (named deterministic tests, independent tests, behavior contracts); security model questions 1-3 + hardcoded secrets rule; safety rules 2/4/6/10 (safe assertions/maps/copies/zero values); error rules 1/2/7/9; design rules 3/5/10/20/21; naming MixedCaps/error-string guidance; code-style early-return and explicit map/slice init rules.
- Review-fix F1 loaded: `golang-documentation` for `docs/migration/conventions.md`; `caveman` for compact handoff. Re-read `AGENTS.md`, issue contract, universal loop, skill routing, GSD adapter, issue #281 body, phase artifacts, migration handoff/conventions/design, engine write/schema files, Twenty writes, and CLI parity reference. Repo-local `.pi/skills/go-implementation/SKILL.md` still unavailable (`ENOENT`).

## GSD adapter path

Required command attempted:

```bash
scripts/gsd doctor && scripts/gsd list | head -80 && scripts/gsd prompt programming-loop init --phase twenty-s4-writes --dry-run
```

Result: doctor/list OK, then `scripts/gsd: unknown GSD command: programming-loop` (exit 1). Per S4 task contract, using manual GSD loop fallback and recording it here, in TDD, verification, and PR body.

Review-fix command evidence:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase twenty-s4-writes --dry-run
scripts/gsd prompt gsd-quick "S4 #281 PR #304 review-fix correction for Twenty batch body shape"
```

Result: doctor OK; programming-loop command still unavailable; `gsd-quick` prompt rendered official GSD quick-task instructions. Manual loop remains active.

## Preflight / dependency evidence

```text
research preflight:
true
0
true
field manifest preflight:
28
546
api surface GET preflight:
56
api surface covered streams preflight:
56
```

Branch evidence:

```text
feat/281-twenty-writes
HEAD 1a86cc1aa9e644b82ddf2f9a46e5be785abb65b1
local parent branch feat/277-twenty-connector-parity present at b697053c; worker HEAD includes S1/S2/S3 commit 1a86cc1a.
parent PR #285 open/draft -> main.
```

## Red validation evidence before production edits

```text
initial red counts:
writes.json missing (expected before S4)
56
56
0
```

Meaning: no S4 writes exist; `api_surface.json` has 56 total GET rows and 0 write rows, not target 140/84.

## Review fix F1 plan (PR #304)

Accepted finding: Twenty `POST /rest/batch/{objects}` expects the request body itself to be a bare JSON array. Current S4 batch actions use `body_fields:["records"]`, producing `{ "records": [...] }`, which conflicts with official Twenty source evidence: controller routes `batch/*path` to `createMany`, `parseRequestArgs` assigns `data: request.body`, `CreateManyQueryArgs.data` is `Partial<ObjectRecord>[]`, and the runner calls `args.data.length` / `args.data.forEach`.

Correction slices:

1. Planning slice: record F1 acceptance, GSD fallback, loaded skills, safety constraints, and exact required gates before production edits.
2. Red-test slice: add focused engine test `TestWriteRawBodyField` proving `body_field:"records"` must send a top-level JSON array, then run it red before production code.
3. Engine slice: add `WriteAction.BodyField` + schema support; for JSON writes only, use `rec[action.BodyField]` as raw payload and error if absent; keep `body_fields`, default JSON, form, graphql, and none behavior unchanged.
4. Twenty slice: update all 28 `batch_` actions from `body_fields:["records"]` to `body_field:"records"`; preserve outer `record_schema.required == ["records"]` for validation/preview gates.
5. Docs slice: update `docs/migration/conventions.md` write-body paragraph with action-scoped, schema-gated raw body-field support and Twenty bare-array example; do not expose generic/raw HTTP writes.
6. Verification slice: required jq/Python/focused engine/connectorgen/conformance/focused packages/vet/build/gofmt/full tests/`scripts/verify-gsd-workflow 1a86cc1a` all passed locally; commit/push `fix(twenty): send batch records as raw array (Refs #281)` pending.

## Original S4 slice plan

1. Planning/red evidence slice (this artifact set): manual GSD fallback, preflight, red counts, safety plan.
2. Derivation slice: generate `writes.json` from `streams.json` schema refs + S2 schemas + `FIELD-MANIFEST.json`; prune immutable/system fields exactly per S4 contract; fail if any object has zero writable fields.
3. Surface slice: append 84 S4 rows to `api_surface.json` without row-level unknown keys; preserve 56 GET rows unchanged.
4. Batch gate: local research marks `POST /rest/batch/{objects}` as reverse_etl with batch max 60 records/request from official docs. Model batch as one Polymetrics input record containing `records` array, using `body_fields:["records"]`; no aggregation, raw HTTP, credentials, or live execution.
5. Verify slice: jq parse/count/shape checks, connectorgen validate, twenty conformance, focused Go tests, gofmt/vet/build/full tests if feasible, `scripts/verify-gsd-workflow 1a86cc1a` after evidence committed.
6. Delivery slice: commit `feat(twenty): add reverse ETL write actions (Refs #281)`, push worker branch, open stacked PR to parent branch with `Refs #281` and `Refs #277`.

## Safety / non-goals

- No live Twenty credentials; no credentialed checks; no raw HTTP calls; no reverse-ETL plan/preview/approval/execute.
- No DELETE actions/API rows in S4; S5 owns delete/destructive semantics.
- Review fix allows narrowly-scoped engine dialect/schema changes only: `WriteAction.BodyField`, writes schema, JSON body execution, and focused engine tests. No dependencies.
- CLI/help/docs/website parity is intentionally out of S4 scope; S6/S7 own `cli_surface.json`, docs, website, generated help/manual artifacts.

## Execution decisions

- plan: `local_critical_path` — user explicitly said do not spawn subagents; this worker owns one isolated cwd/branch.
- tdd-gate: `local_critical_path` — red validation captured inline before production edits.
- review-fix-plan: `local_critical_path` — user instructed no subagents; inline worker owns F1 correction on same branch/cwd.
- review-fix-red-test: `local_critical_path` — focused engine test failed before production support with unknown `WriteAction.BodyField`.
- review-fix-execute: `local_critical_path` — added `body_field` JSON dialect, migrated Twenty batch actions, documented convention.
- review-fix-verify: `local_critical_path` — all required local gates passed; no subagents by instruction.
