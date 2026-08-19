# Discussion log — issue 4261

## Resolved scope

Captain approval on 2026-08-19 fixes the following otherwise ambiguous choices:

1. Port the foundation as a direct PR from current `origin/main`, rather than
   merge or rebase the retired integration branch.
2. Treat `ac2944115` as the content source only. It does not dictate the
   generated GitHub sweep bytes on the new base.
3. Regenerate the sweep and re-run live proof on this branch.
4. Keep only issue-4261 delivery artifacts; never transplant or delete
   historical `.planning` evidence.
5. `internal/app/sync_modes.go` is outside `ac2944115` and is explicitly out
   of scope.

## Inline lifecycle fallback

The canonical contract permits no GSD-role delegation. The `scripts/gsd prompt`
outputs for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` were resolved and are executed inline by this
single worker.
