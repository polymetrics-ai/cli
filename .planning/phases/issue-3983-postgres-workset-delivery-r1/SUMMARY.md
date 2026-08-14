# Summary — Issue 3983

Implemented the sealed workset-to-managed-target bridge. `ChangeDeliveryPlan`
derives the only admitted `incremental_upsert` write plan from a verified
immutable workset, mapping, and asserted PostgreSQL target. Its one-shot
approval binds the workset bytes as well as target control identity.

The controller delivers only the workset delta and explicit tombstones through
the existing `DatabaseWriteExecutor`. It stores a candidate baseline only after
the target commit and delivery ledger yield `DeliveryReceiptV1`; a failed or
unknown commit retains the prior baseline and preserves the immutable workset
identity for replay.

The opt-in PostgreSQL dbtest proves rows omitted from a later source projection
remain in the target, while a sealed explicit tombstone deletes its composite
mapped key. It also proves receipt and candidate-baseline persistence.
