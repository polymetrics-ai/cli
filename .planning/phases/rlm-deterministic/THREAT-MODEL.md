# THREAT-MODEL — Phase 2: RLM Deterministic Backend

## Scope

This phase is fully offline (no network, no credentials). Threat surface is limited to:
- Local filesystem reads/writes (warehouse NDJSON files).
- User-supplied spec file (JSON, parsed by stdlib).
- User-supplied table names and paths (CLI flags).
- Ledger write (local NDJSON or Postgres, reusing existing `internal/ledger`).

---

## Threats and mitigations

### T1 — Path traversal via `--in` / `--out` table names

**Threat:** A caller passes `../../etc/passwd` as an InTable or OutTable name; the engine constructs a path like `<warehouseDir>/../../etc/passwd.ndjson` and reads/writes an arbitrary file.

**Mitigation:**
- Validate that InTable and OutTable values are bare identifiers (alphanumeric, underscore, hyphen; no `/`, `.`, or path separators).
- Use `filepath.Base` on the resolved path and assert it matches the original table name + `.ndjson` extension.
- Return a sentinel `ErrInvalidTableName` on validation failure; tests cover this.
- Reuse pattern from `internal/safety` package (SSRF validation style for identifiers).

### T2 — Malformed or adversarial spec JSON

**Threat:** A spec file with extremely large weights (overflow), deeply nested structures, or oversized feature lists could cause integer overflow, OOM, or slowdown.

**Mitigation:**
- `ParseSpec` enforces: `len(Features) <= 1000`, each weight within `[-1e9, 1e9]` (negative weights rejected as validation error, very large weights capped at parse time via validation).
- Weights are `float64`; normalization divides by sum of absolute weights — no integer overflow possible.
- Feature field names validated to be non-empty strings of bounded length (<= 256 chars).

### T3 — Corrupt or adversarial InTable NDJSON

**Threat:** A malformed NDJSON line causes a panic or silent data corruption.

**Mitigation:**
- Each line is parsed with `json.Unmarshal` into `map[string]any`; on error, increment `RecordsFailed` and continue (no panic).
- Log malformed lines to stderr with line number.
- No `unsafe` usage anywhere in `internal/rlm`.

### T4 — OutTable clobbers an existing table unexpectedly

**Threat:** Caller passes an OutTable that is also used as an InTable for another pipeline step; RLM overwrites it.

**Mitigation:** This is a user configuration concern, not a security threat. Documented in RUNBOOK.md. RLM does not take any lock on InTable; it writes OutTable atomically. The flow engine (Phase 0) is responsible for dependency ordering.

### T5 — Spec file read from untrusted path

**Threat:** `--spec` points to a file that an attacker can control (e.g., a world-writable temp file).

**Mitigation:** File is read once, parsed, validated. No code execution from spec content. Spec is pure data (weights and field names). No eval, no template execution, no shell expansion.

### T6 — Model stub inadvertently called in production

**Threat:** A code path accidentally selects the `model` backend when `--mode deterministic` was intended, leaking no data (stub returns error) but misleading the user.

**Mitigation:** Backend selection is an explicit string switch with no fallback to model. Unknown mode returns `ErrUnknownMode`. `ModelAnalyzer.Run` always returns `ErrNotImplemented` — confirmed by a dedicated test (`TestModelStubReturnsNotImplemented`).

### T7 — Score value used as an execution path

**Non-threat (invariant maintained):** `_rlm_score` is a float in a materialized warehouse table. It is never passed to `exec`, `os/exec`, or any connector write primitive. Scoring output is data; actions require the separate reverse-ETL approval gate (Phase 1 / not part of this phase).

---

## Security properties guaranteed by this phase

1. No network calls. Verified by: no `net/http`, `net`, or external package usage in `internal/rlm`.
2. No credential access. Verified by: no `internal/vault` import.
3. No shell execution. Verified by: no `os/exec` import.
4. Secrets never logged. (No secrets are present in this phase.)
5. Spec content is pure data — not executable.
