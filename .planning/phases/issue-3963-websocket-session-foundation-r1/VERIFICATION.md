# Verification Checklist — closed WebSocket session operation foundation, R1

- [x] GSD/manual fallback, source resolution, required skills, and inline discussion recorded.
- [x] Inherited loader RED rerun and captured before production edits.
- [x] Operation schema rejects unknown execution blocks and invalid WebSocket declarations.
- [x] Loader enforces fixed GET/relative path/subprotocol/closed initial schema/positive bounds.
- [x] Loopback transport proves handshake, auth boundary, masking, bounds, cancellation, close,
  redaction, and malformed-frame rejection.
- [x] Commandrunner permits only matching implemented operation declarations and rejects
  caller-controlled WebSocket transport controls.
- [x] The CLI preserves explicitly selected generic controls so a closed session rejects them rather
  than silently accepting an ignored executor-specific option; static conformance recognizes only
  implemented GET `covered_by.websocket_session` operation coverage.
- [x] No new dependency, credentialed call, reverse-ETL execution, or generic tool surface.
- [x] Targeted tests, full `internal/cli`, vet, build, declarative surface validation, and clean
  generated-surface regeneration are recorded after the `f96a47e80` main rebase.
- [ ] Code review, stacked-PR handoff, and consumer built-binary proof recorded.

## Main rebase re-gate — 2026-08-10

- The foundation was replayed onto rebased parent `3212be755` (which is based on
  `origin/main` `f96a47e80`). The only source conflicts were combined deliberately: plural
  `covered_by.writes`, direct-write coverage, and closed WebSocket-session coverage all remain
  recognized by generator, validator, batch inventory, and static conformance.
- Post-rebase checks passed: focused engine/connsdk/commandrunner/conformance/connectorgen tests,
  `go test -count=1 -timeout 20m ./internal/cli`, `go vet` for the changed packages,
  `go build ./cmd/pm`, `connectorgen validate` (552 bundles), and `surface-sync --check` (552
  bundles, zero drift).
- `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` and
  `pnpm --dir website run gen:website-data` were rerun after the parent regeneration and left the
  worktree clean. A built-binary WebSocket command remains intentionally impossible until the
  #3935 Zoom AI Services consumer declares one; that consumer must perform the final reachability
  proof with the non-zero unresolved-command behavior from #3964.
