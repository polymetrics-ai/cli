# Issue #4292 — parity batches 8–10 context

## Locked decisions

- Scope is the thirty connector bundles named in issue #4292, divided into three
  ten-connector commit increments. This is a declaration-map lane: no engine,
  command, API-surface, transport declaration, credential, or provider-call
  change is authorized.
- Use the exact source-lock/crosswalk/disposition shape accepted in batch one at
  `fa36a676e:internal/connectors/defs/gitlab/sources/gitlab-declaration-disposition.json`.
  Every documented source operation has one primary parity class and one
  foundation disposition.
- Only a documented engine limitation may be a `foundation-gap`. Connector-local
  absence of a typed operation, command, or CLI binding is `declaration-pending`.
  Elevated permission is enabled runtime metadata, never a rejection reason.
- ETL is declaration-capable only where the connector satisfies the definition
  source contract. A typed write action is a `direct_write` primary parity
  class, including DELETE; it is never itself reverse ETL. Reverse-ETL is an
  eligibility attribute on each direct-write operation. That attribute is
  currently a foundation gap for every connector:
  `generic-typed-destination-executor`, evidenced by
  `internal/app/issue_label_warehouse_transport.go:85-95` at `acb85dc03`.
  The required minimal change is: register a connector-neutral typed destination
  `DefinitionFactory` selected by the definition, with per-connector evidence,
  explicit source bindings, acknowledgement and per-mode apply strategies.
  No `transport_binding` action will be invented.
- No live certification claim is allowed. Each summary records
  `live_certification: pending` and sources must be public and credential-free.

## Discussion record

The repaired autonomous brief, issue #4292, batch-one specification, and the
transport correction make the classification rules and scope explicit. No
product decision remains to be requested; the compatible Pi/GSD runtime is not
available in this worker, so the required discussion is recorded inline.
