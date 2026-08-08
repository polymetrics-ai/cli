# Verification Checklist — Zoom Tasks parity, R1

## Planned checks

- [ ] Live artifact, operation count, byte count, hash, and ledger comparison recorded before RED.
- [ ] RED capture committed before all production declaration/foundation changes.
- [x] Redirect-safe multipart foundation proves bearer replay, multipart snapshot rebuild, fixed
  base/bearer/suffix boundaries, 307/308-only behavior, hop caps, and signed-target redaction.
- [ ] All 17 command paths pass real commandrunner preflight.
- [ ] Six direct reads and eleven direct writes run against isolated exact fixtures.
- [ ] Four DELETEs and task PATCH assert 204 status-only semantics; DELETEs require destructive
  confirmation.
- [ ] Endpoint ledger reconciliation is confined to `provider_module=tasks`; zero rows are
  `unsafe_or_disallowed`.
- [ ] Generated CLI docs/site output retains Zoom-only changes after whole-file generation.
- [ ] Fresh `pm` binary reaches base, namespace, provider group, and all 17 command help routes.
- [ ] Scoped local gates, inline verify-work, and manual code review complete.
