# Verification — Front parity resume

Status: **connector-local review remediation passed; shared composition integration pending**.
The current worktree passed JSON parsing, Front validation, surface synchronization, focused Front
conformance, and implemented-command runtime preflight after the conditional-contract changes.

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

Composition note: exact Front request rules are preserved in the connector-owned draft-07
`schemas/request_contracts.json`, and fixed typed CLI variants construct provider-valid requests.
Current main does not evaluate composed keywords, so full generic record validation depends on the
in-flight shared normalized-schema composition lane. No shared validation/runtime file is changed
by this remediation.

Review-remediation evidence: `connectorgen validate` reported zero Front findings,
`surface-sync --check` reported zero drift, Front conformance passed in 1.661s, and the real
implemented-command runtime preflight passed in 1.362s. JSON parsing covered the Front API surface,
operations, writes, CLI surface, draft-07 request contracts, and run state.

Provider-schema correction evidence: `git diff --check` and JSON parsing passed; focused assertions
confirmed the removed template minimum, non-null parent call ID, custom-channel Twilio-secret
prohibition, empty-array fixture, and one-mapping CLI help. Front validation reported zero findings,
surface synchronization reported zero drift, and focused Front conformance passed in 1.626s.
