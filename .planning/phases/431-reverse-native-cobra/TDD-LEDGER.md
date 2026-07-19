# Phase 431 TDD Ledger

Issue: #431 — nativize reverse namespace.
Invocation session: `issue-431-pi-openai-codex-gpt-5.6-sol-high-20260719T010451Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `0b03361e3ec5082d54c416a31715851f71e845fa`

## GSD and skills

Doctor/list passed; `plan-phase 431 --skip-research` was generated and is executed inline. The adapter lacks `programming-loop`, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-documentation`, `golang-spf13-cobra`.

## RED / GREEN / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity/exact start before test or production edits | Complete |
| 1 | RED | `go test ./internal/cli -run 'TestReverse(Command|Local|PlanJSON|FirstOperand|HelpTrailing|ExactExit|Cancellation)' -count=1` | Failed as required before production edits: `internal/cli/reverse_native_cobra_test.go:23:9: undefined: newReverseCobraCommand` |
| 2 | GREEN | Native reverse tree + typed handlers + reverse-only normalization/private operands; remove wrapper/parser use | Pass: focused contract `28.527s`; existing reverse/router safety suite `62.562s` |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI/reverse app and exact-start differential | Pass |
| 4 | Full gate | gofmt, vet, full tests, build, established ordered `make verify` | Pass |
| 5 | Parity/delivery | Runtime help, generated docs/website/golden checks, scope/dependency guards, commit/push | Pass; final evidence prepared for commit/push |

## RED contract

- Native `reverse` owns `list`, `plan`, `preview`, `run`, `status`, and hidden positional `help`; the `run` node owns explicit approval/confirmation flags and no reverse legacy wrapper remains.
- Plan flags `source-table`, `destination`, `map`, `action`, and `limit`, plus run flags `approve` and `confirm`, are `StringArray` (not comma-splitting), `NoOptDefVal=true`, repeatable, and accept assigned/spaced/bare legacy forms. Last value wins except repeated mappings, which all accumulate with later duplicate keys winning.
- Bare namespace and `pm help reverse`, `reverse --help`, `reverse -h`, `reverse help`, and JSON manual routes preserve the canonical reverse manual and exit 0.
- Trailing help, literal `--`, short flags, and unknown flags retain legacy compatibility rather than becoming accidental Cobra controls.
- Invalid actions remain usage exit 2. Leading unknown/help-like tokens cannot discover and execute a later valid action.
- Plan name and preview/run/status IDs are owned by the first token after the action, even when it looks like `--help`, `-h`, literal `--`, an unknown flag, or an internal-carrier-shaped token. Later operands cannot replace them.
- Assigned global `--json`, `--plain`, and `--no-input` preserve validation and placement behavior.
- Usage failures exit 2; malformed mappings/endpoints/integers exit 3; local missing objects/runtime failures remain exit 1. JSON failures produce one Error envelope on stdout and a redacted diagnostic on stderr.
- Human plan output may return an approval value, but JSON plan, preview, list, run/status output, stderr, errors, telemetry/log files, and committed fixtures never contain that value.
- Local fake execution cannot occur before plan creation, explicit preview, valid approval, and required typed confirmation. Invalid/missing approval and confirmation produce zero writer calls. Confirmed execution produces exactly one local fake/temp-state write, then replay is rejected.
- No external HTTP, SQL, connector, runtime, credential, or destructive write is performed.

## Exact RED

Captured after the complete focused test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/reverse_native_cobra_test.go:23:9: undefined: newReverseCobraCommand
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing native constructor is intentional. The tests specify native ownership and every current flag; local plan/preview/approval-bearing run/status/list; all manual routes; trailing help/literal/unknown/action-discovery behavior; strict first-operand ownership; exact exit taxonomy; token nondisclosure; typed confirmation; single-use approval; cancellation; and no local fake writer invocation before all gates pass. Tests use only built-in local connectors and temporary state. No external write, service, credential, dependency, or token value entered the evidence.

## Focused GREEN

`newReverseCobraCommand` now owns list/plan/preview/run/status/help with typed `StringArray` flags, unknown tolerance, no-file completion seams, invocation-private first-operand capture, and reverse-only legacy-tail normalization. Typed handlers preserve output and safety gates; reverse is absent from legacy wrappers and its two `parseFlags` calls are removed. The focused contract passed in `28.527s`; existing reverse/router/validation safety tests passed in `62.562s`. The ordering test performed only a temporary local outbox fake write after preview, valid approval, and typed confirmation; all earlier attempts produced zero writes. No external request or token value entered test output or artifacts.

## Final GREEN / refactor evidence

- Focused contract `28.527s`; repeated ×5 `143.255s`; focused `-race` `322.315s`.
- Existing reverse/router/validation safety suite `62.562s`; router/golden/manual gate `7.754s`; tracked fixture unchanged.
- Reverse app tests `4.615s`; reverse app `-race` `29.405s`; smoke-order safety test `0.390s` and full safety package `0.411s`.
- Full CLI `396.560s`; full app `29.601s`.
- Exact-start binary differential against `0b03361e3ec5082d54c416a31715851f71e845fa`: 21/21 exit/stdout/stderr transcripts match across manual routes, invalid/action-discovery cases, list tails, help-like operands, missing objects, malformed flags, and global booleans.
- Runtime `pm help reverse`, bare `pm reverse`, long/short help are byte-identical with empty stderr. JSON manual, invalid-action usage, generated `docs/cli/reverse.md`, and website docs generation pass with no tracked delta.
- `gofmt -w cmd internal`, `go vet ./...` (`3.104s`), `go test -timeout 20m ./...` (`6m44.294s`; CLI `399.524s`, certify `340.870s`), and `go build ./cmd/pm` (`1.865s`) pass.
- Final `make verify` passed in `6m56.086s`: CLI `388.922s`, lint 0 issues, 547 connector definitions/0 findings, and only its established temporary-root plan → preview → approval → run smoke.
- Scope/dependency guards pass: no `go.mod`, `go.sum`, connector definition, docs/website/golden, generated, or unrelated namespace delta; production has no reverse internal argv carrier; dynamic connector `parseFlags` remains.

No approval value was copied into this ledger, artifacts, JSON, diagnostics, logs, or handoff. No external write, service, live credential, dependency, PR, or review was used.

## Review compatibility correction — session `issue-431-parser-compat-20260719T022304Z`

Exact correction start: `c8f5b9e97a2f71f25cdb362af0055c1c31dc8420` (clean and equal to upstream at kickoff).

GSD doctor/list passed. Required `scripts/gsd prompt programming-loop init --phase 431 --dry-run` failed because `programming-loop` is absent from the 69-command adapter registry, so the recorded manual universal-runtime-loop fallback remains active. Loaded for this correction: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-spf13-cobra`; CLI help/docs parity was reviewed and is not applicable because the public surface is unchanged.

Execution decision: `local_critical_path` — this is a single bounded parser correction in the already-isolated #431 worktree; no subagent tool is exposed, and the user prohibited external writes/services/dependencies/PR/review.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| C0 | Planning | Reopen PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY at exact correction start before test or production edits | Complete |
| C1 | RED | `go test ./internal/cli -run '^TestReverseMalformedUnknownFlagsPreserveLegacyActionOutcomesAndNoEffects$' -count=1` before production edit | Failed as required: all 50 malformed/action combinations returned pflag usage exit 2 instead of baseline action outcomes (list 0, plan validation 3, preview/run/status missing-object 1); no state/outbox effect or approval output |
| C2 | GREEN | Normalize only malformed unknown reverse tail tokens before pflag; preserve known flags, operands, approvals, and legal unknown behavior | Pending |
| C3 | Verify | 324-case exact-start differential; focused/reverse/race/full CLI; no approval output; gofmt/vet/build/diff | Pending |
| C4 | Delivery | Finalize artifacts, commit, and push; no PR/review | Pending |

### Correction RED evidence

The table-driven differential covers five reverse actions × ten malformed unknown forms: `--=x`, `--=`, `--==x`, `---x`, `---x=y`, `---`, `----x`, `----x=y`, `----`, and `-----x`. Every one of the 50 subtests failed at the exact differential assertion before any production edit: pflag returned usage exit 2 with different stdout/stderr, while the corresponding no-tail baseline retained the legacy action outcome. State bytes remained unchanged, no outbox appeared, and no approval output was emitted. This directly reproduces the 50 mismatches reported in `/tmp/pm-397-review-431.log`.
