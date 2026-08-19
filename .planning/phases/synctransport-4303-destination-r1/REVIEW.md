---
status: clean
depth: standard
files_reviewed: 16
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
review_mode: inline_manual_fallback
---

# Refs #4303 — Code review

The normal `gsd-code-reviewer` dispatch is unavailable in this single-worker
run: compatible isolated agents cannot be provided and the canonical delivery
contract forbids role spawning. The review was therefore completed inline at
standard depth for:

- `internal/app/issue_label_warehouse_transport.go`
- `internal/app/transport_composition_test.go`
- `internal/app/app.go`
- `internal/app/transport_dispatch.go`
- `internal/app/types.go`
- `internal/connectors/connectors.go`
- `internal/connectors/engine/write.go`
- `internal/connectors/engine/direct_write.go`
- `internal/connectors/sync_transport.go`
- `internal/synctransport/orchestrator.go`
- `docs/sync-transport-definition.md`

## Disposition

The prior review findings were legitimate and have been remediated in the
shared declaration and selected-action preflight boundaries:

- raw input mappings now defer property admission to the exact selected record
  schema without identifier normalization;
- required top-level action properties must all be mapped before source or
  stage work;
- the generic destination rejects `full_overwrite` until it has a run-scoped
  protocol;
- duplicate `writes.json` action names fail at definition load; and
- successful bodyless `output_policy: none` results retain status and headers.

Focused remediation verification passed:

`go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/app -run '^(TestBundleLoadRejectsDuplicateWriteActionNames|TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies|TestValidateRecordSchemaFieldMappingRequiresExactCompleteFields|TestDeclarativeTypedDestinationSourceBindingsUseExactSelectedActionSchemaFields|TestDeclarativeTypedDestinationPreflightRejectsIncompleteMappingAndFullOverwriteBeforeIO|TestDirectWriteCommandHonorsDeclaredJSONAndNoneResponsePolicies)$'`

The review confirmed that the generic destination takes only declaration-owned
action metadata and source field mappings; validates action schemas before
provider I/O; verifies plan/preview approval before source I/O; rejects
tombstones; requires keyed replay; and cannot select a specialized
`transport_binding` action. The common factory loop collects each exact
declaration's conformance evidence without connector-name selection. GitHub
retains the dedicated typed issue-label executor solely for its provider-state
read-back.

The reconciliation review additionally confirmed that `destination_action` is
an exact persisted stream identity, never an invocation argument; multiple
connector-owned actions cannot cross-select. Schema field admission is checked
against the exact selected action, including schema-valid camelCase spelling.
The result path captures only successful responses from already-named typed
actions, persists acknowledgements before later read-back, and carries every
ordinary response field through App and CLI output. Standard credential headers,
configured secret echoes, and declaration-owned direct-write response secrets
are represented in place by an explicit mask marker; no field is removed for
scope, rarity, destructiveness, paid tier, or unfamiliarity. Focused,
affected-package, real-provider, and full repository verification are recorded
in `VERIFICATION.md`.
