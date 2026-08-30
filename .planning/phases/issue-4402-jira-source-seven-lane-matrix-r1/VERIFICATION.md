# Jira Track A verification record

## Required checks

| Check | Purpose | Result |
| --- | --- | --- |
| `jq empty internal/connectors/defs/jira/sources/jira-source-lane-matrix.json` | Matrix is valid JSON. | passed |
| `go test ./internal/connectors/defs/jira -run '^TestJiraSourceLaneMatrixRetainsEveryLockedOperationAndLane$' -count=1` | Lock binding, 617-row reconciliation, all lane dispositions, exact source facts, legacy backlinks, and deliberate malformed cases. | passed |
| `go test ./internal/connectors/defs/jira -count=1` | Full connector-local Go test package. | passed |
| `go run ./cmd/connectorgen source-import jira --check --read-projection-only` | Detect whether the legacy Jira lock can participate in current check-only importer validation without mutation. | blocked by pre-existing `sources/jira-retained-artifacts.json` absence; no matrix or runtime file was changed |
| `go run ./cmd/connectorgen source-materialize jira --check` | Check v4 materialization eligibility without writing. | blocked by the legacy v2 lock; command requires a v4 lock with one materialization block |
| `go run ./cmd/connectorgen surface-sync internal/connectors/defs --connector jira --check` | Check source projection without writing. | blocked by pre-existing missing canonical source descriptor |
| `jq empty internal/connectors/defs/jira/sources/jira-source-lane-matrix.json` | Matrix is valid JSON. | passed |
| `git diff --cached --check` | Whitespace integrity of scoped work. | passed |
| Changed-path audit | Confirms only the five issue planning files, Jira matrix, and Jira local test changed. | passed |

## Runtime boundary

No implementation proof is claimed. The only `missing_foundation` mapping is `jira.rest.registerDynamicWebhooks` / `sync_transport`, recorded as `cli-webhook-event-surface-foundation-r1` after `transport.sync-contract.v1` lookup; the absent receiver would have to stage through DuckDB. Future runtime work must begin with captain approval and source-specific proof design.
