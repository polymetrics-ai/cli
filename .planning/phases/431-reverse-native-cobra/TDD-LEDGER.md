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
| 2 | GREEN | Native reverse tree + typed handlers + reverse-only normalization/private operands; remove wrapper/parser use | Pending |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI/reverse app and exact-start differential | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, established ordered `make verify` | Pending |
| 5 | Parity/delivery | Runtime help, generated docs/website/golden checks, scope/dependency guards, commit/push | Pending |

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

Exact GREEN, refactor, parity, and final gate output will be appended as commands run. Token values must never be copied into this ledger.
