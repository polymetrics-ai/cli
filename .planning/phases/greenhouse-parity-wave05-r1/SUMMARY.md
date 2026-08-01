# Summary — Greenhouse parity wave05 r1

Implemented connector-local Greenhouse parity refresh from the official Harvest HTML documentation.

## Final counts

- Official documented operations enumerated: 135
- Implemented/fixture-tested: 133 = 69 streams + 64 typed writes
- Blocked/planned: 2 binary attachment upload operations
- Excluded/not-applicable: 0
- Certified/live-safe: 0

## Notable changes

- Made `base_url` version-neutral (`https://harvest.greenhouse.io`) and made all bundle paths explicitly `/v1` or `/v2`.
- Added fixture-backed typed writes for v2 job posts, v2 scheduled interviews, v2 user edit/disable/enable, v2 opening destroy, and candidate-tag destroy.
- Removed unsupported attachment upload write actions and marked those official operations blocked due unbounded base64/external-URL binary payload semantics.
- Corrected hiring-team write paths to the official `/v1/jobs/{id}` endpoints while leaving the read stream on `/v1/jobs/{id}/hiring_team`.
- Tightened destructive/admin schemas and retained explicit destructive confirmation where appropriate.
- Updated Greenhouse docs, generated manual/skill, and connector catalog counts.

## Verification

See `VERIFICATION.md`; final `make verify` passed. No live provider calls, credentials, pushes, PRs, or no-mistakes pipeline runs were performed.
