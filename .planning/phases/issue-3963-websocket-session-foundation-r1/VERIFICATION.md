# Verification Checklist — closed WebSocket session operation foundation, R1

- [x] GSD/manual fallback, source resolution, required skills, and inline discussion recorded.
- [x] Inherited loader RED rerun and captured before production edits.
- [x] Operation schema rejects unknown execution blocks and invalid WebSocket declarations.
- [x] Loader enforces fixed GET/relative path/subprotocol/closed initial schema/positive bounds.
- [x] Loopback transport proves handshake, auth boundary, masking, bounds, cancellation, close,
  redaction, and malformed-frame rejection.
- [x] Commandrunner permits only matching implemented operation declarations and rejects
  caller-controlled WebSocket transport controls.
- [x] No new dependency, credentialed call, reverse-ETL execution, or generic tool surface.
- [x] Targeted tests, vet, build, declarative surface validation, and generated-surface drift check
  recorded. The first full `internal/cli` run is recorded in `TDD-LEDGER.md` as inherited Zoom CRC
  golden drift; it must be rerun after the parent regenerates the relevant golden artifacts.
- [ ] Code review, stacked-PR handoff, and consumer built-binary proof recorded.
