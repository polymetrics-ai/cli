# GitHub certification candidate generation — context

**Delivery:** direct PR to `integration/4015-mvp-flat-r1` from
`fm/cli-candidate-generation-perishable-r1`.

**Target connector:** GitHub only. The production implementation is limited to
the connector-owned definition and the generic certification tooling that reads
declared surface metadata. No shared connector runtime policy may name GitHub
identifiers.

## Locked decisions

- Generate certification candidates from declared `cli_surface.json` metadata;
  hand-authored candidates remain only for a named shape the generator cannot
  express.
- A candidate is not a pass. Live certification requires a real invocation,
  produced-value assertions, provider evidence (or an exact non-pass reason),
  independent cleanup where a fixture is created, and accepted evidence from
  the separate evidence-importer lane.
- The proof slice is the 192 trial-perishable commands: Advanced Security 50,
  Copilot 50, Enterprise 46, Codespaces 46. The connector-owned literal
  cohort declaration is the auditable source of membership. Its 97 direct
  reads are generated and executable; its 95 reverse-ETL commands are
  explicitly deferred to the separate mutation-fixture lifecycle delivery.
- Advanced Security is live-runnable: the fixture repository now has Code
  Security, secret scanning, push protection, validity checks, non-provider
  patterns, and Dependabot security updates enabled. A code-scanning 404 with
  `no analysis found` is a missing-analysis fixture result, not an entitlement
  blocker.
- Provider refusals and product defects are separate terminal non-pass states.
  A failed required path-flag declaration is a product defect, not an API
  refusal.
- Requests are serial and resumable. Credentials are loaded only into an
  exported environment variable by command substitution and never printed,
  persisted, committed, supplied in argv, or written to this phase material.

## Manual lifecycle fallback

`scripts/gsd doctor`, command-source resolution, and all five generated GSD
prompts were completed. This repository has no roadmap phase corresponding to
this direct-PR task and the delivery contract prohibits role spawning, so the
phase uses the documented inline/manual GSD fallback. The historical
`task-delivery-header-template.md` path is absent on the integration base; this
context plus PLAN.md is the delivery header fallback.

## Required skills

Loaded before implementation: `golang-how-to`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, and `golang-structs-interfaces`.

## CLI help/manual/website parity

Not applicable to the public `pm` command surface: the change is internal
connector certification tooling and connector definition metadata. If a new
user-facing `pm` command, flag, help topic, or output shape becomes necessary,
stop and add the full CLI-help/docs/website parity work rather than hiding it
behind the generator.
