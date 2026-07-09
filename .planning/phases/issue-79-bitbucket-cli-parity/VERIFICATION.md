# Verification: Bitbucket CLI Parity Parent

Date: 2026-07-09

## Preflight

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd
scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run
```

## Results

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass; 69 registered commands.
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd`: pass; prompt generated.
- `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run`: blocked; `programming-loop` command unavailable. Manual GSD fallback recorded in `PLAN.md`, `TDD-LEDGER.md`, and `RUN-STATE.json`.

## Required local gates before handoff

Run after implementation slices, not yet complete:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity checklist

- [ ] Runtime help checked or exemption recorded.
- [ ] Bare namespace behavior checked or exemption recorded.
- [ ] `pm <command> --help` checked or exemption recorded.
- [ ] `docs/cli/**` updated or exemption recorded.
- [ ] `website/**` updated/regenerated or exemption recorded.
- [ ] Generated help/manual artifacts updated or exemption recorded.

Current parent phase status: planning seed only; implementation verification pending #90.
