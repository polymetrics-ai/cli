# TDD ledger — Issue 3985 connector canon

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Runtime executability gate | `make connector-runtime-preflight` must fail before the target exists. | Named target runs `TestEveryImplementedCommandPassesRuntimePreflight`; `make verify` invokes it. | Planned |
| Canon integrity gate | `make connector-canon-check` must fail before the target/canon exists. | Deterministic script checks source reports, archive metadata, required procedure sections, and Makefile integration. | Planned |
| Archive correctness | Existing source entry points lack a current/superseded index. | Hash-verified copies exist and retired entry points point to the archive. | Planned |
| Outward documentation | Current README/docs/website do not link a single canon or state the zero-artifact certification baseline. | All three layers link the canon and distinguish readiness/preflight from certification. | Planned |

## Guard meanings

- Runtime preflight is a no-network, no-credential **admission** test. It prevents declared-but-
  blocked commands, but does not prove fixtures or a live provider.
- The certification matrix/live-proof implementation is #3984. This issue only makes the current
  zero baseline and the required procedure unambiguous.
