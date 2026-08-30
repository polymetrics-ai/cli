# Review — GitLab Track A source-to-seven-lane matrix

## Verdict

Pass for the scoped Track A mapping delivery. The change set is connector-local source retention, a source-lane matrix, one reconciliation test, and issue-scoped planning evidence only. It adds no executor, importer, generator, command, stream, write definition, or shared Foundation Atlas implementation.

## Reviewed invariants

- Exact transferred source bytes are bound by Git blob SHA-1 and size; retained rendered-reference artifacts are additionally checked by SHA-256 and byte count.
- The matrix preserves all 1,754 retained source rows and only classifies the three crosswalk-only identities as `not_source_row` boundary evidence.
- Each row has precisely the seven required cells. GET, POST/PUT/PATCH/DELETE, binary, explicit pagination, and hook-registration decisions are recomputed from locked source facts in the test.
- Direct-write and reverse-ETL cells remain independent and every DELETE is retained as a mutation candidate.
- No applicable cell is promoted to `implemented`; the three `/hooks` source-registration cells remain `missing_foundation` with a typed, source-listed Atlas lookup.
- `rest.path_bridge` and the two malformed required-path facts remain source-visible mapping restrictions, never foundation claims and never row-removal justifications.

## Verification outcome

- Focused test and edge variants: pass.
- JSON syntax and package vet: pass.
- Canonical agent-contract check: pass.
- `source-import --check --read-projection-only` and connector validation both stop at the pre-existing `path_bridge` parser acceptance gap. This delivery records it but intentionally does not change shared code.

## Remaining review boundary

The shared source-lock parser must later accept and preserve the established `rest.path_bridge` field before its generic source-import/projection checks can pass. That is a mapping-control repair outside this issue's connector-local scope.

## Semantic repair continuation — review outcome

Pass for the connector-local semantic correction, pending fresh independent review of the pushed commit.

- The repair replaces only the invalid lane heuristics: direct read is no longer GET-only, and ETL is no longer created by request pagination controls alone.
- Semantic POST reads require all three retained facts: POST method, a provider-documented query/lookup-style summary (or execute+query/GraphQL wording), and a 2xx/3xx source response. This is a rule over source facts, not a fixed operation-ID list; arbitrary mutation POSTs remain writes.
- Pagination recognizes explicit provider request controls (`page`, `per_page`, `page_token`, cursor forms, `after`, or `offset`) even where descriptions are absent. It requires the matching successful-response successor before an ETL cell is emitted, avoiding description rigidity without allowing a request-only false positive.
- The two true pairs are visible with their exact request/response fields. The 257 request-control-only rows retain source evidence but truthfully remain non-ETL. This correction does not change an existing stream or claim runtime execution.
- All source-documented provider mutations/deletes remain paired direct-write and reverse-ETL candidates. Three source-cited hook registrations remain the sole sync candidates; list pagination cannot create sync.
- Safety review: no provider I/O, credentials, generic transport, path handling, subprocess invocation, or runtime parser change is added. JSON comes only from already locked source artifacts, and safe map assertions avoid assuming a request-body `content` block exists.
- The only scoped validation failure is unchanged: generic `connectorgen validate` rejects `rest.path_bridge` as an unknown field. The repair does not bypass it, delete a source row, or modify shared parser/certification/mapping controls.
