# Verification — Google Search Console documented-operation parity resume

Status: PASS — focused current-main verification completed before commit.

## Required gates

- [x] `go run ./cmd/connectorgen surface-sync` — 550 connectors scanned; 0 fields filled or corrected.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 550 connectors scanned; 0 fields filled or corrected.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-search-console` — 1 connector checked; 0 findings.
- [x] `go test ./internal/connectors/hooks/google-search-console` — pass.
- [x] `go test ./internal/connectors/conformance/...` — pass.
- [x] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` — pass.
- [x] `go test ./internal/cli/...` — pass in 387.038s after the repository's golden-transcript generator added the exposed namespace to the root manual fixture.
- [x] targeted `go vet ./cmd/connectorgen ./internal/connectors/hooks/google-search-console ./internal/connectors/commandrunner ./internal/cli` — pass.
- [x] `go build ./cmd/pm` and `go build ./cmd/connectorgen` — pass.
- [x] `go run ./cmd/pm docs validate` — connector manuals validated.
- [x] `cd website && pnpm run gen:website-data` — generated connector data refreshed.
- [x] `pm help google-search-console`, bare `pm google-search-console`, and `pm google-search-console direct url-inspection inspect --help` — pass.

## Runtime reachability evidence

A local fixture server, configured only with a disposable environment credential, confirmed that
`search-analytics by-date` dispatches the safely path-escaped Google property request, while the
two bounded direct reads (`url-inspection inspect` and `mobile-friendly-test run`) and `sites list`
return successful fixture responses. No request was sent to Google and no reverse-ETL plan was
executed.

The recovered direct Search Analytics duplicate was deliberately removed after the same fixture
test showed that current-main safety rejects the provider's documented URL-valued `siteUrl` before
dispatch. Search Analytics remains reachable through its five typed ETL streams, so this retires a
false `availability: implemented` claim without creating an operation gap.

## Citation and base check

`git fetch origin main` immediately before final verification resolved `origin/main` to
`36b431cf152eca02c782cf84ff155c33c2175007`; this branch is two commits ahead and zero behind.
That base does not contain a shared canonical request-field citation bundle convention. Per the
resume contract, no competing syntax was invented: the auditable provider-owned evidence remains
in `research/REQUEST-FIELD-MATRIX.md` with 32/32 operation-specific field paths cited.

Full `go test ./...` and `make verify` are intentionally not run as a single bounded command;
the connector-resume contract requires focused gates on current main instead.
