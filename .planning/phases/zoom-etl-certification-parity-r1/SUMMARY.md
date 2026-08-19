---
coverage:
  - id: D1
    description: Generated Zoom sweep records the existing ETL/capability projection as fixture-required.
    verification:
      - kind: unit
        ref: internal/connectors/defs/zoom/command_surface_test.go:TestCertificationSweepProjectsExistingETL
        status: pass
      - kind: other
        ref: go run ./cmd/connectorgen certification-sweep --connector zoom --check
        status: pass
    human_judgment: false
  - id: D2
    description: Existing Zoom ETL commands retain their sanitized fixture and runner behavior.
    verification:
      - kind: unit
        ref: go test -timeout 20m ./internal/connectors/defs/zoom -count=1
        status: pass
      - kind: integration
        ref: go test -timeout 20m ./internal/connectors/conformance -run TestConformance/zoom -count=1
        status: pass
    human_judgment: false
  - id: D3
    description: The wave remains pending central certification scope and cannot be called certified.
    verification:
      - kind: other
        ref: go run ./cmd/agentcontractgen certification-gate --connector zoom --transition integrate_sub_pr
        status: pass
    human_judgment: false
---

# Zoom ETL certification parity — Wave #4266 summary

The wave adds generated, connector-local parity inventory for the three existing Zoom ETL commands and the declared read capability. It adds a regression test that protects their implemented/fixture-required status without changing streams, provider ledger, command surface, authentication, engine code, or central certification configuration.

## Delivery fallback

The canonical issue-first lifecycle was followed inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were resolved through `scripts/gsd prompt`. The delivery contract forbids GSD role spawning, so the artifacts and review were completed manually and recorded here.

## Certification boundary

No credential or live provider call was used. Per the 2026-08-19 captain decision, the expected `capability/zoom/missing` certification HALT is centrally owned by firstmate. This is implemented-and-pending-certification fixture evidence only; it is not a PROCEED result and the sub-PR must not be integrated until central scope and live proof permit it.
