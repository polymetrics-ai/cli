# TDD ledger — Issue #4380 Asana source-to-seven-lane matrix

## Red

- Remove one `lanes.sync_transport` cell from a copied matrix candidate. The focused structural assertion must reject the missing cell and name the source ID/lane.
- The baseline contract's ETL source counts (`expected: 12`, `unmapped_mapping: 12`) are insufficient against the descriptor's 64 pagination facts; a focused assertion must expose that mismatch before the contract correction.

## Green

- The finished matrix has 249 unique rows, each with all seven lane keys and exact source citations.
- Descriptor pagination identifies 64 ETL candidates: 12 implemented declaration bindings and 52 `mapped_unproven` rows; no other GET is promoted merely because it is GET.
- The lock/API-surface/write evidence resolves 119 direct reads, 130 direct writes, and 130 reverse-ETL cells. `createAttachmentForObject` retains both action variants but remains one source operation.
- The event contract resolves exactly `getEvents`, `getTask`, and `getTasks` for sync transport.

## Refactor

- Stable-sort rows by source ID and lane keys in the fixed canonical order. Keep repeated disposition reasons in the matrix vocabulary rather than hiding rows behind aggregate counts.
