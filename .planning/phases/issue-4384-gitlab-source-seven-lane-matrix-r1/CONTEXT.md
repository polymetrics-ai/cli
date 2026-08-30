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

## Semantic mapping repair continuation — 2026-08-31

### Task Delivery Header

- Issue: Refs #4384 — GitLab Track A source-to-seven-lane matrix semantic repair.
- Base branch: `feat/4384-gitlab-track-a-matrix-r1` at `4e3944ec07ed438da0d33051ff123b057186c741`.
- Merges into: this repair remains a pushed review branch for the existing #4384 parent work; no PR or merge is opened by this repair.
- Delivery: a pushed, independently reviewable connector-local semantic-mapping commit with fresh proof on #4384.
- Working branch: `fix/4384-gitlab-semantic-read-r1`.
- Task: correct only GitLab’s retained source-lane matrix and local reconciliation test so lane decisions follow source semantics rather than GET-only or request-control-only heuristics.
- Verification: focused and full GitLab package tests, race, vet, JSON syntax, agent-contract check, staged-diff review, remote SHA read-back, and a documented `connectorgen validate` shared-parser result.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Source-documented bounded HEAD and query/search POST reads are direct-read candidates | live | The real retained matrix has three HEAD and thirteen POST semantic-read rows; each has a source-lock backlink and `mapped_unproven` direct-read cell, while its POST write cell is not fabricated. |
| Mutation POSTs remain excluded from semantic direct reads | live | The in-memory matrix mutation-POST promotion fails the real reconciliation test before an invalid cell could be accepted. |
| ETL needs an actual request-to-success-response continuation pair | live | The retained matrix contains two such pairs and 257 request-control-only rows whose ETL cell is explicitly `not_applicable`; a page/per-page no-response case is asserted. |
| Sync does not follow pagination | live | Only the three source-cited webhook-registration contracts remain `sync_transport` candidates; pagination contributes no sync cells. |
| No shared behavior is introduced | live | Changed implementation paths are the GitLab matrix/test and #4384 GSD evidence only; source locks, crosswalks, runtime, transport, certification, shared mapping controls, and Atlas are unchanged. |

- `direct_read` is now 763 `mapped_unproven` / 991 `not_applicable`: the original 747 source GET reads plus three bounded HEAD metadata reads and thirteen source-documented POST query/lookup reads. No row is marked `implemented`.
- `direct_write` and `reverse_etl` are each 991 `mapped_unproven` / 763 `not_applicable`: every retained provider mutation remains paired, while the thirteen semantic POST reads are no longer fabricated as writes.
- ETL is 2 `mapped_unproven` / 1,752 `not_applicable`: only `page_token → next_page_token` and `after → endCursor` plus `hasNextPage` are retained request/response continuation contracts. The other 257 rows retain their request-control facts and explicitly state why no ETL candidate is claimed.
- `sync_transport` remains exactly three webhook-registration `missing_foundation` records. Pagination was not allowed to create a sync candidate.
- The unchanged generic `connectorgen validate internal/connectors/defs/gitlab --json` still reports `parse source lock: json: unknown field "path_bridge"`. That is the pre-existing shared source-projection parser restriction, not a reason to omit or defer a source row; it is deliberately out of this connector-local repair.
