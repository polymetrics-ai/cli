# Summary — Issue #3992

Implemented authorized schedule firing on top of the landed durable
`app.AuthorizationRecord` scope identity.

## Delivered

- Schedule manifests require an opaque `auth_<id>` reference and reject token-like
  values on both save and load.
- All backends render `pm schedule fire <name> --authorization <reference> --json`.
- `schedule fire` acquires a durable non-replay lease, executes the existing
  connector-backed flow action route without a single-use approval token, captures
  opaque flow receipt IDs, and exposes safe status/inspect state.
- Scope drift, revocation, expiry, rate limits, ambiguous outcomes, cleanup faults,
  overlap, and interrupted fires halt or park rather than replay.
- CLI help, generated manual, golden transcripts, website documentation, and
  generated website docs data describe the authorization reference and status.
- The existing sample certification schedule round-trip now supplies and asserts
  a safe opaque reference, preserving its isolated backend-cleanup proof while
  the dedicated fixture certifies the actual authorized action fire.

## Key evidence

- `internal/cli/schedule_fire_test.go` provides the isolated installed-schedule
  round-trip certification and pre-dispatch safety checks.
- `internal/schedule/fire_test.go` proves lease persistence, crash/lock-loss halt,
  overlap prevention, parking, and active-fire removal refusal.
- `internal/schedule/render_test.go` proves backend payloads contain only the safe
  reference and no fixture token/secret.

See `TDD-LEDGER.md` and `VERIFICATION.md` for the Red/Green commands and gates.
