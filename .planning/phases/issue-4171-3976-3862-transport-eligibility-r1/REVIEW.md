---
status: clean
depth: standard
files_reviewed: 20
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
review_mode: inline_manual_fallback
---

# Code review — transport source eligibility club

The lifecycle reviewer role was not spawned: this direct-PR issue club requires a single autonomous
worker and the project contract forbids lifecycle role spawning. The official code-review prompt was
resolved and applied inline.

## Reviewed invariants

- `Registry.Preflight` remains definition- and exact-reference-led. Its positive stream allowlist
  refusal happens before executor lookup, destination plan, source read, warehouse stage, apply, or
  checkpoint mutation.
- GitHub's eligibility is concrete, duplicate-free, no-wildcard, and exactly matches executable
  declaration streams. `commits` reaches the definition-owned declarative executor through
  `app.Open`; `ISSUES` and other absent values stay typed refusals.
- The general declarative source keeps provider pagination in the engine, bounds emitted batches,
  and reconstructs resume only from its own candidate identities.
- PostgreSQL performs typed catalog discovery before building an identifier-safe lexicographic query
  over cursor plus every primary-key tie field. Invalid/nullable/lossy cursor or key shapes reject
  before reads. The native adapter invokes `PollingPreflight` and the shared polling executor rather
  than introducing connector-name routing or a generic query seam.
- The orchestrator still stages, reopens, applies, reads back, and only then commits a checkpoint;
  cancellation and replay tests cover the acknowledgement boundary.

## Result

The review identified a missing case-equivalent stream edge assertion. It was added to
`TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess` and rerun under normal
and race detection. No critical, warning, or informational finding remains. The external live
certification is tracked as pending verification evidence, not a code-review finding.

## Gap G1 follow-up review

`stages_transport_internal_test.go` continues to prove both sides of the declaration registration
contract: the successful probe must report the exact active source and destination references, and
the no-factory probe must report that same active source reference as unregistered. Updating those
two literals from the replaced `issue_label_source` to `declarative_stream_source` preserves the
guard's observable success and failure assertions; it neither relaxes nor removes them.
