# Summary — #3979 PostgreSQL gap-free bootstrap

## Delivered

- PostgreSQL-private `BootstrapCDC` creates an `EXPORT_SNAPSHOT` logical slot, imports that stable snapshot into the existing bounded typed page plan, and only starts pgoutput-v2 after snapshot durability and initial checkpoint persistence.
- The opaque bootstrap checkpoint binds initial LSN, system ID, timeline, publication, relation, and schema fingerprint. Resume rejects drift with a typed rebootstrap outcome.
- The snapshot receiver receives cloned records plus a full stageable candidate checkpoint. Existing pgoutput-v2 transaction staging continues to supply durable receipt before checkpoint/acknowledgement.
- PostgreSQL 14+ dbtest blocks snapshot delivery, mutates rows concurrently, and proves snapshot plus changefeed equals the final source relation exactly. It also injects snapshot and checkpoint failures and proves explicit rebootstrap rather than silent slot reuse.

## Deferred

- #3983 owns target delivery; #4094 and #4095 own keyed target/history and CDC-delete binding. This source-only coordinator does not construct a destination path.
