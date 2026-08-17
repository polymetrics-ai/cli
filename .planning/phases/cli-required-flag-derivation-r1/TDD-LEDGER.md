# TDD Ledger — CLI required-flag derivation r1

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Repository required-path invariant | New all-bundle test fails, enumerating optional flags mapped to required REST path parameters | Same test returns zero findings after generic derivation and regeneration | planned |
| Surface-sync derivation | Fixture expects required path flag while current output omits it | Generator derives `required: true`, but leaves optional query parameters optional | planned |
| Typed pre-I/O refusal | Missing path flag test proves the current late path-variable error or absence of validation | Typed usage error names the CLI flag and transport call count is zero | planned |
| GitHub P1 count | Generated sweep reports 92 GitHub findings on the base | Same sweep reports zero GitHub findings after generated surface update | planned |
| Unsupported declaration audit | No verifier exists for the 50 declared entries | Programmatic report names all declarations and flags any contradiction without mutation | planned |

## Constraints

- No connector-specific source identifiers or boundary allowlist edits.
- No hand-editing generated artifacts.
- No credentialed provider operation is necessary for deterministic derivation or typed pre-I/O refusal coverage.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-documentation`.
