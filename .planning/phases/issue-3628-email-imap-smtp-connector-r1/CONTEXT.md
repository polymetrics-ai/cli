# CONTEXT — issue #3628 Email IMAP/SMTP native connector

## Phase mapping

Parent issue #3620 tracks provider-parity work. Issue #3628 owns the protocol-level Email
connector and integrates its dependency-ordered children #3660 (bundle/contract), #3662 (IMAP
read), #3664 (SMTP typed write), #3666 (CLI/help/docs parity), and #3668 (validation/evidence)
in one branch and one eventual PR. Issue #3627 is a duplicate and contributes no code.

This is a Tier-3 native connector: IMAP and SMTP are wire protocols, not REST APIs. API-based
Gmail and Outlook connectors remain separate and are not copied or repurposed.

## Locked decisions

- The captain approved `github.com/emersion/go-imap/v2@v2.0.0-beta.8` for this connector only
  (decision key `imap-dependency`, 2026-08-06). Its transitive `go-message` and `go-sasl` modules
  are authorized; no other new dependency or version jump is authorized. The beta API-instability
  risk is accepted. Record the measured `pm` binary-size delta in the PR body.
- IMAP is the read side only. SMTP is send-only and is exposed only as the `send_message` typed
  reverse-ETL write action. No SMTP list/fetch/search stream or read capability is declared.
- The messages cursor is a deterministic, fixed-width encoding of mailbox identity,
  `UIDVALIDITY`, and `UID`. For the same mailbox, its lexical order is the RFC 9051 ordering of
  the `(UIDVALIDITY, UID)` tuple, so the existing scalar stream-state store can safely persist it.
  A changed mailbox encoded in prior state is rejected rather than silently reusing a cursor.
- A changed `UIDVALIDITY` resets the UID lower bound. RFC 9051 requires a client to discard prior
  UID mappings when it changes; it also requires a replacement UIDVALIDITY to be greater than the
  prior value. This connector does not use received/internal dates as cursors.
- Polling cannot observe hard deletes. A removed message simply stops appearing; no tombstone is
  emitted. This limitation must appear in connector docs and command help.
- The write is externally visible and irreversible after SMTP accepts `DATA`. It is non-batchable,
  requires plan → preview → approval → execute and the closed typed `--confirm destructive` value.
  Preview includes the actual SMTP envelope and complete RFC 5322/MIME payload with no masking.
- Real secret values are only `x-secret` configuration; they must never appear in errors,
  documentation, previews, logs, generated artifacts, or fixtures. Local protocol test doubles
  use clearly fake non-secret values only.
- IMAP IDLE/push/subscription lifecycle is excluded. The seam is documented for #3614; this lane
  implements polled reads only.
- Configuration is client-like: IMAP/SMTP host, constrained port, closed security enum, username,
  `x-secret` password, optional SMTP username, optional from address, optional connection timeout,
  and a mailbox/default body bound needed for the bounded stream contract.

## Scope and ownership guard

- Connector-owned production: `internal/connectors/defs/email/**`,
  `internal/connectors/native/email/**`, and the explicit native wiring needed to make the
  connector registered in `internal/connectors/native/nativeset/factories.go`.
- Generated native imports are changed only through `go run ./cmd/connectorgen gen`.
- Do not alter shared operation/command schemas, engine paths, or command-runner functions. In
  particular, #3771 owns the command-runner redaction block/Run paths, #3775 owns required flag
  materialization, and #3769 owns direct-read validation. #3773, #3761, #3745, #3795 are named
  concurrent foundations and are not absorbed.
- Do not fabricate REST endpoints. `api_surface.json` records actual IMAP/SMTP protocol commands
  and their stream/write coverage. The native protocol surface is executable through the native
  connector, not an HTTP executor.
- No live mail server, credential, message, or actual send is used. IMAP/SMTP tests inject local
  protocol fakes and temporary payload files only.

## GSD execution note

Adapter evidence: `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`,
`execute-phase`, `verify-work`, and `code-review`, and `go run ./cmd/agentcontractgen check` all
passed. `scripts/gsd prompt discuss-phase 3628 --auto` and `scripts/gsd prompt plan-phase 3628
--tdd` resolved the official workflow, but #3628 is not a numbered generic ROADMAP phase. The
documented inline/manual fallback is therefore used: the captain's issue tree and decision are the
discussion source, and this phase directory records the same TDD, verification, and review gates.
No GSD role is spawned.

## CLI help/docs/website parity disposition

This adds the `pm email` command tree, its flags, rendered help, connector docs, and website/CLI
reference entries. Runtime help, bare namespace help, every individual command, docs, website
references, and command-surface validation are all applicable and must be checked before handoff.
