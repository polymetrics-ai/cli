# Summary — engine configuration-time spec-constraint validation

## Delivered

- `App.AddCredential` now invokes declared connector configuration validation
  after the existing local-path safety check and before vault/state mutation.
- The engine compiles property `format` and validates supplied top-level
  credential-map fields against their declared `format`, `pattern`, and `enum`
  constraints.
- `engine.Connector`, `engine.Base`, and promoted native definition wrappers
  expose the same optional contract. Its declaration signal is false when no
  configuration constraint exists, so constraint-free connectors retain their
  prior behavior.
- Failures name the field and declared constraint without echoing the input.

## Evidence

- The red GitHub `base_url=not-a-uri` reproduction was committed before the
  implementation and is retained in `TDD-LEDGER.md`.
- The matrix covers URI, date, date-time, both declared patterns, engine and
  Tier-3 Base enums, no persistence on rejection, a constraint-free connector,
  and the promoted-native forwarder.
- All scoped tests and individual repository gates passed; see
  `VERIFICATION.md`.

## Survey and deliberate omissions

The current 550 connector specs declare 554 URI formats, 81 date-time formats,
20 date formats, 2 patterns, and 16 string-only enums across 673 fields. No
constrained field is secret, and no spec declares a numeric, string, array, or
object bound. Bounds were therefore not invented: a future string bound needs
explicit semantics/tests, while a numeric bound also needs a parse/coercion
contract. Requiredness, types, unknown fields, defaults, and secret handling
remain outside this flat configuration-map feature.

## Handoff

This is merge group two. The branch was rebased cleanly onto current `main`
(`d30dd4905`) and the focused app/engine/native-set matrix passed afterward.
The three group-one foundations were not yet on `main`; rebase once more when
they land, then rerun that focused matrix. No connector bundles, vault/storage
code, CLI surface, user-facing docs, or website files were changed.
