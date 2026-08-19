---
coverage:
  - id: D1
    description: GitHub issues reached the definition-owned issue-label destination through the durable warehouse.
    verification:
      - kind: e2e
        ref: fresh pm binary live proof in VERIFICATION.md
        status: pass
    human_judgment: false
  - id: D2
    description: The applied label was independently read from GitHub and acknowledgement preceded the checkpoint.
    verification:
      - kind: e2e
        ref: gh-axi read-back and sealed warehouse/checkpoint evidence in VERIFICATION.md
        status: pass
    human_judgment: false
  - id: D3
    description: Refusals, empty/missing mapping, replay, and delete-availability semantics remain fail-closed.
    verification:
      - kind: integration
        ref: internal/synctransport and internal/app focused regressions in VERIFICATION.md
        status: pass
    human_judgment: false
---

# API → API GitHub route proof summary

The shipped `pm` binary proved the narrow GitHub `issues` → warehouse →
`add_issue_labels` route in the retained private repository
`karthik-sivadas/pm-parity-proof-api-to-api`. It used the connector's supported
token bearer production path, not the revoked runbook token and not an
installation-only claim.

The result is route evidence only: GitHub write coverage remains two actions of
607. Full details, redaction controls, independent read-back, receipt,
checkpoint, replay, cleanup, and local gates are in `VERIFICATION.md`.
