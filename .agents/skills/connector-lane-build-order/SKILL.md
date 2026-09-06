---
name: connector-lane-build-order
description: >-
  Build or review schema-4 source.lock vNext connectors across direct read,
  direct write, binary transfer, ETL, reverse ETL, and sync transport lanes.
user-invocable: false
metadata:
  internal: true
---

# Connector lane build order

Use this procedure whenever a connector lane or source lock changes. The compact
`source.lock.json` is immutable authoring and evidence input. The runtime reads
only the deterministically rendered execution JSON bundle.

## 1. Capture provider facts

Use immutable, reviewable provider facts: an official OpenAPI document, GraphQL
schema, discovery document, or cited reference. Do not fetch mutable provider
documentation during a deterministic render. Preserve citations in
`provider_evidence` or per-operation `source`; neither is runtime permission.

Author `source.lock.json` directly. There is no importer that reconstructs a
lock from execution JSON and no predecessor-format reader.

## 2. Author the canonical operation once

For each provider operation, use one stable operation ID and add only the
execution projections it genuinely supports:

- `stream` for warehouse ETL reads;
- `write` for approval-gated reverse ETL;
- `operation` for direct REST, GraphQL, multipart, or binary execution;
- `commands` for CLI reachability;
- `schema_refs.request`, `response`, and `record` for shared schemas.

One operation may populate several lanes. Never turn a direct operation into a
warehouse pipeline merely to claim ETL coverage.

## 3. Declare all seven lanes

Every lock must explicitly mark exactly these lanes `implemented` or
`unsupported`:

1. `direct_read`
2. `direct_write`
3. `binary_download`
4. `binary_upload`
5. `etl`
6. `reverse_etl`
7. `sync_transport`

The declaration must agree with authored content. Keep a provider mapping when
a genuine executor is missing, mark the lane unsupported, and name the concrete
foundation gap. Do not implement a new shared runtime foundation without the
required architecture approval.

## 4. Render and validate

Run:

```bash
go run ./cmd/connectorgen lock-render <connector>
go run ./cmd/connectorgen lock-render <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs
```

The render path is:

```text
source.lock.json
  -> canonical per-operation descriptor + shared schema registry
  -> deterministic execution JSON bundle
  -> existing engine/commandrunner/transport executors
```

Rendered output may contain only execution facts. It must not contain source
citations, source-operation mappings, review evidence, artifact hashes,
declaration ledgers, or external proof state.

## 5. Prove real execution reachability

For each implemented lane, prove the actual encoder or executor without mutable
provider I/O:

- direct REST/GraphQL/multipart/binary commands reach the credential and
  approval boundary through a fake server;
- ETL materializes through the existing DuckDB/warehouse path;
- reverse ETL preserves plan, preview, approval, execute, acknowledgement, and
  read-back semantics;
- sync transport uses a registered compatible source/destination executor and
  mode;
- malformed required execution JSON is rejected;
- discovery works from an execution-only filesystem.

An unsupported lane needs an explicit empty declaration, not a placeholder
runtime route.

## 6. Runtime rejection boundary

Runtime validation may reject only actual execution-invalid input: missing or
malformed required JSON, ambiguous binding, missing encoder/executor, invalid
bounded route/schema, invalid invocation/approval/auth, or incompatible sync
executor/mode. It must not suppress a documented command because a citation,
hash, importer, review record, or external proof is absent.

Offline diagnostics may report lock citation quality or render drift, but they
must remain non-binding and must not read predecessor descriptor formats.

## 7. Review checklist

Before delivery, verify:

- the lock is the only mutable authoring source;
- all seven lanes are explicit and truthful;
- repeated renders are byte-identical;
- runtime embedding excludes `source.lock.json`;
- no alternate reader, fallback, feature flag, or second execution route exists;
- focused fake-server/DuckDB tests and affected broader Go suites are green;
- documentation points authors to
  `docs/connector-canon/SOURCE-LOCK-VNEXT.md`.

A provider primary key must be a documented stable identifier or documented
composite identity. A content/version hash is not a primary key or cursor merely
because it changes. Use concurrency tokens, cursors, and watermarks only for the
role the provider documents; otherwise choose an explicit full-refresh strategy
or retain the missing incremental capability.

Never invent an idempotency header, provider replay guarantee, or retry-safe
claim. The engine's internal write idempotency key is local correlation, not
provider-side idempotency.

Absence of provider idempotency does not by itself block the first approved
write. Such an action may remain usable when the runtime demonstrably issues one
request per row, disables automatic retry and implicit replay, consumes approval
before provider I/O, captures the bounded provider response, and reports partial
failure. It is then explicitly **single-attempt**, not retry-safe. An ambiguous
result must not be automatically replayed; recovery requires operator review or
a source-backed reconciliation/read-back strategy.

Use the closed `declarative_single_attempt_destination` adapter only when the
destination descriptor declares `delivery.idempotency: "single_attempt"`, an
exact `full_append`/`append` input-fields binding, bounded per-record delivery,
and durable-warehouse acknowledgement. It must force retries off even when the
underlying documented action is otherwise retryable, and it must remain distinct
from `declarative_typed_destination`, whose retry-safe keyed/read-back contract
is stricter. Do not broaden this adapter into a generic HTTP writer or provider
replay path.

Use `batchable: false` only when one warehouse row cannot deterministically map
to one bounded request or the bulk runtime cannot execute that action. Do not
use it merely because provider idempotency, a read-back endpoint, or scheduled
automation is absent. Provider-documented idempotent delete and missing-ok
semantics may support their declared retry policy, but must be cited.

## 9. Add managed transport only after saved execution works

Do not confuse generic saved reverse-ETL execution with an optional managed
destination transport. Prove the plan -> preview -> approval -> run route for
each action first, against an isolated warehouse and capture provider where
possible.

A source transport must bind executor references and stream allowlists to exact
modes, scope/resource evidence, continuation, checkpoints, ordering, delete
semantics, and durable acknowledgement. A managed destination transport must
name exact eligible actions and mode-to-apply bindings, enforce one row to one
bounded request, and state acknowledgement, partial-failure, ordering,
idempotency, retry, and ambiguous-result behavior.

An action can be available for single-attempt saved reverse ETL while still
lacking unattended retry/reconciliation semantics required by a stronger
managed transport. Record that distinction rather than advertising an empty
transport or suppressing the already executable action. Any unsupported managed
mode must fail deterministically before provider I/O.

## 10. Prove usable surface at the credential boundary

A command counts as usable only when the built binary or actual App route reaches
the credential boundary. Declaration reconciliation is necessary but not enough.
For direct commands, exercise command resolution, record construction, preview,
approval, and the real engine request. For saved ETL/reverse ETL, use an isolated
local warehouse and prove plan, preview, approval, execution, checkpoint, and
request-count behavior appropriate to the lane.

Reconcile every locked operation into exactly one accounted state: implemented
lane, planned mapping with a named blocker, or unsupported with source evidence.
Classify saved targets and interactive commands independently: one source
operation may truthfully support ETL and direct read, or reverse ETL and direct
write, without becoming two provider operations. Count aliases separately from
provider operations so a CLI alias never inflates source coverage. Record exact
commands, source-lock identifiers, projected actions/streams,
credential-boundary results, and remaining gaps in delivery evidence.

Use explicit timeouts for changed suites and leave memory-heavy aggregate tests
to CI. Judge merge readiness by usable surface and truthful remaining gaps, not
by source provenance or generated-file counts alone.
