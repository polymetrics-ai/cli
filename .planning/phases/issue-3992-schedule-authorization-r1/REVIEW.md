# Inline code review — Issue #3992

## Scope

Reviewed the schedule persistence/fire lifecycle, CLI route, backend rendering,
flow receipt propagation, certification harness compatibility, tests, generated
manual, and website documentation.

## Findings

No unresolved Critical, Warning, or Info findings.

The review specifically confirmed:

- persisted or rendered schedule data never contains approval tokens, credentials,
  payloads, or raw provider errors;
- an active/running state blocks a replay even after lock-file loss;
- terminal state is written before lock release, and cleanup failure parks;
- scope drift, revocation, and expiry remain typed errors before connector calls;
- all scheduler backends invoke the `schedule fire` gate instead of bypassing it;
- the legacy non-action certification path carries a fixed opaque reference only,
  while the dedicated certification fixture proves the real authorized action fire;
- created/install/remove/list/inspect/status output preserves only the safe reference
  and fire status; and
- docs/golden artifacts match the runtime help.
