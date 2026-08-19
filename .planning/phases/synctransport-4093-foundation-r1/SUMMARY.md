---
coverage:
  - id: D1
    description: Definition-selected source and destination evidence composes through shared production adapters.
    verification:
      - kind: unit
        ref: internal/app/transport_composition_test.go: TestDefinitionTransportFactoriesRegisterSharedSourceOnce
        status: pass
      - kind: unit
        ref: internal/synctransport/definition_composition_test.go: TestDefinitionConformanceVerifierAcceptsEvidenceSelectedByEachDefinition
        status: pass
    human_judgment: false
  - id: D2
    description: Existing GitHub transport route retains its executable lifecycle.
    verification:
      - kind: e2e
        ref: internal/cli/github_transport_binary_test.go: TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle
        status: pass
    human_judgment: false
---

# Summary — Refs #4093

Production transport composition now collects conformance evidence from each
definition declaring the shared adapter instead of reusing GitHub constants.
The declarative source and closed typed issue-label action destination are
registered once for each exact executor reference and select the source or
destination from the preflighted request. No connector-name branch is used.

The legacy App factory wiring is retired; the retained issue-label code is the
closed typed action and approval contract, not a connector-selection path.
`docs/sync-transport-definition.md` supplies the connector authoring recipe.
Destination declarations now refuse `change_capture`; source-only capture uses
the existing connection-warehouse path before transport dispatch.
