# Inline code review — issue 4329

## Scope and safety review

- Exactly one operation-model enum member was added: `read_only`.
- The meta-schema remains closed; an unknown model is rejected.
- The test round-trips every prior operation model as exact JSON bytes.
- Source projection binds the marker to the same source method/path and rejects
  any mutating source method before it can suppress coverage.
- A declared executable route is contradictory, not ignored.
- Evidence reports valid read-only rows separately, with connector-keyed
  rollups; it does not emit a foundation gap for them.
- No connector definition, source lock, GitHub rate limit, credential, runtime
  state, generic writer, or generated artifact was changed.

## Review route

Open the direct main-targeted PR as ready for review. The trusted-author route
is `claude_auto`; collect and disposition its review records after GitHub
starts the workflow.
