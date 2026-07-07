# Test Plan

- Engine unit tests:
  - GraphQL variables resolve `query.*` values.
  - GraphQL template variables with `omit_when_empty` are omitted on empty cursor.
  - Loader rejects non-boolean `omit_when_empty`.
- Bundle tests:
  - GitHub loads new GraphQL project/discussion streams.
  - CLI surface maps read commands to streams instead of operation executor rows.
- Validation:
  - `jq` parses JSON files.
  - `connectorgen validate` passes for GitHub.
  - GitHub conformance passes or records any dynamic GraphQL fixture limitation.
