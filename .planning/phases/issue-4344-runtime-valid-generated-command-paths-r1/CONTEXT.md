# #4344 Context — Runtime-valid generated command paths

## Decision record

- The source projection will derive a new command identity from the operation's
  method and HTTP path, never from `SourceID`.
- The identity encoding must be injective so literals and path parameters (and
  different parameter names) cannot collide. It will use only the command
  parser's accepted identifier alphabet.
- Existing *valid* legacy generated paths are retained. They are reachable
  public surface and must not be silently renamed. Only a legacy generated
  path that the runtime rejects is migrated to the new generated identity.
- Existing path parameter flags remain projected from the action contract;
  changing command identity must not alter their `maps_to` binding.

## Scope and safety

- Issue: Refs #4344 — fix(connectorgen): derive runtime-valid generated command paths.
- Base: `main`; merge destination: `main`.
- No credentialed provider calls, provider writes, dependencies, or reverse-ETL
  execution.
- This is a foundation repair, not a connector-local workaround. The Bitbucket
  and GitLab source-import artifacts are independent input branches; no broad
  import of their unrelated declaration changes is permitted here.

## Inline GSD fallback

The repo-local adapter resolved `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review`. Compatible isolated GSD
workers are not available and the canonical single-worker contract forbids
role spawning, so this issue executes those prompts inline.
