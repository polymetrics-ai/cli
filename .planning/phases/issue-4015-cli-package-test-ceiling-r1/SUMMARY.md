# Execution summary — CLI package test-ceiling foundation

The real-binary test helper previously linked one identical `pm` image for each of 18 callers. It is now a lazy package-scoped fixture guarded by `sync.Once`; `TestMain` removes its temporary directory after all tests run. Every test keeps its original independent project root and real subprocess assertions.

Red/green evidence is captured in `TDD-LEDGER.md`. The full test-name inventory remains exactly 263 runnable names before and after. The changed verbose suite took 532.694s package time / 537.29s wall, down 90.434s package time from the local baseline. No test was deleted, skipped, shortened, tagged out, or moved.

The final rebased `make verify` reports `internal/cli` in 847.535s under the unchanged 20-minute timeout (29.4% headroom). The GSD lifecycle runs inline because this task is outside Pi and its canonical contract prohibits role spawning.
