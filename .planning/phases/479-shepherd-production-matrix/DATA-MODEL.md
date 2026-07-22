# Data model

The durable executable plan remains schema 2. Bootstrap uses a schema-1 semantic proposal with a source
revision digest and `Omit<ProductionChildSpec, "issue">` children. Returned canonical GitHub issue
numbers are inserted only by the host. Verification results retain stable command ID/status/failure kind;
raw bounded output is transient and excluded from durable controller state.

