# Issue #397 PM Orchestrator Extension Verification

Status: captain-authorized correction round 5 planned for R1–R3 only; implementation/verification/re-review pending; no round 6
`verificationPassed`: false (round-5 candidate does not exist yet)

## Identity and scope

- [x] Isolated worktree verified.
- [x] Additive branch/PR #495 starts at reviewed head `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`.
- [x] Current main `873cd7b251f70c4a35a607a0d4e86051ea0fbd15` and parent `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f` pinned after fetch.
- [x] PR #493 head/file ownership inspected at `e21e56339390c5e1946eb4cfaf276eb80a889f29`.
- [x] No PR #493-owned path changed.
- [x] No #408/TUI/product/runtime/dependency change.

## Forward route

- [x] `/pm-orchestrate` is the required owner for parent/stacked work when `programming-loop` is absent.
- [x] PM route preserves PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE.
- [x] State reconciliation, worker isolation, bounded correction, machine contracts, credential safety, and human merge authority remain explicit.
- [x] Fresh-context local Codex review binds exact base/head and dispositions every finding.
- [x] Any head change requires re-verification and fresh local Codex re-review.
- [x] Shepherd is independent and required after review/before integration.
- [x] Claude/Copilot are not required, requested, or fallback coverage in current/future PM route.
- [x] Legacy Claude/Copilot docs/agent remain discoverable only with migration/deprecation pointers.
- [x] Historical completed evidence remains unchanged; only current extension, authoritative #397 queue/run state/summary, and narrow Wave 1 evidence are updated.
- [x] Authoritative #397 state blocks #408 until Wave 1 and PR #493's canonical PM routing migration both integrate.
- [x] Current PM dependency paths use PM-specific handoff/disposition templates with `local_codex`, Shepherd, correction-budget, and human-gate fields.
- [x] Autonomous state uses stable exact-base/candidate-lineage counters across replacement heads; the old counter is read-only migration input.
- [x] Scoped #408 subissue, ready queue, top-level gate, summary, and machine branch identity all require the same Wave1-plus-PR #493 transition.
- [x] Canonical review statuses/dispositions are aligned across schema/workflow/prompt/contracts/templates.
- [x] Both autonomous drivers use one classifier: correction-cap and canonical missing/unknown-kind `human_gate` stop as blocked human decisions; explicit parent readiness and detected legacy missing-kind remain human-ready.
- [x] Canonical producers persist `schema_version`, plain `human_gate`, and the required sibling kind; all other structured gate detail remains outside the terminal string.
- [x] Exact six-value finding-disposition enum is machine-specified and parsed identically across all applicable PM files.

## Round 5 authorization and scope

- [x] Durable decision read: `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pr495-round5-override-2026-07-23.md`.
- [x] Existing lineage retained; counter advances from 4 to 5 without reset.
- [x] Scope limited to R1 missing worker skill references/traversal, R2 unsupported schema fail-closed classification, and R3 actual canonical disposition rows.
- [x] Audit-backed PM review system, product changes, PR #493 files, and round 6 excluded.
- [x] `pm-gsd-worker` references only existing required files and harness-available skills through `required-skills-routing.md`.
- [x] Active dependency traversal rejects missing required references.
- [x] Explicit unknown schema plus missing kind is a blocked human decision; unknown kind remains blocked; absent-schema legacy remains readable.
- [x] Every actual `REVIEW-DISPOSITION.md` row uses the exact canonical enum; F2/N1 remain deferred with rationale.

## Focused validation

- [x] RED captured before canonical guidance changes.
- [x] `scripts/tests/pm-orchestrator-contract.sh`.
- [x] `scripts/tests/pi-model-routing.sh`.
- [x] YAML/JSON parse checks, including replacement-head/correction-cap, parent-ready, canonical-missing-kind, and historical human-gate fixtures.
- [x] Shell syntax and terminal-classifier behavior for canonical blocked, canonical ready, and detected legacy human gates.
- [x] PR #493 changed-path disjointness.
- [x] no dependency delta.

## Full gates

- [x] `gofmt -w cmd internal`
- [x] `git diff --exit-code -- cmd internal`
- [x] `git diff --check`
- [x] `go vet ./...`
- [x] `go test -timeout 20m ./...`
- [x] `go build ./cmd/pm`
- [x] `go mod verify`
- [x] `go mod tidy -diff`
- [x] `make verify`

## Review and delivery

- [x] Initial fresh-context local Codex review completed at `3c88fc78062ba0a3437f79bc88c395286c228c65`; five findings dispositioned in `REVIEW-DISPOSITION.md`.
- [x] Fresh exact-head local Codex re-review at `0665ad7aad1ec083f4bb0572a88ac1a38f417a35`; F2/F3/F5 confirmed and F6–F8 accepted for correction round 2.
- [x] Captain-authorized Gong follow-up created at https://github.com/polymetrics-ai/cli/issues/497 without product changes here.
- [x] Fresh exact-head local Codex re-review at `3af7910528d234d1a1d886a6778d7817495e6321`; N2–N5 accepted and N1 deferred under the no-product boundary.
- [x] Fresh exact-head local Codex re-review at `1d4acf4f633e4f8940ba637f2099723369b2ed30`; N4-R/N3-R accepted for maximum correction round 4.
- [x] Fresh exact-head local Codex review at `32dbda9daf432a5c3f45e7b7753a41eaa96bd915` returned R1–R3 and blocked after round 4/4; Shepherd did not run.
- [ ] Fresh exact-head local Codex re-review clean after authorized correction round 5.
- [ ] Every finding dispositioned; final changed head re-reviewed. Any new actionable finding is a human blocker; no round 6.
- [ ] Exact-head Shepherd/trajectory validation passes.
- [ ] Branch pushed normally without force.
- [ ] PR #495 title/body updated with extension scope and exact head.
- [ ] PR #495 remains draft; PR #438 remains draft/unchanged.
- [ ] No Claude/Copilot requested or counted.
- [ ] Required PR checks green at exact head.
