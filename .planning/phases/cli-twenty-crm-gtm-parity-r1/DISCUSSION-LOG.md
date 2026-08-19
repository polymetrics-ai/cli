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
