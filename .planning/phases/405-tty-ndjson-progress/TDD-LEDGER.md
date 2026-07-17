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
