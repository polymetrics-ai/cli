# TDD ledger — Asana connector parity (#380)

## Setup

- Isolation verified in disposable worktree: `/Users/karthiksivadas/.treehouse/cli-83d592/18/cli`.
- Branch: `fm/cli-asana-parity-wave02-r1`.
- `no-mistakes doctor`: pass before implementation; no daemon stop/restart/update performed.
- `scripts/gsd doctor`: pass.
- `scripts/gsd list`: pass.
- Manual-GSD fallback: `scripts/gsd prompt programming-loop init --phase issue-380-asana-parity --dry-run` is unavailable in this adapter (`unknown GSD command: programming-loop`).

## Red / pre-edit validation

### Official operation inventory and ledger semantics are stale

Credential-free public source comparison against the pinned Asana OpenAPI source:

```text
RED asana official operation inventory
official_count 249 local_count 250 missing_count 0 extra_count 1 duplicate_count 0
official_methods {'GET': 119, 'POST': 81, 'PUT': 26, 'DELETE': 23}
local_classifiers {'covered_by': 26, 'excluded': 224, 'operation': 0}
sample_extra [('GET', '/users/me')]
```

Interpretation: the pre-edit file enumerated the official paths plus a non-OpenAPI `/users/me` alias and used legacy `excluded` rows instead of an operation-ledger-mode planned/blocked ledger. This failed the parent #380 requirement to represent every documented operation exactly once and to keep destructive/admin operations in scope rather than blanket-excluded.

### Connector validation command shape note

The subissue checklist's single-connector invocation currently treats `fixtures/` and `schemas/` as connector roots:

```text
go run ./cmd/connectorgen validate internal/connectors/defs/asana
fixtures: metadata.json: [missing_file] load bundle fixtures: missing required file metadata.json
schemas: metadata.json: [missing_file] load bundle schemas: missing required file metadata.json
connectorgen validate: 2 connector(s) checked, 2 finding(s)
```

Use full defs-root validation or a temporary root containing only `asana/` for real validation evidence.

## Green targets

| Slice | Green evidence |
|---|---|
| Issue policy addendum | #380 and #381-#387 each contain exactly one Asana captain-policy addendum marker. |
| API surface | `api_surface.json` has 249 unique official `(method,path)` rows, missing=0, extra=0, duplicate=0, and no legacy `excluded` rows. |
| Operation metadata | `operations.json` validates through `connectorgen validate`, with every official operation represented once and destructive/admin writes carrying approval and typed-confirmation notes. |
| CLI metadata | `cli_surface.json` validates and uses fixed-target implemented/planned command metadata; no raw provider passthrough command is introduced. |
| Existing executable writes | 13 write fixtures pass conformance; delete actions have `confirm: "destructive"`, idempotent 404 handling, and redaction. |
| Docs/certification | `docs.md`, connector manual/skill docs, and `certification.json` state fixture-only uncertified status and shared blockers truthfully. |

## Evidence log

### Green inventory

```text
GREEN asana official operation inventory
official_count 249 local_count 249 missing_count 0 extra_count 0 duplicate_count 0
official_methods {'GET': 119, 'POST': 81, 'PUT': 26, 'DELETE': 23}
official_lanes {'etl_read': 109, 'reverse_etl_write': 124, 'binary_file': 4, 'cdc_changefeed': 8, 'excluded_not_applicable': 1, 'direct_read_query_search': 3}
local_classifiers {'covered_by': 25, 'operation_blocked': 224, 'excluded': 0}
operations_json 249 cli_commands 249 unique_operation_ids 249
destructive_operations 57 delete_methods 23
```

### Green connector validation/conformance

- `go run ./cmd/connectorgen validate internal/connectors/defs` → pass, `549 connector(s) checked, 0 findings`.
- `go test ./internal/connectors/conformance -run "TestConformance/asana" -count=1` → pass.
- `go test ./internal/connectors/conformance -count=1` → pass.
- `go test ./internal/connectors/engine -count=1` → pass.

### Green CLI/docs/build checks

- `go build ./cmd/pm` → pass.
- `pm help asana`, `pm asana`, and `pm asana tasks delete --help` are help-only and credential-free; delete help includes the typed `--confirm` challenge.
- `pm connectors inspect asana --json` reports 12 streams, 13 write actions, and `write=true` without credentials.
- `.tmp/pm docs validate --connectors-dir docs/connectors` → pass.
- `go test ./internal/cli -run 'Test(CobraRouterShellPreservesDynamicConnector|DynamicConnector|RootHelpListsDynamicConnectorCommands|GoldenTranscripts|GoldenDocsGenerateMatchesTrackedCLIManuals|DocsGenerateAndValidateConnectorDocs|ConnectorInspect|ConnectorCatalog)' -count=1 -timeout 240s` → pass.
- `go vet ./...` → pass.
- `make connector-boundary` → pass (`outcome: clean`).
- `git diff --check` → pass.

### Timeboxed/baseline-heavy verification note

- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` timed out when it included certification CLI tests that repeatedly load every connector bundle.
- `go test ./...` timed out under the default 10-minute package timeout in `internal/cli` and `internal/connectors/certify` certification sweep tests. Other packages already completed or continued successfully in the captured log. No Asana credentials or live provider calls were involved.
