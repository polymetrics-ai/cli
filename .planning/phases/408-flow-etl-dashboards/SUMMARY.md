# Summary — Phase 408 flow/ETL dashboards

Status: correction complete; execute completion false pending Shepherd handoff and independent VERIFY.

## Shepherd correction state — 2026-07-20

- Retry, not reimplementation: preserve prior behavior/docs and replace the custom headless-only production session substitute where required with a real inline Bubble Tea v2 program.
- Accepted authority: ADR-0003 + parent Stage 10 + #408 approve exact Bubble Tea `v2.0.8`, Bubbles `v2.1.1`, Lip Gloss `v2.0.5`, and test-only teatest pseudo-version `v2.0.0-20260720091843-3eef36eaaa28`; no other direct modules.
- Strict RED captured; GREEN/full non-race/focused race passed at implementation commit `c70ecf64`; correction evidence was committed and pushed at `64f1a920`. No independent VERIFY/REVIEW/INTEGRATE claim.
- Preserve race timeouts and prior local-temp reverse-smoke dispatch-boundary deviation exactly; do not rerun full race or `make verify` in CORRECT.

## Correction delivered

- Added exact authorized direct pins: Bubble Tea `v2.0.8`, Bubbles `v2.1.1`, Lip Gloss `v2.0.5`, teatest/v2 `v2.0.0-20260720091843-3eef36eaaa28`; only Go-produced transitives changed.
- `run.Model` now implements current v2 `tea.Model`; deterministic `Update` owns event/cancel/resize/key transitions and all runner/event/cancel waits execute as `tea.Cmd`.
- `Session.Execute` uses a real inline `tea.Program`; custom `inlineRenderer`/select-loop substitute removed; no alt screen; truthful final frame remains.
- Real teatest success/failure/cancel, navigation/help, responsive frames, sanitation/redaction, plus existing bypass/accessibility/bounded-event tests pass.
- Vet, build, full non-race, module integrity/tidy, and focused races pass. Full race and `make verify` intentionally not rerun in CORRECT.

## Current state

- Required docs and skills loaded.
- GSD adapter healthy for `doctor`, `list`, and `plan-phase` prompt generation.
- `programming-loop` prompt is unavailable in `scripts/gsd`; manual universal-loop fallback recorded.
- Worker branch fast-forwarded from `5b603788` to parent `b77d8f49` before production edits.
- Issue-local phase artifacts created.
- EXECUTE resumed at `361a6bec0af1ed9cf84d5bdfdd10f16458d9da4d`; all 19 existing dirty entries adopted intact.
- Focused GREEN/race and the full non-race suite pass for the correction. The earlier `make verify` pass and both full-race timeouts remain preserved evidence; independent VERIFY is the active gate.

## Delivered so far

- Issue-local GSD artifacts.
- RED tests for dual-TTY detection, flow/ETL dashboard activation and bypasses, dashboard frames, cancellation, layout/accessibility, sanitization/redaction, and bridge throttling.
- Minimal GREEN implementation:
  - stdin+stdout TTY detection (`RunOptions.StdinIsTerminal`, `DetectOptions.StdinTTY`);
  - `cmd/pm` auto mode while `cli.Run` stays plain;
  - `internal/ui/run` dashboard model, lifecycle-preserving throttle bridge, event-driven session, live inline refresh, and final scrollback frame;
  - `pm flow run` / `pm etl run` dual-TTY dashboards with parent/SIGINT cancellation propagated to engine contexts;
  - runtime help, docs/cli, and website parity updates.

## Next

1. Artifact-only correction checkpoint `64f1a920` is pushed; control returned to Shepherd.
2. After Shepherd validates CORRECT, a separate independent VERIFY stage owns the preserved full-race disposition and may run `make verify` only under explicit bounded local-temp smoke authority with plan → preview → approval → execute.
3. Do not open a sub-PR or invoke REVIEW/INTEGRATE from CORRECT.

## Blockers / human gates

- No new human dependency gate: accepted ADR-0003 + Stage 10 + #408 authorize the exact four direct pins now present.
- New dependencies beyond that exact budget remain a hard stop; NTCharts, huh, glamour, beta OTel logs, and unrelated direct modules remain forbidden.
- Independent VERIFY remains outstanding; execute completion stays false.
- Preserved full-race evidence: default 10m full race and 20m `internal/cli` retry timed out without race findings. CORRECT did not rerun them.
- Prior `make verify` passed but crossed the narrower dispatch boundary through a local temporary reverse fixture. The required sequence was preserved and no credential, remote, production, or persistent write occurred; this is a recorded prior deviation, not a fabricated verification failure.
