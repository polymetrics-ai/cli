# Executable operation foundation design — PR #4294

Captain rejected a blocked-command placeholder on 2026-08-20. A documented
operation counts only when the installed CLI can execute its exact,
connector-owned provider contract after typed inputs and a credential are
supplied. A `BlockedCommandError` is not reachability.

The companion machine record,
`EXECUTABLE-OPERATION-CAPABILITY-AUDIT.json`, reclassifies all 3,366
source-locked operations that lack a declared surface binding. It preserves the
exact source operation ID, method, path, source location, source URL, prior
operation rejection, and next executable capability for every row. The other
1,012 of 4,378 source operations already have a surface binding.

`source_id` is the connector-owned identity. If a provider omits an operation
ID, it is deterministically derived from the pinned connector name, fixed
lower-case method, and fixed path template (`<connector>.rest.<method>_<path>`).
The provider-owned `operation_id` remains recorded separately, including an
empty value; derivation is identity bookkeeping from the lock, never schema
inference.

## Non-negotiable operation contract

Each command must bind one source operation and only that operation:

- connector-owned source operation identity, fixed HTTP verb, and fixed
  connector-relative path template;
- exact provider-declared path, query, header, and body schemas imported from
  the hash-matched source artifact;
- bounded typed values, with no caller-controlled raw URL, method, header,
  body, arbitrary JSON, or generic HTTP transport;
- the existing credential, auth, rate-limit, retry, redaction, plan, preview,
  approval, and destructive-confirmation policies; and
- actual provider I/O through the installed App/CLI path after no-credential
  preflight has proved the command is dispatchable.

The lock currently stores the source URL, bytes, digest, and operation
inventory, not the raw OpenAPI body. Any promotion must first retrieve the
public artifact and prove its digest and byte count equal the lock. A mismatch
is a source-lock refresh decision; it is never permission to infer a schema.

## Capability allocation

| Capability | Operations | Owner | Result |
| --- | ---: | --- | --- |
| Fixed REST read contract materialization | 1,389 | Connector batches, using existing runtime after source import | Executable direct-read command |
| Fixed REST write contract materialization | 1,828 | Connector batches, using existing runtime after source import | Executable plan/preview/approval/execute command |
| Bounded binary transfer | 120 | Shared binary contract plus connector declarations | Executable bounded download/upload command |
| Status operation-kind registration | 10 | Shared engine | Executable fixed HEAD/status command |
| Provider contract has no bounded typed schema | 19 | Provider schema or separately evidenced bounded contract | Not promotable until the exact input contract exists |

The 491 typed actions are separately reconciled. Their schemas contain 224
nested object/array action contracts; those are downstream consumers of #4305,
not a reason to expose a raw JSON fallback. Four fixture-bound destination
mappings exist; the remaining action mappings await action-scoped source
binding after #4304's persisted App/CLI dispatch work lands.

## Dispatchable implementation slices

### F0 — hash-matched source artifact rehydration and import

Owner: shared `connectorgen` source-import foundation plus connector batch
workers.

Read only the lock URL without credentials, verify exact bytes/SHA-256, then
derive an operation descriptor for every source operation. The descriptor must
retain source identity/location and emit fixed method/path, accepted
path/query/header/body schemas, request media type, response/output policy,
pagination, byte limits, and documented auth scope metadata. It must reject
external references, ambiguous/request-unbounded schemas, callback-only paths,
and any artifact drift before a command is generated.

This is an importer, not a caller-configurable transport.

### F1 — fixed REST operation materialization

Owner: connector-local batches after F0.

For each source contract that fits the existing closed runtime, write an exact
`operations.json` `rest_read` or `rest_write` record and a derived
`cli_surface.json` command. `surface-sync` derives owned mappings; the
installed command proves preflight and then uses real provider I/O only when a
credential is supplied. The first batches cover the 1,389 read and 1,828 write
candidates, but each individual promotion remains gated by the imported schema
form—these counts are allocation, not an unsupported bulk-enable claim.

### F2 — declared typed request-header binding

Owner: shared commandrunner/engine foundation.

`RESTOperationSpec` currently has no operation-scoped typed header input and
`operationDirectReadOverrides` admits only `path.`, `query.`, and `body.`
targets. Add a closed, provider-declared header parameter shape with its own
schema and `header.<name>` mapping. The runtime must validate the name/value
before I/O and merge only admitted non-auth headers. `Authorization`, `Cookie`,
`Host`, proxy headers, and transport metadata remain runtime-owned and cannot
be supplied or overridden by callers.

Tests must prove header/path/query/body isolation, unknown-header refusal,
schema/size refusal before I/O, and preservation of credential/auth policy.

### F3 — structured body materialization (#4305)

Owner: #4305.

Use one schema-aware, declaration-bound object/array materializer for both
fixed REST operations and typed write actions. It consumes only the imported
operation/action schema; validates required, unknown, malformed, and oversized
values before I/O; preserves the exact fixed endpoint; and binds the canonical
structured payload into preview/approval/confirmation. The downstream contract
is recorded in the capability audit. Do not add a second body builder in this
connector lane.

### F4 — bounded binary, multipart, status, and text execution

Owner: shared engine, then connector-local declarations.

Promote the 120 binary rows only through a named binary/multipart contract with
an exact endpoint, approved local input or output target, byte cap, media-type
policy, redirect policy, and confirmation where mutating. Register the existing
`rest_status` and `text_export` kinds in the bundle loader/validation path for
the ten status rows and any source-proven text export. No inline raw bytes,
unbounded files, or arbitrary content type may enter the command surface.

### F5 — persisted reverse-ETL application dispatch and action selection

Owner: #4304 for persisted App/CLI dispatch; follow-on shared source-binding
work for action selection.

First merge and prove the latest #4304 head through the installed App/CLI path.
Then make destination source mappings action-scoped, so each typed action has
its own exact source fields and selected action rather than one executor/stream
mapping. Keep binary transfer separate from REST/reverse-ETL writes.

### F6 — connector batch execution and proof

Owner: this connector lane once F0–F5 dependencies are available.

Work by bounded connector commits. For each promoted operation: validate the
lock hash, generate the command/manual/catalog artifacts, run
`connectorgen validate`, `surface-sync --check`, certification sweep, and the
installed binary in its own initialized project. A dispatchable command reaches
`missing --credential` before provider I/O; a credentialed run is deferred to
live certification. Finish with `connector-boundary`, generated checks, and
`make verify` before any push.

## Explicit non-solutions

There is no disabled-command target, raw body flag, generic HTTP command,
caller-selected method/path/header, arbitrary JSON flag, or fabricated schema
in this design. Destructive, privileged, rare, paid, or elevated-scope
operations remain executable when their typed provider contract is represented;
the existing policy gates govern execution rather than command reachability.
