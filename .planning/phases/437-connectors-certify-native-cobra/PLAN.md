# Phase 437 Plan — connectors and certify native Cobra

Issue: #437
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/437-connectors-certify-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`
Invocation: `issue-437-pi-sol-high-20260719T095145Z`; profile `Sol`; thinking `high`.
Execution decision: `local_critical_path` — this is the final assigned serialized Phase 9 namespace in an isolated worktree; router changes collide with sibling units, no subagent tool is exposed, and the user requested implementation/commit/push with no PR/review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: pass (69 commands).
- `scripts/gsd prompt plan-phase 437 --skip-research`: generated and executed inline.
- Required `programming-loop` command is absent from the adapter registry (`unknown GSD command`), so the manual fallback is `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` with these six pre-production artifacts and strict RED → GREEN → refactor.

## Required context and skills

Read issue #437 and parents #407/#397; `AGENTS.md`; GSD/manual/issue/parent contracts; connector migration handoff, conventions, v2 design, certification design/contracts; CLI Architecture v2 plan/execution prompt; ADR-0002; CLI help/docs/website parity; current connectors/certify/router/golden/manual/website code.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## Scope

- Replace only the legacy top-level `connectors` wrapper and `runConnectors`/`runCertify` namespace parsers with native Cobra commands.
- Native connector actions: `list`, `catalog`, `inspect`, hidden positional `help`, and compatibility aliases `man`/`docs`; preserve current metadata-only behavior and output.
- Native nested certify actions: single connector, `--all`, and `--sweep`, declaring every currently consumed flag: `credential`, `from-env`, `config`, `stream`, `limit`, `modes`, `skip`, `rate-limit`, `budget`, `record`, `replay`, `live-all-modes`, `allow-production-writes`, `keep-workdir`, `write`, `full`, `credentials-file`, `parallel`, `resume`, `older-than`.
- Preserve legacy repeated-last, bare, assigned, space, operand, ignored trailing, unknown-flag, literal `--`, malformed unknown, action/operand discovery, invocation-global, text/JSON, and error-category behavior where current handlers consume it.
- Preserve `cli.Run` in-process re-entrancy, certify exit 0/1/2/3 mapping, context cancellation, bounded cross-connector concurrency, event sequence, telemetry span names/status, and secret/credential-value exclusion.
- Keep dynamic `pm <connector> <path...>` dispatch and its legacy `parseFlags` path exactly sanctioned and unchanged.
- Certify verification is fixture/replay/local only. No live credential checks or writes.

Excluded: connector defs, connector runtime behavior, new certify semantics/flags, live tests, credential values, external services, new dependencies, dynamic parser removal, other namespaces, Phase 16 dashboard, Phase 19 help-tree churn, PR/review.

## TDD and checkpoints

1. Commit/push these six planning artifacts before test or production edits.
2. RED: add focused tests for native tree shape, complete current flags, connector/certify actions and operands, bare/text/JSON/topic/positional/trailing help, literal separator/malformed unknown/action discovery/globals, exact outputs/errors, exits 0/1/2/3, re-entrancy, cancellation/concurrency/events/telemetry, and planted credential-value absence. Capture failure caused by absent native constructors/runtime seam; commit/push.
3. GREEN: introduce the smallest typed flag structs, native constructors, handler adaptation/runtime seam, and compatibility normalization. Remove only connectors/certify namespace `parseFlags` calls and connectors legacy registration.
4. Refactor: focused ×10, race, router/golden/full CLI/certify; exact-base differential; connector validation; docs/manual/website generation; runtime smoke; gofmt/vet/test/build/make verify.
5. Finalize six artifacts, commit/push, no PR/review.

## Parity and safety

Bare `pm connectors` must render the canonical manual and exit 0; invalid actions remain usage exit 2. Update the canonical connectors manual to document certify commands and 0/1/2/3 exits, regenerate `docs/cli/connectors.md`, and mirror the bounded surface in `website/content/docs/cli-reference.mdx`; generated website data follows existing scripts. Completion registration is unchanged and Phase 15/19 work is not pulled forward.

All certify command tests use sample fixtures, replay/local fakes, `t.TempDir`, injected runner/sweeper seams, synthetic credential variable names, and planted non-secret sentinel values solely to assert absence. Never print, summarize, or persist a credential value. No real env-secret resolution, live connector check, write, sweep against external systems, reverse ETL execution, model/runtime service, dependency, or broad path.
