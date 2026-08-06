# CONTEXT — caller-supplied identifier sets

## Locked decisions

- The declaration lives on `operations.json`'s `rest` block as the closed
  `caller_supplied_identifier_sets` array. It is an operation input, not a
  connector config key and not a stream or catalogue declaration.
- Each entry has a bundle-owned `name`, `element_shape`, `wire`, `min_items`,
  and mandatory positive `max_items`. This phase supports exactly
  `opaque_string` and `chain_address` shapes, and exactly `query_comma_separated`,
  `query_repeated`, `body_json_array`, and `path_segment` wires.
- A matching command flag is a required `string_array` mapped to
  `identifier_set.<name>`. Operation bounds and shape are authoritative; the
  CLI flag does not restate them. Explicit blank flags preserve a literal empty
  array when `min_items` is zero; absence remains absence.
- Validate cardinality and every element in the engine before it constructs
  the request. Diagnostics name only parameter, element position, declared
  shape, and limit; they never render an identifier value.
- `path_segment` is deliberately one-element only (`max_items: 1`). It binds a
  named path placeholder and remains URL-path encoded. There is no generic
  list-join/path dialect.
- A JSON map of identifier to timestamp array is out of scope. The r1 value is
  a flat identifier set only; nested arbitrary batch payloads remain blocked
  until a separate closed declaration and executor are designed.
- Test-only bundles under `internal/connectors/engine/testdata/` prove all four
  wire encodings. No production connector, API-surface ledger, or `covered_by`
  entry changes in this foundation slice.

## GSD execution note

The repository GSD roadmap has no phase for this firstmate-scoped foundation
task, so `gsd-sdk query init.phase-op caller-supplied-identifiers-r1` reports
`phase_found:false`. The required discuss/plan/execute/verify/review prompts
were generated and are executed inline using the repository's established
manual fallback. The canonical single-worker contract also forbids role
spawning. TDD, verification, and review requirements remain intact.

## CLI documentation scope

The only command declarations are test-only bundles. Runtime generated help,
manual pages, website docs, and bare namespace behavior have no production
surface to change. The connector-authoring convention is the applicable
documentation surface.
