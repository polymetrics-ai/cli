# SUMMARY — Issue 405 TTY gate and NDJSON progress

Status: implementation and local verification complete; stacked PR pending.

## Delivered

- Deterministic `internal/ui` TTY gate per ADR 0003.
- `cli.RunWithOptions`; existing `Run` delegates with plain mode.
- Global `--plain`, `--no-input`, and `--progress ndjson` parsing.
- `events.NDJSON` progress wired to stderr only.
- `internal/ui/styles` palette/glyph foundation with no-color and ASCII fallback.
- Runtime help, `docs/cli/**`, website CLI reference, and golden help fixtures updated.
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

## Pending

- Commit/push final artifact update.
- Open stacked PR to `feat/cli-architecture-v2` with `Refs #405` and `Refs #397`.
- Automated review coverage pending after PR creation.
