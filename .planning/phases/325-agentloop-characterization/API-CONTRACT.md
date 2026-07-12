# API Contract: Replay and Safety

## Go surface

`internal/agentloop` exposes concrete data types and pure functions only:

- `LoadFixture(path string) (Fixture, error)`
- `LoadFixtures(dir string) ([]Fixture, error)`
- `ValidateFixture(Fixture) error`
- `Replay(Fixture) (ReplayResult, error)`
- `ReplayAll([]Fixture) ([]ReplayResult, error)`
- `CurrentSafetyStatus() SafetyStatus`
- `TrackedEntrypoints() []string` (defensive copy)
- `GuardDriver(path string) GuardResult`

No interface, mutable package global, external adapter, subprocess, network, git, or GitHub API is
introduced in Phase 0.

## Stable JSON result fields

Replay result field order:

1. `schema_version`
2. `incident_id`
3. `violation_code`
4. `legacy_decision`
5. `required_decision`
6. `required_exit_class`
7. `matched_expectation`

Safety status field order:

1. `schema_version`
2. `state`
3. `run_enabled`
4. `resume_enabled`
5. `code`
6. `exit_class`

Guard result appends `entrypoint` after the common safety fields. JSON encoder HTML escaping is
disabled only if required by existing CLI convention; newline termination remains deterministic.

## Process exit classes

- 0: help, status, entrypoint inventory, or successful replay-oracle match.
- 64: command/argument error or untracked entrypoint.
- 65: malformed fixture or replay expectation mismatch.
- 78: tracked driver denied because Phase 0 safety is closed.
