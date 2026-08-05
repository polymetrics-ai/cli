# Test plan — Recurly parity-resume r1

1. Use focused `connectorgen validate` as the red/green check for recovered bundle integrity and
   required command input mappings.
2. Run fixture-backed Recurly conformance for streams and typed write request shapes.
3. Exercise commandrunner preflight globally and Recurly binary commands specifically with bounded
   fixture/replay transport; do not contact Recurly or use credentials.
4. Build `pm`, inspect generated help, and run representative safe commands with synthetic config.
5. Regenerate and verify only Recurly-owned documentation/website surfaces after command metadata
   stabilizes.
