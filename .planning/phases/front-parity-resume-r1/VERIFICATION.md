# Verification — Front parity resume

Status: **passed locally** on commit `a2b7de01a`. Front-specific current-main gates, the full
focused CLI suite, generator checks, and built-CLI help/inspect smoke checks passed. The first full
CLI run exposed only stale root golden transcripts; deterministic regeneration fixed them and the
full rerun passed.

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

Final local evidence:

- `surface-sync --check`: 550 scanned, zero changes.
- Front validation: one connector checked, zero findings.
- Focused Front conformance and all-connector runtime preflight: pass.
- `pm docs generate` plus `pm docs validate`: pass; only Front manual/skill output retained.
- Website generator: pass; Front catalog/command-surface data regenerated.
- Built CLI: `pm front` and `pm front binary download-attachment --help` render the implemented
  command and runtime-owned `--dest-root`, `--file-name`, and `--max-bytes` flags.
- An isolated temporary project let `pm front binary download-attachment` reach the expected
  missing-credential boundary without a provider request. No credentials or live provider calls
  were used.
- Full focused CLI suite: `ok polymetrics.ai/internal/cli 418.708s`.
- `go vet` for `internal/connectors/conformance`, `internal/connectors/commandrunner`, and
  `internal/cli`, plus `go build ./cmd/pm`: pass.

Citation-convention note: a fresh `git fetch origin` found no landed shared machine-readable
field-citation convention on `origin/main`. The phase therefore retains the contract-approved
research matrix and does not invent a competing bundle schema.
