---
description: Run one isolated GSD worker subtask
agent: gsd-worker
subtask: true
---

Execute exactly one bounded worker assignment.

Arguments: `$ARGUMENTS`

Required in the assignment:

- issue or scope
- parent issue/PR when applicable
- branch and base branch
- worker directory or read-only shared checkout permission
- allowed write paths
- required skills
- red/green/refactor evidence requirements
- verification commands
- handoff destination

Return only the worker handoff. Do not merge parent PRs or edit shared parent artifacts unless the
assignment explicitly grants that scope.
