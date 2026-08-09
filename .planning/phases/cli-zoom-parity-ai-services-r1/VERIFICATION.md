# Verification Checklist — Zoom AI Services documented-operation parity, R1

## Planned checks

- [x] Fresh artifact URL, timestamp, HTTP status, byte count, SHA-256, OpenAPI/server provenance,
  and zero source-to-ledger delta recorded before RED; continuation re-fetch on
  `2026-08-09T20:57:31Z` was HTTP 200 / 87,750 bytes / identical SHA-256.
- [x] RED command-surface test captured before production declaration or foundation work; commit/push pending this checkpoint.
- [ ] The 22 provider operations are all covered, including the `101` Live Scribe WebSocket;
  zero Zoom rows use `unsafe_or_disallowed`.
- [ ] The protocol foundation has standalone red/green evidence and no new dependency.
- [ ] Every REST read/write has loopback executable fixture evidence; all three deletes are
  status-only and require typed destructive confirmation.
- [ ] The Live Scribe operation proves fixed handshake/subprotocol/auth/session-update/PCM frame
  ordering, bounds, redaction, cancellation, and fail-closed rejection behavior.
- [ ] A freshly built binary reaches base, namespace, AI Services group, and every exact command
  help route.
- [ ] Surface sync/reconciliation, generated docs/site output, endpoint-ledger locality, and
  website-catalog locality are verified.
- [ ] Scoped local gates, inline manual verify-work, and code-review are complete.
