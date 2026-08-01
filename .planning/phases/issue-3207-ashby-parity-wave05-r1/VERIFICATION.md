# Verification Checklist — Ashby parity wave05-r1

## Results

- [x] Official Ashby inventory generated from the public ReadMe/OpenAPI page without credentials.
  - Source counts: REST operations=185, OpenAPI webhook events=27, total=212.
  - Implemented counts: streams/changefeed=72, bounded direct reads/search/file metadata=7, reverse-ETL write actions=101.
  - Blocked ledger rows=32 (27 webhooks, 4 inbound assessment-partner endpoints, 1 presigned external upload-handle workflow).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  - `connectorgen validate: 550 connector(s) checked, 0 findings`.
  - Note: `connectorgen validate internal/connectors/defs/ashby` treats `ashby/` as a defs root and tries to validate `schemas/` and `fixtures/` as connector directories; the usable command for this repo-local tool is the full defs root.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/ashby' -count=1`
  - `ok   polymetrics.ai/internal/connectors/conformance`.
- [x] Ashby native/hook tests
  - `go test ./internal/connectors/native/ashby ./internal/connectors/hooks/ashby -count=1` passed.
- [x] CLI help/docs/website parity
  - Read `pm help docs` before docs generation.
  - `go run ./cmd/pm docs generate --dir docs/cli` passed; Ashby connector docs restored as the only connector manual changes.
  - `npm run gen:website-data` passed under `website/`.
  - `./pm docs validate --connectors-dir docs/connectors` passed.
  - `./pm help connectors`, `./pm ashby`, `./pm ashby candidate --help`, `./pm ashby candidate search --help`, `./pm connectors inspect ashby --json`, and docs/website grep smoke passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout 10m`
  - `ok   polymetrics.ai/internal/cli 271.948s`.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` passed.
- [x] `go build ./cmd/pm` passed.
- [x] `make connector-boundary` passed with clean outcome and existing documented exceptions only.
- [x] `git diff --check` passed.
- [x] `make verify` passed in this worktree.
  - Verified cwd: `/Users/karthiksivadas/.treehouse/cli-83d592/50/cli`.
  - Final line: `MAKE_VERIFY_EXIT:0`.
  - After final docs-count and `--message-body` flag-safety regenerations, `go run ./cmd/connectorgen validate internal/connectors/defs`, `go test ./internal/cli -run 'TestGolden' -count=1`, `go build ./cmd/pm`, `git diff --check`, `./pm docs validate --connectors-dir docs/connectors`, and Ashby command-help smoke checks still passed.
- [ ] One gh-axi final update across parent/subissues with counts and evidence.
- [ ] Clean commit only; no push, no PR, no no-mistakes final pipeline.
