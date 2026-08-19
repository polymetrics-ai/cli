# Discussion log — issue 4264

The supplied issue brief resolves the relevant design choices:

1. Operation-backed mutations need their true action kind, not an empty value.
2. The operation surface is the second declarative source after `writes.json`;
   connector-specific naming heuristics are out of scope.
3. An undetermined mutation is a product defect expressed as a generator error,
   never an empty or silently substituted `custom` value.
4. The proof is deterministic generation over GitHub and Zoom bundles, not a
   credentialed or destructive connector run.

No further product decision is required for this narrow shared foundation fix.
