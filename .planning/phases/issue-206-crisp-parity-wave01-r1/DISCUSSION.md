# Crisp provider parity — Wave 1 discussion record

The supplied task fixes the product and implementation decisions, so no human decision was opened.

| Decision | Evidence | Result |
|---|---|---|
| Inventory authority | Provider-owned rebuild R1 report | Use its full 234-row ledger unchanged for the first commit. |
| First implementation slice | Report Wave 1 plan | Basic credential/check plus 21 core conversation GET streams only. |
| Read transport | Existing declarative GET stream engine | Model each documented GET independently as a bounded stream and CLI command. |
| Unsafe/later operations | Report per-operation classifications | Keep them blocked in the ledger; do not represent them as executable. |
| Provider interaction | Task boundary | Fixtures only; no credentials or live Crisp traffic. |
| Sensitive output | Captain policy | Full runtime output; no redaction metadata is introduced. |

The remaining 213 operations are intentionally not implementation scope for this wave. In particular, 14 HEAD checks remain blocked on the named typed HEAD direct-read/check foundation; one media listing remains blocked on RTM result correlation; and Generate Bucket URL remains blocked on RTM result correlation plus a bounded signed-upload workflow.
