# Discussion log — reverse-ETL API-surface derivation r1

## Resolved inputs

| Question | Resolution | Evidence |
| --- | --- | --- |
| What is the authoritative scope? | Generic connectorgen surface synchronization plus generated artifacts; GitHub is the measured target only. | Task brief and `surface-sync` causal trace. |
| Which empty entries are defects? | Only implemented commands whose summaries encode a connector-relative method and path. | 214 matching summaries in the base sweep. |
| How are aliases protected? | The parser accepts only an endpoint address prefix and optional action annotation; ordinary friendly summaries do not match. | Exactly 14 nonmatching, named aliases. |
| Can a status change be inferred? | No. Moving endpoint metadata changes routing, not certification status. | Baseline bucket arithmetic remains 1,571. |
| Are live writes required? | No. This is deterministic generator and generated-artifact work; reverse-ETL execution remains out of scope. | Repository safety overlay and task constraints. |
