# GSD discussion log — #277

- `discuss-phase 277` was resolved through `scripts/gsd prompt discuss-phase 277`.
- The task contract supplies the material decisions: one Twenty-only bundle; no
  foundation edits; no contact with the captain's instance; a disposable,
  self-hosted instance is explicitly authorized for live proof; deletes remain
  typed and approval-gated.
- Recovery confirms the exact source commit and records its current-main gaps in
  `CONTEXT.md` before import.
- Decision: use the current provider REST surface for executable contracts. The
  documented GraphQL endpoint is provenance context only unless a provider-owned
  contract requires it; it is not duplicated as an invented raw command surface.
- Decision: do not claim certification merely because a matrix or fixture exists.
  The current allowlist is a separate foundation boundary.
- Firstmate, not the captain, directed the switch from UI setup to a headless
  API/seed route. The captain's standing requirement remains a real-instance,
  real-data proof before merge. This distinction must be retained in the PR
  body and all delivery evidence.
- Reconciliation decision: PR #4304 is the temporary typed-destination
  foundation. It was merged without history rewrite and PR #4298 was retargeted
  through the GitHub API to `fm/cli-reverse-etl-destination-r1`.
- Reconciliation decision: declare all 28 Twenty streams as the closed
  declarative source and the reversible `create_companies` action as the one
  executable typed destination. The adapter rejects tombstones, so delete
  propagation is not claimed; all destructive Twenty delete commands remain
  implemented and user-reachable through their existing typed confirmation.
- Reconciliation decision: attachment resource CRUD is JSON REST metadata, not
  a documented file-transfer contract. The seven-surface ledger records zero
  binary read/write and direct-write operations rather than inventing them.
