# PersistIQ artifact materialization pilot - verification

**Status:** passed static gates under the captain complete-inventory policy;
certification withheld

## Timings

| Step | Result | Wall-clock |
|---|---|---:|
| 1. Identify ledger link | pass | 0.03s |
| 2. Map all 21 operations | pass | 0.03s |
| 3. Fetch, digest, parse OpenAPI 3.0.1 | pass | 2.70s |
| 4. Materialize + static gates + real binary reachability | pass | 50.07s |
| 5. Report collation | pass | 0.09s |
| **Total** | **pilot pass** | **52.92s** |

Step 4 sub-timings: batch plan 3.73s, materialize 1.46s, validate 0.90s,
surface-sync derivation 1.48s, `surface-sync --check` 0.78s, batch gate 0.71s,
repository runtime-preflight test 4.54s, binary build 1.68s, 24-command help
sweep 28.73s, bare namespace 2.37s, and three intentional blocked-command
checks 3.69s.

## Static gates

| Gate | Result |
|---|---|
| `connectorgen validate` | pass: 0 findings |
| `surface-sync --check` | pass: no drift |
| `connectorgen batch gate` | pass: 1 included, 0 dropped; 21 runtime-preflight commands |
| `TestEveryImplementedCommandPassesRuntimePreflight` | pass |
| Real `pm` command reachability | pass: 24/24 help paths; 21/21 implemented |
| Not-implemented execution safety | pass: 3/3 blocked before credentials/network, no unknown commands |

## Counts

| Measure | Count |
|---|---:|
| Mapped artifact operations | 21 |
| Fetched | 1 |
| Parsed as OpenAPI 3/Swagger 2 | 1 |
| Materialized candidates | 1 |
| Gated candidates | 1 |
| Implemented commands | 21 |
| Named-dependency commands | 3 |
| Flagged discrepancies | 3 |
| Reachable command paths | 24 |
| Failed candidates | 0 |

## Policy assertions

- Every artifact operation appears in the 24-row generated API surface.
- `GET /v1/mailboxes`, `GET /v1/activities`, and `GET /v1/accounts` remain
  present and each carries
  `discrepancy=present-in-surface-absent-from-artifact`.
- `GET /v1/leads/{id}`, `POST /v1/leads`, and `PUT /v1/webhook_plugin` are
  visible as `not_implemented` commands with `named_dependency=<slug>` notes.
- No command with an unavailable executor is marked `implemented`.
- Materialization does not run the repository-wide preflight per candidate;
  the final gate runs once over the staged result.

## Certification

Certification is withheld. PersistIQ was never exercised against its provider;
no credentials or provider data were used.

## Evidence

The fresh artifact, exact operation map, generated bundle, timings, reports,
and binary sweep outputs are under
`.planning/phases/persistiq-artifact-materialize-pilot-r1/rerun-2026-08-08/`.

## Multi-source extension verification

The required generalization pilots now pass under the captain's source-order
contract. Watchmode (23), DocuSeal (34 including 11 OpenAPI 3.1 webhooks), and
Float (102 after external Swagger path traversal) materialized and reached
45/45, 34/34, and 104/104 real-binary command paths respectively. Copper's 77
Postman operations materialized and passed the same static gate; reachability
is intentionally not claimed because the existing native scaffold has no
embedded command surface.

The one combined staged gate included 4/4 candidates, dropped 0, checked 32
implemented commands, and saw 265 declared operation rows. `connectorgen
validate` returned 0 findings and 0 warnings; `surface-sync --check` reported
no drift. Per-operation source URL/kind/version/hash/date/coordinate and
alternative citations are recorded in the staged `api_surface.json` files and
the operation-map reports under
`generalization-validation-2026-08-08/`.

Final rerun timing evidence: materialize Watchmode 6.07s, DocuSeal 1.75s,
Float 0.94s, Copper 0.99s; shared validate 1.02s, surface-sync derive/check
0.89s/1.05s, batch gate 1.17s, binary build 12.27s; binary reachability
105.88s/79.62s/251.73s. No provider operation was exercised.
