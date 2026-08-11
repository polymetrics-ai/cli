---
phase: issue-4067-acknowledged-completion-rebase-r1
status: focused_verified
---

# #4067 summary

The focused transport correction is verified. Independent Sol audit finding F1 rejected immutable candidate `883a86cf0040d559edcd4777413d1c2de20cd94a`: ordinary successful final completion leaves an acknowledged transport run durably `running` when an unrelated writer advances the whole project revision. The all-seven-mode real JSON-store witness first reproduced that leak, then passed after the transport branch captured its own acknowledged target stream and allowed a latest-state completion only if the still-running target exactly matched it. Changed/missing/terminal targets fail closed with a detectable revision conflict; cancellation remains durable in all modes; definite versus committed/indeterminate persistence results return truthfully; focused race and #4046/R7/R8 regressions pass. Generated-output, heavy validation, review, no-mistakes, and delivery evidence remain pending.
