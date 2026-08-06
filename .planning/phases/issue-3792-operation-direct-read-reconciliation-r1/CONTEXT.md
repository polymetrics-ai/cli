# Context — #3792 operation direct-read preflight and surface reconciliation

## Decision record

The captain retasked this isolated lane to build two shared foundations, in
order, without promoting any Zendesk Support or HubSpot command:

1. Runtime preflight must prove an implemented operation-backed `direct_read`
   command resolves to a bounded, executable `rest_read` or `provider_search`
   contract.
2. A generator tool must derive `api_surface` operation-row coverage and
   blocked reasons from that same runtime preflight result. It may write
   `covered_by` only when a reachable command passes preflight; otherwise it
   computes a precise current blocked reason or refuses an indeterminate row.

The intended first operational report is a check-only scan of the six
connectors with stale #2985 language. It does not modify their ledgers.

## Non-goals

- No Zendesk Support, HubSpot, Asana, Bitbucket, Freshchat, or YouTube
  Analytics command promotion in this change.
- No hand-authored `api_surface` reason update.
- No credentialed request, response fixture containing a secret, new
  dependency, raw request surface, or output-redaction change.

## Lifecycle fallback

`scripts/gsd doctor`, the required command-source lookups, and
`go run ./cmd/agentcontractgen check` passed. The local Pi adapter cannot
provide the compatible isolated GSD workers required by the official prompts,
so this single worker records and performs the discuss/plan/execute/verify
steps inline. The shared-path foundation makes role spawning incompatible with
the repository's single-owner delivery contract.

## Required skills loaded

`golang-how-to`, `golang-cli`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, and `golang-documentation`.

The CLI help/docs/website parity reference was read. `connectorgen` is a
maintainer generator rather than a `pm` end-user command; this phase updates
its help and migration authoring reference, then verifies that no generated
`pm` help/manual/website surface is affected.
