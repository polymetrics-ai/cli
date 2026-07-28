# TDD Ledger

## Slice: synthetic PM Broker `/v1` contract package

Planned red tests before production code:

- `TestSyntheticClientSuccessPinsAcceptedFixtures` should assert the deterministic fake client
  returns accepted compatibility, scoped resources, opaque secret references, and execution-plan
  metadata matching PM Broker PR #35 fixtures.
- `TestSyntheticClientRefusesIncompatibleContractVersion` should assert typed operations with a
  missing or unsupported `PM-Broker-API-Version` get HTTP 426 and exact code
  `incompatible_contract_version` without execution fallback.
- `TestContractSafetyInvariants` should assert safe correlation IDs, opaque references,
  no raw-secret markers, and no generic request escape hatches.

Red evidence:

- `go test ./internal/pmbroker/contract/v1` failed before production code with `no non-test Go files`, proving the contract package API was absent.

Security-review hardening:

- Added negative coverage for unsafe display hints and unsafe contract-version header values.
- Enforced non-exportable opaque references, safe display-hint markers, and pinned broker-profile connector-kind metadata.

Green evidence:

- `go test ./internal/pmbroker/contract/v1` passed.
- `go test ./internal/pmbroker/...` passed.
- `git diff --check` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed.
