# Summary — issue 4264 operation-backed mutation action kinds

## Delivered

- Extended the generated parity classifier to use the linked operation only
  when no `writes.json` action is present.
- Operation metadata now yields explicit create/update/delete action classes,
  REST `PUT` yields upsert, GraphQL mutations yield custom, and ambiguous REST
  mutations fail into a concrete sweep product defect.
- Regenerated the GitHub certification sweep; its checked artifact contains
  derived action kinds and remains byte-current.
- Verified the Zoom lane's remote declarations read-only: all 18 REST DELETE
  operations are direct-write command targets and therefore take the new shared
  delete derivation without definition changes in this foundation PR.

## Coverage

- Unit tests cover operation-backed delete, create/update/upsert/custom,
  existing writes-backed delete behavior, and the indeterminate product-defect
  path.
- Full `make verify` passed.

## Deferred by ownership

Zoom definitions and a Zoom sweep artifact are not present on this `main`
base and are owned by the concurrent Zoom parity lane. This PR deliberately
does not edit them.
