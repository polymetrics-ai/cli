# TDD Ledger — CLI required-flag derivation r1

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Repository required-path invariant | `go test -timeout 20m ./cmd/connectorgen -run TestRequiredRESTPathParametersAlwaysMapToRequiredCLIFlags -count=1` reported 92 GitHub violations | Same test returns zero across all 552 bundles after derivation and regeneration | green |
| Surface-sync derivation | Focused fixture omitted path-flag requiredness and failed | Fixture verifies both a corrected `required: false` and filled field become `true`, while a required query flag stays optional | green |
| Typed pre-I/O refusal | The GitHub fixture previously reached late `missing path variable "pull_number"` validation | `MissingRequiredFlagError` produces `category=usage`, `code=usage_error`, exit 2, and zero fake-provider calls | green |
| GitHub P1 count | `certification-sweep --connector github --check` reported 92 product defects on the base | Regenerated sweep reports zero product defects (104 flag fields in 92 commands) | green |
| Unsupported declaration audit | No verifier existed for the 50 declared entries | Audit lists every entry without mutation: 26 `unsupported_api` declarations contradict the source lock; 23 `unsupported_local` declarations hold | green |
| Website generated-data gap | PR 4209 `Website generated data` CI check failed | Two repository-generator passes leave `website/**` clean. The first pass produced no website diff, so it could not absorb unrelated or required-flag changes. | green |

## Constraints

- No connector-specific source identifiers or boundary allowlist edits.
- No hand-editing generated artifacts.
- No credentialed provider operation is necessary for deterministic derivation or typed pre-I/O refusal coverage.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-documentation`.
