# Plan — GitLab Track A source-to-seven-lane matrix

## GSD execution record

- Inline/manual fallback: generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were inspected. Compatible isolated GSD role execution is unavailable in this runner and the active team policy prohibits spawning extra agents; this single worker performs the equivalent source audit, TDD, verification, and review inline.
- Required skills used: `connector-lane-build-order`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## Slice plan

1. Retain the exact approved GitLab provider-source inputs from `dc481bac` and verify their bytes/digests before relying on them. Do not carry forward declarations, runtime files, generated documentation, or importer code.
2. Add a focused reconciliation test first. Red must demonstrate that a missing matrix fails; edge variants must reject a hidden source row, a missing lane, collapsed mutation decisions, dropped crosswalk-only identity, and an unsupported promotion to `implemented`.
3. Materialize a connector-local matrix from locked facts only. Each retained source row gets all seven lane cells; ETL is decided from explicit collection/pagination evidence and every POST/PUT/PATCH/DELETE gets independent direct-write and reverse-ETL cells. Supplemental binary source rows remain source rows with their rendered-reference citations.
4. Use the Atlas snapshot before recording any real runtime gap. Record `rest.path_bridge` as a source-preserving mapping restriction with the exact affected facts; do not change its importer.
5. Run only focused checks due constrained Go-cache space, review the staged scope, commit/push this branch, verify the remote SHA, and post a no-checkbox Track A proof to #4384.

## Red–Green–Refactor

| Stage | Expected evidence |
| --- | --- |
| Red | The new reconciliation test fails while the matrix file is absent. |
| Green | The matrix makes exact source IDs, seven cells, lane criteria, boundary records, binary evidence, and typed restrictions pass. |
| Edge | In-memory malformed variants fail for hidden rows, missing cells, missing write pairing, boundary loss, and unsupported `implemented` claims. |
| Refactor | Normalize deterministic ordering and review that only connector-local source/matrix/test/planning paths changed. |

## Non-goals

- No execution, provider I/O, credentials, certification, CLI surface, generator, importer, shared foundation, or merge.
- No treatment of `rest.path_bridge` as a reason to omit a retained source row.
