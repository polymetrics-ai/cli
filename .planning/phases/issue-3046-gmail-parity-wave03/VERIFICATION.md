# Verification checklist — Gmail parity wave03

Fixture-only local gates completed on branch `fm/cli-gmail-parity-wave03-r1`.

## Required gates

- [x] `no-mistakes doctor` — passed; daemon already running, no lifecycle commands used.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/gmail` — `connectorgen validate: 1 connector(s) checked, 0 findings`.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gmail' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed (`ok polymetrics.ai/internal/cli 479.466s` on final run).
- [x] `go build ./cmd/pm` — passed.
- [x] `make connector-boundary` — passed, outcome clean.
- [x] `make verify` — passed on rerun after a CPU-contention timeout in `internal/connectors/certify`; final run completed all fmt/tidy/vet/test/build/docs/smoke/lint/connectorgen/boundary/release gates.
- [x] `git diff --check` — passed.

## CLI/help/docs parity checks

- [x] `go run ./cmd/pm help docs` and `go run ./cmd/pm help skills` reviewed before generation commands.
- [x] `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` executed; retained Gmail MANUAL/SKILL plus catalog/README outputs to keep the branch scoped.
- [x] `cd website && npm run gen:website-data` executed.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` executed after CLI metadata changes.
- [x] `go run ./cmd/pm help gmail` rendered the Gmail manual.
- [x] `go run ./cmd/pm gmail` rendered the provider-style command surface and exited successfully.
- [x] `go run ./cmd/pm gmail send-as get --help` showed `availability=planned` for an email-path direct-read blocker rather than advertising it as executable.

## Safety checks

- [x] No real credentials, tokens, personal email addresses, message bodies, certificate blobs, or private key material in docs/fixtures/issues.
- [x] Reverse ETL write commands remain plan -> preview -> approval -> execute only.
- [x] No generic API/raw method/path/body command is introduced.
- [x] Issue addendum posted with actual counts to #3046-#3053 via `gh-axi`; live certification count remains 0.
