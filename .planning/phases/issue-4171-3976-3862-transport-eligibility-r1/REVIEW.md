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

# Code review — transport source eligibility club (#4171, #3862; #3976 deconflicted)

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
- PostgreSQL polling is not advertised prematurely. The real CLI inspection guard remains exact:
  it requires `planned` plus a blocking reason until a shipped preflight can bind source, object,
  and destination. The overlapping native/shared-poller work is absent from this branch and owned
  by PR 4175.
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

## Gap G2 follow-up review

The review confirmed that `app.Open` composes only the outward closed transport. The attempted
`engine.PollingPreflight` occurred inside source `ReadTransport`, after authentication admission and
typed-catalog I/O. Keeping an implemented declaration would therefore violate the inspection
invariant; restoring `planned` and not duplicating PR 4175 is the narrow correction.
