---
phase: synctransport-4303-destination-r1
issue: 4303
status: complete
key_files:
  created:
    - .planning/phases/synctransport-4303-destination-r1/VERIFICATION.md
    - .planning/phases/synctransport-4303-destination-r1/REVIEW.md
  modified:
    - internal/app/issue_label_warehouse_transport.go
    - internal/app/transport_composition_test.go
    - docs/sync-transport-definition.md
---

# Refs #4303 — Typed reverse-ETL destinations

One declaration-selected factory now composes connector-owned typed
destination actions using exact per-definition evidence. It never accepts an
endpoint, HTTP method, body template, or operation name from a caller.

The generic adapter requires an ordinary schema-backed `writes.json` action,
explicit `input_fields` source bindings, keyed replay, no tombstone deletes,
durable acknowledgement, and the existing approval lifecycle. GitHub’s
issue-label route shares the definition-evidence/factory composition but keeps
its specialized adapter and independent provider-state read-back.

Synthetic tests prove arbitrary first and second declarations, evidence
isolation, malformed/unknown/wrong-role/capture refusal, approval refusal, and
pre-write workset rejection. The real GitHub proof passed before the prior
bespoke factory branch was retired.

The persisted application path now carries a stable `destination_action` on a
stream configuration, so one connector can expose multiple named typed actions
without action inference or runtime selection. Exact selected-action schema
properties accept both snake_case and camelCase bindings and reject all other
names before I/O. Acknowledged typed writes preserve complete ordinary provider
response status, headers, and bodies in persisted App/CLI results; only the
credential boundary masks values in place with an explicit marker.
