# Verification — Jira Parity Wave 01 R1

## Passed

- Red baseline before production edits:
  - Jira current surface/write invariant exited `1` because the old bundle had 15 surface rows and no `writes.json`.
- Official source inventory:
  - Atlassian Jira Cloud OpenAPI v3 fetched without credentials.
  - `sha256:8439da27e1b2dd7b013a0ae721b8aeaa7746bc8e2d816fa28aa1a582e8597501`
  - `md5:ae49a3d84a12210d4686315cb36442be`
  - 616 operations inventoried.
- Green ledger invariant:
  - 616 `api_surface.json` rows.
  - 286 executable `writes.json` actions.
  - 10 blocked reverse-ETL shared-foundation gaps.
  - 86 DELETE actions; every DELETE action declares `confirm: "destructive"`.
  - 103 total destructive-confirmed write actions.
  - endpoint `(method,path)` rows are unique.
- `find internal/connectors/defs/jira -name '*.json' -print0 | xargs -0 jq empty`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - `548 connector(s) checked, 0 findings`
- `go test ./internal/connectors/conformance -run 'TestConformance/jira' -count=1`
  - `ok polymetrics.ai/internal/connectors/conformance`
- Focused CLI dynamic/connector tests:
  - `go test ./internal/cli -run 'TestDynamicConnector|TestConnectorInspectJSONIncludesManifest|TestConnectorCatalog|TestRootHelpListsDynamicConnectorCommands' -count=1`
- Runtime help / bare namespace:
  - `go run ./cmd/pm help jira` exited 0.
  - `go run ./cmd/pm jira --help` exited 0.
  - `go run ./cmd/pm jira` exited 0 and rendered help.
- Credential-free connector inspect:
  - `go run ./cmd/pm connectors inspect jira --json` emitted valid JSON with Jira manifest/write actions.
- `go vet ./internal/connectors/... ./internal/cli/...`
- `go build ./cmd/pm`
- `make connector-boundary`
  - outcome `clean`
- `git diff --check`

## Timed out / narrowed

- `go test ./internal/cli -run 'Connector|Dynamic|Golden|Jira' -count=1` timed out at 120s because the regex selected broad certify/golden tests. Replaced with the focused dynamic/connector subset above.

## Safety evidence

- No secrets requested, printed, summarized, or stored.
- No live Jira API calls with credentials.
- No provider writes or certification execution.
- Destructive Jira DELETE/write actions implemented in `writes.json` require typed `destructive` confirmation through the existing reverse ETL plan -> preview -> approval -> execute path.
- Unsupported shared-foundation gaps remain blocked and are not counted as executable writes:
  - bounded binary download executor for attachment content/thumbnail direct binary reads;
  - typed integer/object array or whole-object body flags for bounded direct reads;
  - raw scalar/binary request body write dialect for watcher/preference/avatar operations;
  - required `Atlassian-Transfer-Id` request headers for App Migration writes;
  - repeated `columns` form-field request bodies for Jira column defaults.
- Fixture-only/replay metadata does not claim live certification.
