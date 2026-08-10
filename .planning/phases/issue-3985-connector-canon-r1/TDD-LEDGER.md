# TDD ledger — Issue 3985 connector canon

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Runtime executability gate | `make connector-runtime-preflight` failed with `No rule to make target 'connector-runtime-preflight'`. | Named target runs `TestEveryImplementedCommandPassesRuntimePreflight`; it passed on 2026-08-10, while `make verify` and `make verify-parallel` each cover the same sweep once through `test`. | Green |
| Canon integrity gate | `make connector-canon-check` failed with `No rule to make target 'connector-canon-check'`. | Deterministic script checks source reports, archive metadata, required procedure sections, corrections, and outward GitHub status wording; it passed on 2026-08-10. | Green |
| Archive correctness | Existing source entry points lacked a current/superseded index and the original GitHub coverage report was unmarked. | Source-pinned reports and preserved local originals exist under the archive; retired entry points point to the canon and the actively wrong reports carry markers. Hash/link audit passed. | Green |
| Outward documentation | README and public GitHub connector/website pages claimed stale catalog/certification status or had no single canon. | README, connector docs, architecture, website, and generated GitHub manual/status text link the canon and distinguish preflight from live certification. Docs/website checks passed. | Green |
| PostgreSQL parity tree correction | Live #3972 had 11 children, abbreviated direct-looking route descriptions, and no pre-certification owner for the four-flow/seven-mode warehouse matrix. | REST audit created #3987, attached it under #3972, updated #3972/#3978, and recorded durable warehouse-route amendments on affected children. The r2 report, final REST read-back, hash audit, and canon check make the 12-child graph visible. | Green |

## Guard meanings

- Runtime preflight is a no-network, no-credential **admission** test. It prevents declared-but-
  blocked commands, but does not prove fixtures or a live provider.
- The certification matrix/live-proof implementation is #3984. This issue only makes the current
  zero baseline and the required procedure unambiguous.
- The first canon-script run after implementation caught a case-sensitive assertion mismatch
  (`Warehouse` versus `warehouse`); the fixed target then passed. This was a test-harness repair,
  not a relaxation of the gate.
