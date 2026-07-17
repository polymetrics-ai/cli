# Phase 423 TDD Ledger

Issue: #423 — nativize perf namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-code-style`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `golang-how-to`: CLI command tree routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: preserve exit codes, stdout/stderr discipline, CLI unit tests, and machine-readable output.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior/public contract over implementation-only details.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context when propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording; application CLI help is primary documentation.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags guidance for `StringArray`, `NoOptDefVal`, and unknown-flag compatibility.
- `golang-security`: trust-boundary questions #1-#3; no secrets; command args are untrusted; no runtime services started for tests.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.
- `golang-code-style`: early returns, clear small helpers, semantic line breaks.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 423 --skip-research >/tmp/gsd-plan-phase-423.prompt
scripts/gsd prompt programming-loop init --phase 423 --dry-run >/tmp/gsd-programming-loop-423.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-423.prompt` (10664 bytes).
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | Pending | Pending | Add focused native-perf subtree and behavior tests before production code. |
| 2 | Green | Pending | Pending | Implement minimal native perf Cobra subtree. |
| 3 | Refactor | Pending | Pending | Run focused/golden gates. |
| 4 | Full gate | Pending | Pending | Run full local gates and CLI parity checks. |

## Planned red tests

- `TestPerfCommandIsNativeCobraSubtree`: current wrapper should fail because `perf` remains legacy; expected native `compare`/`sync-modes` subcommands, declared `StringArray` flags, `NoOptDefVal="true"`, unknown-flag whitelist, and no-file completion seams are missing.
- `TestPerfCompareFlagFormsPreserveLegacySemantics`: current metadata path should fail until pflag declarations and normalization exist; behavior cases cover space/equals forms, repeated scalar last-wins, bare bool/value sentinels, unknown flags, extra args, late globals, JSON envelope preservation, and runtime config endpoint use.
- `TestPerfSyncModesFlagFormsPreserveLegacySemantics`: records space/equals/repeated/bare-value semantics and output envelope preservation.
- `TestPerfBareAndInvalidActionSemantics`: bare namespace help must exit 0 and invalid action must remain usage exit 2.

## Exact red outputs

Pending — capture before production code.

## Exact green outputs

Pending.
