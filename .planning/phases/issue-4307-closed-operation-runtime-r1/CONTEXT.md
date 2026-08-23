# Context: Issue #4307 closed operation runtime

## Locked decisions

- This is a shared engine/commandrunner/CLI foundation, not a connector implementation. No
  production connector definition may be edited and no connector-name branch is permitted.
- Each command remains bound to one exact declaration-owned operation. Callers may provide only
  typed values mapped to declared path, query, header, or body fields; the four namespaces never
  cross-bind.
- Header inputs are limited to a declared non-auth header schema and an exact
  `header.<declared-name>` mapping. Auth, proxy, cookie, host, connection, forwarding, transport,
  and runtime headers remain unreachable after canonicalization.
- Binary, multipart, status, and text are named, bounded operation kinds. They do not become a
  generic HTTP transport, a raw body mechanism, a file copier, or a reverse-ETL record writer.
- The exact declared response contract preserves every ordinary admitted provider status, header,
  body, and field element. Scope, rarity, tier, destructiveness, and unfamiliarity never filter an
  admitted result. Existing credential/transport-secret masking remains mandatory, preserves the
  field presence, and marks the masked value explicitly.
- #4305 owns the shared structured JSON body materializer and #4304 owns persisted typed-destination
  dispatch. This issue composes only through their published interfaces and does not copy either
  implementation.
- Existing main already contains portions of F4 (#4297), so the implementation must extend its
  existing closed contracts and loader/runner mirrors rather than introduce parallel builders.

## GSD discussion record

`scripts/gsd prompt discuss-phase 4307 --auto` was generated and executed inline. The firstmate
brief supplies the product decisions normally gathered interactively: bounded typed declarations,
no generic escape hatches, two synthetic identities, zero-I/O rejection proof, and no credentialed
provider execution. Inline/manual execution is required because #4307 is not a numbered ROADMAP
phase and the task worktree does not permit spawning GSD roles.
