---
coverage:
  - id: D1
    description: Native IMAP mailbox listing; message reads remain blocked pending #3810.
    verification:
      - kind: integration
        ref: go test ./internal/connectors/native/email -count=1
        status: pass
      - kind: other
        ref: go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
        status: pass
    human_judgment: false
  - id: D2
    description: SMTP-only typed send through the destructive preview/approval gate without masking payload.
    verification:
      - kind: integration
        ref: TestSendPreviewIsUnmaskedAndAttachmentDriftCannotDispatch; TestSendRequiresTypedApprovalBeforeSMTPDispatch
        status: pass
      - kind: unit
        ref: TestMessageSendCommandBuildsTypedUnmaskedPreview
        status: pass
    human_judgment: false
  - id: D3
    description: CLI/manual/website command-surface parity for every Email command.
    verification:
      - kind: unit
        ref: TestDynamicConnectorHelpAndBareNamespace
        status: pass
      - kind: other
        ref: go run ./cmd/pm docs validate --connectors-dir docs/connectors
        status: pass
    human_judgment: false
---

# Inline execution summary — issue #3628

The repository's Pi adapter cannot execute a non-ROADMAP issue as a numbered GSD phase and this
task forbids role spawning. The approved inline/manual fallback was used. `scripts/gsd prompt
execute-phase 3628`, `verify-work 3628`, and `code-review 3628` were resolved on 2026-08-06; this
directory holds their equivalent plan, TDD, validation, UAT, and review records.

Delivered work: the `email` native bundle and executor, IMAP mailbox listing, SMTP-only destructive
send, real command-runner preflight coverage, credential constraint regression coverage, generated
CLI/manual/catalog/website data, and the approved IMAP module. See `VERIFICATION.md` for exact
commands and binary evidence.
