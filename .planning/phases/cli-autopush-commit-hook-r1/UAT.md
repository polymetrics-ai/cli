# UAT — Time-boxed rebase-safe post-commit push

Manual `verify-work` fallback: all deliverables are deterministic local-Git
behaviors, so no human judgment is required.

| Deliverable | Automated evidence | Result |
| --- | --- | --- |
| Refusal and replay safety | `scripts/tests/post-commit-autopush.sh` runs a real two-commit rebase, manual and clean squash/cherry-pick/revert completion, delayed children during and after a manual merge, stale/default-pushurl, detached, and opt-out cases. | Pass |
| Rate and worktree safety | The same harness proves a shared linked-worktree rate ref, concurrent expired-rate compare-and-swap, ten-minute skip, expiry catch-up, and a receive-delayed detached push. | Pass |
| Non-force rejection | The harness creates a real diverged feature branch; the remote rejects the local update, remains unchanged, and the commit exits zero with one local log line. | Pass |
| Operator controls | `docs/GUIDE.md` documents manual hook-path opt-in and `PM_NO_AUTOPUSH=1`. | Pass |
