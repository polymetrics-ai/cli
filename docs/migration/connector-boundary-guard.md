# Connector boundary guard

`connectorgen boundary` is a developer guard for connector-definition ownership. It scans shared production Go for connector-specific policy that should live in `internal/connectors/defs/<connector>/`, `internal/connectors/hooks/<connector>/`, or `internal/connectors/native/<connector>/`.

```bash
go run ./cmd/connectorgen boundary .
go run ./cmd/connectorgen boundary . --json
go run ./cmd/connectorgen boundary . --base origin/main
make connector-boundary
```

`--base <ref>` limits the primary scan to Go files changed from that git ref, including untracked
files, while exception-ledger contracts are still checked against the whole tree.

## Output and exit status

- Exit `0`: clean boundary report.
- Exit `1`: policy violations or exception-ledger contract failures.
- Exit `2`: invalid invocation or scanner configuration.

JSON output uses the `ConnectorBoundaryReport` envelope with stable arrays for `findings`, `warnings`, and applied `exceptions`.

## Exception ledger

The baseline ledger is `internal/connectors/boundary/exceptions.json`. Each row must bind an exact `rule`, `connector`, `path`, `match`, `reason`, `migration_issue_url`, `owner`, `expires_on`, and `max_matches`.

The guard fails when an exception expires, stops matching, or matches more findings than `max_matches`. Free-form `approved_by` prose is ignored and does not count as approval.

## Review disposition

When the guard fails, prefer moving behavior into connector definitions. Add an exception only for current residue with a public follow-up issue, an owner, an expiry, and a bounded match count. This guard does not need connector credentials or live provider calls.
