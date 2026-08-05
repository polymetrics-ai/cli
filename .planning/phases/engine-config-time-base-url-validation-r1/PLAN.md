# PLAN — engine configuration-time spec-constraint validation

## Scope

Foundation lane on branch `fm/cli-engine-config-time-base-url-validation-r1`.

The only product change is configuration-time validation of constraints declared
by connector `spec.json` files, at `App.AddCredential`. The change must reject
an invalid declared `base_url` before a credential is persisted and give an
actionable, secret-safe error naming both the field and the violated constraint.

Owned paths:

- `internal/app/**` — the `AddCredential` boundary and its regression tests.
- `internal/connectors/**` / `internal/connectors/engine/**` only where needed
  for a definition-driven validation seam and its unit tests.
- `.planning/phases/engine-config-time-base-url-validation-r1/**`.

Out of scope:

- credential storage, vault behavior, secret values, or secret validation;
- connector bundles and connector-local policy;
- runtime request execution, overlays, command surfaces, help/docs, and
  website generation;
- the three first-group foundations (`rest_write`, schema-deriver union arms,
  and connector mechanisms). This branch will rebase onto those before merge.

## GSD mode and required skills

- `scripts/gsd doctor`: passed.
- `scripts/gsd prompt plan-phase engine-config-time-base-url-validation-r1 --skip-research`:
  generated and executed inline.
- `scripts/gsd prompt programming-loop init --phase engine-config-time-base-url-validation-r1 --dry-run`:
  unavailable because the adapter does not register `programming-loop`.
- **Fallback:** manual GSD universal programming loop, with strict red → green
  evidence, the TDD ledger, and verification artifacts retained. The fallback
  weakens no quality or human gate.

Loaded per `.agents/agentic-delivery/references/required-skills-routing.md`:
`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, and
`golang-testing`.

## Evidence-first gate

The brief requires reproducing the gap before selecting the production seam.
The first source change is therefore a red app-level regression test using the
real bundled GitHub spec, whose `base_url` declares `format: "uri"`. It must
show that `AddCredential` currently accepts `base_url=not-a-uri`, then prove
the credential was not refused before persistence.

No implementation design or production edit may start until that red evidence
is recorded in `TDD-LEDGER.md`.

## Verified starting state

| Claim | Evidence |
| --- | --- |
| `AddCredential` validates only a local `path` special case before `vault.Put` and state save | `internal/app/app.go` |
| The declarative engine compiles `spec.json` but `newRuntime` only materializes defaults/interpolates | `internal/connectors/engine/{bundle,read}.go` |
| `format` is currently accepted as an annotation, not evaluated by `Schema.Validate` | `internal/connectors/engine/schema.go` |
| Current `spec.json` declarations contain URI/date/date-time formats, 2 patterns, and 16 enums | read-only inventory over `internal/connectors/defs/*/spec.json` |
| No current `spec.json` declares numeric/string bound keywords (`minimum`, `maximum`, `minLength`, `maxLength`) | same inventory |

## Planned sequence

1. **Red — reproduce the real boundary failure.** Add and run the app-level
   regression test; record its exact failure before production edits.
2. **Survey — establish the actual declared constraint set.** Record every
   constraint family and whether it is configuration-shaped. Do not invent
   validation rules for keywords absent from connector specs.
3. **Design after red evidence.** Reuse the loaded engine schema rather than
   duplicate connector-specific rules. The selected seam must let a connector
   with no declared configuration constraint remain genuinely unconstrained;
   it must not make a no-op forwarder look like validation coverage.
4. **Green — enforce only the surveyed declarative constraints at
   `AddCredential`, before the vault or state mutation.** Errors must be
   field/constraint-specific and must never echo values.
5. **Regression matrix.** Prove URI (`base_url`), pattern, enum, each declared
   format sibling, unconstrained inputs, and no-persistence-on-rejection. Add
   any narrowly needed schema unit tests.
6. **Refactor and verify.** Run focused app/engine/connectors tests, the
   executable configuration-time path, and the repo-required scoped gates.

## Constraint and compatibility principles

- Connector definitions are the only source of policy. No hard-coded provider
  names or field names are permitted.
- A constraint-free connector/configuration must retain today's acceptance
  behavior. Requiredness, unknown-field rejection, typed storage conversion,
  and secret handling are not silently bundled into this feature.
- Validation happens before `vault.Put` and before credential state mutation;
  a rejected configuration leaves no credential record behind.
- Constraint errors disclose the field and declared constraint, not the value.
- Bounds unsupported by the credential map representation, or absent from all
  connector specs, will be reported explicitly rather than represented by a
  misleading generic success path.

## Rebase plan

This is merge group two. Before handoff, rebase onto the three group-one
foundations and rerun the focused regression matrix. Keep the diff limited to
the configuration-validation seam and tests so conflicts remain mechanical.
