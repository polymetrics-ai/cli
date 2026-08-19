# GitHub Certification Matrix Drift — TDD Ledger

**Exact starting base:** to be recorded after the mandatory RED run.

| Stage | Command / assertion | Expected result | Status |
| --- | --- | --- | --- |
| RED | `make connectorgen-certification-matrix` | Rejects the stale generated GitHub shard before regeneration | PENDING |
| Source audit | Generator and GitHub bundle input inspection | Identifies whether the artifact alone is stale | PENDING |
| GREEN | Canonical GitHub-only matrix generation | Changes only `internal/connectors/defs/github/certification-matrix.json` | PENDING |
| GREEN check | `make connectorgen-certification-matrix` | Passes after regeneration | PENDING |
| Edge | Run the GitHub-only generator twice | Second run leaves the worktree byte-identical | PENDING |
| Happy / bad path | Focused existing generator tests | Matching artifacts pass and intentionally stale artifacts fail | PENDING |

## Test boundary

The RED assertion is the real generated-artifact drift gate, not an invented unit test. It is observable: an unchanged stale artifact must fail the canonical check, while a regenerated artifact must pass it. The edge assertion observes the second generation produces no byte change. A regression test is added only if focused existing coverage does not already exercise those three contracts.
