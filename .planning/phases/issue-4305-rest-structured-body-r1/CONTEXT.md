# Issue #4305: Declaration-bound REST structured body

## Discuss-phase outcome

The accepted scope is a shared foundation, not a connector lane. The provider
declaration remains the complete authority for request method, route, media
type, metadata mapping, body schema, nesting, and limits. A caller may provide
only values for declared input fields; it cannot supply a raw body, a JSON
template, request metadata, or a different command/action identity.

The existing typed-write approval and confirmation workflow remains in place.
The approval identity must include the canonical materialized structured
payload, so an input changed after planning cannot be executed.

## Decisions

- Use a single schema-aware body materializer for the CLI command runner and
  typed-action executor.
- Represent nested input only through declared typed fields; do not add a raw
  JSON escape hatch.
- Validate and bound payload shape before the request is constructed or
  transport is called.
- Use a source-like synthetic test bundle rather than modifying a production
  connector definition owned by a connector lane.
- Document the downstream composition rule for reconciliation lanes.

## Manual GSD fallback

Issue #4305 is not a phase in the repository's current roadmap, and this
Firstmate task forbids compatible isolated role spawning. The generated
discuss-phase, plan-phase --tdd, execute-phase, verify-work, and code-review
prompts are therefore executed inline with their evidence recorded in this
phase directory.
