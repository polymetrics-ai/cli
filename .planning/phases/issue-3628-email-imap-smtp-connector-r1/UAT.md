# UAT — issue #3628 Email IMAP/SMTP connector

Date: 2026-08-06. Automated local protocol-double UAT passed.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| IMAP mailbox list is executable and limited | `TestMailboxesAreReachableThroughCommandRunner`; `TestMailboxListHonorsRequestedLimit` | pass |
| IMAP message reads are blocked until shared state support exists | `TestMessagesReadIsBlockedAndUndeclared` | pass |
| SMTP remains send-only and approval-gated | `TestSendRequiresTypedApprovalBeforeSMTPDispatch`; command-runner fleet preflight | pass |
| Preview is complete and unmasked | `TestSendPreviewIsUnmaskedAndAttachmentDriftCannotDispatch`; `TestMessageSendCommandBuildsTypedUnmaskedPreview` | pass |
| Serialized plan data remains executable | `TestSendMessageAcceptsPersistedJSONArrays` | pass |
| Help/manual/website data stay reachable | CLI help test, generated docs validation, and website script tests | pass |

No judgment-only live-mail test is performed: the task expressly prohibits live credentials and
external delivery during development. The captain's post-landing mailbox test remains the human
acceptance step.
