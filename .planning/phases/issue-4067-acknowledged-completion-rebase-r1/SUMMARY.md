---
phase: issue-4067-acknowledged-completion-rebase-r1
status: in_progress
---

# #4067 summary

Implementation has not begun. This phase was created after independent Sol audit finding F1 rejected immutable candidate `883a86cf0040d559edcd4777413d1c2de20cd94a`: ordinary successful final completion can leave an acknowledged transport run durably `running` when an unrelated writer advances the whole project revision. The planned repair is a strict, acknowledged-target-only completion rebase; #4046 typed stale-writer failure behavior remains untouched.

