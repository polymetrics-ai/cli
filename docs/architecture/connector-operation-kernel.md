# Connector operation kernel

The operation kernel executes bounded connector commands from the rendered
execution JSON bundle. `commandrunner` and `internal/connectors/engine` are the
runtime authority. The authoring pipeline is documented in
[`../connector-canon/SOURCE-LOCK-VNEXT.md`](../connector-canon/SOURCE-LOCK-VNEXT.md).

`operations.json` is optional execution metadata for commands that are not a
stream or write action. `cli_surface.json` binds a command to exactly one
target: `stream`, `write`, or `operation`. Runtime discovery and execution read
those rendered JSON files only.

## Closed operation kinds

The schema recognizes a closed set of operation kinds, including bounded REST
reads and writes, status checks, provider search, fixed-document GraphQL,
binary download, text export, and file upload. A kind is executable only when
the runtime has its actual encoder/executor and the command binding selects it
unambiguously. Unsupported local-git, local-file, browser, XML, composite, or
other declared kinds remain non-executable until a genuine shared executor
exists.

There is no generic shell, unrestricted HTTP, arbitrary SQL write, or
caller-supplied GraphQL executor.

## Runtime contract

- Routes, verbs, GraphQL documents, request schemas, response schemas, output
  policies, pagination, headers, redirect rules, and byte caps are fixed by
  the rendered execution bundle.
- Caller values are accepted only through schema-bound parameters. Unknown,
  ambiguous, or malformed bindings fail before provider I/O.
- Direct reads and status checks cross the normal credential boundary and
  execute one bounded request.
- Direct writes and mutations retain plan, preview, approval, authorization,
  and execute. Destructive commands require the declared confirmation.
- Binary downloads, text exports, and uploads use the existing bounded file
  encoders and path containment rules.
- Fixed GraphQL operations use their rendered document and checked variables;
  the caller cannot substitute a document.
- Successful provider response facts follow the declared output policy while
  runtime-generated diagnostics remain secret-taint-safe.

Validation rejects true runtime-invalid inputs: malformed or missing required
execution JSON, ambiguous bindings, missing encoders/executors, invalid bounded
routes or schemas, invalid invocation/approval/auth, and incompatible sync
executor/mode pairs. It does not suppress a documented command because of an
external evidence ledger.

## Authoring and proof

Authors declare the operation lane in immutable `source.lock.json`, including
shared request/response schema references. The canonical descriptor validates
and sorts it; the deterministic renderer writes `operations.json`, referenced
schemas, and `cli_surface.json`.

Proof must include:

1. deterministic `lock-render --check`;
2. discovery from execution JSON with no source-lock read;
3. fake-server reachability through credentials without live provider I/O;
4. happy, malformed, ambiguous, approval/auth, and bounded-output cases; and
5. warehouse/sync tests when the same source operation populates ETL or sync
   lanes.

If an operation has no genuine shared executor, keep its source mapping and
declare that lane unsupported. Do not add a connector-specific bypass.
