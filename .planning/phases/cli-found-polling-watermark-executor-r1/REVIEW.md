---
phase: cli-found-polling-watermark-executor-r1
status: clean
depth: standard
files_reviewed: 14
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
review_mode: inline-manual-fallback
---

# Code review — polling-watermark changefeed executor

The generated `scripts/gsd prompt code-review 3855` workflow was reviewed and
executed inline. The workflow would normally spawn `gsd-code-reviewer`, but
this isolated issue worker is prohibited from role spawning; the manual
fallback preserves the same standard-depth source, security, and cross-file
review scope.

## Scope

- declaration schema and `ChangefeedDescriptor` semantic validation;
- shared executor ordering, checkpoint commit ordering, cancellation, and
  record extraction paths;
- capability projection through definition, list, catalog, and manifest;
- test-only bundle and regression coverage;
- migration authoring documentation.

## Resolved during review

1. The executor initially accepted `DeletionRecords` from a source even where
   the declaration had no `deletion_endpoint`, which could have made a
   `not_available` hard-delete contract emit tombstones. It now fails closed
   before emitting or committing, with a direct regression test.
2. An uncheckpointed timestamp read initially derived a boundary from the
   current clock. That could silently skip historical records. An initial read
   now passes no boundary to the closed source adapter; safety lag applies only
   after a committed timestamp exists. A direct regression test locks this.

## Result

No remaining critical, warning, or informational findings. The focused engine
and connector suites pass after the review fixes; broader local gates are
recorded in `VERIFICATION.md`.
