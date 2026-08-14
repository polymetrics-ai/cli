---
name: pm-reverse-etl
description: Plan, preview, approve, and execute reverse ETL.
---

# pm-reverse-etl

- Run `pm reverse plan` before any write.
- Run `pm reverse preview <plan-id> --json` before approval.
- For destructive plans, obtain the approval token only after preview and pass the closed `--confirm destructive` value.
- Pipe the approval token as one line into `pm reverse run <plan-id> --approval-token-stdin` only after explicit approval.
