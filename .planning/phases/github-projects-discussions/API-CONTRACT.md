# API Contract

## Engine

GraphQL variables may reference:

- `config.*`
- `secrets.*`
- `record.*` for writes only
- `cursor`
- `incremental.lower_bound`
- `fanout.id`
- `query.*` for read command/query parameters

Template object shape:

```json
{
  "template": "{{ query.number }}",
  "type": "integer",
  "omit_when_empty": true
}
```

`omit_when_empty` only applies to template objects. It omits the variable if interpolation resolves
to an empty string. It does not suppress unresolved-key or type-conversion errors.

## GitHub Streams

All new streams use `POST /graphql` with fixed documents and JSON output. No generic GraphQL escape
hatch is added.

