# Local Verification

- CI detected: no
- Local harness required: yes

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Install or lockfile validation | missing | TBD | No matching command was detected in the repo profile. |
| Format check | missing | TBD | No matching command was detected in the repo profile. |
| Lint | configured | pnpm --filter cli-polymetrics-ai lint |  |
| Typecheck or static analysis | missing | TBD | No matching command was detected in the repo profile. |
| Unit tests | missing | TBD | No matching command was detected in the repo profile. |
| Integration tests | missing | TBD | No matching command was detected in the repo profile. |
| E2E or smoke tests | missing | TBD | No matching command was detected in the repo profile. |
| Build | configured | pnpm --filter cli-polymetrics-ai build |  |
| Dependency vulnerability scan | missing_optional_tool | TBD | Add or configure dependency scanning before production release. |
| Secret scan | configured | git diff --check |  |
| Accessibility check | missing_optional_tool | TBD | Required for user-facing UI work. |
| Load or benchmark | missing_optional_tool | TBD | Required for performance-sensitive paths. |

