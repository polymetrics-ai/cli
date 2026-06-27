# THREAT-MODEL — Flow Engine (Phase 0)

## Scope

`internal/flow` package and `pm flow` CLI subcommand. Phase 0 covers read-only steps (sync,
query) only. No network writes, no credential transmission.

## Trust boundaries

- The flow manifest YAML/JSON is authored by the local user or an agent. It is read from disk.
- Step SQL is user-supplied and executed against the local warehouse only (no external network SQL).
- Lock files are local filesystem objects.
- Ledger entries are local JSONL files.

## Threats and mitigations

### T1 — Path traversal in manifest `--file` flag
- Attack: `--file ../../etc/passwd`
- Mitigation: Resolve path with `filepath.Abs` then `filepath.Clean`; reject if outside
  project root (reuse `internal/safety` which already has path validation helpers).

### T2 — Malicious SQL in query step
- Attack: manifest `sql` field contains `DROP TABLE` or exfiltration queries.
- Mitigation: Phase 0 SQL runs against the local JSONL warehouse only (no network DB). The
  local SQL engine is read-only by design (`app.QuerySQL` returns records; it does not mutate
  warehouse files). Document this invariant in the engine.

### T3 — Lock file race (TOCTOU)
- Attack: two concurrent processes both check lock before either creates it.
- Mitigation: `state.FileLock` uses `O_EXCL` which is atomic on POSIX. Already mitigated by
  existing implementation.

### T4 — Lock file left behind on crash
- Attack: process is killed; lock file remains; subsequent runs are permanently blocked.
- Mitigation: Lock file contains PID. Engine checks on startup whether the PID in the lock
  file is still alive (`os.FindProcess` + signal 0); if dead, removes stale lock. Document
  and implement stale-lock recovery in `engine.acquireLease`.

### T5 — Secret leakage in flow manifest
- Attack: user includes a credential value in the manifest `sql` field or step config.
- Mitigation: Engine never logs or prints manifest field values to stdout/stderr other than
  step IDs and statuses. SQL is not echoed in output. `--json` output contains only
  `RunResult` fields, not raw manifest content.

### T6 — Checkpoint store tampering
- Attack: user manually edits checkpoint file to mark a failed step as success.
- Mitigation: No cryptographic integrity check (out of scope for local-first tool). Document
  that `--force` is the recovery path.

### T7 — Infinite manifest (DoS)
- Attack: manifest with thousands of steps exhausts memory.
- Mitigation: Validate step count <= 1000 in `ValidateManifest`. Return `ErrManifestInvalid`.

## Out of scope (Phase 0)

- Network-write safety (Phase 1)
- Credential exposure in action payloads (Phase 1)
- Model output injection (Phase 4)
