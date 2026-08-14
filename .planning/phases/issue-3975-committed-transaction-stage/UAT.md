# UAT — Issue #3975 committed-transaction staging and durable receipts

## Automated acceptance result

**PASS for this non-interactive private foundation slice.** There is no UI,
provider mutation, credential, database, source checkpoint, or human-judgment
step in the #3975 scope.

Automated tests directly observe the durable state transitions rather than
relying on exit status: staged chunks are unreadable by the receiver before
commit; commit delivers exactly one ordered whole transaction; a receipt is
read back from a newly opened stage root before it can derive a downstream
acknowledgement; and abort/cancel/fault paths leave no deceptive receipt or
temporary residue. Restart recovery retains sealed receipt-less work only and
cleans it once a valid receipt is present.

This accepts only the source-agnostic stage/receipt seam. It does not certify
PostgreSQL CDC, authorize source LSN acknowledgement, enable a capability, or
authorize push, PR creation, CI handling, or merge.
