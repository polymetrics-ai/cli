# GitHub certification candidate generation — discussion log

The binding brief settles the implementation decisions:

1. Candidate generation is the mechanism; bulk hand-authoring is disallowed.
2. The declared CLI surface is the source of command path, flags, intent,
   availability, stream, and API surface. Connector-owned configuration may
   supply only fixture bindings that cannot be inferred from the surface.
3. Assertions must target produced response values below `/response`; exit
   status is never certification evidence.
4. All 1,571 declared commands remain exactly once in the generated sweep;
   every state has a concrete machine-checkable reason.
5. The live proof starts with the perishable cross-scope families and stops
   with an explicit `needs-decision` only if the measured command shapes cannot
   be generated without a new, materially different design.

No open product choice remains before the small generated-candidate measurement.
