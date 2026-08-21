# Twenty CRM shared-foundation handoff

## Boundary

This is a provider-neutral transport capability gap. It must not be implemented
in the Twenty connector lane. Twenty's all-operations acceptance criteria are
unchanged: all 55 remaining record-shaped mappings, all 28 documented batch
array-envelope actions, and all 28 typed destructive tombstone actions remain
required and CLI-reachable.

The local Twenty checkpoint is `7059fe85a`; it and the protected disposable
runtime/vault state remain unpushed and untouched. This handoff contains no
credential, record identifier, endpoint override, request body, or runtime
state.

## Source traces

| Source | SHA-256 | Exact role |
| --- | --- | --- |
| `internal/connectors/defs/twenty/write_eligibility.json` | `8b9ac70d6f07a44edc708cabaaa9f2c216f14f922440fdef3d2f81b16842b8da` | Action membership and exact candidate input-to-source-field mappings. |
| `internal/connectors/defs/twenty/writes.json` | `5188fe0dea80e93471b476e91c75a33e2b6ca7fc8c9835144b87591acf7535f8` | Provider-owned typed action schemas, methods, paths, and body fields. |
| `internal/connectors/defs/twenty/sync_transport.json` | `d4b10c0d26ce159d8c6c9fb7d2e4c740332adc7e28aeee869f19388eaf51bc58` | Current declarative source/destination contract and its one binding. |

Provider source remains Twenty's documented API revision
`e14694f4ff9ca51b791ba6b09639fed0944c5ad7`. The source-locked REST cohort is
168 operations: 28 ETL streams, 28 direct reads, and 112 typed reverse-ETL
actions. The documented REST inventory has no direct-write or binary-transfer
operation; GraphQL/metadata schema is workspace-generated and was not inferred.

## Current contract and executable red witness

`DestinationSourceBinding` has only `executor`, `eligible_streams`, and one
`record_mapping` (`internal/connectors/sync_transport.go:143`). Validation
prohibits two bindings for the same executor (`:436`), and
`SourceBindingFor(source, stream)` takes no selected action (`:450`). The App
first resolves the persisted action, then retrieves that action-independent
binding (`internal/app/declarative_typed_destination_approval.go:274`).

The connector-local witness is:

```text
go test -count=1 -timeout 20m ./internal/connectors/defs/twenty \
  -run '^TestPerActionSourceBindingFoundationHandoffWitness$'
```

Its initial red expectation failed before provider I/O: a declaration extended
in memory with `update_companies` successfully selected that action, but its
only binding was `name` → `name`; `update_companies` requires `id` → `id` and
`name` → `name`. The committed witness now asserts this negative fact while
also proving the happy `create_companies` mapping and the edge refusal of an
undeclared action.

This is not a request for generic action, HTTP, body, shell, or SQL input. The
persisted connection action remains declaration-owned. The missing contract is
the action-sensitive selection of an already declaration-owned source mapping.

## Required foundation contract

The foundation must let one destination use multiple declared actions from the
same source executor and stream, with each action selecting its own closed
`input_fields` mapping. It must retain these invariants:

- The action is persisted in the connection stream; callers cannot supply or
  override action, URL, method, body, source mapping, connector, or evidence.
- The selected action resolves exactly one matching declaration-owned mapping.
- Every mapped input exists in that selected action's typed schema and every
  required selected input is mapped from the declared source record contract.
- Missing, foreign, duplicate, stale, cross-connector, cross-action, or
  mismatched selections fail before source or provider I/O.
- The sealed approval, acknowledgement, output projection, and reopened-App
  state retain the exact action and mapping identity.

The foundation lane owns the data-shape design. This handoff requires the
observable contract above, not a specific field name or JSON representation.

## Dependency membership

| Gap | Exact Twenty selector | Count | Required closure evidence |
| --- | --- | ---: | --- |
| Per-action source bindings | `$.actions[?(@.disposition == 'eligible_pending_foundation_multiplicity')]` | 55 | At least two actions on one executor/stream with different mappings each select their own mapping; all 55 become definition-bound and persist/reopen with exact acknowledgements. |
| Batch array-envelope delivery | `$.actions[?(@.disposition == 'semantic_array_envelope_incompatible')]` | 28 | Bounded grouping produces the provider-owned `records` array; empty, over-limit, malformed, cross-action, and post-approval mutation fail before I/O; acknowledgement accounts for every member. |
| Tombstone workset delivery | `$.actions[?(@.disposition == 'semantic_tombstone_incompatible')]` | 28 | Typed confirmation, plan/preview/approval seal, workset ownership, idempotency, acknowledgement, and independent absence verification for only declared records. |

All 112 reverse-ETL commands remain `implemented`; the three groups are a
foundation execution gap, never an operation exclusion. Destructive,
privileged, or uncommon actions remain reachable and retain their existing
typed safety gates.

## Foundation-lane test matrix

| Case | Assertion |
| --- | --- |
| Happy | Two declared actions sharing one source executor/stream but with different `input_fields` resolve their own action-specific mapping, execute through the persisted App path, and retain acknowledgement after reopen. |
| Bad | Missing action, duplicate mapping, foreign action, mismatched selected-schema input, missing required input, and a forged persisted selection fail before source/provider I/O. |
| Edge | Empty action continues to work only for an unambiguous single-action mode; multi-action modes require the persisted selection and cannot default by mode. |
| Batch | Exact declared maximum envelope succeeds; empty, 61st member, malformed member, and post-approval mutation fail before I/O. |
| Tombstone | Missing destructive confirmation, stale approval, cross-workset replay, and unknown action fail before I/O; a successful lane-owned delete verifies independent provider absence. |

## Re-entry criteria for Twenty

Resume the connector lane only after a published, reviewed foundation head
satisfies the matrix, is normally merged with an ancestor proof, and its own
checks are green. Then bind every 55 mapping, refresh the ledger hash, run the
fresh credentialed disposable-instance proof, and complete the all-operations
certification. No fixture-only substitute is acceptable.

## Handoff verification

- `jq` validates `FOUNDATION-GAPS.json` and `FOUNDATION-HANDOFF.json`.
- The current `write_eligibility.json` SHA matches the snapshot plus each of
  the three exact gap selectors (55/28/28).
- `TestPerActionSourceBindingFoundationHandoffWitness` passes as the stable
  negative witness after its recorded red run.
- Focused Twenty bundle tests, the persisted-destination App tests,
  `connectorgen validate`, `surface-sync --check`, and the Twenty certification
  sweep all pass.

No full verification, push, live credential use, or runtime mutation occurs in
this handoff-only slice; final certification remains gated on the foundation.
