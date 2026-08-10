# Issue #3993 — offline harness completion summary

## Delivered

- The case generator now requires an immutable `Polymetrics-Cert` boundary.
  It accepts a frozen case ledger when available to calculate reason-family
  movement, but a live case file no longer depends on a stale checked-in
  ledger. It no longer knows the
  former personal owner, repository, or commit SHA. Its output carries the
  run-owned organization/repository immutable IDs and reports the
  attemptable/blocked movement by reason family from the supplied ledger.
- The live sweep rejects case files whose slug or immutable IDs do not match
  the boundary, queues all eligible operations behind one barrier, records the
  barrier release in its report, and requires an independent read-back for a
  write before a write process may run.
- The generated lab manifest and bootstrap probe inventory now derive from all
  **1,521** current implemented GitHub commands. Their current cohorts are
  900 run-owned repository, 308 run-owned organization, 33 GitHub App/
  installation, and 280 feature/entitlement commands. Both self-checks pass.

## Honest certification status

No credentialed GitHub operation or warehouse flow was run in this worktree.
The reported measurement artifact is not present here; the checked-in older
case ledger contains 1,081 commands and cannot be substituted for the
captain's stated frozen 1,521-command measurement. The generator therefore
accepts a frozen ledger as an explicit optional comparison input and reports
the live case classification independently.

The required ephemeral credential/proof path is owned by #3989 and the shared
REST/GraphQL rate-admission path by #3990. Both are prerequisites for a real
barrier release, quota measurement, cleanup/residue proof, and the requested
attemptable/blocked movement report. No vault credential was created and the
revoked fine-grained token was not touched.

The inbound GitHub → warehouse → DuckDB leg is consequently **not reproven**
in this branch. The outbound warehouse → GitHub leg remains intentionally
unimplemented pending #3994's approved-action path and #3992's schedule
execution; `pm flow` still identifies action steps as approval-gated.

Certification remains unclaimed until those prerequisites land and a
credentialed run records real provider results, per-operation read-backs,
cleanup, residue, and quota-bucket failures.
