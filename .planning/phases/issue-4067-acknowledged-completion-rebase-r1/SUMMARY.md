---
phase: issue-4067-acknowledged-completion-rebase-r1
status: green_checkpoint
---

# #4067 summary

The RED/GREEN core is complete. Independent Sol audit finding F1 rejected immutable candidate `883a86cf0040d559edcd4777413d1c2de20cd94a`: ordinary successful final completion leaves an acknowledged transport run durably `running` when an unrelated writer advances the whole project revision. The all-seven-mode real JSON-store witness first reproduced that leak, then passed after the transport branch captured its own acknowledged target stream and allowed a latest-state completion only if the still-running target exactly matched it. #4046 typed stale-writer failure behavior remains untouched; fail-closed target, cancellation, state-store-outcome, race, regression, generated, and delivery evidence remain pending.
