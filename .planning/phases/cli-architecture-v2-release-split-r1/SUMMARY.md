# Summary — CLI Architecture v2 Cobra/Viper release split

Status: draft PR #499 open; authorized release-asset audit in progress before returning to the canonical PM review-system wait.

The approved five source squashes are reconstructed on exact latest-main base `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`. The single `internal/cli/cli.go` conflict retained current Gong behavior, Cobra was adapted to the current JSON-help signature, and only affected golden/docs outputs were regenerated.

Focused and full tests, focused race checks, vet, build, module integrity, lint, docs validation, the ordered local reverse smoke, connector validation, `make verify`, and target-toolchain vulnerability scanning passed. Seventeen deterministic current-main/candidate invocations matched exit code, stdout bytes, and stderr bytes exactly.

The candidate explicitly excludes TUI, events, logging, OpenTelemetry, PR #493 routing work, and PM review-system implementation. ADR 0002 and an additive release-split state record describe the exact boundary and parent consequences.

Captain timing authorization allowed the exact green candidate to be pushed and opened as draft PR #499 before canonical review. The PR prominently records all pending gates and is not ready to merge or release.

After the bounded release-asset audit, delivery returns to the external wait because `chore/pm-first-round-review-system-r1` is at `355510f5b0827f800dcbaa232d5804f8fcd3b407` (`docs(orchestration): record round two red evidence`) with no PR, so it is not integrated or clean. No manual/Claude/Copilot substitute is allowed. Exact-version PM packets, synthesis, independent Shepherd, no-mistakes, GitHub CI/Snyk, RC version decision, and release publication remain pending.
