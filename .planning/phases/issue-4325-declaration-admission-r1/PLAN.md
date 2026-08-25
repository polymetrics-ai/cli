# Plan — issue 4325 declaration-admission foundation

## Task Delivery Header

- Issue: Refs #4325 — source-declaration admission certification.
- Base branch: main.
- Merges into: main.
- Delivery: Pull request #4351 is open against `main`; this expanded slice is
  committed, pushed to its existing branch, and locally verified.
- Working branch: fm/cli-declaration-admission-certification-r1.
- Task: Make declaration-first mapping semantics enforceable: every deferred
  source declaration must name an evidenced missing implementation component,
  and its command stays discoverable. A policy-only block must not be accepted
  as a deferred foundation.
- Verification: Focused declaration-admission and commandrunner tests,
  actual no-credential deferred preflight, formatting, and the applicable
  generator/docs/canon checks.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Deferred source mapping has a CLI projection | fake | A hermetic cited bundle has a deferred delete command that resolves to a typed missing-foundation error. A provider call cannot establish this schema invariant. |
| Policy-only suppression is rejected | fake | A hermetic declaration bundle mutates only the foundation component to a policy value; it is rejected before any provider request. A provider call cannot establish this schema invariant. |
| Existing implemented delete remains runnable | live | GitHub's cited `label delete` declaration admits and the no-credential binary dispatch reaches `missing --credential`. |
| Deferred command cannot hide behind a policy or excluded endpoint | fake | A hermetic bundle tries to declare a deferred command against an excluded/policy endpoint. The shared runtime resolver refuses it before it can become `missing_foundation`; no provider call can establish this local structural invariant. |
| Deferred foundation is machine-verifiable and round-trippable | fake | A hermetic source cohort, declaration catalog, and command surface carry the same typed component, evidence, exact source/binding identity, destructive semantic, and canonical target. Tests mutate each independently and observe an admission/preflight failure. |
| Citation follows the existing safe source-publication policy | fake | A hermetic source row using HTTP, userinfo, a private literal, a fragment, or a credential-shaped query is rejected by the shared local citation validator. |

## Scope boundary

This is a shared tooling/schema PR. It does not convert a connector, refresh
a provider artifact, call a provider, add credentials, or perform a
write/delete. It must not weaken `commandrunner.Preflight`, `surface-sync`,
source-lock verification, runtime certification, or live certification.

Captain clarification 007 supersedes the prior Stripe conversion direction.
The uncommitted mapping work at
`internal/connectors/defs/stripe/cli_surface.json` and
`internal/connectors/defs/stripe/sources/stripe-declaration-admission.json`
is intentionally preserved for `cli-batch1-repair-r1`; it must not be staged
or committed by this PR.

## Audit repair context (2026-08-25)

Captain inbox 008 carries an independent audit of frozen commit `3d39cc1fc`.
DA-001 through DA-007 were confirmed, but DA-008 found that a deferred command
could bypass exact API-surface/executor mapping through an excluded or policy
row, and DA-009 found that admission citations used a weaker URL policy than
source import. This repair remains a generic foundation change: it adds no
provider mapping and leaves the preserved Stripe files untouched.

The normal GSD adapter path was resolved with `scripts/gsd doctor`,
`scripts/gsd sources`, and generated `discuss-phase`, `plan-phase --gaps`,
`execute-phase --gaps-only`, `verify-work`, and `code-review` prompts. This
non-Pi single-worker session cannot run the Pi role runtime, and the canonical
contract forbids role spawning here, so the discuss/plan/execute/verify/review
work is recorded and performed inline. Required skills loaded: `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`, plus `golang-lint` for final static verification.

## Audit repair waves (TDD)

1. **Red — independent denominator and command identity:** Extend the
   hermetic catalog tests to reject duplicate exact provider operation
   identities, noncanonical/non-round-trippable command paths, and weak or
   secret-shaped citation URLs. The source row itself remains the independent
   completeness denominator; no source lock, imported bytes, or hash is read.
2. **Red — shared deferred resolver:** Add adversarial bundles whose deferred
   target is excluded, policy-only, mismatched, or multi-endpoint. The current
   resolver returns `missing_foundation` before inspecting any target, so these
   tests must fail first.
3. **Green — typed target and foundation evidence:** Carry a typed foundation
   component/evidence and exactly one target through the declaration catalog,
   CLI schema, engine projection, and commandrunner. The engine's no-network
   deferred preflight validates its target against a blocked ledger operation;
   `connectorgen declaration-admission` calls the real `commandrunner.Preflight`
   with canonical command segments and accepts only typed
   `system/missing_foundation` for a valid deferred row.
4. **Green — semantic destructive metadata:** Require DELETE rows to declare
   `kind=delete`; require any other exact target classified as
   `destructive_action` to carry destructive metadata; reject a non-destructive
   target falsely labelled delete; and require an implemented row's exact
   write/operation binding to retain the declared destructive semantics.
   Deferred metadata remains a declaration, not a runtime claim.
5. **Refactor/document/verify:** Reuse the source-import publication URL
   validator for admission citations, document all three certificates, run
   the focused red/green tests and repository gates, then conduct the inline
   code review and update the PR body.

## TDD execution slices

1. **Red — admission contract:** Add focused table-driven tests for a cited
   runnable read; deferred reverse-ETL write/delete; deferred binary
   download/upload; importer/descriptor gap; missing/duplicate/stale/base-path
   mismatch; false implementation; and an all-deferred zero-runnable bundle.
   The tests should fail because `declaration-admission` and its schema/type do
   not exist.
2. **Green — shared declaration checker:** Add the required, versioned
   `declaration_admission_sources.json` independent cohort and separate
   `declaration_admissions.json` catalog plus a
   `connectorgen declaration-admission [defs-dir] [--json]` command. It checks
   a nonzero expected source denominator, deterministically cross-links source identity, lane,
   canonical endpoint, command, destructive/delete metadata, and runtime
   state. It never fetches provider data or requires source artifact bytes,
   hashes, request schemas, or fixtures.
3. **Green — explicit deferred command state:** Extend the command surface’s
   shared deferred/foundation metadata only as needed so an admitted deferred
   command stays discoverable and `commandrunner` returns a typed
   missing-foundation refusal before any executor. Keep implemented preflight
   rules unchanged.
4. **Refactor/document/gate:** Add the Make target and concise certification
   design/canon documentation distinguishing declaration admission from
   runtime and live certificates. Run formatting, targeted tests, relevant
   generator/check targets, review, and full feasible local verification.
5. **Red/green — enforceable deferred mapping:** Require every deferred
   declaration to name an evidenced missing implementation component rather
   than a method, risk, confirmation, approval, blocked-by-default, source
   retention/hash, or live-certification policy. Prove a hermetic deferred
   delete command returns `system/missing_foundation`, while GitHub's `label
   delete` remains implemented and reaches the no-credential boundary.

## Exact-SHA re-audit repair (2026-08-26)

The independent Codex audit of `683a3c76e` reopened DA-001 through DA-006,
DA-010, and DA-011. The repair uses two required root catalogs rather than
optional connector sidecars; binds source, declaration, command, and runtime by
stable binding identity rather than method/path alone; derives destructive
semantics from the independent source row; and moves implemented binding
resolution into one engine function shared with runtime preflight. Deferred
targets now include source/binding identity and use the compact source cohort
embedded in `defs.FS`, so production preflight does not depend on the omitted
`api_surface.json`. Typed `missing_foundation` survives oversized metadata,
the App plan path preflights before credential resolution, and the public CLI
preserves its machine-readable code. Structural endpoint checks apply to new
source/deferred identities without retroactively changing legacy implemented
provider-specific command paths.

## Exact-SHA R2 gap closure (2026-08-26)

The independent Codex R2 audit of `c99e40a315b20b776a3b8653b54fc682a8469844`
found four remaining High blockers. This gap wave stays within the shared
admission/runtime seam and preserves every uncommitted Stripe and Docker Hub
path byte-for-byte.

1. **Red — provenance-only source uniqueness (DA-002):** duplicate one exact
   provider citation under different local source IDs and different runtime
   bindings. Both the authoring check and compact production ledger must reject
   it before declaration or runtime evaluation. The source key comprises only
   canonical source URL, exact document location, protocol, raw provider
   operation ID, method, and canonical provider endpoint identity; binding
   uniqueness remains a separate invariant.
2. **Red — fail-closed canonical endpoint equivalence (DA-003):** reject
   same-binding wrong-method/wrong-path aliases for stream, write, REST, and
   binary bindings. Add a locked real-bundle census for the 243 non-GraphQL
   aliases and an end-to-end GitHub `discussion list` admission case. The green
   resolver may prove only named transformations: declared base-URL path
   composition, absolute-transport normalization, query separation, positional
   placeholder equivalence, the closed operation-variant annotation grammar,
   a registered hook transport route, and GraphQL operation identity to
   `POST /graphql` transport. It must retain canonical and transport endpoints
   separately.
3. **Red — reject runnable commands relabelled deferred (DA-004):** exercise
   the actual implemented `commandrunner` preflight against a synthesized
   implemented form of each deferred command. Every foundation component must
   fail for the runnable GitHub label delete and GraphQL read/write controls.
   Remove the idempotency component because no runtime executor requires it,
   and make response-descriptor evidence REST/binary-specific.
4. **Red — exact production ledger inventory (DA-012):** require
   `declaration_admission_sources.json` to have one exact root-artifact class,
   byte attribution, and deterministic inventory while continuing to reject
   the declaration catalog, API surfaces, fixtures, and non-exempt source
   locks.
5. **Green/refactor/gate:** implement the smallest shared changes, correct the
   stale Makefile wording, update the certificate-separation docs and this
   ledger, run focused packages plus every applicable generated/static gate in
   a clean committed checkout, then execute `verify-work` and `code-review`
   inline. The generated GSD gap prompts are executed inline because the
   canonical single-worker contract and launch brief forbid role spawning.
6. **Post-repair merge integration gate (captain steer 014):** load the real
   Outreach bundle and synthesize only its missing discovery projection in the
   test. Prove the existing `/api/v2`-based `prospects` ETL stream and
   destructive `delete_account` write through declaration admission,
   canonical/transport resolution, and no-I/O commandrunner preflight. This
   supplements, and does not replace, the exact 243-alias fleet census and
   GitHub GraphQL/delete controls. No provider request or connector definition
   change is permitted.

Required skills for this gap wave: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`,
`golang-documentation`, and `golang-lint`.

Closure: all six steps are complete on clean code SHA `f97dede07`. The fleet
census proves 243 non-GraphQL and 4 GraphQL aliases, the real Outreach gate
passes, all 477 connectorgen tests/examples and every directly integrated Go
package pass, and the whole-tree boundary scan reports 0 findings without a
new exception. `VERIFICATION.md` and `REVIEW.md` contain the exact commands and
R2 disposition.

## CLI docs parity

`connectorgen declaration-admission` is an internal generator command, not a
new `pm` command. The applicable docs are its `connectorgen` usage and the
connector certification/design canon. `pm help`, bare namespace behavior,
`docs/cli/**`, website pages, manual generation, and shell completion are not
applicable. Deferred connector commands retain normal `cli_surface.json`
discovery and are covered by commandrunner tests.

## Commit checkpoints

- Plan/TDD evidence checkpoint before production changes.
- Red-test checkpoint when the repository’s test convention permits it.
- Green implementation/documentation checkpoint after targeted gates.
- Review-fix checkpoint if inspection finds an actionable defect.
