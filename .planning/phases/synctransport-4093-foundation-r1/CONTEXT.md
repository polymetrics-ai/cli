# Refs #4093 — Discussion context

The issue already fixes the architectural decision: sources and destinations must be connector-neutral, definitions select their evidence, and destination execution remains a closed typed action adapter. The worker therefore makes no product decision and uses the supplied acceptance criteria as the discussion result.

Non-negotiable constraints: fail closed before I/O; no connector-name switch in source/destination registration; no generic HTTP/SQL write; GitHub parity before deletion; `change_capture` is source-only; and a post-commit kill must reconcile rather than duplicate effects.
