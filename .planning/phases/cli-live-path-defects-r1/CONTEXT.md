# Context: live-path defects r1

The Firstmate launch brief and issues #4119, #4125, and #4169 define this
non-interactive corrective phase. The three defects are deliberately delivered
as independent, reviewable commits on a single branch so that a larger-than-
expected slice cannot delay the other two. The task explicitly excludes #4158
(the PostgreSQL managed-target control assertion).

No product decision remains open:

- rate-limit admission must use the canonical route actually sent, including
  redirects;
- a declared shared-rate-limit window must be rejected before either duration
  or Redis-TTL conversion can overflow;
- an upstream authentication rejection must be surfaced as a credential error,
  while a true internal failure remains internal.

The phase runs as an inline/manual GSD fallback: this runtime cannot provide a
compatible isolated GSD worker and the task contract prohibits role spawning.
That records the execution mechanism, not a waiver of the lifecycle or TDD.
