# Observability

This slice does not add runtime telemetry. Existing read errors carry connector, stream, page, and
record index context. GraphQL `errors[]` responses fail closed with summarized provider messages.
