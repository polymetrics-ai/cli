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

#90 green slice completed these required local gates:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity checklist

- [x] Runtime help checked or exemption recorded (`./pm help connectors`; no `pm bitbucket` dispatcher in #90).
- [x] Bare namespace behavior checked or exemption recorded (#90 metadata-only; no new namespace command).
- [x] `pm <command> --help` checked or exemption recorded (#90 metadata-only; #91 owns renderer).
- [x] `docs/cli/**` updated or exemption recorded (no command behavior changed).
- [x] `website/**` updated/regenerated (`cd website && pnpm run gen:website-data` twice).
- [x] Generated help/manual artifacts updated or exemption recorded (`docs/connectors/bitbucket/**` and catalog docs added; `./pm docs validate` passed).

Current parent phase status: #90 verified green slice complete; remaining lanes #91-#96 pending.

## #90 focused evidence

```bash
jq . internal/connectors/defs/bitbucket/*.json
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -count=1
go build ./cmd/pm
./pm help connectors
./pm connectors inspect bitbucket --json
cd website && pnpm run gen:website-data
cd website && pnpm run gen:website-data
git diff --check
go test ./internal/cli ./internal/connectors/bundleregistry -count=1
go vet ./...
go test ./...
./pm docs validate --connectors-dir docs/connectors
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: passed; connector validation reported `connectors_checked=548`, `findings=0`, `warnings=0`; `make verify` passed. Commit/push checkpoint: `0e359d76` pushed to `feat/79-bitbucket-cli-parity`.

CodeRabbit actionable finding disposition: accepted and fixed by tightening the #90 test so `direct_write` commands must be `unsafe_or_disallowed`. Post-review-fix `go test ./...` and `make verify` passed.
