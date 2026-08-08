# TDD Ledger — scoped operation-surface reconciliation, R1

## RED — captured before production code

The red test invokes `surface-reconcile --notes-contains provider_module=healthcare` against a
synthetic bundle row whose provider-module provenance lives only in `operation.notes`. The installed
tool failed before implementation because it did not recognize that selector:

```text
--- FAIL: TestRunSurfaceReconcileNotesContainsScopesOperationRows (0.01s)
    surfacereconcile_test.go:67: surface-reconcile unmatched --notes-contains exit = 2, want 0; stdout= stderr=connectorgen surface-reconcile: unknown flag "--notes-contains"
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.741s
FAIL
```

## GREEN — captured

Add the selector without changing existing `--reason-contains` behavior. Both selectors must be
conjunctive when present, so callers cannot accidentally broaden a category-scoped reconciliation.

The implementation adds `--notes-contains`, searches the durable `operation.notes` provenance, and
keeps it conjunctive with `--reason-contains`. Green evidence:

```text
$ go test -count=1 ./cmd/connectorgen
ok  	polymetrics.ai/cmd/connectorgen	12.348s
```
