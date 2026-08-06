# DISCUSSION LOG — issue #3628 Email IMAP/SMTP connector

Date: 2026-08-06.

The parent and all children (#3620, #3628, #3627, #3660, #3662, #3664, #3666, #3668, and #3614)
were read with `gh-axi issue view <number> --full` before implementation. Their dependency order
is followed in `PLAN.md`.

No unresolved product question remains:

- Captain decision `imap-dependency` authorizes go-imap v2 beta.8 and only its required modules.
- Native wire-protocol commands have an established contract: `etl`/`reverse_etl` surfaces are
  covered by actual protocol operation labels in `api_surface.json`; REST-shaped rows are neither
  needed nor permitted.
- The existing preview model exposes connector warnings. The native writer will use that established
  unmasked preview channel for the exact SMTP envelope and MIME bytes, while
  `engine.PreparedWrite` binds the same bytes into the approval digest. No shared preview schema
  change is needed.
- The original mailbox + UIDVALIDITY + UID cursor decision is superseded for this slice: sparse
  UID scan continuation and full-refresh enforcement need #3810's shared mode validation and
  checkpoint state, so message reads remain unavailable until #3810 lands.

Canonical sources: RFC 9051 (IMAP4rev2) and RFC 6409 (message submission), both retrieved
2026-08-06; `docs/migration/conventions.md`; `docs/architecture/connector-architecture-v2-design.md`;
`internal/connectors/native/postgres`; `internal/connectors/native/dynamodb`; and the captain
dependency ruling outside this repository.
