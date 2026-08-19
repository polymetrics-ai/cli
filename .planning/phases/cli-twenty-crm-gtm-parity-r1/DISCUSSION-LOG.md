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
- Source-audit confirmation: Twenty documents both REST and GraphQL as
  workspace-schema-generated Core APIs. The source-locked #277 inventory is
  the workspace-specific REST contract; no GraphQL or Metadata API operation
  is inferred from its REST counterpart. This boundary is recorded in the
  machine-readable ledger.
- Completeness decision: `write_eligibility.json` gives all 112 typed actions
  one explicit disposition. `create_companies` is bound as the reversible
  declaration proof; 55 create/update actions have schema-intersecting,
  closed candidate mappings but await #4304's exact-action multiplicity;
  28 batch actions require a `records` array envelope and 28 deletes require
  a tombstone workset, neither representable by the current single-record,
  no-tombstone contract. These are transport semantics, never safety,
  privilege, or destructive exclusions. Every action remains directly
  CLI-reachable through its implemented reverse-ETL command.
- Captain decision (2026-08-20): keep all 55 eligible actions required. Wait
  for #4304's persisted multi-action selection and exact per-action bindings;
  do not introduce a Twenty-local workaround. The final live gate must use the
  dedicated credentialed instance through the registered `pm-twenty`
  Keychain/secret-store reference, test the built CLI and persisted App path,
  and treat the prior direct-stream run as historical evidence only.
- Runtime discovery result (2026-08-20): captain authorized read-only Docker,
  Podman, and protected-tree inspection. There is no deployed Twenty container
  or Compose project on the local Docker daemon, Podman is unavailable with
  both local machines stopped, and the protected tree exposes no non-exporting
  credential or deployment reference. The lane remains blocked on a dedicated
  disposable running instance plus an execution-time stdin/FD/approved
  secret-store handoff; no production or unrelated local container may be used.
- Official-recipe review result (2026-08-20): current upstream commit
  `e14694f4ff9ca51b791ba6b09639fed0944c5ad7` declares a self-contained
  production Compose topology, but its production entrypoint migrates the
  database only. The official setup path requires the first workspace through
  the interactive application flow and does not document noninteractive
  workspace or API-key issuance. The upstream development seed is not part of
  the production recipe. Decision: do not start a stack, edit its database,
  bypass auth, or use an unreviewed browser-extracted credential. Remain
  blocked until an authoritative noninteractive disposable bootstrap or a
  registered non-echoing credential reference is available.
