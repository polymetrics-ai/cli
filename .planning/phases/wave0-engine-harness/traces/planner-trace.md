# Agent Trace: planner

Role: gsd-loop-planner (model: fable) · Date: 2026-07-02 · Phase: wave0-engine-harness

## Rendered Prompt Or Prompt Reference

Coordinator dispatch: "write the wave0 phase planning artifacts into
.planning/phases/wave0-engine-harness/" per the GSD Universal Programming Loop planner role;
prompt library: `docs/prompts/universal-programming-loop-prompts.md` (Kickoff + model/skill
policy). Snapshot in `.planning/phases/wave0-engine-harness/PROMPTS.md`.

## Files Inspected

- `.planning/{PROJECT.md,ROADMAP.md,STATE.md,config.json}` — acceptance criteria, model policy,
  tdd_mode=true, use_worktrees=false.
- `docs/plans/universal-programming-loop-prd.md`, `docs/migration/orchestration-plan.md`,
  `docs/prompts/universal-programming-loop-prompts.md`,
  `docs/architecture/connector-architecture-v2-design.md`,
  `docs/architecture/connector-certification-design.md`, `docs/architecture/repo-profile.json`.
- Code grounding: `internal/connectors/connectors.go` (RegisterFactory:60, RegisterNativeLive:49,
  NewRegistry:273, Connector interface:256), `internal/connectors/manifest.go`,
  `internal/connectors/connsdk/{http,paginate,auth,extract,state,schema}.go`,
  `internal/connectors/native_conformance.go`,
  `internal/connectors/stripe/{stripe,streams,write}.go`,
  `internal/connectors/searxng/searxng.go` (+streams.go head),
  `internal/connectors/postgres/postgres.go` (head), `internal/cli/cli.go` (Run:24,
  runConnectors:156, appRegistry:983), `internal/app/sync_modes.go`, `Makefile` (verify:60,
  smoke:41), `cmd/registrygen/main.go` (skip map:30), `go.mod` (go 1.25.4; pgx present).
- Existing phase scaffold: PRD-COVERAGE.md, RUN-STATE.json, TDD-GATE.json, VERIFICATION.md,
  AGENT-ORCHESTRATION.json, AGENTS.md, PROMPTS.md.

## Actions Taken

Wrote planning artifacts (no implementation code, no tests):
SPEC.md · PLAN.md (21 tasks: 18 T/B TDD pairs + 1 docs-only + 1 verification; dispatch waves
A–H with `wave:` labels) · TEST-PLAN.md · THREAT-MODEL.md · RUNBOOK.md · DATA-MODEL.md ·
API-CONTRACT.md · OBSERVABILITY.md · EVAL-PLAN.md · PRD-COVERAGE.md (updated) ·
docs/adr/0001-connectors-as-data.md · docs/plans/POSTMORTEM-TEMPLATE.md · this trace.

## Commands Run

Read-only inspection (ls/grep/wc/head over the paths above). No build/test commands — planning
only; verification commands are specified per task in PLAN.md.

## Findings

1. **Coexistence decided as recommended**: goldens' legacy packages stay compiled + registered;
   engine-backed versions are constructed only in tests from `defs.FS`; `native/postgres` gets NO
   init() registration (guard test T-17). Registry flip is wave6. Single legacy edit in wave0:
   `cmd/registrygen/main.go` skip map += {defs, engine, hooks, native, conformance, certify},
   with a byte-identical registryset regen guard (T-16).
2. **Engine gap found and closed in the spec**: stripe's pagination
   (`stripe/stripe.go:147`) is `starting_after` = last-record-id + `has_more` stop — not the
   design's body-token cursor. PaginationSpec extended with `last_record_field`/`stop_path`
   (DATA-MODEL §2); both cursor variants in the 6-type termination matrix; connsdk untouched.
3. **Empty-header omission rule** specified so stripe's conditional `Stripe-Account` header is
   expressible declaratively.
4. **`streams.json` optional iff `capabilities.dynamic_schema`** so the postgres Tier-3 bundle is
   honest (runtime discovery from information_schema).
5. **Certify scope pinned to certification-design steps 1–2**: report.go + cliharness +
   stages_source proven from a Go test against `sample`; no CLI wiring, no write/flow/schedule,
   `--credential` etl fix deferred to wave1 (sample flows through connections like `make smoke`).
6. **Validator keyword set frozen**; stripe's "email or name" write rule approximated by
   `minProperties: 1` — first entry of the conventions.md parity-deviation ledger.
7. **`defs/defs.go` uses one `//go:embed all:*`** so optional files can't break compilation with
   only 3 bundles.
8. **Definition introduced additively** (`internal/connectors/definition.go` +
   `DefinitionProvider`); core interface/manifest/app/cli untouched in wave0.

## Handoff Summary

Coordinator can dispatch Wave A immediately (4 parallel Sonnet tasks T/B-01..04). B-12 (certify
base) has no engine dependency and may float earlier. Wave-close procedure and path-guard in
RUNBOOK §5. TDD gate: every B-task requires RED evidence in TDD-LEDGER.md first.

## Verification Evidence

Planning-phase only: artifacts exist at the paths above; PRD-COVERAGE rows all present or
not-applicable with reasons (design-direction: backend-only, frontend flag is the docs website;
release-notes: deferred to wave6). Gate commands per task are listed in PLAN.md and EVAL-PLAN.md.

## Unresolved Risks

1. **golangci-lint acquisition** (SPEC §5 human flag): binary not on PATH here; pick pinned
   `go run …golangci-lint@<version>` (no go.mod change) vs local install before B-18.
2. **inventory loc counting convention** (all .go incl. tests vs non-test) — confirm before B-19.
3. **api_surface depth for goldens**: minimal honest surfaces with `out_of_scope` exclusions in
   wave0; full surfaces are Pass B. Confirm reviewer checklist alignment.
4. Engine expressiveness risk beyond the goldens is deferred by design (ENGINE_GAP protocol);
   watch for repeated gaps during Wave F per EVAL-PLAN §7.
