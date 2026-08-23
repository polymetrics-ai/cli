# Context — issue 4338 source-cited non-executable mutation disposition

## Decided scope

- This foundation models source operations that **do mutate** but have no
  complete executable action. It is connector-neutral across the current
  Asana (25), Jira (16), Sentry (34), and Vercel (159) consumer inventories.
- The disposition is per locked source operation and must retain provider trace
  and a concrete runtime missing-foundation gap. It cannot manufacture an
  action, command, record contract, or implemented capability.
- It is deliberately distinct from read-only coverage: read-only applies only
  to operations that cannot mutate (GET, HEAD, or no mutating provider
  contract), and rejects POST, PUT, PATCH, and DELETE. A source-cited mutation
  disposition is the only deferred path for those methods.
- Validation accepts the disposition only if no complete executable action
  exists and no matching command claims implementation. A complete action
  remains ordinary implemented coverage; an existing command is never
  downgraded to `partial` by this work.
- Vercel's 159 rows are not a blanket connector waiver. Batch 1 must decide
  each cited provider operation: author a complete action when it is intended
  to be executable, or retain this per-operation source-cited gap while it is
  intentionally non-executable. The foundation accepts neither a connector
  switch nor a count-based suppression.
- Do not modify connector definitions. Batch 1 owns Asana/Jira declarations;
  the read-only lane owns its own source declarations. Sentry may correctly use
  both foundations for different operations later.

## Coordination

`cli-read-only-source-coverage-r1` concurrently edits source-projection read
coverage in the nearby executable-coverage branch. This change must be a small,
separate mutation-only predicate/data path, with no helper renames or drive-by
refactor. Merge `origin/main` before push and again if the sibling lands first;
never rebase or force-push the published branch.

## Inline GSD fallback

The adapter prompts are resolved and followed inline. This direct-PR session has
no compatible isolated GSD runtime, and the repository contract requires one
inline delivery worker.
