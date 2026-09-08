# Recurly historical regression inputs

Exact unmodified bytes from the source_commit in provenance.json (before0b214b79 deleted fixtures). The selection is the complete input family consumed by the four surviving regression roots: all93 declared stream first pages including all five legacy streams, list_sites second page, the two mutation-query writes, and the operation-scoped retry research record. Other write/check fixtures are not restored.

Each file is bound to its historical path, size and SHA256. Source facts and sanitized expected records are historical independent expectations, not regenerated from current schemas. The retry record references23 operation coordinates and three parameter coordinates within an OAS whose SHA it records; those references are document coordinates, not filesystem dependencies. The raw OAS is not retained in Git, so this restores the recorded historical claims, not proof of current provider accuracy. No provider request was made. No DELETE idempotency is inferred.

These files are test-only and absent from defs.FS. No runtime artifact or source.lock is introduced. The existing tests still validate current execution schemas, query controls, legacy projection and operation-specific retry declarations against these independent historical inputs.
