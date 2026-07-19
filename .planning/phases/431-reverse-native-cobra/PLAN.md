# Phase 431 Plan — Reverse native Cobra namespace

Issue: polymetrics-ai/cli#431
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/431-reverse-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `0b03361e3ec5082d54c416a31715851f71e845fa`
Invocation session: `issue-431-pi-openai-codex-gpt-5.6-sol-high-20260719T010451Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — #431 is the next serialized Phase 9 namespace in its assigned isolated worktree; central router writes collide with sibling units, no subagent tool is exposed in this session, and the user bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed.
- `scripts/gsd prompt plan-phase 431 --skip-research`: generated and executed inline.
- `scripts/gsd prompt programming-loop init --phase 431 --dry-run`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Required reading and skills

Read issue #431 and parents #407/#397; `AGENTS.md`; GSD, issue, parent-orchestration, worker-handoff, CLI parity, Stage 9, ADR-0002, architecture plan §5/§9, reverse manual/website/router/app/tests, and `internal/safety/smoke_makefile_test.go`.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only the `reverse` legacy wrapper with a native Cobra subtree for `list`, `plan`, `preview`, `run` (the approval/confirmation execution step), `status`, and hidden positional `help` compatibility.
- Declare every current reverse flag with typed repeated pflags and legacy bare/assigned/repeated behavior: plan `--source-table`, `--destination`, `--map`, `--action`, `--limit`; run `--approve`, `--confirm`.
- Preserve first-operand ownership for plan name, preview/run plan ID, and status run ID before Cobra normalization; help-like, literal `--`, short, unknown, and carrier-shaped operands must fail closed exactly as legacy behavior does.
- Adapt reverse handlers to typed values and remove only reverse's `parseFlags` calls. Dynamic connector dispatch remains on `parseFlags`.
- Preserve exact exit taxonomy, stdout/stderr and one-envelope behavior, approval-token nondisclosure in JSON/logs/errors, typed confirmation, single-use approval, and strict plan → preview → approval → execute ordering.
- Use only local fakes, built-in local connectors, and `t.TempDir` state. No external write or service call.

Excluded: other namespaces; connector bundles; dynamic connector parser; dependencies; services; credentials; external HTTP/SQL writes; Phase 18 guided session; Phase 19 help/man churn; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — focused tests specify:
   - native reverse tree/action ownership and all current flag definitions/types;
   - list/plan/preview/approval-bearing run/status using temporary local state only;
   - bare/text/JSON/long/short/positional help;
   - trailing help, literal `--`, first-operand ownership, unknown flags/actions, action-discovery boundaries, and assigned global booleans;
   - exact usage/validation/internal exit taxonomy and stdout/stderr/JSON envelopes;
   - token absence from JSON, diagnostics, and local logs;
   - typed confirmation and strict plan → preview → approval → execute order, with no fake writer invocation before all gates pass.
   Capture focused failure before production edits; commit/push tests.
3. **GREEN checkpoint** — add the smallest reverse command and typed handlers; remove reverse from `cobraLegacyCommands`; add only reverse-specific normalization/private operand state; remove only reverse parser calls.
4. **Refactor/parity checkpoint** — focused/repeated/race/router/golden/full CLI and reverse app tests; exact-start differential for preserved parser edges; runtime help and docs/website generation checks.
5. **Final checkpoint** — gofmt, vet, full tests, build, and established ordered `make verify` gate; finalize six artifacts; commit/push; no PR.

## CLI parity stance

Command names, flags, manual bytes, JSON envelopes, docs, website content, generated artifacts, and golden fixtures should remain unchanged. Checked-in docs/website/golden edits are not applicable unless a real mismatch is found; Phase 19 owns deliberate help-tree churn. Verify `pm help reverse`, bare `pm reverse`, `pm reverse --help`, JSON manual, invalid actions, generated `docs/cli/reverse.md`, and website generation.

## Safety

No secret values, external connectors, credentialed checks, optional services, dependency changes, unrestricted writes, destructive/admin actions, or production deploys. New tests may execute only local fake writers or temporary local outbox state and must prove no execution before plan, preview, approval, and any required typed confirmation. Approval values stay in memory and are never included in committed artifacts, failure messages, JSON, or logs. The only allowed broader smoke is the repository's already-established ordered reverse gate inside final `make verify`.

## Completion

Completed and verified at implementation head `f5aeafb7bb7a6702077382e98acb790d3865073f`. Strict RED preceded production edits. Native reverse list/plan/preview/run/status/help, all current flags, invocation-private operand ownership, exact error/output behavior, token nondisclosure, typed confirmation, and ordered local fake execution pass focused, repeated, race, router, golden, reverse-app, full CLI/repository, 21-case exact-start differential, runtime help, generated docs/website, gofmt, vet, build, scope/dependency, and `make verify` gates. Only reverse parser calls were removed. No external write, service, live credential, dependency, unrelated change, PR, or review was used.
