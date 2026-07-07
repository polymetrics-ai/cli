# Eval Plan

Success criteria:

- `pm github project list --json` routes to a stream-backed GraphQL read.
- `pm github project item-list --project-id <id> --json` routes `project-id` into GraphQL variables.
- `pm github discussion list --json` routes to a stream-backed GraphQL read.
- `pm github discussion view --number <n> --json` routes `number` into GraphQL variables.
- Project/discussion create commands remain blocked/planned with risk notes.
