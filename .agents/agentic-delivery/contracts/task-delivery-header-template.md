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
- Delivery:
  <!-- State the concrete completion condition for this task, for example `pull request open against the stated base with its checks green` or `branch committed and ready, no push`; committing locally is not delivery, and verify the pull request exists rather than assuming it. -->
- Working branch:
  <!-- Use the exact branch this agent creates, for example `fix/execution-schema-version`. -->
- Task:
  <!-- Describe what this agent will build in its own words and list the acceptance criteria that make it done. -->
- Verification:
  <!-- List the concrete commands, tests, or review evidence that will prove the acceptance criteria. -->

## Evidence Table

<!-- Live evidence against the real connector or implementation is the default when one applies. Add one row for every acceptance criterion; mark it `live` or `fake`, and give every fake its own written reason. Never summarize exceptions as "mocked where needed". -->
| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| <criterion> | <live \| fake> | <state change asserted by the live test, or why this individual fake is genuinely necessary> |

## Assertion Rule

<!-- Every live test must assert an observable state change that would be absent if the code did nothing; `no error returned` and `did not panic` are not evidence. -->
```

## Worked Example

```markdown
## Task Delivery Header

- Issue: Refs #4097 — Production MVP
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green.
- Working branch: docs/task-delivery-header
- Task: Add a reusable task-start header and require agents to declare the exact PR base before implementation; the PR template and agent instructions must point to the same check.
- Verification: Review the template fields and guidance, run `git diff --check`, and read the opened PR's base from GitHub.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The template declares an exact base branch and landing path | live | A repository-content check reads the actual template and asserts the completed `Base branch` and `Merges into` values are present; without this change, they are absent. |
| The pull-request template requires API base verification | live | A repository-content check reads the actual PR template and asserts the required base field and API read-back command are present; without this change, they are absent. |
| Agents must complete the header before work and verify the PR base after opening | live | A repository-content check reads `AGENTS.md` and asserts its pointer requires both actions; without this change, that directive is absent. |

## Assertion Rule

Every live check asserts the required template text exists; a successful command alone is not treated as proof.
```

## Evidence Rules

Live evidence against the real connector or real implementation is the default when one applies.
For every acceptance criterion, the evidence table must contain one row marked `live` or `fake`. A
fake is an exception: enumerate each one separately and write why it is genuinely necessary. Never
replace those reasons with a collective statement such as “mocked where needed.”

## The Assertion Rule

Every live test must assert an observable state change that would be absent if the code did
nothing. “No error returned” and “did not panic” are not evidence.

This is not satisfied merely because a real dependency was used. A shared rate-limit observer once
ran its Lua script against the real dependency, but that script errored on every call. The feature
was therefore a no-op, while every test passed because it asserted only that the call did not panic.
Live tests must assert the effect they claim to cover.

A skipped test reports as a pass: `go test` counts `SKIP` toward success. “Checks green” therefore
never, by itself, proves live coverage.

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
