# Dual-Mechanism Connector Foundations (P0) Summary

## Local recovery outcome

- Added `driver.NewFlow`, the missing public browser-session credential flow.
  It launches the controlled real browser, waits for the user-completed login,
  captures only declared cookies, records an expiry hint/browser fingerprint,
  and closes the browser. It adds no form-fill or password-entry surface.
- Confined OAuth callback binding to literal loopback IP addresses.
- Enforced exactly one browserauth credential outcome and provided explicit
  expiry-driven reauthentication decisions for OAuth and browser sessions.
- Preserved the full declared `metadata.mechanism` governance block through
  engine synthesis, catalog/inspect JSON, and connector manuals.
- Local tests and individual repository gates are green; see
  `VERIFICATION.md` for commands.

## Pending external gate

The branch is older than current `main`, where the required
`connectorgen surface-sync` command exists. No surface-sync result is claimed
yet. After no-mistakes rebases the branch, run its `--check` command locally,
record its actual result, then continue the pipeline/PR workflow.
