# TDD Ledger — scoped operation-surface reconciliation, R1

## RED — pending

The red test invokes `surface-reconcile --notes-contains provider_module=healthcare` against a
synthetic bundle row whose provider-module provenance lives only in `operation.notes`. The installed
tool must fail before implementation because it does not recognize that selector.

## GREEN — pending

Add the selector without changing existing `--reason-contains` behavior. Both selectors must be
conjunctive when present, so callers cannot accidentally broaden a category-scoped reconciliation.
