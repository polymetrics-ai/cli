# GitHub Certification Matrix Drift — TDD Ledger

**Exact starting base:** `origin/main` = `51dd6d468e4a40ece70c36efb81df4fdede8a8b6`.

| Stage | Command / assertion | Expected result | Status |
| --- | --- | --- | --- |
| RED | `make connectorgen-certification-matrix` | Rejects the stale generated GitHub shard before regeneration | NOT REPRODUCIBLE — exited 0: `certification shards are current: connectors=3 capability_complete=0 certified=0` |
| Source audit | Generator and GitHub bundle input inspection | Identifies whether the artifact alone is stale | PASS — `generateCertificationMatrix` rebuilt the GitHub shard from its loaded declaration bundle and certification inputs, and the emitted bytes match the checked-in file. |
| GREEN | Canonical GitHub-only matrix generation | Changes only `internal/connectors/defs/github/certification-matrix.json` | PASS — `go run ./cmd/connectorgen certification-matrix --connector github` completed with no matrix diff. |
| GREEN check | `make connectorgen-certification-matrix` | Passes after regeneration | PASS — already passing before either generation. |
| Edge | Run the GitHub-only generator twice | Second run leaves the worktree byte-identical | PASS — both invocations succeeded and `git diff --exit-code -- internal/connectors/defs/github/certification-matrix.json` remained clean after each. |
| Happy / bad path | Focused existing generator tests | Matching artifacts pass and intentionally stale artifacts fail | PASS — `TestCertificationScopedGenerationLeavesOtherShardByteIdentical`, `TestCertificationScopedCheckDoesNotReadGlobalStatusOrOtherConnectorShards`, and `TestCertificationShardDriftFails` passed. |

## Test boundary

The RED assertion is the real generated-artifact drift gate, not an invented unit test. It is observable: an unchanged stale artifact must fail the canonical check, while a regenerated artifact must pass it. The edge assertion observes the second generation produces no byte change. A regression test is added only if focused existing coverage does not already exercise those three contracts.

## Outcome: no stale artifact on current main

The required RED observation is absent on the latest fetched default branch.
`make connectorgen-certification-matrix` returned exit 0 before any generated
file was written. `generateCertificationMatrix` uses `buildCertificationShards`,
which derives the allowlisted shard from `buildCapabilityMatrixForConnectors`
and `buildFlowMatrixForConnectors`; the GitHub bundle is loaded through
`engine.Load` from `internal/connectors/defs` with its scoped runtime operation
endpoint ledger. The canonical GitHub-only writer emitted exactly the
checked-in bytes twice.

The existing focused tests independently cover the requested bad path (a stale
GitHub shard is rejected), the happy scoped check, and scope-preserving
determinism. There is therefore no generator, source, or artifact correction
to make. The smallest provider-neutral correction plan is **none**: do not
change certification data; rerun this task only from a newly identified remote
SHA that fails the canonical check.
