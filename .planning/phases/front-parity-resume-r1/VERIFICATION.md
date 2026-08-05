# Verification — Front parity resume

Status: **in progress**. Front-specific current-main gates and generated-doc checks pass. The first
full CLI run exposed only stale root golden transcripts; their deterministic update passed the
focused golden test. Re-run the full focused CLI suite and the final generator/validation sequence
before setting `verificationPassed`.

Required gates:

- `go run ./cmd/connectorgen surface-sync`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/connectorgen validate internal/connectors/defs/front`
- `go test ./internal/connectors/conformance/... -run 'TestConformance/front'`
- `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`
- `go test ./internal/cli/...`
- focused `go vet` / `go build ./cmd/pm`
- `cd website && pnpm run gen:website-data`
- built `pm front --help`, representative Front commands, and `pm connectors inspect front --json`

Do not substitute a historical green run for these gates.

Current evidence:

- `surface-sync --check`: 550 scanned, zero changes.
- Front validation: one connector checked, zero findings.
- Focused Front conformance and all-connector runtime preflight: pass.
- `pm docs generate` plus `pm docs validate`: pass; only Front manual/skill output retained.
- Website generator: pass; Front catalog/command-surface data regenerated.
- Built CLI: `pm front` and `pm front binary download-attachment --help` render the implemented
  command and runtime-owned `--dest-root`, `--file-name`, and `--max-bytes` flags.
- A credential-free execution cannot reach Front in this disposable worktree because it has no
  `.polymetrics` project; no credentials or live provider calls were used.
