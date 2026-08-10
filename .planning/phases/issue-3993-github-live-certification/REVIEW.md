# Inline code review — Issue #3993

## Scope reviewed

- `scripts/github-live-cases.mjs`
- `scripts/github-live-proof-sweep.mjs`
- `scripts/github-live-lab-manifest.mjs`
- `scripts/github-live-bootstrap-probes.mjs`
- focused Node tests and regenerated source-derived artifacts

## Result

No blocking implementation finding remains in the offline harness slice.

- Boundary admission is default-deny through `validateLabBoundary`, requires
  exactly one run-owned `Polymetrics-Cert` organization and repository, and
  binds both slugs and immutable IDs before dispatch.
- The sweep has no local concurrency cap and releases its entire applicable
  operation list through one barrier. It preserves terminal accounting and
  redacts invocation/output-derived data from reports.
- A write cannot be marked proven without an independent read-back, and a
  mismatched owner/repository case fails before a PM process starts.
- Manifest/bootstrap checks derive from current source rather than the stale
  pre-skip count. The historic ledger remains archival context only.

## Follow-up outside this change

Credentialed provider execution is blocked by #3989/#3990 and must be
reviewed again after those shared foundations land. That is a certification
evidence gap, not a reason to weaken the default-deny or read-back checks.
