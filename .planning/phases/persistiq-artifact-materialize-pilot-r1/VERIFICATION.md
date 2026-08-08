# PersistIQ artifact materialization pilot - verification

**Status:** failed at materialization coverage; no generated bundle was installed

## Timings

| Step | Result | Wall-clock |
|---|---|---:|
| 1. Identify ledger link | pass | 0.02s |
| 2. Map 21 operations | pass | 0.04s |
| 3. Fetch, digest, parse | pass | 2.75s |
| 4. Materialize and static gates | failed closed at coverage | 16.92s |
| 5. Binary reachability/report | report completed; generated sweep not applicable | 0.01s |
| Total | failed pilot | 19.74s |

## Static gates

| Gate | Result |
|---|---|
| `connectorgen validate` | source-only pass: 0 findings; no generated candidate |
| `surface-sync --check` | fail: source runtime endpoint ledger drift=true |
| `TestEveryImplementedCommandPassesRuntimePreflight` | repository test pass; not generated PersistIQ evidence |
| `connectorgen batch gate` | fail: source legacy v0 provenance refusal |
| Real `pm` command reachability | baseline fail: `unknown command "persistiq"`; generated sweep 0 |

## Counts

| Measure | Count |
|---|---:|
| Ledger operations | 21 |
| Fetched | 1 |
| Parsed as OpenAPI 3/Swagger 2 | 1 |
| Materialized | 0 |
| Gated | 0 |
| Reachable | 0 |
| Failed | 1 connector candidate |

## Certification

Certification is withheld. The pilot is never exercised against PersistIQ with
credentials or provider data.

## Failure evidence

`connectorgen batch materialize` returned exit 1 and wrote a drop report with:

```text
executable coverage GET /v1/mailboxes is absent from the cited artifact
```

The existing source bundle also has two further executable legacy streams absent
from the fetched artifact (`GET /v1/activities` and `GET /v1/accounts`). The
materializer stopped at the first coverage mismatch. No source bundle was
rewritten and no gate was weakened.
