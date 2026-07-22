# Runbook and rollback

If bootstrap fails before publication, correct the issue/authority ambiguity and rerun `start`; no durable
controller state is created. An existing invalid or conflicting plan is never overwritten automatically.
If verification fails, inspect the bounded command result and let the correction budget run. To roll back
this correction, revert its commits before #479 integration; do not delete human-authored plans or merge
the parent PR. The parent/default-branch merge remains a human action.

