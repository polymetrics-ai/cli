# VERIFICATION — issue-3175 Chargebee parity wave05 r1

## Focused gates

- PASS: `go test ./cmd/connectorgen -run 'TestChargebeeAPISurfaceOperationLedgerMetrics' -count=1`
- PASS: focused temp-root `go run ./cmd/connectorgen validate <tmp-containing-chargebee>` -> `1 connector(s) checked, 0 findings`
- PASS: `go test ./internal/connectors/conformance -run 'TestConformance/chargebee' -count=1`
- PASS: `go test ./internal/cli -run 'Test(DynamicConnectorHelpAndBareNamespace|ConnectorInspectJSONIncludesManifest|RootHelpListsDynamicConnectorCommands|GoldenTranscripts)$' -count=1`
- PASS: `go build ./cmd/pm`
- PASS: `./pm docs validate --connectors-dir docs/connectors`
- PASS: `make connector-boundary`
- PASS: `make connectorgen-validate`
- PASS: `go vet ./...`
- PASS: `make lint`
- PASS: `make smoke`
- PASS: `make release-workflow-check`
- PASS: `git diff --check`

## Full gate note

- BLOCKED/TIMEOUT: `make verify` reached the repository `go test -timeout 20m ./...` package timeout in `polymetrics.ai/internal/connectors/certify` (`TestCleanupVerifyFailureRecordsLeak` was the running test when the package hit 20m). Earlier focused Chargebee tests, full connector validation, vet, build, docs validation, smoke, lint, connector-boundary, release-workflow-check, and diff-check passed. `polymetrics.ai/internal/cli` did pass under the same 20m timeout when run directly but is slow (`1136.081s`).
- BLOCKED/TIMEOUT: raw `go test ./...` and `go test -timeout 20m ./...` also hit long-running repository package timeouts unrelated to Chargebee assertions.

## CLI/docs parity checks

- PASS: `go run ./cmd/pm help docs`
- PASS: `./pm help connectors`
- PASS: `./pm connectors`
- PASS: `./pm connectors inspect chargebee --json` loaded the updated connector manifest (125 streams, 264 write actions in the manifest payload).
- PASS: `go run ./cmd/pm docs generate --dir docs/cli` regenerated connector docs/catalog surfaces, then non-Chargebee connector doc churn was reverted.
- PASS: website connector data regenerated with `npm run gen:website-data` under `website/`.

## Safety verification

- PASS: no live provider calls.
- PASS: no credentials requested or printed.
- PASS: no generic raw write/read/query escape hatch added.
- PASS: reverse ETL remains plan -> preview -> approval -> execute.
- PASS: destructive delete-style actions include typed confirmation/idempotent-missing semantics where applicable.
- PASS: fixture-only remains uncertified.

## Repo knowledge script

- NOTE: `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` was run after planning artifacts were created; it reported an existing repo conflict because both `AGENTS.md` and `CLAUDE.md` are real files. No repo instruction files were changed.
