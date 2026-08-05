# Specification — destructive-write confirmation foundation

## Objective

Provide one shared runtime contract for every connector delete or destructive mutation:

`plan -> preview -> explicit approval -> execute`

The operation remains exposed only through its canonical public connector command. The foundation
does not add connector aliases, synthetic duplicate commands, or a generic HTTP write surface.

## Required behavior

1. Destructive intent is derived from closed metadata, not prompt text: HTTP `DELETE`, delete or
   destructive mutation class, or the existing `confirm: "destructive"` declaration.
2. The JSON bundle dialect accepts a closed confirmation object whose only supported kind is
   `destructive`; unknown keys and values are rejected. The existing closed string declaration is
   normalized for compatibility with already shipped bundles.
3. A destructive plan does not become executable merely because it exists. A real no-network
   preview must materialize every canonical request and bind all records, query/body construction,
   definition/hook identity, concrete target, and secret-safe credential revision.
4. A vault-authenticated plan seal binds plan lifetime, identity, mode, connector/action,
   credential/configuration revisions, batchability, and confirmation before preview. The persisted
   grant adds the exact preview and concrete target without carrying secrets.
5. App execution reloads under a revisioned state update and creates a monotonic authenticated
   vault consumption marker before dispatch. The opaque engine evidence is independently one-shot
   even when copied or when state JSON is rolled back.
6. The shared prepared-write gate validates evidence before invoking an executor callback.
   Declarative writes and native Amazon SQS use this seam, and a future `rest_write` executor can use
   it unchanged.
7. Generic reverse ETL and canonical connector-command execution both supply evidence only after
   the stored plan, preview, approval token, typed confirmation, and current plan hash pass.
8. Direct destructive engine writes without evidence fail before HTTP dispatch.
9. `batchable: false` remains authoritative at both plan and execute time; confirmation never
   converts a non-batchable action into a bulk action.

## Scope

- `internal/connectors/engine/` write gate, bundle types, schemas, and tests.
- Connector write request/preview types needed by the shared engine/app seam.
- `internal/app/` reverse-plan preview, authenticated grant, and atomic execution transition.
- Native Amazon SQS adoption of the provider-neutral prepared-write seam.
- Connector command runner/CLI lifecycle plumbing and focused help/docs only when behavior changes.
- GSD phase artifacts and fixture-only tests.

## Non-goals

- Binding the remaining connector operations.
- Implementing the `rest_write` HTTP executor.
- Editing connector runtime definitions or bundle JSON, especially under `defs/zendesk-support/`
  or `defs/asana/`.
- Live provider calls, credentialed tests, new dependencies, generic write tools, or main merge.

## Coverage statement

The gate is reusable for all 623 measured DELETE operations across the 14 shipped connectors. The
171 already bound operations retain their canonical command/action mapping. The remaining 452
become technically bindable through this shared contract; authoring their typed operation/action
schemas, canonical `cli_surface` mapping, request fixtures, and connector-specific dispositions
remains connector-lane work.
