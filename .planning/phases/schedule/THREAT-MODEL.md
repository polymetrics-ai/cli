# THREAT-MODEL — Phase 3: Scheduling

Date: 2026-06-27

---

## Assets

1. Schedule manifest files (`<root>/schedules/*.json`) — contain flow names and cron expressions.
2. OS timer files (launchd plist, systemd unit, crontab) — control what executes on the host.
3. `pm` binary path embedded in timer payloads.
4. Temporal workflow registrations (opt-in).

---

## Threats

### TH-1 — Cron expression injection into OS unit files

**Vector:** Malicious cron string containing newlines or special characters that corrupt plist/systemd/crontab file format.
**Impact:** Arbitrary command injection at scheduled execution time.
**Mitigation:**
- `ParseCron` validates against a strict allowlist (`[0-9*,/-]` per field) before any file is written.
- All values written into templates are escaped/quoted; no shell interpolation.
- Launchd plist uses `<array>` of args (not shell string), so there is no shell expansion.
- Systemd service uses `ExecStart=` with quoted args (no `sh -c`).
- Crontab line's cron fields are passed through only after `ParseCron` succeeds.

### TH-2 — Flow name injection

**Vector:** Flow name containing shell metacharacters placed into timer payloads.
**Impact:** Command injection during scheduled execution.
**Mitigation:**
- Flow name validated as slug (`[a-z0-9][a-z0-9-]*`) before manifest is saved.
- Launchd plist encodes args as XML array elements, not a shell string.
- Systemd service: flow name is a `<array>` arg (via `ExecStart=` positional args, not `sh -c`).
- Crontab backend: flow name is shell-quoted using `strconv.Quote`-equivalent quoting.

### TH-3 — Path traversal in manifest name

**Vector:** Schedule name containing `../` sequences to write manifests outside the schedules dir.
**Impact:** Overwrite arbitrary files on the filesystem.
**Mitigation:**
- Name validated as slug (no `/`, `.`, or other path chars).
- File path constructed as `filepath.Join(schedulesDir, name+".json")` and then `filepath.Clean`-checked to confirm it is still under `schedulesDir`.

### TH-4 — pm binary path not validated

**Vector:** If `os.Executable()` returns a writable path, an attacker who can write there controls what runs on each schedule.
**Impact:** Privilege escalation or persistent code execution.
**Mitigation:** `pm schedule install` prints the resolved binary path in its output; user is responsible for ensuring the binary is trusted. This phase does not run as root; it installs to user-owned directories only.

### TH-5 — Temporal credential exposure

**Vector:** Temporal connection string contains credentials (embedded in `POLYMETRICS_TEMPORAL_ADDR`).
**Impact:** Credential leakage in logs or output.
**Mitigation:**
- The address is passed to `runtimecheck`, which masks credentials in `CheckResult.Endpoint` output.
- `pm schedule install` output shows only "temporal" as backend kind, not the full address.

### TH-6 — File permission on plist/unit files

**Vector:** Plist or systemd unit file is world-writable; attacker modifies payload.
**Impact:** Arbitrary command execution at next scheduled run.
**Mitigation:** Files written with mode `0600` (`os.WriteFile(path, data, 0600)`).

---

## Out of scope

- Root-level cron/launchd daemons (this phase installs user-level only).
- Temporal RBAC / TLS configuration (handled by existing `runtimecheck` / user Temporal setup).
- Windows (out of scope for this phase).
