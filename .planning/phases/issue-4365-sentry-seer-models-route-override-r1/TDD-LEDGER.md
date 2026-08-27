# Issue #4365 TDD ledger

## Red — planned before production edits

| Slice | Test | Expected initial failure | Status |
| --- | --- | --- | --- |
| Happy | `TestSentrySeerModelsSourceBoundRoute` | Sentry has no operation, CLI command, named route, or direct-read ledger row. | pending |
| Bad | `TestSentrySeerModelsRouteRejectsIdentityDrift` | The source-bound route contract is absent, so mismatch cases cannot resolve a materialized Sentry operation. | pending |
| Edge | `TestSentrySeerModelsRoutePreservesPathAcrossBaseSlashForms` | Sentry has no typed route operation to invoke against the local server. | pending |
| CLI boundary | `TestSentrySeerModelsCLIRequiresCredentialBeforeTransport` | The generated Sentry command does not yet exist. | pending |

## Green/refactor record

Update each row with its exact test command, failure output summary, minimal
declaration change, and green result. The final row must record the built-binary
command, exact stderr assertion, and spy request count.
