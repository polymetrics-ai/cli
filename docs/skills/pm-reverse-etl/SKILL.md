---
name: pm-reverse-etl
description: Plan, preview, approve, and execute reverse ETL.
---

# pm-reverse-etl

- Run `pm reverse plan` before any write.
- Run `pm reverse preview <plan-id> --json` before approval.
- Run `pm reverse run <plan-id> --approve <token>` only after explicit approval.
