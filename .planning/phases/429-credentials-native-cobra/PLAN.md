# Phase 429 Plan — Credentials native Cobra namespace

Issue: polymetrics-ai/cli#429
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/429-credentials-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting parent HEAD: `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`
Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`
Explicit invocation profile: `model=openai-codex/gpt-5.6-sol`, `thinking=high`
Execution decision: `local_critical_path` — ninth serialized Phase 9 namespace unit is assigned to this isolated branch/worktree. Central router files collide with later units, this session exposes no subagent tool, and the user bounded delivery to #429 with no PR or external review.

## Required reading complete

- Issue #429 via `gh`; parent #397; umbrella #407; draft parent PR #438; prior native Cobra patterns through #428.
- `AGENTS.md`; issue-agent, parent-orchestrator, and worker-handoff contracts; GSD Pi adapter and universal/manual programming loop.
- `.planning/config.json`, `PROJECT.md`, `ROADMAP.md`, `STATE.md`; universal programming-loop PRD/prompt library.
- Required-skills routing and CLI help/docs/website parity.
- CLI Architecture v2 plan §5/§9; execution prompt Stage 9; ADR 0002.
- Credentials manual/website surface, router/global parser/error funnel, app credential lifecycle, safety path validation, vault implementation/tests, golden tests, and adjacent namespace tests.

## GSD adapter and fallback

- `scripts/gsd doctor` — pass.
- `scripts/gsd list` — pass; 69 commands listed.
- `scripts/gsd prompt plan-phase 429 --skip-research` — generated and executed inline.
- `scripts/gsd prompt programming-loop init --phase 429 --dry-run` — unavailable: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual fallback: execute `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` inline with strict RED → GREEN → refactor evidence and all six issue-local artifacts.

## Required skills loaded

- `gsd-core`.
- `golang-how-to` first, then `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, and `golang-spf13-cobra`.
- Applied constraints: fresh Cobra trees; typed repeated pflags; controlled stdin via Cobra input; raw user-selected environment lookup only; strict identifier/path checks; stable error mapping and stdout/stderr separation; no secret values in argv, logs, output, artifacts, or handoff.

## Scope and exclusions

Allowed:

- Register native `credentials` and `add|list|inspect|test|remove` nodes plus positional `credentials help` compatibility.
- Declare current `add` flags using repeated `StringArray` values and preserve bare, assigned, repeated-last-wins/all-values, unknown-flag, extra-positional, trailing-help, and literal-`--` behavior where legacy requires it.
- Adapt only `runCredentials` away from `parseFlags`; use Cobra input for `--value-stdin` without interactive entry.
- Add credentials-only action-discovery boundaries so invalid leading tokens cannot reach a later mutating action.
- Add focused security/router tests using only opaque synthetic redaction fixtures that are never logged or printed.
- Preserve strict credential/connector identifiers and existing warehouse/outbox path-containment and explicit external-path opt-in behavior.

Excluded:

- Interactive secret entry; credential rotation/export; vault format/crypto changes; connector checks requiring credentials; other namespaces; dynamic connector parsing; connector bundles; dependencies; Phase 15 completion implementation; Phase 19 help/man churn; shared parent artifacts; PR or external review.

## Existing contract to preserve

- Bare `pm credentials`, `pm help credentials`, `pm credentials --help`, `pm credentials -h`, and positional `pm credentials help` render the canonical manual; JSON routes emit `CommandManual/credentials`.
- `add` accepts spaced/assigned/repeated/bare current flags: `--connector`, `--from-env`, `--value-stdin`, and `--config`; user-selected env names remain raw data plumbing, not Viper config.
- Environment and stdin are the only secret-value intake paths; config-only credentials remain valid; repeated environment fields are supported and only the final stdin field is selected as before.
- List/inspect/test/remove output contains metadata only. Secret values never appear in stdout, stderr, state metadata, docs, artifacts, or logs.
- Invalid actions are usage errors; global `--root`, `--json`, `--plain`, and `--no-input` continue before/after namespace in bare/assigned forms.
- Existing warehouse/outbox local-write path containment and `allow_external_path=true` opt-in remain exact; file-source paths retain read-path behavior.

## TDD slices and checkpoints

1. **Planning checkpoint** — create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY before production edits; commit and push.
2. **RED tests first** — specify native tree/flags/completion seam; list/add/remove operations; help matrix; unknown/invalid/global booleans; strict names/path containment; env/stdin selection; redaction; trailing-help/literal compatibility; and no action-discovery bypass. Capture focused RED before production edits; commit/push.
3. **Smallest GREEN** — remove `credentials` from `cobraLegacyCommands`; add native nodes and typed flags; credentials-only pre-Cobra normalization/boundary; typed handlers; Cobra input seam; identifier/source validation; remove only the credentials parser call.
4. **Refactor/parity** — focused/repeated/race/security/router/golden/full CLI tests; exact start-vs-head differential for preserved cases; built binary help/list/error/global checks in temporary roots; docs/website/generated diff checks.
5. **Full gates/delivery** — `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`; scope/dependency/diff guards; finalize six artifacts; coherent commits/push; no PR.

## CLI help/docs/website parity stance

Parser ownership changes but command names, flags, output envelopes, canonical manual, checked-in `docs/cli/credentials.md`, website CLI reference, generated help, and goldens should remain unchanged. Runtime bare/text/JSON/positional help, temporary docs generation diff, website generation diff, and golden tests prove parity. Focused subcommand help/man churn remains deferred to Phase 19; completion values remain deferred to Phase 15.

## Safety

No real secret values, credentialed checks, interactive secret entry, optional services, dependencies, destructive/admin actions, generic write tools, production deployment, PR, external review, or merge. Tests use opaque synthetic redaction fixtures only, never include fixture content in failure output, and run in temporary project roots. No `credentials test` invocation may contact an external endpoint. The required existing local verification gates run without credentialed external checks.
