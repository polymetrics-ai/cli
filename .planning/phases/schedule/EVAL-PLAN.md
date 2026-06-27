# EVAL-PLAN — Phase 3: Scheduling

Date: 2026-06-27

---

## Evaluation approach

This is a CLI/backend phase. Evaluation is test-driven and deterministic, not probabilistic. No LLM calls in this phase; no prompt evals needed.

---

## Correctness evals

### E-1 Cron parser correctness
Method: table-driven unit tests (Group A in TEST-PLAN.md).
Pass criterion: all 12 A-cases pass; `Next()` returns exact expected times.

### E-2 Unit file rendering accuracy
Method: golden file comparison.
Pass criterion: rendered plist/service/timer/crontab-line is byte-for-byte identical to golden fixture.
Update golden: `go test ./internal/schedule/... -update` (implement `-update` flag in tests).

### E-3 Backend selection logic
Method: unit tests with injected OS + probe function (Group D in TEST-PLAN.md).
Pass criterion: all 5 D-cases select correct backend kind.

### E-4 CLI round-trip
Method: `cli.Run` integration tests (Group E in TEST-PLAN.md) against temp root.
Pass criterion: all 7 E-cases pass with correct exit codes and JSON output.

---

## Regression evals (phase gate)

```bash
export GOTOOLCHAIN=auto
gofmt -w internal/schedule internal/cli
go vet ./...
go test ./internal/schedule/... ./internal/cli/...
go build ./cmd/pm
make verify
```

All must exit 0 for the phase to be complete.

---

## Prompt evals

Not applicable — this phase contains no LLM-generated output or prompt-driven logic.

---

## Performance

No explicit benchmark targets for this phase. The schedule operations are one-shot (file writes + OS exec) and not on a hot path. If `CronExpr.Next()` is called in a tight loop (e.g., generating a list of upcoming runs), it should complete in < 1ms per call — no benchmark needed at this stage.
