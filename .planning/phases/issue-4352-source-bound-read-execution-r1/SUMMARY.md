# Summary — #4352 source-bound read execution foundation

## Delivered

- Added a closed `source_operation` declaration edge: exact source ID, GET method, and connector-relative path.
- Source projection now considers source GET operations and only promotes an already declared fixed REST read or an already semantic stream. It never creates a URL, header, arbitrary method, curl-style request, or ETL claim.
- Direct reads require bounded JSON response output and imported required typed path/query inputs. Incomplete contracts remain non-implemented with a stable `missing_foundation=source-bound-read-execution-r1:` note.
- Source-bound ETL remains a normal named stream command. Engine preflight proves one declared `stream_etl` composite owns that stream, exact source route, record schema, and pagination.
- Materialized four Asana controls from the authoritative mapping inventory: three bounded direct reads and the existing paginated workspace stream.

## Honest scope

The three new direct commands are `access-requests get-access-requests`, `agents get-agents-for-workspace`, and `agents get-agent`. `workspaces list` is only an exact-source-bound ETL control; it was already executable.

This does not claim any of Batch-1's 100 planned Asana GET rows are newly
implemented. The retained source import binds the nine already executable
controls (three bounded direct reads and six declared streams) to their locked
source identities. Batch 1 must still materialize an exact declaration-owned
operation/command, typed required path/query/body inputs, bounded direct-read
output or proven stream records/pagination semantics, and an executable
runtime preflight for each of the 100 planned GET rows before promotion. Any
row without those source-backed semantics stays declaration-pending or has a
named missing foundation.

The same shared mapping closure now records every retained Asana mutation
without changing a command's executor: 21 absent actions are source-cited
non-executable declarations, 65 existing reverse-ETL actions name the missing
typed request-schema foundation, and 4 legacy delete actions name their
source-to-local path-parameter alias foundation. Other connector batches reuse
this by retaining their own locked descriptor plus exact operation-granular
dispositions; no connector-local shim or runtime certification is required for
declaration admission.

## Integration prerequisites

Do not merge this branch. After the certification/mapping foundations merge, integration needs fresh exact-head tests, an independent audit, and real connector-shaped evidence for Outreach and each other connector that uses this shared edge. Those connector changes may retain source-mapped rows that are honestly deferred for named foundations.
