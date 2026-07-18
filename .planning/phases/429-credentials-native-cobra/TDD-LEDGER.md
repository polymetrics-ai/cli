# Phase 429 TDD Ledger

Issue: #429 — nativize credentials namespace.
Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 429 --skip-research` generated a prompt and is being executed inline. `scripts/gsd prompt programming-loop init --phase 429 --dry-run` failed because the command is absent, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-spf13-cobra`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six issue-local phase artifacts before test or production edits | Complete |
| 1 | RED | Add focused credentials tree/operation/help/security tests; run `go test ./internal/cli -run 'Credentials|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected before production edits: `undefined: newCredentialsCobraCommand`; package build failed |
| 2 | GREEN | Native tree, typed flags/handler, controlled input, action boundary, strict validation | Pass: focused credentials/router `25.475s`; focused race subset `111.267s` |
| 3 | Refactor | Focused/repeated/race/security/router/golden/full CLI and exact legacy differential | In progress: golden pass `5.513s`; exact start differential 28/28; full CLI pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Parity/delivery | Built help/list/error checks, docs/website/generated diff, scope/dependency guards, commit/push | Pending |

## Planned RED coverage

- `credentials` is absent from legacy wrappers and native Cobra owns `add`, `list`, `inspect`, `test`, `remove`, and hidden positional `help`.
- `add` declares `stringArray` flags `connector`, `from-env`, `value-stdin`, and `config`, all with legacy bare-value `true`, repeated semantics, unknown tolerance, and no-file completion seam.
- Add/list/remove preserve text and JSON output, spaced/assigned/bare/repeated flags, extra positionals, unknown flags, and fresh-tree re-entrancy.
- Bare/text/JSON/long/short/positional help remains canonical; action-tail help and literal `--` retain legacy behavior.
- Invalid actions and invalid assigned global booleans retain usage/validation categories; valid assigned global booleans remain effective.
- Credential, connector, secret-field, environment-variable, and config-key names reject control/path-traversal input before persistence or environment/stdin reads.
- Warehouse/outbox paths cannot escape the temporary root without explicit existing opt-in; allowed local and file-source paths remain valid.
- `--from-env` supports repeated mappings, detects malformed/missing/empty named sources, and never prints values. `--value-stdin` reads only controlled Cobra input, trims only trailing CR/LF, and final repeated field selection remains compatible. Config-only credentials remain valid; no interactive input path exists.
- Opaque synthetic env/stdin fixtures are absent from stdout, stderr, and state metadata after success and error paths; tests never log fixture content.
- Leading unknown, short, assigned help-like, and literal-boundary tokens cannot discover or execute later add/remove actions; temporary state remains unchanged.

## Exact RED

Captured after the complete focused test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/credentials_cli_test.go:22:9: undefined: newCredentialsCobraCommand
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing native constructor is intentional. The same test-only checkpoint specifies native ownership/flags, list/add/remove, help/global behavior, strict identifiers/path containment, controlled env/stdin source handling, redaction, legacy action-tail compatibility, and fail-closed action discovery. No test executed and no stdin/environment credential fixture was read during this RED build failure.

## Focused GREEN

```text
$ go test ./internal/cli -run 'Credentials|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t25.475s

$ go test -race ./internal/cli -run 'CredentialsCommandIsNative|CredentialsSecretSources|CredentialsOutputsAndErrors|CredentialsLeadingInvalid' -count=1
ok  \tpolymetrics.ai/internal/cli\t111.267s
```

Native Cobra now owns the complete credentials subtree. Current flags are typed `StringArray` values with bare compatibility, stdin comes only from `cmd.InOrStdin`, invalid action heads are bounded before Cobra discovery, and all identifier/config/path checks run before named environment or stdin values are read. Opaque fixture tests pass without emitting fixture content. The golden transcript test passes and a 28-case start-vs-head differential matches exact exit/stdout/stderr for preserved help, list, add flag forms, unknown/extra/tail/literal inputs, invalid heads, and globals.

## Evidence policy

GREEN evidence must include exact command/result and no secret material. `verificationPassed` remains false until the complete declared verification including `make verify` exits 0.
