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

## Slice: PR #595 convention CI repair

Planned checks before workflow edits:

- `branch-name` should reproduce the CI failure for `fm/cli-pm-broker-contract-fixtures-r1`
  because the workflow only allowed Conventional Commit prefixes and a few exact exceptions.
- `require-linked-issue` remains governed by `.github/workflows/pr-issue-guard.yml`; this branch-name
  repair does not add an `fm/*` exemption to that guard.

Red evidence:

- CI `branch-name` failed with `Invalid branch name: fm/cli-pm-broker-contract-fixtures-r1`.
- Local `PR_BODY="$(gh pr view 595 --json body --jq .body)" go run ./cmd/prissueguard --title "feat(pmbroker): add synthetic broker v1 contract fixtures"` exited `1` with `PR body must reference an issue`, confirming the issue-first guard stayed enabled.

Green evidence after workflow edits:

- Branch policy shell check passed for `fm/cli-pm-broker-contract-fixtures-r1`, still accepted
  `feat/valid-branch`, and still rejected `feature/not-valid`.
- Local issue-guard behavior still required issue-first linkage for `fm/*`, matching
  `.github/workflows/pr-issue-guard.yml`.
- `go test ./cmd/prissueguard ./internal/coordination/issueguard` passed, confirming the issue
  guard binary/package behavior itself was not loosened.
- `git diff --check` passed.
