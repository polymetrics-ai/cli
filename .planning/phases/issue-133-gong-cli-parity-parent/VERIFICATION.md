# Verification — Gong CLI Parity Parent (#133)

## Preflight

| Command | Result |
|---|---|
| `scripts/gsd doctor` | pass |
| `scripts/gsd verify-pi` | pass |
| `scripts/gsd list --json` | pass |
| `scripts/gsd prompt plan-phase 133 --skip-research --tdd` | pass |
| `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run` | fail: adapter lacks `programming-loop`; manual-GSD fallback recorded |

## Required local gates before parent handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## #144 targeted gates

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1
```

Result: pass for all four commands on 2026-07-09.

## Full gate attempt

| Command | Result |
|---|---|
| `gofmt -w cmd internal` | pass |
| `go vet ./...` | pass |
| `go test ./...` | fail: `internal/connectors/certify` exceeded the default 10m package timeout in existing certify tests |
| `go test -timeout 20m ./...` | pass |
| `go build ./cmd/pm` | pass |
| `make verify` | pass |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | pass (also included in `make verify`) |

## CLI help/docs/website parity checklist

Applies to later CLI-visible lanes (#141-#143, #145-#147). #144 is connector metadata/docs only unless CLI-rendered metadata changes.

- [ ] `pm help <topic>` checked when CLI topics change.
- [ ] Bare namespace behavior checked when command groups change.
- [ ] `pm <command> --help` checked when command help changes.
- [ ] `docs/cli/**` updated or exemption recorded.
- [ ] `website/**` updated or exemption recorded.
- [ ] Generated help/manual artifacts updated or exemption recorded.

## Safety verification

- No secrets used.
- No credentialed Gong requests used.
- No reverse ETL execution.
- No destructive/admin external actions.
- No new dependencies.


## 2026-07-10 integrated lane verification

- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/commandrunner -count=1`
- [x] `go test ./cmd/connectorgen -run Gong -count=1`
- [x] `go test ./cmd/connectorgen -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`

## 2026-07-10 broader gates

- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `go test -timeout 20m ./...`
- [x] `make verify`
- [x] `./pm help connectors`
- [x] `./pm connectors inspect gong`
- [x] `./pm connectors inspect gong --json`
