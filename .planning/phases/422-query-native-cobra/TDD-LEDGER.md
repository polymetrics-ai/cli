# Phase 422 TDD Ledger

Issue: #422 — nativize query namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-database`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `golang-how-to`: CLI command tree routes to `golang-spf13-cobra` + `golang-cli`; database/SQL code routes to `golang-database` + `golang-security`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: preserve exit codes, stdout/stderr discipline, CLI unit tests, and machine-readable output.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior/public contract over implementation-only details.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context when propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording; application CLI help is primary documentation.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags guidance for `StringArray`, `NoOptDefVal`, and unknown-flag compatibility.
- `golang-security`: trust-boundary questions #1-#3; no secrets; command args/SQL are untrusted; no generic SQL write.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.
- `golang-database`: #2 parameterized/no user-input SQL concatenation, #3 context propagation, #5 close rows where DB rows are used, #14 no schema writes. Query-engine behavior stays unchanged in this phase.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 422 --skip-research >/tmp/gsd-plan-phase-422.prompt
scripts/gsd prompt programming-loop init --phase 422 --dry-run >/tmp/gsd-programming-loop-422.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-422.prompt` (10668 bytes).
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | Pending: `go test ./internal/cli/ -run 'Query|CobraRouterShell' -count=1` | Pending | Will fail until `query` is native and invalid action avoids project open. |
| 2 | Green | Pending | Pending | Native query parser green. |
| 3 | Refactor | Pending | Pending | Gofmt + focused gate. |
| 4 | Full gate | Pending | Pending | Required issue verification and parity checks. |

## Planned red tests

- `TestQueryCommandIsNativeCobraSubtree`: current wrapper should fail because `query.DisableFlagParsing` is true, no native `run` subcommand exists, native flags are missing, unknown-flag whitelist is absent, and no completion/no-file seam exists.
- `TestQueryRunFlagFormsPreserveLegacySemantics`: current metadata path should fail until pflag declarations and normalization exist; behavior cases cover space/equals forms, repeated scalar last-wins, repeated/comma `--fields` accumulation, bare bool sentinel values, unknown flags, extra args, late globals, and JSON envelope preservation.
- `TestQueryInvalidActionIsUsageBeforeProjectOpen`: invalid actions must remain usage exit 2 and must not open `.polymetrics` first.
- `TestQueryRunRejectsWriteSQL`: `--sql` write attempts must continue to be rejected by the query engine guard; no generic SQL write exposed.

## Exact red outputs

Pending capture before production edits.

## Exact green outputs

Pending.
