# GSD discussion log — Issue 3973

`scripts/gsd prompt discuss-phase 3973` was generated and executed inline.

The supplied issue and firstmate brief answer the only implementation-sensitive
questions: use the existing typed target control record and closed mode
vocabulary; require preview-derived approval; retain one session through all
batches; roll back on failure/cancellation; surface indeterminate commit
without retry; require durable receipt before checkpoint authority; and defer
all PostgreSQL driver/DDL/SQL work to #3982. No user-facing CLI surface changes.
