# TDD LEDGER — Issue 405 TTY gate and NDJSON progress

## Loaded skills

- `gsd-core` — repo-local GSD adapter workflow.
- `caveman` — compact final handoff only.
- `golang-how-to` — required Go skill router.
- `golang-cli` — CLI flags, stdout/stderr discipline, run options.
- `golang-spf13-cobra` — current router/help flag integration.
- `golang-testing` — red/green table tests and CLI contract tests.
- `golang-error-handling` — validation errors and existing exit taxonomy.
- `golang-security` — no secrets in progress/logs/docs; terminal sanitization preserved.
- `golang-safety` — deterministic defaults, nil writer/env handling, no panic-prone assumptions.
- `golang-documentation` — runtime help, docs/cli, website parity.
- UI/design source docs: `docs/design/tui-ux-design.md`, `docs/adr/0003-interactive-tui-layer.md`.

Stack implementation skill note: `.pi/skills/go-implementation/SKILL.md` was requested by worker instructions but is absent in this checkout (`ENOENT`); loaded `gsd-core` plus required global Go skills from `.agents/agentic-delivery/references/required-skills-routing.md` instead.

## GSD command evidence

```bash
scripts/gsd doctor
```

Result: pass.

```bash
scripts/gsd list
```

Result: pass; 69 commands listed.

```bash
scripts/gsd prompt plan-phase 405 --skip-research
```

Result: generated official `/gsd-plan-phase 405 --skip-research` prompt.

```bash
scripts/gsd prompt programming-loop init --phase 405 --dry-run
```

Result: fail, adapter gap: `scripts/gsd: unknown GSD command: programming-loop`.

Fallback: loaded `.pi/prompts/pm-gsd-loop.md` and running the universal programming loop inline/manual; decision `local_critical_path`.

## Red / Green ledger

| Slice | Test / validation | Red evidence | Green evidence | Refactor evidence |
|---|---|---|---|---|
| 1 ui detection/styles | `go test ./internal/ui/... -count=1` | fail (build): undefined `DetectOptions`, `Mode`, `ModeTUI`, `ModePlain`, `ResolveGlyphs`, `ResolvePalette`, `Options`, `ProfileNone`, `TokenOK` | pass: `ok polymetrics.ai/internal/ui 0.155s`; `ok polymetrics.ai/internal/ui/styles 0.286s` | `gofmt -w internal/cli internal/ui` |
| 2 CLI run options/global flags | `go test ./internal/cli/... -run 'TestRunWithOptions|TestGlobalUI|TestProgress' -count=1` | fail (build): undefined `RunWithOptions`, `RunOptions`, `ModePlain`, `ModeAuto` | pass: `ok polymetrics.ai/internal/cli 1.033s` | `gofmt -w internal/cli internal/ui`; test adjusted to preserve existing JSON-error stderr diagnostic contract |
| 3 stderr-only NDJSON | `go test ./internal/cli/... -run TestProgressNDJSON -count=1` | fail (build): same undefined run-options API as slice 2 | pass: `ok polymetrics.ai/internal/cli 1.045s` | no extra refactor |
| 4 docs/help parity | `go test ./internal/cli/... -run 'Test.*Help|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1` plus docs/website grep | fail: `TestGoldenDocsGenerateMatchesTrackedCLIManuals` drifted for `config.md` after docs-map change; help output initially missed new flags | pass: `ok polymetrics.ai/internal/cli 1.448s` after `pm docs generate`, golden transcript update, website update | docs regenerated; golden transcripts updated intentionally |
| focused package | `go test ./internal/cli/... -count=1` | pending after slices | pass: `ok polymetrics.ai/internal/cli 169.642s` | no extra refactor |
| final gates | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify` | pending after focused green | pass: combined gate exited 0; `go test ./...` included `internal/cli 173.250s`, `internal/connectors/certify 346.045s`; `make verify` passed with `internal/cli 174.571s`, `internal/connectors/certify 348.926s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings` | `go mod tidy` in `make verify`; no post-gate working-tree changes |

## Red test capture rule

Before production edits, add focused failing tests only. Capture exact command and failure output here before implementing each slice.

### Initial red evidence — 2026-07-17

```bash
go test ./internal/ui/... -count=1
```

Result: fail/build. Key output:

```text
internal/ui/detect_test.go:8:8: undefined: DetectOptions
internal/ui/detect_test.go:9:8: undefined: Mode
internal/ui/detect_test.go:14:10: undefined: ModeTUI
internal/ui/styles/styles_test.go:9:13: undefined: ResolveGlyphs
internal/ui/styles/styles_test.go:21:11: undefined: ResolvePalette
FAIL	polymetrics.ai/internal/ui [build failed]
FAIL	polymetrics.ai/internal/ui/styles [build failed]
```

```bash
go test ./internal/cli/... -run 'TestRunWithOptions|TestGlobalUI|TestProgress' -count=1
```

Result: fail/build. Key output:

```text
internal/cli/ui_options_test.go:21:14: undefined: RunWithOptions
internal/cli/ui_options_test.go:21:61: undefined: RunOptions
internal/cli/ui_options_test.go:21:78: undefined: ModePlain
internal/cli/ui_options_test.go:35:20: undefined: ModeAuto
FAIL	polymetrics.ai/internal/cli [build failed]
```

```bash
go test ./internal/cli/... -run TestProgressNDJSON -count=1
```

Result: fail/build on same missing run-options API. This is the required red evidence before production edits.

## Final gate evidence — 2026-07-17

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Result: pass. Key output:

```text
ok  polymetrics.ai/internal/cli 173.250s
ok  polymetrics.ai/internal/connectors/certify 346.045s
...
make verify
ok  polymetrics.ai/internal/cli 174.571s
ok  polymetrics.ai/internal/connectors/certify 348.926s
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.U8utLpbB3Q
0 issues.
connectorgen validate: 547 connector(s) checked, 0 findings
```

Targeted parity:

```bash
./pm --help
./pm help config
./pm etl --help
./pm flow --help
./pm etl
./pm flow
./pm --root "$tmpdir" flow bogus
rg -n -- '--plain|--no-input|--progress ndjson' docs/cli website/content/docs/cli-reference.mdx
```

Result: pass. `flow bogus` exit `2` with `error: flow: unknown subcommand "bogus"`.

## Review disposition ledger

No automated review findings yet. Stacked PR pending; record Claude/Copilot/human route status after PR creation.

## Review-fix cycle — PR #457 head `3702318efa5514b8fad20c99bba2e3281164bec7`

### Additional loaded skills / references

- `golang-lint` — review-fix gate includes `go vet ./...` and static-quality checks.
- `golang-spf13-viper` — config/env layering reviewed for CLI env capture implications.
- `vercel-react-best-practices`, `vercel-composition-patterns` — website docs/generated data checked; no React component code in scope.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` — website architecture doc touchpoint loaded.

### Accepted findings mapped to red tests

| Finding | Planned red test / validation | Red evidence | Green evidence |
|---|---|---|---|
| TTY env semantics | `go test ./internal/ui/... -run TestDetectModeUsesADRGate -count=1` with `CI=0/false`, `PM_NO_TUI=0/false` | fail: `PM_NO_TUI=0`, `PM_NO_TUI=false`, `CI=0`, `CI=false` returned `Mode="tui"`, want `plain` | pass: focused `go test ./internal/ui/...` review gate included ADR env cases; `internal/ui 0.172s`, `internal/ui/styles 0.317s` |
| Color degradation | `go test ./internal/ui/... -run TestDetectCapabilitiesDegradeForColorAndASCII -count=1`; `go test ./internal/cli/... -run TestInvocationEnvCapturesColorControls -count=1` | fail: `NO_COLOR=0`, `NO_COLOR=false`, `CLICOLOR=0` returned `Color=true`; `invocationEnv` missing `CLICOLOR` | pass: focused UI and CLI review gates passed; CLI gate `internal/cli 6.686s` |
| ANSI16 dim SGR | `go test ./internal/ui/styles -run TestPaletteANSI16DimUsesBrightBlackSGR -count=1` | fail: `ANSI16 dim style = "\x1b[38mdim\x1b[0m", want bright-black SGR 90` | pass: focused UI/styles review gate included `TokenDim` SGR 90 case |
| Human terminal controls | `go test ./internal/cli/... -run 'TestFlow.*Sanitizes.*Human' -count=1` | fail: human flow plan/list output contains unsafe terminal rune `U+001B` from step ID / filename payloads | pass: focused CLI review gate sanitized human plan/list output and preserved raw JSON output |
| Docs/runtime wording + mixed stderr | `go test ./internal/cli/... -run TestGlobalUIFlagsDocumentedInHelp -count=1`; docs/website grep | fail: root help missing `Future TTY renderers`; config help missing `invalid UI/progress flag`; etl/flow help missing `stderr may also include the final error diagnostic` | pass: focused CLI review gate passed; docs/website grep found `--progress ndjson`, future TTY wording, mixed-stderr diagnostics, and exit-code wording |
| Website/docs parity | `cd website && pnpm run gen:docs`; generated diff check | pending until docs edits | pass: `cd website && pnpm run gen:docs` wrote 11 docs pages to `website/lib/docs.generated.ts`; focused CLI golden docs/transcripts passed |

### Red test capture rule for review fixes

Do not edit production Go/docs for these findings until the focused red tests above fail and exact output is captured here. Preserve JSON output semantics for flow outputs; sanitize only human stdout fields unless a test proves JSON changed unexpectedly.

### Review-fix red evidence captured — 2026-07-17

```bash
go test ./internal/ui/... -run 'TestDetectModeUsesADRGate|TestDetectCapabilitiesDegradeForColorAndASCII|TestPaletteANSI16DimUsesBrightBlackSGR' -count=1
```

Result: fail. Key output:

```text
--- FAIL: TestDetectModeUsesADRGate/pm_no_tui_zero_still_forces_plain
    Detect(... Env:map[PM_NO_TUI:0 TERM:xterm-256color]).Mode = "tui", want "plain" (reasons=[])
--- FAIL: TestDetectModeUsesADRGate/pm_no_tui_false_still_forces_plain
    Detect(... Env:map[PM_NO_TUI:false TERM:xterm-256color]).Mode = "tui", want "plain" (reasons=[])
--- FAIL: TestDetectModeUsesADRGate/ci_zero_still_forces_plain
    Detect(... Env:map[CI:0 TERM:xterm-256color]).Mode = "tui", want "plain" (reasons=[])
--- FAIL: TestDetectModeUsesADRGate/ci_false_still_forces_plain
    Detect(... Env:map[CI:false TERM:xterm-256color]).Mode = "tui", want "plain" (reasons=[])
--- FAIL: TestDetectCapabilitiesDegradeForColorAndASCII/no_color_zero_disables_color
    Color/ASCII = true/false, want false/false
--- FAIL: TestDetectCapabilitiesDegradeForColorAndASCII/no_color_false_disables_color
    Color/ASCII = true/false, want false/false
--- FAIL: TestDetectCapabilitiesDegradeForColorAndASCII/clicolor_zero_disables_color
    Color/ASCII = true/false, want false/false
--- FAIL: TestPaletteANSI16DimUsesBrightBlackSGR
    ANSI16 dim style = "\x1b[38mdim\x1b[0m", want bright-black SGR 90
FAIL
```

```bash
go test ./internal/cli/... -run 'TestInvocationEnvCapturesColorControls|TestFlow.*Sanitizes.*Human|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1
```

Result: fail. Key output:

```text
--- FAIL: TestFlowPlanSanitizesUnsafeStepIDsInHumanOutput
    human output contains unsafe terminal rune U+001B in "Flow: unsafe-flow  status=ok\n  1. score\x1b]0;owned\a\u202edone\n"
--- FAIL: TestFlowListSanitizesUnsafeFilenamesInHumanOutput
    human output contains unsafe terminal rune U+001B in "nightly\x1b]2;owned\a\u202eflow\n"
--- FAIL: TestInvocationEnvCapturesColorControls
    invocationEnv missing CLICOLOR in map[string]string{"CI":"", "NO_COLOR":"0", "PM_ASCII":"", "PM_NO_TUI":"", "TERM":"xterm-256color"}
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/root_help
    help output missing "Future TTY renderers"
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/config_help
    help output missing "invalid UI/progress flag"
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/etl_help
    help output missing "stderr may also include the final error diagnostic"
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/flow_help
    help output missing "stderr may also include the final error diagnostic"
FAIL
```

### Review-fix green evidence captured — 2026-07-17

```bash
go test ./internal/ui/... -run 'TestDetectModeUsesADRGate|TestDetectCapabilitiesDegradeForColorAndASCII|TestPaletteANSI16DimUsesBrightBlackSGR' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/ui	0.172s
ok  	polymetrics.ai/internal/ui/styles	0.317s
```

```bash
go test ./internal/cli/... -run 'TestInvocationEnvCapturesColorControls|TestFlow.*Sanitizes.*Human|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/cli	6.686s
```

```bash
go test ./internal/cli/... -count=1
```

Result: pass.

```text
ok  	polymetrics.ai/internal/cli	169.138s
```

```bash
cd website && pnpm run gen:docs
```

Result: pass.

```text
Wrote 11 docs pages to lib/docs.generated.ts.
```

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Result: pass. Key output:

```text
ok  	polymetrics.ai/internal/cli	170.511s
ok  	polymetrics.ai/internal/connectors/certify	340.438s
make verify
ok  	polymetrics.ai/internal/cli	171.287s
ok  	polymetrics.ai/internal/connectors/certify	342.514s
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.CnpIQlCCHU
0 issues.
connectorgen validate: 547 connector(s) checked, 0 findings
```

## Review-fix cycle #2 — PR #457 head `2195a66659be9d62bf99bfc8e2506e77da81e02f`

Loaded skills unchanged for this docs/help fix: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-lint`, `golang-spf13-cobra`. Stack implementation skill note remains: `.pi/skills/go-implementation/SKILL.md` missing (`ENOENT`), so repo/global Go routing skills are loaded instead.

Planned red validation before production docs/help edits:

- `rg -n "CLICOLOR_FORCE" docs/design/tui-ux-design.md` must show stale docs claims.
- Update `TestGlobalUIFlagsDocumentedInHelp` expectation first, then run `go test ./internal/cli/... -run 'TestGlobalUIFlagsDocumentedInHelp' -count=1`; expected red: root/ETL/flow help missing exit `3` invalid UI/progress flag wording.

Review-fix #2 red evidence captured — 2026-07-17:

```bash
rg -n "CLICOLOR_FORCE" docs/design/tui-ux-design.md
```

Result: fail/stale docs validation. Matches remain at `docs/design/tui-ux-design.md:63` and `:376`.

```bash
go test ./internal/cli/... -run 'TestGlobalUIFlagsDocumentedInHelp' -count=1
```

Result: fail. Key output:

```text
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/root_help
    help output missing "3 validation error"
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/etl_help
    help output missing "3 validation error"
--- FAIL: TestGlobalUIFlagsDocumentedInHelp/flow_help
    help output missing "3 validation error"
FAIL	polymetrics.ai/internal/cli	0.538s
```

Review-fix #2 green evidence captured — 2026-07-17:

```bash
tmp_connectors=$(mktemp -d); go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$tmp_connectors"; rm -rf "$tmp_connectors"
```

Result: pass.

```text
Generated docs in docs/cli and connector docs in /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.fztdKsog5X
```

```bash
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1
```

Result: pass; `ok polymetrics.ai/internal/cli 9.869s`.

```bash
go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1
```

Result: pass; `ok polymetrics.ai/internal/cli 6.724s`.

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Result: pass. Key output:

```text
go test ./...: ok polymetrics.ai/internal/cli 170.546s; ok polymetrics.ai/internal/connectors/certify 339.739s
make verify: ok polymetrics.ai/internal/cli 171.209s; ok polymetrics.ai/internal/connectors/certify 342.470s
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.JYQwR5QWMh
0 issues.
connectorgen validate: 547 connector(s) checked, 0 findings
```

## Review-fix cycle #3 — final docs-writer P2 at PR #457 head `1c1ae22dbeb333fe11abb34029e896e0523ee723`

Loaded skills/references for this docs-only website fix: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-documentation`, `vercel-react-best-practices`, `vercel-composition-patterns`, `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`, and `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`. Stack implementation skill note: `.pi/skills/ts-website/SKILL.md`, `.pi/skills/go-implementation/SKILL.md`, and `.pi/skills/design-ui/SKILL.md` are absent in this checkout; loaded available routing skills instead.

Planned red validation before production docs edits:

- `rg -n -- '--website-dir|embedded help|checked-in CLI markdown|website MDX|pm docs validate' website/content/docs/cli-reference.mdx website/content/docs/architecture.mdx` must show stale website claims.
- `./pm docs` must show the runtime synopsis only supports `pm docs validate [--connectors-dir <path>]`.

Review-fix #3 red validation captured — 2026-07-17:

```bash
rg -n -- '--website-dir|embedded help|checked-in CLI markdown|website MDX|pm docs validate' website/content/docs/cli-reference.mdx website/content/docs/architecture.mdx
```

Result: fail/stale docs validation. Matches include unsupported `--website-dir`, embedded-help / checked-in CLI markdown / website MDX validation claims, and `pm docs validate` release-gate overclaim in `website/content/docs/cli-reference.mdx` and `website/content/docs/architecture.mdx`.

```bash
./pm docs
```

Result: runtime truth captured. Key output:

```text
SYNOPSIS
  pm docs generate --dir <path>
  pm docs validate [--connectors-dir <path>]

DESCRIPTION
  Writes embedded command documentation as markdown files. Generation also
  writes connector manuals under a connector docs directory. By default, when
  --dir is docs/cli, connector docs are written to docs/connectors.

  Validation checks every registered connector has a generated MANUAL.md with
  required human and agent workflow sections. This is intended for CI hooks and
  local preflight checks before adding or changing connectors.
```

Review-fix #3 green evidence captured — 2026-07-17:

```bash
cd website && pnpm run gen:docs
```

Result: pass.

```text
Wrote 11 docs pages to lib/docs.generated.ts.
```

```bash
rg -n -- '--website-dir' website/content/docs website/lib/docs.generated.ts
```

Result: pass/no matches (command exits 1 because `rg` found no matches).

```bash
go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1
```

Result: pass; `ok polymetrics.ai/internal/cli 6.642s`.

```bash
make verify
```

Result: pass. Key output:

```text
./pm docs validate --connectors-dir docs/connectors
Validated connector docs in docs/connectors
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.zTZPadGuaI
0 issues.
connectorgen validate: 547 connector(s) checked, 0 findings
```

```bash
cd website && pnpm run typecheck
```

Result: blocked local environment, not product failure: `sh: tsc: command not found`; `node_modules` missing.
