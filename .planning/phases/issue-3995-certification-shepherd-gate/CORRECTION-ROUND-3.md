# Correction round 3 of 5 — issue #3995

## Tracking

- Child #4028 owns malformed-pointer coordinates, the shared producer/consumer flow-kind inventory,
  and status derivation from matched evidence.
- Child #4030 owns the explicit canonical non-symlink `certification-gate --root` boundary and
  its nonzero help/usage behavior.
- Child #4024 remains unchanged; this round creates no child and does not rewrite rounds 1 or 2.

## RED-to-GREEN record

Before production edits, tests were added for omitted, added, and remapped flow kinds; forged
completion/status/baseline reports; unsafe and safe-record malformed pointers; `live_tested` with
no evidence; every protected transition with missing, relative, traversing, symlinked, or
symlink-contract roots; and help without a `PROCEED` verdict. After all source and generated
projection updates, focused verification passed:

```sh
go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
go run ./cmd/agentcontractgen sync --root "$PWD"
go run ./cmd/agentcontractgen check --root "$PWD"
go run ./cmd/connectorgen certification-matrix --check
go test -timeout 20m ./cmd/agentcontractgen -run '^TestRunCertificationGate(BlocksEveryProtectedTransition|RejectsUntrustedRootsForEveryTransition|HelpBlocksWithoutProceedVerdict)$' -count=1
```

The projection sync updated all eight registered Claude, Codex, Pi, and OpenCode projections.
No full repository suite, lint suite, CI, PR, provider, credential, evidence-creation, network,
or outer-pipeline action ran in this review phase.

## Disposition

The producer and consumer share one flow-kind catalog. The consumer validates every structurally
valid cell against its accepted record before deriving capability, workflow, sync-mode, flow, and
final status reports, and halts any inconsistent aggregate. Pointer failures retain the trusted
cell coordinate and only a canonical safe evidence record. The command accepts only an explicit,
canonical absolute non-symlink root and never exits zero without encoded `PROCEED`.

The checked-in GitHub baseline must remain deterministic `RETRY` with
`capability/github/capability:check/live_evidence`; a complete producer-valid fixture may
`PROCEED`; malformed, mismatched, or escaped input halts; evaluation remains read-only.
