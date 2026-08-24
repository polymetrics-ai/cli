# Discussion log — issue 4338

## Inputs and decisions

- Provider source is authoritative: Asana has 25 absent-action mutation cases,
  Jira has 16 incomplete-contract cases, Sentry has 34 POST/PATCH cases, and
  Vercel has 159 POST/PATCH cases. The expanded consumer set is roughly 234
  mutations across four connectors; fixtures prove per-operation behavior and
  no connector-level/count-only waiver exists.
- A mutating provider operation without a complete action must remain a
  source-traced runtime gap; it cannot disappear as a read-only or generic
  waiver classification.
- Read-only is a sibling foundation, not a fallback: it is restricted to
  non-mutating operations and must reject POST, PUT, PATCH, and DELETE.
- The model must be generic so Sentry can later hold read-only and mutation-gap
  dispositions for different operations within one connector. Sentry SCIM
  PATCH and dashboard POST fixtures are explicitly mutation-gap inputs.
- The Vercel handoff is a batch-1 authoring decision, not automatic permission
  to declare 159 writes absent. A cited disposition is correct only for an
  operation batch 1 intentionally leaves non-executable; an operation meant
  to work needs its own complete write action and remains rejected here.
- No connector definition changes belong to this PR. The foundation is limited
  to shared `cmd/connectorgen` code, behavioral tests, and delivery evidence.

## Rejected alternatives

- Treating a mutating source operation as read-only: this bypasses mutation
  coverage and the write-safety boundary.
- Marking a working command `partial`: this weakens a valid executable claim.
- Creating a synthetic write/action contract: this fabricates an unsupported
  request shape instead of recording the source gap.
