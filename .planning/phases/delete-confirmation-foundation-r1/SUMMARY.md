---
coverage:
  - id: D1
    description: Destructive confirmation is a typed closed schema for writes and future operations.
    verification:
      - kind: unit
        ref: internal/connectors/engine/bundle_test.go TestBundleLoadAcceptsClosedDestructiveWriteConfirmation/TestBundleLoadAcceptsClosedDestructiveOperationConfirmation/TestBundleLoadRejectsOpenDestructiveConfirmation
        status: pass
    human_judgment: false
  - id: D2
    description: Destructive writes require plan, real preview, explicit typed approval, and execute in that order.
    verification:
      - kind: integration
        ref: internal/app/reverse_confirmation_test.go destructive plan/preview/approval/execute cases
        status: pass
      - kind: e2e
        ref: internal/cli/reverse_cli_test.go TestGitHubDestructiveCommandRequiresTypedConfirmation
        status: pass
    human_judgment: false
  - id: D3
    description: The lifecycle is exposed on an existing canonical public connector command without a duplicate alias.
    verification:
      - kind: e2e
        ref: pm github repo deploy-key delete --help and TestGitHubDestructiveCommandRequiresTypedConfirmation
        status: pass
    human_judgment: false
  - id: D4
    description: A future rest_write executor can use the provider-neutral gate without modifying it.
    verification:
      - kind: unit
        ref: internal/connectors/engine/write_test.go TestRestWriteOperationUsesSharedDestructiveExecutionGate
        status: pass
    human_judgment: false
  - id: D5
    description: Direct engine, generic bulk, canonical command, state replay, digest drift, and batchability bypasses fail closed.
    verification:
      - kind: integration
        ref: internal/connectors/engine/write_test.go and internal/app/reverse_confirmation_test.go bypass cases
        status: pass
      - kind: integration
        ref: internal/connectors/defs/asana and internal/connectors/defs/zendesk-support reverse-ETL fixture packages
        status: pass
    human_judgment: false
---

# Phase Summary — destructive-write confirmation foundation

## Outcome

One shared engine/app contract now enforces `plan -> preview -> explicit approval -> execute` for
destructive connector writes. HTTP DELETE, delete/destructive mutation metadata,
`destructive:true`, and any non-empty confirmation declaration all fail closed into the gate. The
only accepted confirmation value is the typed `destructive` kind.

No new connector command, alias, generic HTTP writer, or connector operation binding was added.
The end-to-end proof uses the existing canonical public command
`pm github repo deploy-key delete`.

## Runtime contract

- `writes.json` and `operations.json` accept the same closed confirmation object:
  `{"kind":"destructive"}` with unknown values and properties rejected.
- Engine dry-run previews produce a deterministic digest over connector, action, method, path,
  effective config, records, and preview output without making a provider call.
- Destructive plans mint no approval token until a real preview persists that digest.
- A project-vault plan seal authenticates plan lifetime, mode, connector/action, credential and
  effective-configuration revisions, batchability, and confirmation before preview.
- Execution revalidates the stored plan hash and dry-run digest, creates a monotonic authenticated
  vault consumption marker under a revision-CAS state transition, and lets the engine compare the
  complete MAC-bound target again before invoking its executor closure.
- The shared `DestructiveTargetForOperation` plus `GateDestructiveExecution` seam already accepts a
  `rest_write` operation and an executor callback; the future executor needs no gate changes.
- Generic reverse ETL rechecks `batchable:false` before preview and execution. Confirmation never
  turns a non-batchable action into a bulk action.

## Fixture evidence

Captain decision `defs-delete-fixtures` authorized exactly two connector test files for behavioral
evidence:

- Asana `reverse_etl_execute_test.go` obtains destructive evidence through the real app canonical
  command plan, preview, typed confirmation, and execute lifecycle.
- Zendesk Support `reverse_etl_execute_test.go` obtains the same evidence through the real app bulk
  plan, preview, typed confirmation, and execute lifecycle.

Both retain their original HTTP method/path/body/redaction assertions. No bundle JSON or connector
runtime definition changed; documentation housekeeping only corrects stale connector prose.

## TDD result

The ledger records RED then GREEN evidence for declaration/schema, mandatory preview, explicit
approval, the future executor seam, direct/bulk/replay bypass resistance, DELETE inference,
unknown confirmation fail-closed behavior, and the independent `destructive:true` flag. All
changed packages, the full `internal/cli` suite, local smoke, vet/lint/build, connector validation,
surface sync, boundary, docs, and release gates passed.

The follow-up trusted-input review additionally covers caller-key authority rejection, signed plan
lifetime, stale whole-state CAS rejection, rolled-back-state replay, persistent nonce consumption,
and configuration/batchability drift. Its field-by-field disposition is recorded in
`TRUSTED-INPUT-SWEEP.md`.

## Operation coverage

The gate now makes all **623** measured DELETE operations across the 14 shipped connectors
technically bindable through one runtime contract. This foundation does not claim 623 bindings:

- **171** already-bound destructive operations can execute only through the shared lifecycle when
  reached by their current canonical command/action mapping.
- **452** documented operations remain connector work: each still needs its canonical public
  `cli_surface` command, typed closed action/operation schema, request mapping, fixture, and
  connector-specific disposition.

The same gate shape applies to the repository-wide 3,409 DELETE inventory and to the forthcoming
`rest_write` executor, but those connector bindings/executor implementation remain out of scope.
