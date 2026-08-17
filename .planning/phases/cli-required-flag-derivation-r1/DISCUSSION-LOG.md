# Discussion Log — CLI required-flag derivation r1

- 2026-08-17: Used the supplied P1 brief and shared production-parity context as the discussion record.
- Decision: requiredness is inferred once from `operations.json` REST path parameters in `connectorgen surface-sync`; no connector-specific branch, identifier, or boundary allowlist entry is permitted.
- Decision: declarations marked `unsupported_api` or `unsupported_local` are audited and reported, not rewritten by this lane.
- Decision: test-first execution covers the generic generated-surface invariant and the runner's typed, pre-I/O rejection.
