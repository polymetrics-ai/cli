# Issue #4283: Batch-1 source-rigidity R2 plan

## Scope decision

**Authorized bounded cohort.** Captain instruction `006.msg` changes the obsolete delivery policy rather than waiving it. PR #4294 may contain the ten named Batch-1 connectors and only the shared foundations required to preserve their source-lock identities. Every shared change needs an Atlas `reuse`, `constrained_extension`, or captain-approved `actual_gap` disposition. The scope never authorizes unrelated connectors, generic raw transport, credential handling, destructive-operation policy, approval-policy changes, provider-live I/O, or a merge.

## Immutable cohort prerequisites

1. Keep the frozen remote SHA and exhaustive review ledger current.
2. Generate a machine-checkable source-operation-to-lane matrix for all ten locked cohorts, including direct, binary, ETL, reverse-ETL, sync transport, CLI/docs, and typed blocked/unsupported cells.
3. Reconcile each source identity to exactly one `implemented`, `blocked_with_named_foundation`, or `unsupported_with_provider_evidence` disposition; certification is overlay evidence only.
4. Consult the Foundation Atlas before every shared foundation change and update it in the same change when its real contract changes.

## Planned TDD sequence once scope is valid

| Slice | Red | Green | Edge/refactor proof |
| --- | --- | --- | --- |
| Source-normalization foundation (F1/F2) | Representative locked aliases/numeric/default/duplicate/range cases fail. | Presence-aware canonical source facts retain raw evidence and exact semantic status behavior. | `default` alone remains blocked; declared 404/non-2xx never become executable. |
| Retention and bridge foundation (P1–P3) | Legacy v2/v3 locked enrichment and cross-document references cannot map. | Source-contract fragments, citations, and non-binding byte/hash evidence survive deterministic bridging. | Unavailable/zero-operation documents remain accounted without blocking unaffected operations. |
| Existing-lane projection foundation (P4–P6) | Cited scalar/pagination/form/multipart/binary/no-body/GraphQL lanes and blocked rows are omitted. | Existing closed lanes are selected by evidence; commands/docs expose an implemented or typed nonimplemented result. | Method alone cannot select a lane; destructive/reverse-ETL surfaces preserve approval semantics. |
| Serialization/media/schema/response foundations (R1–R4) | Exact source-shaped query/cookie/media/union/recursive/response-variant fixtures fail. | Closed source-cited contracts route only to exact encoders/selectors. | Unknown/open/raw transport remains refused before I/O; provider defaults are not guessed. |
| Bounded Batch-1 cohort | The cohort matrix lacks a one-to-one source identity disposition/reachability record. | Definitions and generated surface reconcile every cited identity across the ten named connectors. | Built binary reaches `missing --credential` for implemented dispatch; blocked/unsupported rows show typed cited outcomes. |
| Delivery-policy correction | Static policy scan still finds a one-connector/separate-foundation requirement in active delivery guidance. | Active delivery guidance requires a bounded named cohort, immutable ledger, Atlas disposition, and preserved gates. | Legacy/archived cleanup guidance stays historical and is not reactivated. |

## CLI parity checklist

- [ ] `pm <connector>` contextual discovery / help exits successfully where a namespace exists.
- [ ] `pm help <connector>` and selected `pm <connector> <path> --help` show generated source-backed outcomes.
- [ ] `docs/cli/**`, `website/**`, generated help/manual indexes, and discovery metadata are updated or explicitly not applicable per connector lane.
- [ ] JSON stdout/stderr, credential boundary, and reverse-ETL plan → preview → approval → execute behavior are asserted without provider I/O.

## Commit checkpoints

1. Policy correction: RED static policy scan → GREEN active-policy scan and document/config parsing → REFACTOR for one consistent bounded-cohort vocabulary.
2. Matrix/report: RED missing or omitted source identities → GREEN source/mapping/reachability report for all ten connectors.
3. Foundation slices: commit each coherent red, green, refactor group with its Atlas and matrix reconciliation evidence.
4. Final exact-SHA review/fix/re-review is required before normal fast-forward push.
