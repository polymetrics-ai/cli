# Pre-integration Review Trace

The independent reviewer blocked integration after running the then-current focused suite. Findings
covered incompatible controller/store statuses, orphaned-run resume, stop overwrite races,
unconfined built-in read tools, missing post-run exact-head validation, shutdown/global-concurrency
gaps, and persistence allowlisting.

Every P0 finding and the applicable P1 findings became a failing regression before the fix. The
resulting design supplies host-verified bounded PR evidence to zero-tool child sessions, recaptures
it after both lanes, serializes only explicit state DTO fields, and owns one active run in the Pi
process. A fresh exact-head review remains required after the live canary and root gates.
