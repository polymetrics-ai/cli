# Task Delivery Header Template

Complete this header before starting work and record it in the task's plan, worker handoff, or
other durable task record. It makes the intended pull-request base explicit before a branch or
pull request can drift.

## Blank Header

```markdown
## Task Delivery Header

- Issue:
  <!-- Use `Closes #123 — issue title` only for completed work landing through the default branch; use `Refs #123 — issue title` for stacked or incremental work. -->
- Base branch:
  <!-- Use the exact, existing pull-request target ref in full, for example `integration/4015-mvp-flat-r1`, never an implied or descriptive branch name. -->
- Merges into:
  <!-- State the final destination and every intermediate branch, for example `integration/4015-mvp-flat-r1 → main`. -->
- Working branch:
  <!-- Use the exact branch this agent creates, for example `fix/certification-schema-version`. -->
- Task:
  <!-- Describe what this agent will build in its own words and list the acceptance criteria that make it done. -->
- Verification:
  <!-- List the concrete commands, tests, or review evidence that will prove the acceptance criteria. -->
```

## Worked Example

```markdown
## Task Delivery Header

- Issue: Refs #4097 — Production MVP
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Working branch: docs/task-delivery-header
- Task: Add a reusable task-start header and require agents to declare the exact PR base before implementation; the PR template and agent instructions must point to the same check.
- Verification: Review the template fields and guidance, run `git diff --check`, and read the opened PR's base from GitHub.
```

## After Opening the Pull Request

Read the base back from the GitHub API; do not trust the branch you intended to set:

```sh
gh api /repos/<owner>/<repo>/pulls/<n> --jq .base.ref
```

The returned ref must exactly equal **Base branch** in the completed task delivery header.

- Do not use `gh pr view --json baseRefName` for this check: it is silently ignored for this
  purpose and must not be used.
- Correct a wrong base with:

  ```sh
  gh api -X PATCH /repos/<owner>/<repo>/pulls/<n> --field base=<ref>
  ```

- If a duplicate pull request already exists for the same head, close the wrongly based pull
  request. GitHub refuses to retarget it when the base/head pair would duplicate an existing PR.
