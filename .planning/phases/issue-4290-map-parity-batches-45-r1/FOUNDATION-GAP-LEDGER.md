# Issue #4290 foundation-gap ledger

Status: **blocked**. Shared gaps are deduplicated by stable ID; their affected provider operations remain visible as individual machine-readable rows.

| Gap | Status | Affected surface | Enumerated / retained rows | Connectors | Owner |
| --- | --- | --- | ---: | ---: | --- |
| application-generic-destination-dispatch | open | reverse_etl | 345 / 49 | 8 | #4304 / fm/cli-reverse-etl-destination-r1 |

| Rollup | Enumerated provider operations | Retained bindings outside current inventory | Connectors | Merge-ready operations |
| --- | ---: | ---: | ---: | ---: |
| batch4 | 321 | 46 | 5 | 0 |
| batch5 | 24 | 3 | 3 | 0 |
| portfolio | 345 | 49 | 8 | 0 |

Each `operation_gap_rows` entry in `FOUNDATION-GAP-LEDGER.json` contains the exact provider method/path/source URL/revision/hash trace, canonical mapping, failing App runtime evidence, owner, status, affected surfaces, and closure commands. No connector-specific workaround is permitted.
