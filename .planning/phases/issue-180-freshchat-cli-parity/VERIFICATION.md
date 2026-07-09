# Verification — Issue #180 Freshchat CLI parity parent

## Preflight already run

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd prompt plan-phase issue-180-freshchat-cli-parity --skip-research
scripts/gsd prompt programming-loop init --phase issue-180-freshchat-cli-parity --dry-run
```

Results:

- `doctor`: pass.
- `verify-pi`: pass.
- `list --json`: pass.
- `plan-phase`: generated prompt followed.
- `programming-loop`: blocked/unavailable (`unknown GSD command: programming-loop`); manual fallback recorded.

## Parent checkpoint verification

Before opening parent PR:

```bash
git status --short --branch
git diff -- .planning/phases/issue-180-freshchat-cli-parity
```

## Required local verification before final parent handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Focused verification for issue #181

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface'
```

## CLI help/docs/website parity verification

For #181, runtime metadata parsing is in scope; help renderer/docs parity is deferred to #182 unless existing generic connector command help changes as a side effect.

Planned no-credential checks after #181 lands:

```bash
go build ./cmd/pm
./pm help connectors
./pm connectors --help
./pm freshchat --help || true
rg -n "freshchat|Freshchat" docs/cli website internal/connectors/defs/freshchat
```

Do not run credentialed `pm freshchat ...` commands or reverse ETL execution.
