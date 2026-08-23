---
coverage:
  - id: D1
    description: CLI stdin credential values preserve valid bytes while rejecting normalized-empty input before persistence.
    verification:
      - kind: integration
        ref: internal/cli/credential_coordination_cli_test.go: TestCredentialsAddStdinPreservesSingleTerminalDelimiterAndRoundTrips; TestCredentialsAddStdinRejectsEmptyNormalizedSecretBeforePersistence
        status: pass
    human_judgment: false
  - id: D2
    description: App/vault and every selected required shared authentication route reject empty credential material without emitting a credential request.
    verification:
      - kind: unit
        ref: internal/app/secret_store_test.go; internal/connectors/connsdk/auth_test.go; internal/connectors/connsdk/refresh_token_test.go; internal/connectors/engine/auth_test.go
        status: pass
    human_judgment: false
  - id: D3
    description: CLI, generated manual, and website documentation describe the stdin normalization contract.
    verification:
      - kind: other
        ref: TestGoldenDocsGenerateMatchesTrackedCLIManuals; docs-check-no-build
        status: pass
    human_judgment: false
---

# Summary — provider-neutral non-empty credential foundation

## Delivered

- Added `internal/credential`, a provider-neutral contract that normalizes at
  most one stdin LF/CRLF delimiter, rejects persistent empty values with a
  typed non-secret error, and rejects blank material at authentication time.
- Enforced it at CLI stdin/environment intake, App and vault persistence, the
  vault-backed runtime secret store, declarative engine selection, static
  bearer/basic/API-key authenticators, and the required OAuth grant material.
- Preserved byte-exact non-empty persistence, long stdin inputs outside argv,
  optional no-auth routing, and OAuth refresh grants with their documented
  optional public-client secret.
- Updated CLI help, generated CLI manual, golden help transcript, website
  source, and generated website docs. No Twenty connector path changed.

## Consumer handoff

Twenty CRM PR [#4298](https://github.com/polymetrics-ai/cli/pull/4298) must
integrate foundation commit **`5321e06f886839b4531aa36b90ef513d2acf3bb7`**
(`fix(credentials): reject empty credential material`). This is the exact code
and verification commit; it has no provider-name special case and touches no
Twenty bundle path.

The current branch adds this handoff record in a follow-up documentation commit.
It is not a merge, push, or PR creation.

## Lifecycle

The issue-first GSD lifecycle was executed inline because the canonical worker
contract forbids spawning lifecycle roles here. `PLAN.md`, `TDD-LEDGER.md`,
`VERIFICATION.md`, and `REVIEW.md` contain the discuss/plan/execute/verify/
review evidence, including the red and green steps.
