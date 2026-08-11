# Correction round 5 of 5 — issue #3995

## Scope

- Child #4028 tracks the final flow-pair endpoint-role invariant and the generated catalog
  bootstrap correction.
- The unchanged producer remains bound to `flow-matrix.json`; the consumer continues to receive
  its importable catalog only through `agentcontractgen sync/check`.

## RED record

Before production edits, coverage was authored for a green flow pair whose source role becomes
not-applicable and for missing, empty, invalid, duplicate, and mutable generated catalog data.
Per this review phase’s one-focused-verification boundary, these assertions were not run
separately; they target the two source-confirmed pre-fix weaknesses.

## GREEN record

After all source changes, the generated catalog file was removed from the working tree and the
first command below compiled the stable catalog package and recreated it from `flow-matrix.json`.
The focused package, unchanged-producer, projection check, and forbidden-path audit then passed:

```sh
go run ./cmd/agentcontractgen sync --root "$PWD"
go test -timeout 20m ./internal/certificationcatalog ./internal/agentcontract ./cmd/agentcontractgen -count=1
go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
go run ./cmd/connectorgen certification-matrix --check
go run ./cmd/agentcontractgen check --root "$PWD"
git diff --check da7747a796049601a179a97c025bfb05f011f1e8
git diff --name-only da7747a796049601a179a97c025bfb05f011f1e8 -- 'cmd/connectorgen/certification*.go'
```

Sync reported one recreated catalog and no projection changes. The final path audit produced no
forbidden producer-path output. The checked-in GitHub artifact remains deterministic `RETRY` at
`capability/github/capability:check/live_evidence`; a complete valid fixture may `PROCEED`; role,
catalog, malformed, mismatched, escaped, and overridden-invalid inputs halt before transition
enforcement can proceed.

The first consumer package run also exposed stale deterministic render hashes after the prior
intentional canonical root-instruction change. Both expected hashes were refreshed from the
current canonical render before the final focused Green run passed.
