# GitHub slice 4 live certification

## Scope

Certify the read-side commands in `slice-4-user-copilot.json` against the
disposable GitHub identity and publish only schema-v2, redacted evidence under
`internal/connectors/certifications/evidence/`. Mutation certification is
explicitly deferred until the captain decides whether an agent may self-approve
a reverse-ETL plan token.

## GSD execution record

This run used the inline/manual GSD fallback because this shell runner cannot
launch the Pi workflow's interactive runtime. The adapter was validated with
`scripts/gsd doctor`; command sources and prompts were resolved for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review`. Required skills used: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`.

## TDD / evidence plan

- Red: a newly captured record must fail the certification-matrix gate until it
  has the schema-v2 scope, observed-operations credential proof, a redacted
  request/response exchange, and a safe run identifier.
- Green: execute the declared `pm github` read against the disposable identity,
  assert a produced response property (never just exit status), write one
  uniquely run-scoped record, and pass `go run ./cmd/connectorgen
  certification-matrix --check` immediately.
- Cleanup: for every contained fixture probe, issue provider-side cleanup
  directly and independently read back the relevant collection. The final
  read-backs for user keys, signing keys, GPG keys, social accounts, blocks,
  and follows were all zero.

## Non-goals

Do not regenerate shared sweep or matrix artifacts. Do not perform any
mutation while the approval-token decision is pending.
