# Local Verification

Phase: issue-392-github-only-auth

| Check | Status | Notes |
| --- | --- | --- |
| GSD adapter | fallback | `scripts/gsd prompt programming-loop ...` is not exposed; manual lifecycle is active. |
| GSD doctor | pending | Run before implementation completion. |
| Red provider test | expected fail | Targeted Vitest run failed on the three-provider array before production edits. |
| Green provider test | pending | Run after implementation. |
| Typecheck | pending | Website TypeScript check. |
| Unit tests | pending | Full website Vitest suite. |
| Build | pending | Website production build. |
| Local server | pending | Updated branch listening on `127.0.0.1:3100`. |
| Browser smoke | pending | GitHub is the only sign-in choice; key blog routes render without console/page errors. |

## Safety

- No secret values may appear in commands, logs, artifacts, commits, or PR text.
- Local OAuth credentials are sourced only from the existing ignored environment file at server startup.
- Production deployment and parent-to-main merge remain human-gated.
