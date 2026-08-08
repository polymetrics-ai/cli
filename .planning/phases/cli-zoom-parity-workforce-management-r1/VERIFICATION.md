# Verification Checklist — Zoom Workforce Management parity, R1

## Planned checks

- [x] Live artifact, operation count, byte count, hash, and ledger comparison recorded before RED.
- [x] RED capture committed before all production declaration/foundation changes.
- [ ] CSV foundation proves valid CSV reaches the provider while malformed or non-`.csv` sources
  are rejected before network dispatch; existing JSON validation remains green.
- [ ] All 18 command paths pass real commandrunner preflight.
- [ ] Eleven direct reads and seven direct writes run against isolated exact fixtures.
- [ ] Both DELETEs assert 204 status-only semantics and require destructive confirmation.
- [ ] Endpoint ledger reconciliation is confined to `provider_module=workforce-management`; zero
  rows are `unsafe_or_disallowed`.
- [ ] Generated CLI docs/site output retains Zoom-only changes after whole-file generation.
- [ ] Fresh `pm` binary reaches base, namespace, provider group, and all 18 command help routes.
- [ ] Scoped local gates, inline verify-work, and manual code review complete.
