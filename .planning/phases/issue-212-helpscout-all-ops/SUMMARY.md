# Summary: issue-212 Help Scout all operations

Status: implemented locally with draft stacked PR open: https://github.com/polymetrics-ai/cli/pull/258 (base `feat/213-helpscout-cli-surface-metadata`, stacked on #236). Review routing/CI are pending.

## Delivered

- Added generic bounded JSON direct-read output policy (`output_policy=json`) to schema, validator, commandrunner, and engine.
- Converted Help Scout operation coverage to GitHub-style full-surface parity:
  - 145 official endpoint rows.
  - 4 stream-backed ETL reads.
  - 73 implemented bounded JSON direct-read commands.
  - 66 typed reverse-ETL write actions.
  - 2 bounded binary/raw payload operation metadata rows that remain operation-gated.
- Added `operations.json` and `writes.json` for Help Scout.
- Updated Help Scout metadata, spec base URL handling, streams/check paths, docs, generated connector manuals/catalog, and website connector data.
- Added runtime help behavior so `pm help help-scout`, `pm help-scout`, and `pm help-scout --help` render the connector manual/command surface without credentials.

## Safety

- No credentialed Help Scout checks run.
- No secrets requested, printed, stored, or summarized.
- No new dependencies added.
- No generic raw HTTP write, shell write, SQL write, or arbitrary mutation command exposed.
- Reverse ETL writes remain plan → preview → approval token → typed confirmation.
- Binary/raw payload endpoints remain bounded operation metadata and do not expose unbounded downloads.

## Verification

Focused Go, connector validation, CLI help, docs validation, website data generation, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` passed. Website typecheck is blocked by missing `node_modules`/`tsc`.
