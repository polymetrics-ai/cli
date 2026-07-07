# Test Plan

- Engine unit tests:
  - GraphQL variables resolve `query.*` values.
  - GraphQL template variables with `omit_when_empty` are omitted on empty cursor.
  - GraphQL template variables with an explicitly-empty `req.Query` value are omitted when
    `omit_when_empty` is true.
  - Loader rejects non-boolean `omit_when_empty`.
  - Loader rejects GraphQL variable `default` values that do not match the declared `type`.
- Bundle tests:
  - GitHub loads new GraphQL project/discussion streams.
  - CLI surface maps read commands to streams instead of operation executor rows.
- Conformance:
  - `connectorgen validate` passes for GitHub.
  - GitHub conformance passes or records any dynamic GraphQL fixture limitation.
  - `read_query` fixture replay covers parameterized reads and passes (or records limitation).
