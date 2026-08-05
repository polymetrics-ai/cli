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
| Current `spec.json` declarations contain 554 URI, 81 date-time, and 20 date formats; 2 patterns; and 16 enums | read-only inventory over every `internal/connectors/defs/*/spec.json` root property |
| No constrained property is a secret | same inventory (`x-secret: true` count: 0) |
| Every declared enum value is a string | same inventory (`non-string enum` count: 0) |
| No current spec, including nested nodes, declares a numeric, string, array, or object bound keyword | recursive read-only inventory over every `spec.json` (`minimum`, `maximum`, `exclusive*`, `minLength`, `maxLength`, `multipleOf`, `minItems`, `maxItems`, `minProperties`, `maxProperties`: 0) |

## Post-red survey and selected seam

The complete currently declared, credential-map-compatible constraint set is:

| Family | Declarations | Examples used by tests |
| --- | ---: | --- |
| `format: uri` | 554 | `github.base_url` |
| `format: date-time` | 81 | `github.since` |
| `format: date` | 20 | `google-search-console.start_date` |
| `pattern` | 2 | `agilecrm.domain`, `dockerhub.docker_username` |
| `enum` | 16 | `coin-api.environment`, `postgres.mode` |

`postgres.mode` is deliberately included because Postgres is a Tier-3 native
that embeds `engine.Base`; this proves the seam applies to both declarative
and Base-backed connectors rather than only `engine.Connector`.

The selected design is deliberately targeted rather than a call to the
existing full-instance `Schema.Validate`:

1. Compile and retain each property's declared `format`, alongside the
   already compiled pattern and enum.
2. Add a `Schema` configuration validator that visits only supplied top-level
   `map[string]string` fields with a declared `format`, pattern, or enum.
   It sorts keys for deterministic first errors and returns an error that
   names the field and the actual declared constraint without echoing input.
3. Add an optional connector contract whose `HasConfigurationConstraints`
   signal distinguishes a real declaration from an unconditional no-op. The
   app invokes the validator only when that signal is true. `engine.Connector`,
   `engine.Base`, and the promoted-native `definitionConnector` forwarder all
   expose the contract from the same compiled schema.
4. Invoke that contract in `App.AddCredential` after existing local-path
   safety checks and before vault writes or state mutation.

This intentionally does **not** enforce `required`, types, unknown fields,
defaults, or secret values: the accepted credential configuration is a flat
string map today, and those would alter existing behavior or cross the
storage/secret ownership boundary. Bounds are also deliberately omitted:
there are zero such declarations in every current `spec.json`, and the only
already-supported bounds (`minItems`/`maxItems`) describe JSON array instances,
not scalar credential-map strings. A future string bound needs an explicit
semantics-and-test addition; a numeric bound additionally needs a deliberate
parse/coercion contract before it can claim support.

## Planned sequence

1. **Red — complete.** The app-level regression test reproduced the real
   boundary failure before production edits; its exact failure is retained in
   `TDD-LEDGER.md`.
2. **Survey — complete.** The actual declared constraint set and deliberate
   omissions are recorded above; no absent keyword received invented policy.
3. **Design after red evidence — complete.** Reuse the loaded engine schema
   rather than duplicate connector-specific rules. The optional validator
   advertises actual declarations through `HasConfigurationConstraints`, so a
   constraint-free connector remains genuinely unconstrained rather than
   looking auto-satisfied by a no-op forwarder.
4. **Green — complete.** Enforce only the surveyed declarative constraints at
   `AddCredential`, before vault or state mutation. Errors are
   field/constraint-specific and never echo input values.
5. **Regression matrix — complete.** URI (`base_url`), both date formats,
   both patterns, engine/Base-backed enums, unconstrained input, no-persistence,
   and promoted-native forwarding all have red/green coverage.
6. **Refactor and verify — complete.** Focused and scoped app/engine/connector
   tests, the full CLI package, build/vet, and individual repository gates all
   pass; details are in `VERIFICATION.md`.

### Forwarder audit

The post-green registry audit found a second generic seam:
`native/nativeset.definitionConnector` wraps promoted native connectors with
`engine.Base` to supply their bundle definition. It must forward this optional
capability from that Base rather than leaving the wrapper unable to honour a
future declared constraint. Its methods will report the Base's actual
`HasConfigurationConstraints` result, not blanket success. The current survey
shows all 29 promoted-native specs are constraint-free, but the test constructs a
constrained wrapped bundle to lock the contract before it is needed. The
red-forwarder test now passes with that delegation in place.

## TDD gate tracking

- [x] Task: R1 type: tdd — reject a declared configuration constraint before
  credential persistence, beginning with the real GitHub `base_url` URI case.

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

This is merge group two. The branch was rebased cleanly onto current
`origin/main` commit `d30dd4905`; its focused engine/app/native regression
matrix passed again afterward. The three stated group-one foundations were not
yet on `main` at that check, so rebase once more onto them before integration if
they land later. Keep the diff limited to the configuration-validation seam and
tests so conflicts remain mechanical.
