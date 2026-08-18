# TDD ledger — PostgreSQL transport surface truthfulness

| ID | Test class | Red assertion | Green contract |
| --- | --- | --- | --- |
| T1 | Happy path | The real `app.Open` registry has no single proof that every declared PostgreSQL destination mode resolves to registered PostgreSQL source and destination executors. | All five declared modes resolve through production composition; the test asserts both returned executor references. |
| T2 | Bad path | `incremental_dedupe_history` is source-advertised but cannot pair with a PostgreSQL destination. | It is refused as an unsupported **source** mode before executor I/O, with no fallback or destination mutation. |
| T3 | Edge: exact finite mode set | A declaration can contain a source-only mode even when the destination set looks correct. | Source and destination are the same five unique modes; `full_overwrite` remains present. |

The red checkpoint is expected to fail T2/T3 before the bundle edit: preflight says the destination, rather than the source, refuses `incremental_dedupe_history`, proving the source declaration is overbroad.

## Captured red

Red: the production registry refused `incremental_dedupe_history` at the destination even though PostgreSQL's source declaration advertised it.

```text
TestOpenRefusesPostgresUnpairedHistoryModeBeforeExecutorIO:
  destination transport does not support sync mode "incremental_dedupe_history"
TestOpenPostgresTransportDeclarationsAreExactModeIntersection:
  source modes included incremental_dedupe_history beyond the five-mode destination set
```

## Green evidence

Green: after the source declaration removed only that unmatched mode, all five declared modes preflight through `app.Open`, and history mode is rejected at the source declaration before executor I/O.

- Removed only `incremental_dedupe_history` from PostgreSQL's source declaration.
- `TestOpenPreflightsEveryDeclaredPostgresDestinationMode` uses the `app.Open` production composition path and asserts the actual resolved source/destination executor references for all five modes.
- `TestOpenRefusesPostgresUnpairedHistoryModeBeforeExecutorIO` now receives the exact source-mode refusal before an executor is invoked.
- `TestOpenPostgresTransportDeclarationsAreExactModeIntersection` asserts both declaration sides remain the same five unique modes, including `full_overwrite`.
- The older `TestPostgresDefinitionDeclaresResumablePollingTransportSource` expected the unmatched source history mode. It was obsolete by the red preflight evidence above, not relaxed to hide a failure; it now asserts the same executable five-mode contract.
