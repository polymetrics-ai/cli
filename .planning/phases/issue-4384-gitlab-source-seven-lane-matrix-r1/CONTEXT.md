# Context — GitLab Track A source-to-seven-lane matrix

## Task Delivery Header

- Issue: Refs #4384 — GitLab source-to-seven-lane matrix.
- Base branch: `origin/main` at `813f457a925f7ee3fe3bea101a43e445992c8552`.
- Merges into: parent #4325 integration decision (`fm/cli-top100-declaration-batch-r1`) → `main`; this task does not merge or open a PR.
- Delivery: a pushed, independently reviewable source-retention/matrix branch with scoped tests green and a completion-proof comment on #4384.
- Working branch: `feat/4384-gitlab-track-a-matrix-r1`.
- Task: retain the exact GitLab source facts needed on the `origin/main` branch, then account every retained source row across the seven lanes without changing runtime, execution, certification, generators, or another connector.
- Verification: deterministic source-lock/crosswalk/matrix reconciliation; deliberate invalid-row/cell tests; JSON validity; focused GitLab package test and source-definition check modes when available; staged-diff review and remote-SHA confirmation.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Retained source input is byte-identical to the approved Batch snapshot | live | The source test verifies the listed source artifact digests/bytes and exact snapshot binding; a changed or omitted retained input fails. |
| Every retained source ID has all seven lane cells | live | The connector-local test compares the full lock ID set and each seven-cell row, rejecting a hidden row or missing cell. |
| Pageable reads and mutations retain independent ETL/write decisions | live | The test recomputes cited source facts and rejects a missing ETL or direct-write/reverse-ETL cell. |
| Boundary-only identities remain visible | live | The test compares every crosswalk-only identity with the matrix boundary record. |
| No shared runtime behavior is introduced | live | Changed-path review is limited to GitLab sources, matrix, local test, and issue-scoped evidence. |

## Scope boundary

The `origin/main` base has no GitLab `sources/` directory. Parent authorization therefore permits only the byte-identical retained source lock, crosswalk, descriptor, binary evidence, and retained artifact bytes from `fm/cli-top100-declaration-batch-r1@dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`. The current `rest.path_bridge` importer limitation remains a mapping restriction; it cannot remove a row and no shared importer change is in scope.

## Materialized source facts

- The exact copied source artifact set is recorded in `sources/gitlab-source-lane-matrix.json#source_snapshot`: primary lock `d874f0d462bc054d3065a41e32b0bb1b1675a84c` (5,783,052 bytes), crosswalk `20353c196663b425280bc0f63ee09dddb5bdc913` (3,004,301), descriptor `19d74be07d6862ef2492c9fb6ff9e6467b67df96` (18,119,458), binary lock `ee538b7ce20912ea13e95d87854b0c014928231d` (3,058), retained-artifact manifest `1d7104c91be20f4ab73920389e681c7a0fb6bc56` (678), and the two cited rendered-reference bytes.
- The frozen denominator is 1,754 source rows: 1,752 primary OpenAPI rows plus two source-backed rendered-reference binary supplements. The three crosswalk-only Product Analytics identities are explicit `not_source_row` boundary records, not dropped source rows.
- Counts: direct read 747 mapped-unproven; direct write and reverse ETL each 1,004 mapped-unproven; binary download 1 mapped-unproven; binary upload 46 mapped-unproven; ETL 255 mapped-unproven; sync 3 missing-foundation. Every other cell is explicitly `not_applicable`; no cell claims `implemented`.
- The source-visible mapping restriction preserves `rest.path_bridge` and the two malformed required-path facts for `postApiV4GroupsIdDashEpicsEpicIidIssuesIssueId` (`epic_issue_id`) and `getApiV4JobsIdSbomScansSbomScanId` (`sbom_digest`). It proposes no shared importer implementation.
