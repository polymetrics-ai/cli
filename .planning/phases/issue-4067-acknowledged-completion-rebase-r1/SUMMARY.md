---
phase: issue-4067-acknowledged-completion-rebase-r1
status: red_recorded
---

# #4067 summary

The behavioral RED is complete; production implementation has not begun. Independent Sol audit finding F1 rejected immutable candidate `883a86cf0040d559edcd4777413d1c2de20cd94a`: ordinary successful final completion leaves an acknowledged transport run durably `running` when an unrelated writer advances the whole project revision. The new all-seven-mode real JSON-store witness preserved the acknowledged target stream and unrelated writer, then observed the zero returned run and durable `running` lifecycle leak. The planned repair remains a strict, acknowledged-target-only completion rebase; #4046 typed stale-writer failure behavior remains untouched.
