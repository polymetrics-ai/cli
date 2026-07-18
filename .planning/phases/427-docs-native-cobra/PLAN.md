# Phase 427 Plan — Docs native Cobra namespace

Issue: polymetrics-ai/cli#427
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/427-docs-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`
Invocation session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — seventh serialized Phase 9 namespace unit is assigned to this isolated branch/worktree. Central router files collide with later units, the runtime exposes no subagent tool in this session, and the user explicitly bounded this run to #427 with no PR or external review.

## Required reading complete

- Issue #427 via `gh`; parent #397; umbrella #407; adjacent #426 skills implementation/artifact patterns; draft parent PR #438.
- `AGENTS.md`; issue-agent and parent-orchestrator contracts; parent orchestration, stacked-work, and GSD universal-runtime/manual-loop policies; worker handoff template.
- `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`; universal programming-loop PRD and prompt library.
- Required-skills routing, GSD Pi adapter, CLI help/docs/website parity.
- CLI Architecture v2 improvement plan §5/§9; execution prompt Stage 9; ADR 0002.
- Current docs handler/generators, connector-doc filesystem behavior, Cobra router, global/config parsing, focused/agentic/golden tests, embedded manual, checked-in CLI docs, and website generation path.

## GSD adapter and fallback

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- `scripts/gsd sources programming-loop` — command source unavailable because the adapter registry has no `programming-loop` command.
- `scripts/gsd prompt plan-phase 427 --skip-research` — prompt generated successfully and executed inline.
- Required programming-loop probe: `scripts/gsd prompt programming-loop init --phase 427 --dry-run` — unavailable with `scripts/gsd: unknown GSD command: programming-loop` (exit 1).
- Manual fallback: execute `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline with strict RED → GREEN → refactor evidence and all six issue-local artifacts. The adapter is healthy; only the absent programming-loop command requires fallback.

## Required skills loaded

- `gsd-core`.
- `golang-how-to` first, then `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, and `golang-security`.
- Applied constraints: fresh Cobra trees; `RunE`; declared pflags and action nodes; injected writers; stable exit taxonomy/stdout/stderr; observable table tests; canonical help; untrusted output paths scoped to safe temporary roots in tests; no secrets/dependencies/external services.

## Scope and exclusions

Allowed:

- Register native top-level `docs` namespace, native `generate` and `validate` actions, declared `--dir` and `--connectors-dir` flags where currently accepted, plus positional `docs help` compatibility.
- Minimal docs handler adaptation from `parseFlags` to typed values while preserving generated artifact bytes, destination defaults, filesystem errors, and output text.
- Docs-focused/router tests and issue-local phase artifacts.
- Directly applicable help/manual/website/generated artifacts only if behavior intentionally changes.

Excluded:

- Phase 14 `docs view`; Phase 19 focused help/man-page churn; other namespace migrations; dynamic connector parsing; connector bundle changes; generator content or canonical docs-map ownership changes; path-policy redesign; completion implementation; dependencies; services; credentials; ETL/reverse ETL; shared parent artifacts; PR creation or external review.

## Existing contract to preserve

- Bare `pm docs`, `pm help docs`, `pm docs --help`, `pm docs -h`, and positional `pm docs help` render the canonical manual and exit 0; JSON variants emit `CommandManual/docs`.
- Current actions are `generate` and `validate`; invalid actions remain usage exit 2. Missing/empty generate `--dir` retains its current internal-error behavior and text unless a focused RED proves a different contract.
- `generate` accepts `--dir` and `--connectors-dir`; `validate` accepts `--connectors-dir` and `--dir` as fallback. Spaced, assigned, repeated-last-wins, comma-containing, and bare-flag (`true`) forms retain legacy parsing.
- Unknown flags remain tolerated on actions; extra positional arguments and unknown values do not alter declared values.
- Global `--json`, `--root`, `--plain`, and `--no-input` remain accepted before/after the namespace in spaced/assigned forms; assigned booleans remain connected to config precedence, including configured JSON overridden by `--json=false`.
- Generation writes command manuals byte-for-byte from the canonical docs map, connector manuals/catalog/icons beneath the selected connector output directory, and the existing exact success lines. Validation checks the selected connector directory without modifying generator ownership.

## TDD slices and checkpoints

1. **Planning checkpoint** — create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY before production edits; commit and push.
2. **RED tests first** — specify native registration; generate/validate and every current flag/output-directory form; bare/text/JSON/positional help; unknown-flag compatibility; invalid action/error categories; assigned globals/config; generated byte parity; safe temp-root filesystem containment; fresh-tree completion seam. Capture exact RED before production edits; commit/push test checkpoint.
3. **Smallest GREEN** — remove `docs` from `cobraLegacyCommands`; add native namespace/generate/validate/help nodes; declare `StringArrayVar` flags with `NoOptDefVal=true`, unknown-flag whitelist, and no-file completion; normalize spaced path values; adapt `runDocs` to typed action flags and remove its sole `parseFlags` use.
4. **Refactor/parity** — run focused docs/router/golden/full CLI tests; verify built-binary text/JSON/help/error/global/output-dir forms using temporary roots; generate CLI docs to temp and byte-diff; run connector docs validation; run website generator/diff; expect no checked-in help/docs/golden delta.
5. **Full gates/delivery** — `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`; scope/dependency/diff checks; finalize artifacts; commit coherent green and verification checkpoints; push branch without PR.

## CLI help/docs/website parity stance

This changes parser ownership only. Embedded help, `docs/cli/docs.md`, website references, canonical docs-map ownership, generated connector manuals, completion/discovery names, output text, error envelopes, and goldens should remain unchanged. Bare namespace help, all text/JSON help routes, invalid actions, generated CLI byte diff, connector validation, website generator diff, and golden tests provide proof. `docs view` remains Phase 14; focused subcommand help/man generation remains Phase 19; completion implementation remains Phase 15.

## Safety

No secrets, credentials, services, dependencies, generic write tools, destructive/admin actions, quality-gate reductions, external reviews, PR, or merge. All focused generation/validation tests and manual parity checks write only beneath `t.TempDir` or command-created temporary roots and assert generated paths stay local. The required `make verify` may run its existing local sample reverse-ETL smoke only under plan → preview → approval → execute and without external writes.

## Bounded review correction — 2026-07-18

Correction session: `issue-427-review-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T121208Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact correction start `ea93b4bb7a7eb09236ad829d5ad6055b0c00c30d`.

Accepted both medium findings from `/tmp/pm-397-review-427.log`:

1. Preserve legacy action/trailing `--help` and `-h` behavior for `docs generate`, `docs validate`, and invalid actions without enabling Phase 19 focused action help.
2. Preserve legacy continuation after literal `--` so later docs action flags remain parsed.

Correction boundary: add focused differential tests first, then implement the smallest docs-scoped argument-normalization seam before native Cobra parsing. Keep the native docs command/action tree, typed pflags, canonical namespace help, existing outputs/error categories, other namespaces, docs bytes, website, dependencies, and deferred Phase 14/19 behavior unchanged.

Correction TDD/checkpoints:

1. Planning checkpoint: update PLAN/TDD-LEDGER/VERIFICATION/RUN-STATE with this Sol/high session and mark verification incomplete; commit/push.
2. RED checkpoint: add generate/validate/bogus tests for trailing `--help`/`-h`, missing and supplied directory flags, and literal `--`; capture focused failures before production edits; commit/push.
3. GREEN checkpoint: add only a docs-scoped compatibility normalizer that removes action-tail help/separator tokens which legacy `parseFlags` ignored while leaving namespace help and all other namespaces untouched.
4. Verification checkpoint: run focused docs/router/golden, base-vs-head differential binaries, full CLI if indicated, temp docs generation/diff and validation, website generation/diff, gofmt, vet, build, and scope/dependency guards; update all artifacts and push without PR or external review.

Execution decision: `local_critical_path` — one bounded correction in the existing isolated issue branch; the available runtime exposes no subagent tool, and the user prohibited PR/external review.

## Completion note

The original slice reached verified HEAD `ea93b4bb7a7eb09236ad829d5ad6055b0c00c30d`. The accepted bounded review correction above is now in progress; its verification state supersedes the original terminal claim until the new gates pass.
