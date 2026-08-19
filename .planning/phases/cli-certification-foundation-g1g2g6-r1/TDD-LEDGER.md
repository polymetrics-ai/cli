# TDD ledger: connector certification foundation G1/G2/G6

## Planned evidence

| Slice | Red | Green | Refactor / verification |
| --- | --- | --- | --- |
| G1 projection | Existing sweep schema cannot name exact kind/class/action and rejects neither a mismatch nor a delete action. | Classifier derives exact normalized types from operation, intent, capability, and transport descriptors. | Table tests cover happy, bad, edge, database, and delete routes. |
| G2 generated sweep | Existing bytes omit projection fields and can represent no projection as N/A. | Each row has the generated fields and generated artifact checks are current. | Snapshot drift, validation, and GitHub sweep check pass. |
| G6 evidence import | Existing imports can publish a prefix, final files can be read partial, and the live script deletes a valid evidence record after a drift check. | Batch validation publishes all or none via no-replace atomic paths; readers see valid JSON; script generates scoped shard before checking. | Concurrent reader and forced-failure regressions pass. |

## Actual evidence

### 2026-08-19 — planning checkpoint

- Red: pending production tests; the authoritative scout report verified missing sweep projections and direct final-path proof writes.
- Green: pending implementation.
- Manual GSD fallback: lifecycle prompts resolved and executed inline because isolated compatible GSD workers are unavailable and the task contract forbids role spawning.

### 2026-08-19 — G1/G2 projection

- Red: `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification(Parity|SweepProjects|SweepCarries)'` initially failed to build because the normalized classifier and projection fields did not exist.
- Green: the same focused command passed after the single generated classifier was added. It covers REST/GraphQL read and write intent, stream and write references, delete, binary download/upload, CDC, changefeed, managed destination, N/A, and mismatches. It also proves Zoom/GitLab `capability:read` projects `rest_read/direct_read`, PostgreSQL's declared managed destination projects `reverse_etl/reverse_etl` despite `write:false`, and the MySQL-shaped managed-destination contract uses that same rule.
- Green: `go run ./cmd/connectorgen certification-sweep --connector github` regenerated 1,575 GitHub rows (1,571 CLI commands plus four declared capability/transport rows); `go test -timeout 20m ./cmd/connectorgen -run TestCertificationSweepCommandChecksGeneratedGitHubArtifact` passed. Every row contains both generated projection fields, and declared GitHub delete actions are `rest_write/direct_write` with `write_action_kind=delete`.

### 2026-08-19 — rebase reconciliation for declaration rows

- Red: after rebasing onto `origin/integration/4015-mvp-flat-r1`, `go test -timeout 20m ./cmd/connectorgen -run 'TestCertificationSweepEmitsDeclaredManagedDestinationProjection|TestCertificationSweepProjectsExecutableAPIReads'` failed: PostgreSQL had no `cli_surface.json`, and Zoom/GitLab's implemented read capability had no generated row. The helper classifier alone did not make those contracts schedulable.
- Green: the sweep now emits stable capability, changefeed, and source/destination transport rows alongside CLI commands. `go test -count=1 -timeout 20m ./cmd/connectorgen -run 'TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive|TestCertificationSweepEmitsDeclaredManagedDestinationProjection|TestCertificationSweepProjectsDeclaredAPIReadCapabilities'` passed; PostgreSQL's actual `transport destination` row is `reverse_etl/reverse_etl`, and Zoom/GitLab's `capability read` rows are `rest_read/direct_read` despite generic `write:false`.
- Green: `go test -count=1 -timeout 20m ./cmd/connectorgen -run 'TestCertificationParityClassifier(ProjectsOnlyTheNormalizedTaxonomy|RefusesMismatchedOrUnresolvedReferences)$'` passed with planned-changefeed and wrong-transport-role edge cases.

### 2026-08-19 — G6 immutable publication

- Red: `go test -timeout 20m ./cmd/connectorgen -run 'Test(PreparedEvidence|LiveCertificationScript)'` initially failed because no prepared atomic publisher existed. `TestCertificationScopedCheckDoesNotReadGlobalStatusOrOtherConnectorShards` also failed because the old check required PostgreSQL's unrelated shard.
- Green: staged same-filesystem hard-link publication, destination prevalidation, no-replace handling, and the concurrent matrix-reader regression pass. Report, transport, and change-capture imports now prepare every record before any final name is published.
- Green: the live runner test asserts `draft -> import -> scoped generation -> scoped check` and refuses the old accepted-evidence deletion ordering. Scoped GitHub matrix checks pass without reading global status or another connector shard.
- Live green: the authorized read-only GitHub fixture run used Keychain entry `pm-cert-classic` only through `PM_CERT_GITHUB_TOKEN`; `repo read-file` returned HTTP 200, published schema-v2 accepted record `github-direct_read_sweep_repo_read_file-github-1787097821338-6301903bc006.json`, regenerated GitHub's scoped shard, and passed its scoped check. No credential value was written to this ledger.
