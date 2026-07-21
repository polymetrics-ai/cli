# PR #438 Read-only Canary Trace

Date: 2026-07-21
Final hardened extension checkpoint: `c1c5e9e90a4a59b621bdfb94d0500b8124868d45`
Target worktree: `feat/cli-architecture-v2`
Target head: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`

The extension was loaded explicitly from the issue #471 worktree into Pi 0.80.6 while the parent
process ran from PR #438's clean, exact-head worktree. The command was:

```text
/pm-shepherd canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental --timeout-seconds 300
```

Result:

- run ID: `run-8dca003a-bd0f-4a8c-886c-af6f6371f9a2`
- run status: `completed`
- generation: `3`
- combined score: `0.9813225484852143`
- scout: `succeeded`, score `0.9914875553891529`
- validator: `succeeded`, score `0.9712617560666672`
- hard gates: none
- persisted file mode: `0600`
- persisted summaries: fixed `lane_succeeded` categories; no model/provider free text
- child tools: none by enforced SDK contract
- global active lease: released after terminal persistence

Both summaries stayed within the host-supplied PR/check snapshot. They identified the visible state
as open, draft, merge-clean, all substantive checks successful, website deploy skipped, and no
recorded review decision. After completion, `git status --porcelain` remained empty, local and
GitHub heads still matched the exact candidate, and PR #438 remained open, draft, and CLEAN. This
was a full rerun after all ten critical and six warning deep-review findings were corrected; the
earlier generation-1 and generation-2 canaries are superseded as release evidence. No GitHub write,
connector call, credential read, child session persistence, or reverse ETL occurred.
