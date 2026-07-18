# Phase 427 TDD Ledger

Issue: #427 — nativize docs namespace.
Invocation session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 427 --skip-research` generated the official prompt and it was executed inline. `scripts/gsd prompt programming-loop init --phase 427 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | `go test ./internal/cli/ -run 'Docs|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected (`11.332s`) |
| 2 | GREEN | Native namespace/actions/typed flags and docs-only legacy parser removal | Pass (`11.462s`) |
| 3 | Refactor | Focused docs/router/golden + full CLI tests | Pass (`18.453s`; golden `5.470s`; full CLI `227.224s`) |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass |
| 5 | Parity/safety | built binary, docs/website/generated/golden/dependency/scope checks | Pass |

## Planned RED coverage

- `docs` is native and absent from legacy wrappers; `generate` and `validate` are native actions with declared `stringArray` output-directory flags, `NoOptDefVal=true`, unknown-flag whitelists, and no-file completion fallback.
- `generate` preserves `--dir` and `--connectors-dir` spaced/assigned/repeated-last-wins/bare/comma forms, exact output text, exact CLI-manual bytes, and expected connector artifacts.
- `validate` preserves `--connectors-dir`, fallback `--dir`, default semantics where safely applicable, unknown flags, extra positionals, and exact output text.
- Bare/text/JSON/flag/short/positional help all resolve to the canonical docs manual.
- Invalid action remains usage exit 2; missing/empty generate dir and malformed assigned globals retain existing categories; errors are not masked by help.
- Global/config forms cover `--root value`, `--root=value`, late placement, assigned `--json=true|false`, `--plain=false`, and `--no-input=true`.
- Filesystem tests write only under safe temporary roots, assert all generated relative paths remain local, compare CLI docs byte-for-byte to the canonical map/checked-in tree, and validate connector output without secrets.

## Exact RED

Captured after test-only edits and before any production-code edit:

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestDocsCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:213: docs command must use native Cobra flag parsing
FAIL
FAIL\tpolymetrics.ai/internal/cli\t11.332s
FAIL
```

All observable docs behavior, byte-parity, action/flag, error, global/config, and safe temp-root filesystem tests passed through the legacy wrapper. RED isolates the required parser ownership and registration change. Production files remained untouched at this checkpoint.

## Focused GREEN

```text
$ go test ./internal/cli/ -run 'Docs|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t11.462s

$ go test ./internal/cli/... -run 'Docs|CobraRouterShell|Golden' -count=1
ok  \tpolymetrics.ai/internal/cli\t18.453s

$ go test ./internal/cli/ -run '^TestGoldenTranscripts$' -count=1
ok  \tpolymetrics.ai/internal/cli\t5.470s

$ go test ./internal/cli/... -count=1
ok  \tpolymetrics.ai/internal/cli\t227.224s
```

Implementation: `docs`, `docs generate`, and `docs validate` are native Cobra nodes; `--dir` and `--connectors-dir` are `StringArray` flags with legacy bare-flag behavior and spaced-value normalization; unknown action flags remain whitelisted; positional help is a hidden compatibility node. `runDocs` accepts typed action flags directly, so the docs-only `parseFlags` call and legacy wrapper are removed. Focused byte/output/help/global-config/filesystem tests and unchanged goldens pass.

## Full GREEN and parity

| Command / gate | Result |
|---|---|
| Final focused docs/router tests after default-path hardening | Pass (`13.710s`) |
| Final docs/router/golden subset | Pass (`20.158s`) |
| `go test ./internal/cli/... -count=1` | Pass (`227.224s`) |
| `go test ./internal/cli/ -run '^TestGoldenTranscripts$' -count=1` | Pass (`5.470s`) |
| `gofmt -w cmd internal`; `go vet ./...`; `go build -o /tmp/pm-427 ./cmd/pm` | Pass, no diagnostics |
| `go test -timeout 20m ./...` | Pass (`real 347.167s`; CLI `229.851s`, certify `342.890s`) |
| final `make verify` | Pass (`real 249.846s`; CLI `229.679s`, smoke OK, lint `0 issues`, 547 connectors / 0 findings) |
| built binary help/generation/validation/error/config/global matrix | Pass: `help_bytes=818`, exits invalid=`2`, missing=`1`, plain override=`2` |
| temp docs generation + `diff -ru docs/cli`; temp and tracked docs validation | Pass, no diff |
| `npm --prefix website run gen:docs`; website diff | Pass, 11 pages, no diff |
| start-HEAD dependency/docs/website/golden/connector-def and `git diff --check` guards | Pass |

`make verify` used only its existing local temporary sample smoke. Its reverse ETL step followed plan → preview → approval → run; no external service or credentialed connector check was used. Focused generation wrote only beneath temporary roots, compared every generated CLI manual byte-for-byte, asserted generated relative paths stayed local, and validated connector catalog/manual/icon artifacts without reading credentials.
