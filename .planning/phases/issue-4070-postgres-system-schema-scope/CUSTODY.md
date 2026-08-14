# #4070 custody record

Captured before the task worktree changed branches on 2026-08-12.

| Identity | Evidence | Disposition |
| --- | --- | --- |
| Task worktree | `/Users/karthiksivadas/.treehouse/cli-83d592/21/cli`, initially detached and clean at `da7747a796049601a179a97c025bfb05f011f1e8` | Isolated task laboratory confirmed. |
| #3976 remote PR head | `origin/feat/3976-postgres-dynamic-catalog` = `24d0055f5c9421f0bd18d0d33313a3917210ba84` | Remains the draft #4065 head; not moved. |
| Preserved #3976 local head | `/Users/karthiksivadas/.treehouse/cli-83d592/15/cli` branch = `46ee78620dfecb091090e40fc8986025f073d6a9`, clean | Remains immutable and untouched. |
| Parked #3976 pipeline head | no-mistakes run `01KZRCHTJ7QS1SFXNSNHV4H605`, `49a9386d2c629e53594c6bba1dd9a74a05b3bff5` | Run is `pipeline_owned`, `fix_review`, 5/5 exhausted; no response, sync, or rewrite was performed. |
| Structured next action | `continue_active_run: no-mistakes axi status`; `sync --recover` is not offered | This task must not execute the old run's next action. |
| New #4070 branch | Exact local fetch of the preserved commit graph into `fix/4070-postgres-system-schema-scope`, then `git switch` | New branch head is exactly `49a9386`; no existing ref was moved or overwritten. |

The child PR will compare this branch to exactly
`feat/3976-postgres-dynamic-catalog`. Its base is intentionally the draft
#4065 branch, never `main` or `feat/3972-postgres-parity`.
