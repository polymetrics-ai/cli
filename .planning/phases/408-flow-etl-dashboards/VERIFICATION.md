# Verification — Phase 408 flow/ETL dashboards

Status: planned; no production edits yet.

## Required local gates

Run after each coherent green slice where feasible:

```bash
gofmt -w cmd internal
git diff --check
go test ./internal/ui/... ./internal/cli/... ./internal/flow/... ./internal/app/...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Phase-specific target:

```bash
go test -race ./...
```

If a full gate is blocked by time/environment, record exact command, result, and blocker here. Do not mark `verificationPassed` true unless `make verify` exits 0.

## Focused checklist

### TUI/model/event/cancellation

- [ ] Dashboard model success frame.
- [ ] Dashboard model failure frame with sanitized/redacted error.
- [ ] Dashboard model cancellation frame after runner final event.
- [ ] Event throttle/coalesce retains lifecycle events.
- [ ] Ctrl+C cancels engine/run context and waits for Done/final frame.
- [ ] No goroutine/channel leaks under focused race test.

### Layout/accessibility/view hygiene

- [ ] Wide layout (160x45).
- [ ] Standard layout (100x30 and/or 80x24).
- [ ] Compact layout (60-79 width).
- [ ] Size guard below 60x18.
- [ ] No-color frame has no ANSI.
- [ ] ASCII fallback frame.
- [ ] Reduced-motion/static status frame.
- [ ] Accessibility/plain sequential transcript.
- [ ] Control-character sanitation.
- [ ] Secret-like value redaction.

### CLI activation and parity

- [ ] Eligible stdin+stdout TTY activates dashboard for `pm flow run`.
- [ ] Eligible stdin+stdout TTY activates dashboard for `pm etl run`.
- [ ] `--plain` bypass.
- [ ] `--json` bypass.
- [ ] `--no-input` bypass.
- [ ] `CI=1` bypass.
- [ ] `PM_NO_TUI=1` bypass.
- [ ] `TERM=dumb` bypass.
- [ ] stdin-piped fallback.
- [ ] stdout-piped fallback.
- [ ] No ANSI in machine paths.
- [ ] Plain output byte/exit parity for existing behavior.

### CLI help/docs/website parity

- [ ] `pm help flow` checked.
- [ ] `pm flow` bare namespace checked.
- [ ] `pm flow run --help` or equivalent focused help checked.
- [ ] `pm help etl` checked.
- [ ] `pm etl` bare namespace checked.
- [ ] `pm etl run --help` or equivalent focused help checked.
- [ ] `docs/cli/flow.md` updated/verified.
- [ ] `docs/cli/etl.md` updated/verified.
- [ ] `website/**` updated/verified or marked not applicable.
- [ ] Generated help/manual artifacts/goldens updated/verified or marked not applicable.

## Gate results

| Command | Result | Evidence |
|---|---|---|
| `scripts/gsd doctor` | PASS | Plan cycle. |
| `scripts/gsd list` | PASS | Plan cycle. |
| `scripts/gsd prompt plan-phase 408 --skip-research` | PASS | `/tmp/gsd-plan-408.txt`. |
| `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` | FAIL | `scripts/gsd: unknown GSD command: programming-loop`; manual fallback active. |
| `git fetch origin feat/cli-architecture-v2 && git merge --ff-only origin/feat/cli-architecture-v2` | PASS | Fast-forwarded to `b77d8f49`. |

## Manual TTY record

Pending. Will run only non-secret, non-credentialed demo/help/fixture commands.
