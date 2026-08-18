# Review — #3860 polling-watermark truth surfaces

## Method

- Inline/manual fallback: the project-local GSD adapter generated the required
  `code-review` prompt, but this runner has no compatible Pi role runtime and
  the canonical single-worker contract forbids spawning review roles.
- Reviewed the complete production, test, source-doc, generated-doc, and
  planning diff after fresh-binary regeneration.
- Ran `git diff --check`, focused CLI/connector/engine assertions, connector
  validation, `surface-sync --check`, documentation validation, and targeted
  Go vet as recorded in `VERIFICATION.md`.

## Findings

No actionable correctness, security, scope, or generated-artifact findings.

- Planned polling declarations serialize only `status` and `reason`; they
  cannot advertise zero-value `source`/`target` bindings.
- PostgreSQL's logical-replication CDC declaration remains separate from the
  planned bounded-polling declaration.
- The PostgreSQL native protocol inventory still has zero REST endpoints.
- The diff introduces no credentials, provider mutation, generic write tool,
  dependency, or changes to excluded issues #4125, #4136, #4090, or #4154.
