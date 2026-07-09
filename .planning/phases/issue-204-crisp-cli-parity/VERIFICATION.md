# Verification — Issue #204 Crisp CLI parity parent

## Completed preflight

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
```

Result: pass. Evidence saved in Pi bash log `/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/pi-bash-413bdf7d46103c1d.log`.

```bash
scripts/gsd prompt plan-phase 204 --skip-research
```

Result: pass; generated prompt saved to `/tmp/gsd-plan-phase-204.txt`.

```bash
scripts/gsd prompt programming-loop init --phase issue-204-crisp-cli-parity --dry-run
```

Result: blocked; adapter returned `unknown GSD command: programming-loop`. Manual GSD/TDD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

## Planned local gates

Issue #205 targeted gates:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
go run ./cmd/connectorgen validate internal/connectors/defs
```

Parent integrated gates before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Runtime-backed integration is optional and intentionally not part of the non-credentialed Crisp parity gates unless requested:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```

## CLI help/docs/website parity gates

Applicable once Crisp exposes executable CLI command/help surfaces:

```bash
pm help connectors
pm connectors
pm connectors inspect crisp --json
pm connectors inspect crisp --help
rg -n "crisp|Crisp" docs/cli website
```

Current #205 metadata-only scaffold exemption: runtime provider-specific `pm crisp ...` commands are not implemented in #205; `cli_surface.json` is docs/help metadata only and must not dispatch writes.

## Current status

- Full parent verification: pending.
- Targeted #205 validation: pending.
- Automated review route: pending parent PR creation.
