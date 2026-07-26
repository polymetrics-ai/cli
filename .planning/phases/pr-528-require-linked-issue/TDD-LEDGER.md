# TDD ledger - PR 528 require-linked-issue repair

## Red

Command:

```bash
go run ./cmd/prissueguard --title "$(gh pr view 528 --repo polymetrics-ai/cli --json title -q .title)" --body-file /dev/stdin < <(gh pr view 528 --repo polymetrics-ai/cli --json body -q .body)
```

Observed result:

```text
issueguard: blocked
- PR body must reference an issue with Closes #123 for completed work or Refs #123 for stacked/incremental work
exit status 1
```

## Green target

After the PR body is edited to include `Refs #67`, the same guard must report:

```text
issueguard: ok (1 linked issue)
```

Green evidence captured on 2026-07-26:

```text
issueguard: ok (1 linked issue)
```

## Refactor

No guard code change is planned. The root cause is missing PR metadata, not a parser defect.
