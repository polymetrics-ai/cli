# Verification — Monday parity wave02 r1 (#82)

## Targeted gates

| Command | Result |
|---|---|
| `go test ./cmd/connectorgen -run MondayAPISurface -count=1` | red before bundle update on `metadata.capabilities.write=false`; green after update (`ok`, 0.352s) |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | pass: `549 connector(s) checked, 0 findings` |
| `go test ./internal/connectors/conformance -run 'TestConformance/monday' -count=1` | pass (`ok`, 1.911s) |
| `go test ./cmd/connectorgen -count=1` | pass (`ok`, 24.660s) |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` | pass; regenerated root help goldens for Monday command discovery |
| `go test ./internal/cli -run 'TestDynamicConnector\|TestRootHelpListsDynamicConnectorCommands\|TestGoldenTranscripts\|TestGoldenDocsGenerateMatchesTrackedCLIManuals\|TestDocsGenerateAndValidateConnectorDocs\|TestDocsGenerateIncludesConnectorCatalog' -count=1` | pass (`ok`, 61.296s) |
| `go vet ./internal/connectors/... ./internal/cli/... ./cmd/connectorgen` | pass (no output) |
| `go vet ./...` | pass (no output) |
| `go test -timeout 20m ./...` | pass; slow packages included `internal/cli` and `internal/connectors/certify` |
| `go build ./cmd/pm` | pass (no output; produced local `./pm`) |
| `make connector-boundary` | pass: outcome `clean`, 130 checked files, 549 connectors loaded; only pre-existing exceptions listed |
| `git diff --check` | pass (no output) |
| `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` | pass; generated connector docs/catalog |
| `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` | pass: `Validated connector docs in docs/connectors` |
| `make verify` | pass; includes fmt, tidy-check, vet, test, build, docs-check, smoke, lint, connector validation/boundary, and release workflow check |

## CLI smoke checks

Built binary checks passed:

- `./pm help connectors` exit 0.
- `./pm connectors inspect monday --json` exit 0; connector write capability true, 5 streams, 102 write actions.
- `./pm help monday` exit 0; detailed help includes planned query commands, reverse commands, and destructive `delete-board` metadata.
- `./pm monday` exit 0; bare namespace renders concise contextual help.
- `./pm monday --help` exit 0; help renders concise contextual help.
- `./pm monday --bogus` exits 2 with usage error (`error: missing connector command path`).

## Known non-blocking observations

- No live monday.com provider calls, credentials, writes, certification, pushes, merges, or PRs were performed.
- The public docs inventory produced 254 operations (66 queries, 188 mutations). Direct query commands are planned pending shared duplicate `POST /` classifier validation and GraphQL `errors[]` direct-read semantics. The parent issue preserves 292 operations from a live schema source that this worker did not call; absent schema-only operations remain source-blocked/uninvented.
