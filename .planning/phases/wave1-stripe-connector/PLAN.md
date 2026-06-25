# PLAN — Stripe connector on connsdk

Behavior tasks require red-first evidence (TDD gate).

## Wave A — read path
- [ ] id: t-stripe-read type: test — Red httptest test: `stripe.New()` reads `customers` across 2
      pages (has_more cursor) with Bearer auth, maps records, advances incremental cursor.
- [ ] id: b-stripe-read type: behavior — Implement package stripe on connsdk: New/Name/Metadata
      (caps read+write), Check, Catalog (core streams), Read via connsdk.Requester + Bearer +
      data[] extraction + has_more/starting_after pagination + StatefulReader cursor. streams.go
      mappers. Fixture mode.

## Wave B — write path + registration
- [ ] id: t-stripe-write type: test — Red test: ValidateWrite accepts `create_customer`, rejects
      unknown action; DryRunWrite returns staged count; registry resolves `stripe` with Write cap.
- [ ] id: b-stripe-write type: behavior — Implement WriteValidator/DryRunWriter + allow-listed
      write actions + payload builders; `init()` RegisterFactory; add blank import to registryset.

## Wave C — catalog flip + verify
- [ ] id: t-stripe-verify type: test — Parity: `pm connectors inspect stripe --json` → kind
      Connector read+write; conformance still green; `make verify` green.
- [ ] id: b-stripe-catalog type: behavior — Flip `source-stripe` to enabled + pm_connector_name=stripe
      in catalog_data.json; ensure NativeConformanceReports length unchanged and stripe treated live.

## Ordering
A → B → C. Heavy implementation delegated to a backend subagent; orchestrator owns red tests + gates.

## Rollback
Delete the stripe package + registryset import + revert the catalog entry; no other connector affected.
