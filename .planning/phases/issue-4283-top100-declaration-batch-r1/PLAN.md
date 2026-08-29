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
| Legacy canonical-evidence projection | `source.retention-import.v1`, `source.projection-admission.v1`, `runtime.direct-execution.v1` | `constrained_extension` | When a structurally sufficient v2/v3 source lock retains its closed provider `source_contract` plus the exact operation inventory/citations, import that retained contract directly for mapping. Hash and raw-artifact availability are provenance/verification evidence, never admission gates. The importer cross-checks identity, URL, location, method, original provider route, mapped connector route, and counts against the lock. The selected Batch-1 connector projection must materialize only source-complete direct-read and reverse-ETL shapes through the existing closed runtime, while every other operation remains a per-row source-cited foundation or evidence disposition. This is generic source-lock behavior, not a GitLab branch or a new runtime foundation. |
| GitLab complete eligible-lane projection | `source.projection-admission.v1`, `runtime.direct-execution.v1` | `constrained_extension` | GitLab's 1,752-row canonical descriptor materializes only fixed JSON GET/direct-read and no-body reverse-ETL contracts through existing closed lanes. The projection uses mapped connector routes while retaining provider routes/citations, exact 2xx statuses, and finite response bounds; every non-materialized identity retains a cited runtime gap. |
| GitLab provider-key alias candidate | `runtime.provider-parameter-alias.v1` | `actual_gap` (`investigating`; captain approval required) | Fifteen new direct-read candidates require a reversible, closed CLI-name to exact provider query-key alias for punctuation-bearing source keys. The shared engine has no such seam. The descriptor and operation-evidence rows retain the exact request-phase/location/foundation outcome; no runtime/CLI alias, raw transport, or connector-name branch is implemented. An existing `getApiV4Issues` ETL lane remains source-backed with its documented zero-filter request and is not blocked by optional query keys it does not expose. |
| GitLab endpoint-ledger propagation | `source.projection-admission.v1` plus `runtime.direct-execution.v1` | `constrained_extension` of an available foundation, not a planned runtime capability | A generated direct read already has a closed `rest_read` executor; it failed before credential acquisition only because `internal/connectors/defs/operation_endpoint_ledger.json` lacked the generated `GET /admin/active_context/connections` entry. `cmd/connectorgen/surfacesync.go#syncRuntimeOperationEndpointLedgerConnector` now refreshes exactly one selected connector row while preserving every unselected row. The existing runtime consumer is `internal/connectors/engine/operation_endpoint_ledger.go#OperationDirectReadEndpointLedgerEntries`; `TestGitLabGeneratedDirectReadReachesCredentialBoundary` proves the built command reaches `missing --credential` without provider I/O. No engine executor, request encoding, approval, credential, or provider-specific branch is added. |
| P3 source-projection document isolation | `source.projection-admission.v1` | `constrained_extension` | A v3 lock's unavailable document is evidence owned by source import/operation evidence; it must not prevent source-projection from validating and materializing independently retained operations from the lock's other documents. No unavailable identity is promoted or omitted. |

## Planned TDD sequence once scope is valid

| Slice | Red | Green | Edge/refactor proof |
| --- | --- | --- | --- |
| Source-normalization foundation (F1/F2) | Representative locked aliases/numeric/default/duplicate/range cases fail. | Presence-aware canonical source facts retain raw evidence and exact semantic status behavior. | `default` alone remains blocked; declared 404/non-2xx never become executable. |
| Retention and bridge foundation (P1–P3) | Legacy v2/v3 locked enrichment and cross-document references cannot map. | Source-contract fragments, citations, and non-binding byte/hash evidence survive deterministic bridging. | Unavailable/zero-operation documents remain accounted without blocking unaffected operations. |
| Non-binding evidence resilience | Removing valid v2/v3 retained-source, certification-subject, or website evidence aborts accounting or leaves an identity executable. | Every identity remains in lane accounting with the matching named source-retention, projection, or certification gap. | Existing declarations stay visible; missing retained-source input disables only the affected mapping row, while certification/website gaps remain overlays. No provider I/O or connector-specific branch. |
| P3 source-projection document isolation | A v3 lock containing one unavailable document plus one closed available operation makes `checkSourceProjection` return the unavailable-document finding before validating the available operation. | The available descriptor operation validates normally; unavailable-document evidence stays in the source-import/operation-evidence outcome rather than blocking an unrelated projection. | A descriptor mismatch for the available operation still fails; an all-unavailable lock remains represented by its source-import evidence and invents no executable operation. |
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

## GitLab live reconciliation ledger — pending scoped commit

Generated from the refreshed connector-only artifact with `go run ./cmd/connectorgen operation-evidence --connector gitlab --output <temp>/gitlab-operation-evidence.json`, after `source-import gitlab --materialize-eligible-lanes --check`:

| Cell | Measured count |
| --- | ---: |
| Locked source rows | 1,752 |
| Runtime-enabled source rows | 733 |
| Enabled direct-read classifications | 582 |
| Enabled ETL classifications | 4 |
| Enabled reverse-ETL classifications | 147 |
| Blocked rows with one or more named foundations | 1,019 |
| Blocked rows without a named foundation | 0 |
| Evidence-defect rows | 0 |

The per-row gap ledger has 2,306 `source_contract` entries because foundations overlap; it is not a row denominator. Its foundation counts are malformed path parameter 2, request encoding 47, request schema 1,243, closed source operation execution 143, provider parameter alias candidate 15, and non-executable mutation 856. The earlier provisional arithmetic classified individual shapes and included 16 punctuated-query source routes; the refreshed artifact is authoritative: its 1,019 blocked identities plus 733 enabled identities reconcile to 1,752. One punctuated route, `getApiV4Issues`, remains enabled as the existing `issues list` ETL lane with no exposed optional unsafe key, so it is not one of the 15 new direct-read alias-candidate gaps.
