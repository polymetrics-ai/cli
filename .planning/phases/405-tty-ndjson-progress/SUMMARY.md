# SUMMARY — Issue 405 TTY gate and NDJSON progress

Status: review-fix cycle for PR #457 is locally verified; PR body updated and branch push remains.

## Delivered

- Deterministic `internal/ui` TTY gate per ADR 0003.
- `cli.RunWithOptions`; existing `Run` delegates with plain mode.
- Global `--plain`, `--no-input`, and `--progress ndjson` parsing.
- `events.NDJSON` progress wired to stderr only.
- `internal/ui/styles` palette/glyph foundation with no-color and ASCII fallback.
- Runtime help, `docs/cli/**`, website CLI reference, generated website docs data, and golden help fixtures updated.
- Stage-approved dependency: `golang.org/x/term v0.42.0`.

## TDD evidence

- Red tests captured for missing UI detection/styles API and missing `RunWithOptions` API.
- Docs parity red captured via generated docs drift before docs update.
- Green focused tests passed for `internal/ui`, `internal/cli` run-options/progress/help, and full `internal/cli` package.

## Verification

Final combined gate passed:

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Key output:

```text
ok  polymetrics.ai/internal/cli 173.250s
ok  polymetrics.ai/internal/connectors/certify 346.045s
make verify
ok  polymetrics.ai/internal/cli 174.571s
ok  polymetrics.ai/internal/connectors/certify 348.926s
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.U8utLpbB3Q
0 issues.
connectorgen validate: 547 connector(s) checked, 0 findings
```

CLI parity passed: `pm --help`, `pm help config`, `pm etl --help`, `pm flow --help`, `pm etl`, `pm flow`, initialized-root `pm flow bogus` exit 2, docs/website grep for new flags.

Remote Website checks initially failed because `website/lib/docs.generated.ts` needed regeneration. Ran `cd website && pnpm run gen:website-data` and committed the generated docs data. Local `pnpm install --frozen-lockfile` is blocked by `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`; CI install succeeded, so remote rerun is source of truth for website typecheck/tests.

## Review-fix delivered

- `PM_NO_TUI` and `CI` now suppress TUI on any non-empty value, including `0` and `false`.
- `NO_COLOR` now suppresses color on any non-empty value; `CLICOLOR=0` is captured and suppresses color; `TERM=dumb` remains no-color/ASCII.
- ANSI16 `TokenDim` now maps color index 8 to bright-black SGR `90` instead of invalid `38`.
- Human `pm flow` outputs sanitize terminal controls in flow names, step IDs, statuses, and listed filenames while JSON output remains raw.
- Runtime help, `docs/cli/**`, website ETL/architecture/CLI reference, golden transcripts, and `website/lib/docs.generated.ts` updated for future TTY wording, mixed stderr diagnostics, and exit-code 3 wording.

## Review-fix verification

- Focused UI review gate passed: `internal/ui 0.172s`; `internal/ui/styles 0.317s`.
- Focused CLI review gate passed: `internal/cli 6.686s`.
- Full CLI package passed: `internal/cli 169.138s`.
- Final combined gate passed: `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify`.
- Key full-gate output: `go test ./...` included `internal/cli 170.511s`, `internal/connectors/certify 340.438s`; `make verify` included `internal/cli 171.287s`, `internal/connectors/certify 342.514s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`.

## Pending

- Push review-fix slice to `origin feat/405-tty-ndjson-progress`.
- Confirm automated review coverage after fix commit; do not merge parent PR #438 to `main`.
