---
coverage:
  - id: D1
    description: Certification parity projection and safe evidence publication are ported to main.
    verification:
      - kind: unit
        ref: go test -count=1 -timeout 20m ./cmd/connectorgen
        status: pass
    human_judgment: false
  - id: D2
    description: GitHub generated sweep and scoped matrix evidence are current.
    verification:
      - kind: integration
        ref: connectorgen certification-sweep/candidates/matrix checks
        status: pass
    human_judgment: false
  - id: D3
    description: A bounded real GitHub read produced redacted accepted evidence on this branch.
    verification:
      - kind: integration
        ref: scripts/certify-connector-live.mjs github --limit 1
        status: pass
    human_judgment: false
---

# Summary — issue 4261

The certification foundation is ported from `ac2944115` onto the current
`origin/main` base without integration-branch merge drift. The generated GitHub
sweep was recreated from this branch, and a read-only live proof produced a new
schema-v2 redacted record. `internal/app/sync_modes.go`, release metadata, and
historical planning traces are excluded.

The GSD lifecycle was executed inline under the canonical no-delegation
contract. Targeted tests, the full required `make verify`, and the inline code
review all passed without findings.
