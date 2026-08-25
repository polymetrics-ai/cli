# Summary — #4352 source-bound read execution foundation

## Delivered

- Added a closed `source_operation` declaration edge: exact source ID, GET method, and connector-relative path.
- Source projection now considers source GET operations and only promotes an already declared fixed REST read or an already semantic stream. It never creates a URL, header, arbitrary method, curl-style request, or ETL claim.
- Direct reads require bounded JSON response output and imported required typed path/query inputs. Incomplete contracts remain non-implemented with a stable `missing_foundation=source-bound-read-execution-r1:` note.
- Source-bound ETL remains a normal named stream command. Engine preflight proves one declared `stream_etl` composite owns that stream, exact source route, record schema, and pagination.
- Materialized four Asana controls from the authoritative mapping inventory: three bounded direct reads and the existing paginated workspace stream.

## Honest scope

The three new direct commands are `access-requests get-access-requests`, `agents get-agents-for-workspace`, and `agents get-agent`. `workspaces list` is only an exact-source-bound ETL control; it was already executable.

This does not claim all Batch-1 Asana reads are implemented. Of the stated baseline 100 planned GET rows, three are now materialized controls and 97 still need exact source descriptor/typed operation materialization. Any row without a bounded direct-read contract or source-backed stream semantics stays declaration-pending or has a named missing foundation.

## Integration prerequisites

Do not merge this branch. After the certification/mapping foundations merge, integration needs fresh exact-head tests, an independent audit, and real connector-shaped evidence for Outreach and each other connector that uses this shared edge. Those connector changes may retain source-mapped rows that are honestly deferred for named foundations.
