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

Current parent phase status: #90 verified green slice complete; #91-#96 completed inline as one local-critical-path implementation slice pending commit/push/review checkpoint.

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

## #91-#96 final evidence

```bash
jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json
go test ./cmd/connectorgen -run Bitbucket -count=1
go test ./internal/cli -run Bitbucket -count=1
go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
./pm help bitbucket
./pm bitbucket
./pm bitbucket --help
npm --prefix website run gen:website-data
```

Results: passed. `make verify` first exceeded the 900s tool timeout while `internal/connectors/certify` was still running; the rerun with a longer timeout passed. No credentialed Bitbucket checks were run.

CLI/help/docs/website parity:

- [x] `pm help bitbucket` renders Bitbucket connector manual/command surface.
- [x] bare `pm bitbucket` renders contextual help and exits successfully.
- [x] `pm bitbucket --help` renders contextual help and exits successfully.
- [x] Runtime stream command path covered by `pm bitbucket issue list` fixture-backed CLI test.
- [x] Runtime direct-read command path covered by `pm bitbucket repo view` fixture-backed CLI test with clone-link redaction.
- [x] Runtime write-plan command path covered by `pm bitbucket issue create` CLI test with approval-gated JSON and no approval token/raw record leakage.
- [x] `docs/connectors/bitbucket/**`, connector catalog docs, and website generated connector data updated.
- [x] Website CLI reference notes dispatched Bitbucket connector command surface and blocked raw/local workflows.

## Full-surface follow-up evidence

After the user requested full Bitbucket operation implementation, a follow-up GSD phase was added at `.planning/phases/issue-79-bitbucket-full-surface/`.

Results:

- 331/331 official Bitbucket Swagger operations covered by typed surfaces.
- 179/179 GET operations covered by direct-read commands.
- 152/152 POST/PUT/DELETE operations covered by named reverse-ETL write actions.
- 0 blocked `api_surface.operation` rows remain.
- 342 implemented Bitbucket CLI commands are declared.
- Bounded binary/text GET policy added as `bitbucket_binary_base64` with no filesystem writes.
- Full local gates passed, including `make verify` and `go run ./cmd/connectorgen validate internal/connectors/defs`.
