# Verification Checklist — Zoom Tasks parity, R1

## Planned checks

- [x] Live artifact, operation count, byte count, hash, and ledger comparison recorded before RED.
- [x] RED capture committed before all production declaration/foundation changes.
- [x] Redirect-safe multipart foundation proves bearer replay, multipart snapshot rebuild, fixed
  base/bearer/suffix boundaries, 307/308-only behavior, hop caps, and signed-target redaction.
- [x] JSON-only file foundation proves valid JSON reaches the provider while malformed or
  non-`.json` sources are rejected before network dispatch and parsing stays bounded.
- [x] All 17 command paths pass real commandrunner preflight.
- [x] Six direct reads and eleven direct writes run against isolated exact fixtures.
- [x] Four DELETEs and task PATCH assert 204 status-only semantics; DELETEs require destructive
  confirmation.
- [x] Endpoint ledger reconciliation is confined to `provider_module=tasks`; zero rows are
  `unsafe_or_disallowed`.
- [x] Generated CLI docs/site output retains Zoom-only changes after whole-file generation.
- [x] Fresh `pm` binary reaches base, namespace, provider group, and all 17 command help routes.
- [x] Scoped local gates, inline verify-work, and manual code review complete.

## Captured results

- Fresh live source: `https://developers.zoom.us/docs/api/tasks.md`, retrieved
  `2026-08-08T15:22:01Z`, HTTP `200`, `68,605` bytes, SHA-256
  `53920b1c473e4d8ccdad03475d6d14f13b6b0b54ce036127dd644e51850f09be`.
  It contained 17 Tasks operations and matched the inherited ledger before implementation.
- RED was committed in `043db963b`; the first failing run recorded `67/1,775/38/24` against the
  `84/1,758/44/35` target and real `unknown command` preflight results for all 17 exact paths.
- Foundation GREEN commits are `235d6a322` (declared multipart redirect) and `122b8d8d1`
  (declared JSON/extension validation). They remain separate from Tasks connector authoring.
- Connector GREEN: `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...` and the
  scoped Zoom/engine/connsdk/commandrunner/connectorgen suite passed. The six read and eleven
  write fixture lifecycle tests exercise every endpoint; status-only DELETE/PATCH and destructive
  confirmation cases are included.
- Runtime reachability: a freshly built `.tmp/pm` completed `pm help zoom`, `pm zoom`, `pm zoom
  tasks`, and `--help` for every listed Tasks command with exit status zero.
- Generated checks: `connectorgen validate`, surface sync/reconcile check, `go vet`, lint,
  tidy/docs/smoke/agent-contract/connectorgen/boundary/release gates, and `git diff --check`
  passed. Whole-file generators were used; the retained docs/site changes are Zoom-only.
- Manual inline `verify-work` plus `code-review` found no remaining issue. The runtime cannot
  register this provider category and the parent contract forbids role spawning, so the documented
  manual-GSD fallback is the applicable lifecycle route.
