# VERIFICATION — issue #3715 Pi clean project-only workers

Status: implementation verified locally; repository gates, GSD review, and no-mistakes remain.

## Checklist

- [x] The two Pi files are generated from the canonical contract and drift check passes.
- [x] `clean-project` discovers exactly `pm-delivery-worker` and `pm-connector-worker`.
- [x] A hostile global fixture, retained legacy project roles, and the extension bundle are absent
  from a clean worker's discovered roster.
- [x] Generated YAML has required `name` and `description`; the bounded `tools` declaration omits
  `subagent`.
- [x] The extension removes `subagent` from child allowlists even if requested and rejects nested
  depth before spawn.
- [x] Project confirmation/trust behavior remains active for clean-project, including
  noninteractive refusal.
- [x] The extension README accurately cites official Pi format, discovery/scope, and tool docs.
- [x] No legacy `.pi/agents` file, global Pi state, `.gsd`, GSD Pi extension, or GSD skill changed.
- [x] Focused Go/Pi tests pass.
- [x] Applicable repository gates pass.
- [x] Inline GSD verify/review evidence and a reasoned review disposition are recorded.

## Executed verification

- `bash scripts/tests/pi-clean-project-agents.sh`: passed. It imports the real extension entry
  point using the installed Pi dependency graph, verifies clean default discovery/entry behavior,
  project confirmation refusal without UI, user/project precedence, child tool stripping, depth
  rejection, no-extension child arguments, and a hostile project-worker symlink.
- `go test ./internal/agentcontract ./cmd/agentcontractgen -count=1`: passed.
- `go test ./internal/cli -count=1`: passed.
- `go vet ./internal/agentcontract ./cmd/agentcontractgen`: passed.
- `go build ./cmd/pm`: passed.
- `go run ./cmd/agentcontractgen sync`: generated 2 registered Pi projections.
- `go run ./cmd/agentcontractgen check` and `make agent-contract-check`: passed.
- `make tidy-check`, `make lint`, `make docs-check-no-build`, and `make smoke-no-build`: passed.
- `make connectorgen-validate`: 550 connectors, zero findings.
- `make connectorgen-surface-sync`: 550 connectors, zero drift.
- `make connector-boundary` and `make release-workflow-check`: passed.
- `git diff --check`: passed.

The full `go test ./...` / `make verify` monolith was intentionally not run locally under the
repository's per-command timeout policy. The relevant package tests, `internal/cli`, and each
non-monolithic `make verify` gate were run separately.

## GSD fallback and review

The installed Pi adapter exposes the GSD prompts but not a compatible executable
`programming-loop`, and its `execute-phase`/`verify-work` command contract expects numeric phases.
This issue uses an issue-named phase and prohibits spawned roles, so the generated prompts were
executed inline. `SUMMARY.md` provides coverage, `UAT.md` records the automated acceptance evidence,
and `REVIEW.md` records the standard inline review and its fixed symlink finding.
