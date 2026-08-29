# Issue #4283: Batch-1 source-rigidity R2 plan

## Task Delivery Header

- Issue: Refs #4283 — Batch-1 source-rigidity R2.
- Base branch: `main`.
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main`.
- Delivery: Extend existing PR #4294 only; exact reviewed SHA is normal-fast-forward pushed to `fm/cli-top100-declaration-batch-r1`, without merge or new PR.
- Working branch: `fm/cli-top100-declaration-batch-r1` (detached assigned worktree, remote tracked explicitly).
- Task: Preserve all ten Batch-1 source-lock identities through bounded mapping, generated surfaces, and typed truthful dispositions; no provider I/O or generic transport.
- Verification: Focused red/green Go tests, connector-scoped reconciliation report, real no-I/O preflight where a row is implemented, documentation/Atlas validation, exact-SHA review, and relevant local CI gates.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `connector-lane-build-order`.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A scoped evidence bridge cannot omit or add connectors outside its selected source locks | live | `TestOperationEvidenceConnectorFilterIsBounded` checks all 249 Asana source identities and rejects a missing connector even when a valid connector is selected. |
| Legacy provider facts remain retained but non-executable | live | Source-import parser tests assert retained object fragments and reject unknown outer members/non-object fragments before any provider I/O. |

## Scope decision

**Authorized bounded cohort.** Captain instruction `006.msg` changes the obsolete delivery policy rather than waiving it. PR #4294 may contain the ten named Batch-1 connectors and only the shared foundations required to preserve their source-lock identities. Every shared change needs an Atlas `reuse`, `constrained_extension`, or captain-approved `actual_gap` disposition. The scope never authorizes unrelated connectors, generic raw transport, credential handling, destructive-operation policy, approval-policy changes, provider-live I/O, or a merge.

## Immutable cohort prerequisites

1. Keep the frozen remote SHA and exhaustive review ledger current.
2. Generate a machine-checkable source-operation-to-lane matrix for all ten locked cohorts, including direct, binary, ETL, reverse-ETL, sync transport, CLI/docs, and typed blocked/unsupported cells.
3. Reconcile each source identity to exactly one `implemented`, `blocked_with_named_foundation`, or `unsupported_with_provider_evidence` disposition; certification is overlay evidence only.
4. Consult the Foundation Atlas before every shared foundation change and update it in the same change when its real contract changes.

## Foundation Atlas dispositions

| Slice | Atlas ID | Classification | Bounded contract |
| --- | --- | --- | --- |
| P1 legacy enrichment retention | `source.retention-import.v1` | `constrained_extension` | Retain v1/v2 `source_contract` and `source_operation` object fragments as identity-attached, non-executable source evidence. The existing source-import owner remains the only implementation; no connector branch, raw request transport, or runtime admission is added. |
| P2 scoped source-lock mapping report | `source.projection-admission.v1` | `constrained_extension` | Select an explicit connector set when emitting the operation evidence bridge. It reads only their locks, rejects any selected connector without a lock, requires a non-global output, and does not alter the global fixed-100 contract or claim execution. |
| P2 non-binding evidence resilience | `source.retention-import.v1`, `source.projection-admission.v1`, `verification.conformance-certification.v1` | `reuse` | Valid v2/v3 source identities remain in operation evidence when retained-source, certification-subject, or generated website evidence is unavailable. Each affected row receives a named typed gap; no command, executor, credential behavior, or certification-based promotion is created. |

## Planned TDD sequence once scope is valid

| Slice | Red | Green | Edge/refactor proof |
| --- | --- | --- | --- |
| Source-normalization foundation (F1/F2) | Representative locked aliases/numeric/default/duplicate/range cases fail. | Presence-aware canonical source facts retain raw evidence and exact semantic status behavior. | `default` alone remains blocked; declared 404/non-2xx never become executable. |
| Retention and bridge foundation (P1–P3) | Legacy v2/v3 locked enrichment and cross-document references cannot map. | Source-contract fragments, citations, and non-binding byte/hash evidence survive deterministic bridging. | Unavailable/zero-operation documents remain accounted without blocking unaffected operations. |
| Non-binding evidence resilience | Removing valid v2/v3 retained-source, certification-subject, or website evidence aborts accounting or leaves an identity executable. | Every identity remains in lane accounting with the matching named source-retention, projection, or certification gap. | Existing declarations stay visible; missing retained-source input disables only the affected mapping row, while certification/website gaps remain overlays. No provider I/O or connector-specific branch. |
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
