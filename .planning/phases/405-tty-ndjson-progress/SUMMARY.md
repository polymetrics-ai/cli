# SUMMARY — Issue 405 TTY gate and NDJSON progress

Status: review-fix #3 for PR #457 is locally verified; PR body updated and branch push remains.

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

## Review-fix #2 delivered

- Removed `CLICOLOR_FORCE` support claims from `docs/design/tui-ux-design.md`; design docs now state the implemented color controls: `NO_COLOR`, `CLICOLOR=0`, and `TERM=dumb`.
- Added exit code `3` validation-error wording for invalid UI/progress flags to root, ETL, and flow runtime help.
- Regenerated `docs/cli/etl.md`, `docs/cli/flow.md`, and golden transcripts with existing project commands.

## Review-fix #2 verification

- Red validation captured: `rg -n "CLICOLOR_FORCE" docs/design/tui-ux-design.md` found stale claims; `go test ./internal/cli/... -run 'TestGlobalUIFlagsDocumentedInHelp' -count=1` failed for root/ETL/flow missing `3 validation error` wording.
- Focused gate passed: `go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` (`internal/cli 6.724s`).
- Full gate passed: `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify`.
- Key full-gate output: `go test ./...` included `internal/cli 170.546s`, `internal/connectors/certify 339.739s`; `make verify` included `internal/cli 171.209s`, `internal/connectors/certify 342.470s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`.

## Review-fix #3 delivered

- Corrected website docs overclaims: `pm docs validate` is now described as connector-docs validation through `--connectors-dir`, not embedded-help / `docs/cli/**` / website MDX validation.
- Removed unsupported `--website-dir` from website docs examples.
- Regenerated `website/lib/docs.generated.ts` with `cd website && pnpm run gen:docs`.
- Preserved CLI/website parity wording: tests, generators, website data generation, and `make verify` cover those surfaces where applicable.

## Review-fix #3 verification

- Red validation captured: website docs contained unsupported `--website-dir` and `pm docs validate` overclaim; `./pm docs` showed runtime truth `pm docs validate [--connectors-dir <path>]`.
- Website docs data: `cd website && pnpm run gen:docs` passed (`Wrote 11 docs pages to lib/docs.generated.ts`).
- Focused Go gate passed: `go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` (`internal/cli 6.642s`).
- Full gate passed: `make verify` (`Validated connector docs in docs/connectors`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`).
- Website typecheck blocked locally because `node_modules` is missing: `sh: tsc: command not found`.

## Pending

- Push final docs-fix commit, then confirm automated review coverage after fix commit; do not merge parent PR #438 to `main`.
