# Correction round 4 of 5 — issue #3995

## Scope

- The round-three direct catalog dependency in `cmd/connectorgen/certification*.go` is being
  removed. The unchanged producer remains bound to `flow-matrix.json` by
  `connectorgen certification-matrix --check`.
- `agentcontractgen sync` and `check` own an importable generated flow-kind catalog derived from
  that matrix for the consumer; it is not a hand-maintained duplicate allowlist.
- Child #4028 covers immutable override facts and every raw flow pair-set evidence binding.
  Child #4024 covers precision-preserving semantic proof comparison.

## RED record

Before production edits, coverage was authored for an override that promotes immutable facts, an
override that masks safe-missing, mismatched, or wrong-coordinate base evidence, semantic proof
numbers that differ above IEEE-754 exact range, and catalog sync/check drift. Per this review
phase's one-focused-verification boundary, the RED assertions were not run separately; each
targets a pre-fix path that could otherwise reach `PROCEED` or accept generated-catalog drift.

## GREEN record

After all source changes, the focused package checks, unchanged producer matrix check,
`agentcontractgen sync/check`, and a changed-path audit passed:

```sh
go run ./cmd/agentcontractgen sync --root "$PWD"
go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
go run ./cmd/connectorgen certification-matrix --check
go run ./cmd/agentcontractgen check --root "$PWD"
git diff --check da7747a796049601a179a97c025bfb05f011f1e8
```

Sync reported zero projection and catalog drift. The changed-path audit found no changed
`cmd/connectorgen/certification*.go` path relative to the base. The round-three claim that no
such path changed was inaccurate for that intermediate revision and is superseded by this record.
The GitHub baseline remains deterministic `RETRY` at
`capability/github/capability:check/live_evidence`; a complete valid fixture may `PROCEED`; every
malformed, mismatched, escaped, overridden-invalid, or precision-mismatched input halts.
