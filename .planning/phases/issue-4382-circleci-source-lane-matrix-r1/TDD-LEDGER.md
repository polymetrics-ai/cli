# TDD ledger — CircleCI source-lane matrix R1

| Stage | Command / probe | Result | Evidence |
| --- | --- | --- | --- |
| Intake | `scripts/gsd doctor` | pass | Pi-local GSD adapter healthy |
| Red | `go test ./internal/connectors/defs/circleci -run 'TestCircleCISourceLaneMatrix' -count=1` | expected fail | `read CircleCI source-lane matrix: open sources/circleci-source-lane-matrix.json: no such file or directory` before the matrix was materialized |
| Green | same focused test | pass | 111 rows, 777 cells, source-ID parity, seven lanes, paging/mutation dispositions, source facts, and adversarial mutations all pass |
| Edge: hidden row | `TestCircleCISourceLaneMatrixRejectsHiddenSourceRows` | pass | removes an in-memory matrix row and requires `source row absent from matrix` |
| Edge: bad backlink | `TestCircleCISourceLaneMatrixRejectsInvalidArtifactBacklink` | pass | changes an artifact link to an unknown source ID and requires backlink rejection |
| Edge: paging | `TestCircleCISourceLaneMatrixRejectsMissingPagingETLDisposition` | pass | changes a source-documented cursor candidate's ETL cell to not-applicable and requires rejection |
| Edge: mutation | `TestCircleCISourceLaneMatrixRejectsMissingMutationReverseETLDisposition` | pass | changes a mutation candidate's reverse-ETL cell to not-applicable and requires rejection |
| Refactor | `gofmt -w internal/connectors/defs/circleci/source_lane_matrix_test.go` | pass | local validator is formatted; generated JSON is deterministically indented |

The adversarial tests are deliberate red probes run after the fixture becomes valid;
they mutate decoded in-memory data and must be rejected by the same production-local
validator. No production runtime behavior is exercised or introduced.
