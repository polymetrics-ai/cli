# TDD ledger — #3771 command-runner runtime content

| Slice | Red evidence to capture before production edit | Green evidence | Status |
| --- | --- | --- | --- |
| #3782 | Existing runner assertions and a new preservation assertion fail because `***` is produced or `RedactFields` is forwarded. | Focused runner tests preserve record/error values and leave executor request fields empty. | Planned |
| #3784 | New public `PlanConnectorCommand` persistence test fails because the saved/reloaded sample masks a declared nested/token/content field. | Public application path persists and reloads the complete sample without calling the write method. | Planned |
| #3786 | Focused CLI golden/help assertion fails after desired wording is asserted. | Help, manual, golden transcript, and website copy agree on complete connector-command content and scoped source-table distinction. | Planned |
| #3790 | New behavior-level regression matrix fails under the old runner behavior. | Matrix proves records, previews, errors, and executor request fields retain the policy. | Planned |

Existing tests asserting masking are intentionally revised rather than deleted because the captain
overruled the old policy. The final PR must call out this reversal.
