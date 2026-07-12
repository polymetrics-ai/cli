# Observability Plan: Phase 0

- Safety status and replay produce bounded deterministic JSON suitable for local inspection.
- Drivers emit one typed denial to stderr and do not create a log file while closed.
- Violation codes have low cardinality and never include paths, prompts, commands, identities from
  real runs, or sensitive values.
- Tests assert stdout/stderr separation and byte-stable JSON. No telemetry, model call, network
  request, or persistent operational event is added in this phase.
