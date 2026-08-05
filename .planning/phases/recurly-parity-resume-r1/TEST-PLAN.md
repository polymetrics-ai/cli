# Test plan — Recurly parity-resume r1

1. Use focused `connectorgen validate` as the red/green check for recovered bundle integrity and
   required command input mappings.
2. Run fixture-backed Recurly conformance for streams and typed write request shapes.
3. Exercise commandrunner preflight globally and Recurly binary commands specifically with bounded
   fixture/replay transport; do not contact Recurly or use credentials.
4. Build `pm`, inspect generated help, and run representative safe commands with synthetic config.
5. Regenerate and verify only Recurly-owned documentation/website surfaces after command metadata
   stabilizes.
6. For the review-fix round, add focused regressions for decimal command coercion, declarative-write
   retry/idempotency behavior, replayed response headers, and selected OAS contract invariants.
   Apply all fixes before running one final focused `go test` command over only the touched packages.
7. For the retry-and-mutation-controls round, cover explicitly idempotent unkeyed deletes, an
   unmarked delete's single-attempt behavior, write-fixture query matching, typed Recurly query
   schemas and flags, explicit refund/redaction choices, raw operation evidence, and query-bearing
   fixtures in that same final focused command.
8. For the termination-charge round, prove an omitted charge is rejected by validation and preview,
   require the generated boolean query/CLI/evidence contract, and include connector validation plus
   the focused Recurly contract tests in one end-of-round command.
