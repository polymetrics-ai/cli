# Discussion log — Issue 4001

The issue and captain brief supplied the material decisions needed for this foundation:

- The contract must serve both database and API destinations, so it cannot live in a provider or
  database package.
- Configuration failures must be structurally non-retryable.
- Dispatch analysis needs five exact machine-readable kinds but implementation of the call-graph
  analyser remains #3991's scope.
- Certification needs structured, serializable untestable reasons without serializing causes or
  sensitive values.
- #3998 is in flight and must remain the owner of database declaration validation. Its future
  consumer integration is not silently implemented in this branch.

No product choice remained that required a human decision.
