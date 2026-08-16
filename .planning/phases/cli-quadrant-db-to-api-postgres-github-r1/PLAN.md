# Plan — DB → API PostgreSQL → GitHub route R1

## Scope and skills

Target connector: the existing issue-label destination only. PostgreSQL's pre-existing polling-watermark source is exercised as the upstream native source; no PostgreSQL production behavior, generic writer, Arrow fast path, or #4184 run-scoped full-overwrite sequencing changes.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.

GSD lifecycle: `scripts/gsd doctor`, `scripts/gsd sources` for all five required commands, and `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were completed. The generated Pi workflows cannot run compatible isolated role agents in this task runner; execute their required phases inline and record each result in this directory and the PR.

## Definition-derived route

| Source | Destination mode | Descriptor-selected strategy/action | Required row mapping |
| --- | --- | --- | --- |
| PostgreSQL `postgres_polling_watermark` | `full_append` | `append` / `add_issue_labels` | declaration input `target_issue` integer + `label` string → action record `issue_number` + singleton `labels` array |
| PostgreSQL `postgres_polling_watermark` | `incremental_upsert` | `merge` / `set_issue_labels` | declaration input `target_issue` integer + `label` string → action record `issue_number` + singleton `labels` array |
| PostgreSQL `postgres_polling_watermark` | `full_overwrite` | ineligible | Explicit typed rejection: this two-action API destination does not implement a run-scoped overwrite protocol. |

The two source field names are the existing transport-binding input names; the action record fields and their types are taken from GitHub `writes.json`. The destination configuration provides the pre-approved expected value pair for the existing preview/digest gate; execution derives the actual typed record from the reopened PostgreSQL row and refuses a mismatch before any provider write. This keeps plan → preview → approval honest without adding a generic mapper.

## Red / Green delivery slices

1. **R1 — source-bound regression characterization:** Add a fresh-`pm` integration test for PostgreSQL source → GitHub `full_append`; it must fail because the current approval/destination requires the source and destination to be the same connector and synthesizes the destination record instead of mapping the reopened row. Assert exact provider labels, warehouse artifacts, durable receipt/read-back, and checkpoint.
2. **R2 — closed source-row mapper:** Implement the narrow PostgreSQL-to-GitHub mapping at the existing destination adapter. Map only declaration inputs `target_issue`/`label` into their action-owned record fields, validate the row, and require it match the pre-approved preview pair before authorization or provider write; retain non-PostgreSQL behavior. Green tests cover both action declarations, null/malformed rows, and zero provider writes on refusal.
3. **R3 — durability and replay boundaries:** Add binary/container proof for zero rows, `NULL` mapped columns, replay-safe keyed application, interruption/resume via PostgreSQL watermark, and the source's `deletes: not_available` refusal. Every named edge asserts provider state and checkpoint state, not exit status alone.
4. **R4 — real controlled-repository proof:** Build the exact `pm` binary, source the GitHub token only into its process environment, execute plan → preview → approval → run against retained private `karthik-sivadas/pm-parity-proof-db-to-api`, and independently read labels from GitHub. The two dedicated sentinel issue labels remain as verifiable evidence; no third GitHub destination action is invoked for cleanup.
5. **R6 — definition-owned source admission correction:** Replace the shared provider/executor switch with an optional closed `destination_transport.source_bindings` declaration. Each binding names an exact source executor, eligible source streams, and one bounded mapping form. Registry preflight rejects an unlisted source binding before executor access; the issue-label adapter reuses the selected binding for configured-record matching or typed input-field mapping. The connector definition, rather than shared Go, names the two admitted source executors and their row contract.

## Non-goals

- No new GitHub write action (coverage remains two of 607).
- No generic API writer, generic SQL source, generic transport action, or open-ended record mapping. `source_bindings` is a closed descriptor field with two typed mapping forms and exact executor/stream declarations.
- No changes to #4184 atomic full-overwrite behavior, Arrow path, or its benchmark evidence.
- No API→API quadrant edits or broad certification of the remaining 605 actions.

## CLI parity checklist

- [x] The existing `github-issue-label` help/manual, `docs/cli/etl.md`, generated transcript, and website mirror describe both its retained GitHub source slice and the new exact PostgreSQL polling-row contract; flags and bare namespace behavior remain unchanged.
- [x] `pm help etl`, `pm etl transport`, `pm etl transport github-issue-label --help`, docs/website grep, and golden generation/check prove the changed semantics agree.

## Checkpoints

- Commit planning and red/green TDD evidence before production code.
- Run changed-package tests plus `internal/cli` with `-timeout 20m` after each green slice.
- Run opt-in PostgreSQL integration only with the existing shared runtime endpoint; never restart/reconfigure it.
- Before push, run build, vet, lint, generated-file checks, all remaining verification entry points, verify-work, code review, and the independent live read-back.
