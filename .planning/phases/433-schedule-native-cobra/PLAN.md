# Phase 433 Plan — Schedule native Cobra namespace

Issue: polymetrics-ai/cli#433
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/433-schedule-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `ab1c79eede67fa87e1c6b808d6ddba0b27fcf00d`
Invocation session: `issue-433-pi-sol-high-20260719T044819Z`
Explicit invocation profile: `Sol`, `high`
Execution decision: `local_critical_path` — #433 is the assigned next serialized Phase 9 unit in an isolated worktree. Its central router scope collides with sibling migrations, this runtime exposes no subagent tool, and the user bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed (69 commands).
- `scripts/gsd prompt plan-phase 433 --skip-research`: generated and is executed inline.
- `scripts/gsd prompt programming-loop init --phase 433 --dry-run`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Required reading and skills

Read issues #433/#407/#397; `AGENTS.md`; GSD adapter/core/manual universal loop; issue and parent-orchestrator contracts; CLI help/docs/website parity; runtime integration reference; architecture plan §5/§9; Stage 9 execution prompt; ADR-0002; schedule CLI, cron, manifests, backend selection/render/install/remove tests; typed config schedule integration; generated schedule manual, website references, golden transcripts; and adjacent native flow/reverse Cobra patterns.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the `schedule` legacy wrapper with a native Cobra subtree for the current public actions `create`, `list`, `install`, `remove`, and hidden positional `help` compatibility.
- Declare the complete existing local flag surface with native pflags while preserving current repeated, bare, assigned, ignored-unknown, malformed, and operand behavior where applicable: `--name`, `--cron`, `--flow`, and `--crontab`, plus invocation-global `--root`, `--json`, `--plain`, `--no-input`, and `--progress` behavior.
- Adapt schedule handlers to typed values and remove only schedule's `parseFlags` call sites and dispatcher. Dynamic connector dispatch remains on `parseFlags`; other namespaces remain untouched.
- Preserve cron validation, manifest name validation, root propagation into installed payloads, typed-config backend selection, context propagation, install error wrapping, best-effort backend cleanup plus crontab fallback, manifest deletion, exact text/JSON envelopes, stdout/stderr discipline, exit taxonomy, and deterministic list ordering.
- Introduce an invocation-local schedule runtime seam only as needed to inject a fixed clock, executable path, backend selector, and fake backends in tests. Default production behavior remains unchanged.
- Use only temporary schedule roots, redirected temporary crontab files, and injected fake backends. Never invoke `crontab`, `launchctl`, `systemctl`, Temporal, or another external scheduler in tests or verification.

The repository's current schedule contract names deletion `remove`; there is no current `uninstall`, `run`, or `history` action. To preserve the golden contract and #433's migration-only scope, RED coverage will assert those action heads remain usage errors and never reach a backend. They are not new Phase 9 commands. Interactive schedule creation remains Phase 11 (#409).

Excluded: new schedule actions or aliases; Phase 11 wizard/cron presets/flow validation/next-fire preview; other namespaces; connector bundles; dynamic connector parser; new dependencies; services; credentials; real scheduler effects; Phase 19 focused help/man churn; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — focused tests specify:
   - native schedule ownership for create/list/install/remove/help, all native flag types/NoOpt behavior/unknown tolerance/completion seams, and absence from legacy wrappers;
   - every current flag and operand form, repeated last-value behavior, bare/assigned flags, strict first positional ownership, and unchanged project-root payload semantics;
   - bare/text/JSON/long/short/positional help; trailing help; literal `--`; malformed/legal unknown flags; invalid action heads including `uninstall`, `run`, and `history`; no later-action discovery; assigned global booleans;
   - cron/name/not-found/backend validation, context propagation, exact error categories, one-envelope JSON, and deterministic create/list/install/remove outputs;
   - fake install/remove backend calls, non-crontab cleanup fallback, and no backend call on invalid input/action.
   Capture focused failure before any production edit; commit/push tests.
3. **GREEN checkpoint** — add the smallest native schedule command and typed handlers; add only schedule-specific normalization/private operand state needed for exact compatibility; remove schedule from `cobraLegacyCommands`; delete only `runSchedule` and its schedule `parseFlags` uses once unused.
4. **Refactor/parity checkpoint** — run focused/repeated/race schedule/router/golden/full CLI and schedule-package tests, exact-start parser/output differential where deterministic, runtime help, generated docs/website checks, formatting, vet, build, and scope/dependency guards.
5. **Final checkpoint** — run the established full `make verify` without real scheduler effects; finalize all six artifacts; commit/push; no PR or review.

## CLI parity stance

Public command names, flags, manual bytes, output schemas, docs, website content, generated artifacts, and golden fixtures should remain unchanged. Checked-in docs/website/golden edits are not applicable unless verification finds a real mismatch. Phase 11 owns the schedule-create wizard; Phase 19 owns deliberate focused-help/man churn. Verify `pm help schedule`, bare `pm schedule`, `pm schedule --help`, short/positional/JSON manual routes, invalid actions, `docs/cli/schedule.md`, website CLI-reference/architecture pages, generated docs, completion discovery, and golden transcripts.

## Safety

No secrets or approval values, external connectors, credentialed checks, optional services, dependencies, unrestricted writes, destructive/admin actions, reverse ETL, or production deploys. Tests use temp roots/temp crontab files and fakes only. No real crontab, launchd, systemd, or Temporal scheduler call is allowed. Context cancellation must remain propagated and all file effects must remain bounded to temporary directories.

## Completion

DRAFT — production code has not been edited. The RED, GREEN, parser differential, parity, full verification, commit, and push evidence will be appended as executed.
