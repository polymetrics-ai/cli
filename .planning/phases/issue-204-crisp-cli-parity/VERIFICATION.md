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

## Completed targeted #205 gates

Red validation before scaffold:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
```

Result: failed as expected because the Crisp bundle path did not exist yet.

Targeted single-connector validation after scaffold (temp root required because connectorgen validates connector directories under a root):

```bash
tmp=$(mktemp -d); cp -R internal/connectors/defs/crisp "$tmp/crisp"; go run ./cmd/connectorgen validate "$tmp"
```

Result: `connectorgen validate: 1 connector(s) checked, 0 findings`.

Fleet validation:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: `connectorgen validate: 548 connector(s) checked, 0 findings`.

Conformance smoke:

```bash
go test ./internal/connectors/conformance -run 'TestConformance/crisp'
```

Result: pass.

Connector help/inspect/docs smoke:

```bash
go run ./cmd/pm help connectors
tmp=$(mktemp); go run ./cmd/pm connectors inspect crisp --json > "$tmp" && python3 -c 'import json,sys; obj=json.load(open(sys.argv[1])); print(obj["connector"]["name"])' "$tmp"; rm -f "$tmp"
./pm docs validate --connectors-dir docs/connectors
rg -n "crisp|Crisp" docs/cli docs/connectors
```

Result: pass; inspect output `crisp`, docs validate passed, docs grep finds Crisp catalog/manual entries.

Full local gate on #205 branch:

```bash
make verify
```

Result: pass; completed fmt, tidy-check, vet, test, build, docs validate, smoke, lint, and connectorgen validate.

## Planned local gates

Parent integrated gates before final handoff after later slices:

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

- Full parent verification for the current #205 branch: passed via `make verify`.
- Targeted #205 validation: passed.
- Automated review route: pending subissue PR creation/review.
