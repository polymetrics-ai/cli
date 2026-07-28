# CLI PM Broker Contract Fixtures R1

## Objective

Add a CLI-side, synthetic-only PM Broker `/v1` contract foundation that later profile,
context, and execution lanes can import without contacting a live broker or handling credentials.

## GSD and skills evidence

- GSD adapter: `scripts/gsd doctor` succeeded. `scripts/gsd prompt programming-loop init --phase pm-broker-contract-fixtures-r1 --dry-run` is unavailable in this adapter (`unknown GSD command: programming-loop`), so this slice uses the documented repo-local GSD manual fallback with `scripts/gsd prompt gsd-quick "CLI PM Broker contract fixtures and fake-broker client foundation"`.
- Required routing read: `.agents/agentic-delivery/references/required-skills-routing.md` and `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Skills loaded: `gsd-core`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`.

## Scope

- Add an internal Go package for pinned PM Broker `/v1` JSON fixtures, typed contract shapes,
  and a deterministic in-memory fake broker client/transport.
- Pin PM Broker PR #35 compatibility behavior: typed `/v1` operations require
  `PM-Broker-API-Version: 1.0`; missing or unsupported values return HTTP 426 with
  `incompatible_contract_version` and the accepted safe error envelope.
- Cover success, incompatible-version refusal, safe correlation IDs, opaque references, no raw
  secret exposure, and absence of generic request escape hatches with network-free tests.
- Document how later profile/context/execution CLI lanes should consume this package while the
  public authentication registry remains future work.

## Non-goals and safety

- No normal user command behavior changes.
- No provider SDKs, production resources, live GCP/VPS logic, credentials, service-account JSON,
  arbitrary authenticated HTTP, raw-secret retrieval/export, SQL, shell, runtime plugins,
  arbitrary headers/endpoints, or generic JSON/body execution.
- Tests remain synthetic, deterministic, network-free, and credential-free.

## Implementation plan

1. Create `internal/pmbroker/contract/v1` with constants, closed typed JSON structs, validation,
   deterministic fixtures, and an in-memory fake broker transport.
2. Add black-box tests that first fail against the absent package, then pass after implementation.
3. Add package documentation for future CLI profile/context/execution consumers and update the
   PM Broker CLI integration plan with the new package pointer.
4. Run targeted package tests, gofmt, and broader Go verification as practical before commit.
