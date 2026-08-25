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
  Stripe bundle validation, `connectorgen declaration-admission`, the
  no-credential `pm stripe accounts delete --json` boundary, formatting, and
  the applicable generator/docs/canon checks.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Deferred source mapping has a CLI projection | fake | A hermetic cited bundle has a deferred delete command that resolves to a typed missing-foundation error. A provider call cannot establish this schema invariant. |
| Policy-only suppression is rejected | fake | A hermetic declaration bundle mutates only the foundation component to a policy value; it is rejected before any provider request. A provider call cannot establish this schema invariant. |
| Existing implemented delete remains runnable | live | GitHub's cited `label delete` declaration admits and the no-credential binary dispatch reaches `missing --credential`. |

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

## TDD execution slices

1. **Red — admission contract:** Add focused table-driven tests for a cited
   runnable read; deferred reverse-ETL write/delete; deferred binary
   download/upload; importer/descriptor gap; missing/duplicate/stale/base-path
   mismatch; false implementation; and an all-deferred zero-runnable bundle.
   The tests should fail because `declaration-admission` and its schema/type do
   not exist.
2. **Green — shared declaration checker:** Add the optional, versioned
   `sources/<connector>-declaration-admission.json` sidecar and a
   `connectorgen declaration-admission [defs-dir] [--json]` command. It checks
   only opt-in sidecars, deterministically cross-links source identity, lane,
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
