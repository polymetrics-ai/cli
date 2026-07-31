# Verification — issue #188 Front parity

## Required gates

- [x] `go run ./cmd/connectorgen validate` — pass (`549 connector(s) checked, 0 findings`).
- [x] Targeted temp-root Front validation — pass (`1 connector checked, 0 findings, 0 warnings`).
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` — pass.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m` — pass.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` — pass.
- [x] `go build ./cmd/pm` — pass.
- [x] `make connector-boundary` — pass, `outcome: clean`.
- [x] `git diff --check` — pass.
- [x] `gh-axi issue edit` addenda for #188 and #189-#195 — pass; marker appears exactly once in each refreshed body.

Note: `go run ./cmd/connectorgen validate internal/connectors/defs/front` is not a valid targeted invocation for the current tool because it scans children of the provided directory as connectors. The equivalent targeted gate was run by copying `front/` under a temporary defs root, and the full defs-root validation was run successfully.

## Coverage evidence

Generated connector-local coverage counts:

- `api_surface.json`: 255 operation rows from Front `llms.txt` reference pages.
- `streams.json`: 115 read streams (112 ETL reads + 3 GET event/changefeed surfaces).
- `writes.json`: 119 write actions (118 reverse-ETL writes + 1 application event trigger surface).
- `operations.json`: 21 fixed metadata operations (5 direct reads, 5 binary downloads, 10 not-applicable/disallowed operations, 1 duplicate ledger operation).
- `cli_surface.json`: 255 connector-local command metadata entries.
- `fixtures/writes/*.json`: 119 sanitized write fixture shape files.
- `schemas/*.json`: 115 generated permissive stream schemas.
- `certification.json`: fixture-only certification contract; live certification remains unavailable/uncertified.

## CLI/help/docs parity checklist

- Runtime help code was not edited.
- Connector-local `cli_surface.json` is updated for provider-style command discovery.
- Connector-local `docs.md` documents read, write, direct, binary, excluded/not-applicable, fixture-only certification, and safety gates.
- Generated CLI golden transcripts were updated because root help now discovers the Front provider-style command surface.
- Website/manual site regeneration was not run in this worker; generator-ready connector metadata is present and no website files were in the allowed scope.

## Safety verification

- No live Front provider calls.
- No credentials requested, printed, or stored.
- No destructive external action executed.
- No reverse ETL execution; only local validation and fixture replay/capture servers.
- No shared runtime/foundation files edited.
- No new dependencies.
- No no-mistakes, push, PR, merge, VPS, or Thaalam work.
