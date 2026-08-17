# Verification — CLI required-flag derivation r1

## Checklist

- [x] Required-path invariant enumerates every bundle and returns zero violations.
- [x] Focused derivation test covers required path and optional query parameters.
- [x] Command-runner test asserts typed usage error before provider I/O.
- [x] GitHub sweep falls from 92 to zero; cross-connector before/after count is recorded: 104 fields in 92 GitHub commands, zero violations in every other connector.
- [x] All 50 unsupported declarations are listed with verification result; no silent reclassification. The audit reports 26 source-supported `unsupported_api` contradictions, one retained `unsupported_api`, and 23 valid `unsupported_local` entries.
- [x] Generated connector surfaces are regenerated twice and the second run is byte-stable.
- [x] `go test -timeout 20m ./cmd/connectorgen` passes.
- [x] Connector validate/surface-sync/boundary/runtime-preflight gates pass.
- [x] Website docs generator is byte-stable and relevant runtime help is accurate.
- [x] Gofmt, vet, changed packages plus consumers, lint/docs/smoke/agent-contract/release checks pass or an exact base-branch blocker is recorded.
- [x] Review completed with findings/dispositions recorded.

## Commands and results

All commands below exited 0.

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen` | Pass; includes the repository-wide required REST path flag invariant. |
| `go test -timeout 20m ./internal/connectors/commandrunner` | Pass; confirms the typed missing-required-flag error is returned before executor I/O. |
| `go test -timeout 20m ./internal/cli` | Pass; the fake GitHub transport sees zero calls and the JSON failure is `category=usage`, `code=usage_error`, exit 2. |
| `go vet ./...` | Pass. |
| `go build ./cmd/pm` | Pass. |
| `make tidy-check` | Pass. |
| `make lint` | Pass; `golangci-lint` found 0 issues. |
| `make docs-check` | Pass. |
| `make smoke-no-build` | Pass. |
| `make agent-contract-check` | Pass. |
| `make connectorgen-validate` | Pass; 552 connectors, 0 findings. |
| `make connectorgen-surface-sync` | Pass; 552 connectors, 0 fields to fill/correct. |
| `make connector-runtime-preflight` | Pass. |
| `make connector-canon-check` | Pass. |
| `make connector-boundary` | Pass; no allowlist change. |
| `make release-workflow-check` | Pass. |
| `go run ./cmd/connectorgen certification-sweep --connector github --check` | Pass; 1,571 commands current. |
| `jq '.product_defects | length' internal/connectors/defs/github/certification-sweep.json` | `0`; the base sweep had `92`. |
| `go run ./cmd/connectorgen surface-sync` ×2 | Both passes scanned 552 connectors and reported 0 filled/corrected fields. |
| `go run ./cmd/connectorgen certification-sweep --connector github` ×2 | Both passes rewrote the same 1,571-command sweep. |
| `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` ×2 | Both passes succeeded. |
| `pnpm --dir website run gen:docs` ×2 | Both passes wrote 12 pages. |
| `git diff --exit-code` | Pass after the second complete generator pass; byte-stable generated files, including `website/**`. |
| `go run ./cmd/pm github pulls files view --help` | Pass; shows required `--pull-number`. |
| `go run ./cmd/pm github releases assets view --help` | Pass; shows required `--asset-id`. |
| `go run ./cmd/pm help github` and `go run ./cmd/pm github` | Both pass; connector manual and bare namespace help render successfully. |

## Scope note

The repository instructs workers with per-command time limits not to run the
whole `go test ./...`/`make verify` aggregate because the 552-connector suite
can be cut off and look like a hang. The changed packages, their CLI consumer,
and `cmd/connectorgen` (the generated-artifact consumer) were all run above;
CI carries the full aggregate suite.
