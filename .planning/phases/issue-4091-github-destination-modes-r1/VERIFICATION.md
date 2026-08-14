# Issue #4091 verification

**Status:** pending implementation.

| Must-have | Result | Evidence |
| --- | --- | --- |
| Non-additive disabled/unauthorized execution sends zero provider writes | pending | targeted recorder assertions |
| Changed scope fails before provider request | pending | targeted recorder assertions |
| Single-use token cannot replay | pending | persisted approval/recorder assertions |
| Identical authorized scope runs unattended | pending | second run without token plus destination read-back |
| Modes are definition-owned and generated surface is current | pending | bundle and `surface-sync --check` |
| Target packages and repository gates pass | pending | commands and results recorded after implementation |
| Explicit PR base read-back matches integration | pending | GitHub API result recorded after PR creation |

## Live-evidence gap

No GitHub credential or private runbook access was supplied. This worker will not request or expose either. Deterministic in-process GitHub provider tests prove state-changing allowed cases and zero-send refused cases. A separately authorized operator must append credentialed live evidence to #4091 before merge if it remains required.
