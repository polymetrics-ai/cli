# TDD Ledger

| Slice | Red test | Production change | Green gate | Status |
| --- | --- | --- | --- | --- |
| Bookmark identity | Store equivalent body/point anchors for one user/post and require distinct bookmarks | Include `block_type` in bookmark uniqueness and conflict lookup through an additive migration | API bookmark test | Green: two distinct 201 bookmarks and four applied migrations |
| Delete rollback | Reject comment DELETE fetch and require optimistic state restoration without an unhandled rejection | Catch rejected DELETE calls and restore prior state | Playwright comments UI test | Green: note restored and no unhandled rejection |
| Auth hydration | Capture browser errors while the blog right rail resolves an immediate Better Auth session response | Keep the server and first client render in the same loading state | Playwright blog smoke test | Green: no hydration mismatch with immediate session response |
| Secret-free image build | Build the production image without runtime secrets | Give only the builder stage an explicit public placeholder; the runner still requires runtime env | Podman image build and runtime auth initiation | Green: clean image build and GitHub OAuth initiation pass |

## Baseline

- Parent head typecheck: pass
- Parent head unit suite: 10 files, 64 tests pass
- Stacked PR #394 CI: pass
