# Implementing a connector with source.lock vNext

This is the required end-to-end procedure for API and database connectors.
Read the [canon index](INDEX.md) and the complete
[vNext architecture](SOURCE-LOCK-VNEXT.md) first.

## 1. Scope and foundation check

Start with one scoped issue. Record the connector, provider material revision,
intended operations, all seven lane outcomes, required runtime encoders and
executors, and the exact tests that will prove the change.

Search the [Foundation Atlas](foundations/README.md) before changing shared
runtime. Reuse an existing engine, encoder, warehouse, or sync executor when it
matches. If an actual shared capability is absent, keep the source fact in the
lock, declare the affected lane `unsupported`, name the exact missing contract,
and request approval before implementing a new shared foundation. Never fill a
shared gap with connector-specific runtime branching.

## 2. Author the single input

Create or update:

```text
internal/connectors/defs/<connector>/source.lock.json
```

Use schema version 4. Declare exactly the seven required lanes. Put provider
citations and facts in `provider_evidence` or operation-local `source` objects;
they remain authoring-only. Put execution content in metadata, config schema,
HTTP, shared schemas, canonical operations, CLI root, and optional execution
objects.

Model a provider operation once and attach every valid execution consumer to
it. For example, one collection can own a direct-read command and an ETL stream;
one typed mutation can own direct-write and reverse-ETL commands. Keep the
interactive and saved semantics distinct.

Use shared request, response, and record schema references. Requests must be
closed to documented input fields. Recursive read-only response fields do not
become request inputs. Multipart and binary contracts declare exact fields,
media, status, and byte bounds. Pagination and fan-out remain bounded.

## 3. Render execution JSON

Run:

```bash
go run ./cmd/connectorgen lock-render <connector>
go run ./cmd/connectorgen lock-render <connector> --check
```

Review both sides of the diff. Every rendered change must follow from the lock.
No provider citation or authoring evidence may appear in execution JSON.
`--check` must pass without writing and a repeated render must be byte-stable.

For the first conversion of an existing connector, author
`source.lock.json` directly from the retained provider facts. Review the first
render as a replacement of the prior bundle. There is no reverse importer from
execution JSON and no period in which both formats are mutable sources of truth.

## 4. Validate the runtime contract

The runtime filesystem must contain execution JSON only. Prove that
`engine.Load` discovers the bundle when source locks are not present in that
filesystem. Check required artifacts, schema references, route bounds, command
bindings, registered encoders/executors, invocation validation, credentials,
approval, and sync mode compatibility.

Runtime validation is allowed to reject only an actual execution-invalid
condition. It must not consult provider citations, source locks, artifact
digests, review results, or external proof state. Offline authoring diagnostics
are advisory and cannot suppress a command.

Run the structural checks:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run '^TestVNextSourceLock' -count=1
go test ./internal/connectors/defs -count=1
go test ./internal/connectors/engine -count=1
go test ./internal/connectors/commandrunner -count=1
```

## 5. Prove each implemented lane

| Lane | Minimum proof |
| --- | --- |
| Direct read | Command discovery, unambiguous binding, bounded request, fake-server wire shape, and credential-boundary reachability without external provider I/O. |
| Direct write | Closed input validation, exact encoder, preview/approval when required, fake-server request, and response handling. |
| Binary download | Declared media/status/size, bounded destination path, digest/output behavior, and fake binary response. |
| Binary upload | Declared parts/media/size, confined input file, digest/preview, approval, and multipart fake-server proof. |
| ETL | Stream/schema/pagination semantics and durable materialization into the connection-owned DuckDB path. |
| Reverse ETL | Warehouse selection, plan, preview, approval, typed action dispatch, partial-failure/idempotency behavior, and receipt. |
| Sync transport | Registered source/destination executors, compatible modes, warehouse mediation, checkpoint/acknowledgement, and rate-limit semantics when configured. |

An `unsupported` lane needs an explicit empty declaration and a truthful reason
in the issue or connector documentation. It must not have hidden execution
content.

## 6. Warehouse and safety rules

All saved movement is warehouse-mediated:

- API → DuckDB → API
- API → DuckDB → database
- database → DuckDB → API
- database → DuckDB → database

There is no source-to-destination shortcut, including a connector to itself.
Reverse delivery always follows plan → preview → approval when required →
execute. Database work uses the typed database framework; it does not add raw
SQL execution. Retry, resume, idempotency, tombstones, checkpoint advancement,
and acknowledgement must respect the existing warehouse/sync contracts.

## 7. Verification and delivery

Add happy, bad, and edge tests before or with the implementation:

- healthy JSON-only discovery and provider-I/O-free preflight;
- malformed or missing required execution JSON rejection;
- ambiguous or missing runtime binding rejection;
- all seven lane declarations and sync exposure;
- deterministic render and drift detection;
- multi-lane operation without semantic collapse;
- fake-server encoder tests and DuckDB saved-flow tests.

Then run connector-local tests, affected shared-package tests, CLI tests, and a
build. A typical final set is:

```bash
go run ./cmd/connectorgen lock-render <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/cli -count=1
go build ./cmd/pm
```

Use a local fake server by default. Use provider credentials only when already
non-interactive, separately authorized, and necessary. Never fetch mutable
provider documentation as part of a deterministic render or test.

Commit connector-local cohorts only after all required checks are green. Push
normally; do not force-push. A changed lock, its complete rendered outputs, and
the corresponding tests travel together.
