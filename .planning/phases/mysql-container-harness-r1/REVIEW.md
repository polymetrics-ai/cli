# Review — MySQL container harness R1

`scripts/gsd sources code-review` and its generated prompt were resolved. The official workflow
expects a GSD reviewer worker, which this task's single-worker contract does not permit, so the
review was completed inline over the changed Go, bundle, test, generated-doc, and dependency files.

## Review focus

- Docker commands always receive a caller-supplied context; no global context mutation or Podman
  command remains in the harness.
- Generated resource names are owned by the run; cleanup is idempotent, continues after an error,
  and removes an image only when this run pulled it.
- Colima reset follows Docker cleanup and remains an explicit destructive opt-in incompatible with
  keeping the image.
- The native MySQL path validates identifiers, uses parameters for cursor values, never exposes
  caller configuration in identifier/connection errors, and has no write operation.
- The binlog declaration, Go closed vocabulary, executor descriptor, checkpoint timing, and live
  row-event test all agree.
- The only new direct dependency is the approved MySQL client/replication module; final module
  verification and vulnerability scanning are clean.

## Findings

The follow-up review repaired native registry installation, ambiguous Docker resource ownership,
complete primary-key-tiebroken read paging, binlog row-format fail-closure, per-row CDC dedupe
state, and CDC readiness synchronization. The focused repair test is recorded by this gate; the
outer pipeline owns the remaining test, lint, build, and review phases.

Verdict: review findings addressed; GitHub automated-review routing remains a PR-stage responsibility.
