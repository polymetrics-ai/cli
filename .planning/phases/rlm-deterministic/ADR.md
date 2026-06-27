# ADR — Phase 2: RLM Deterministic Backend

---

## ADR-001: Spec file format is JSON, not YAML

**Status:** Accepted

**Context:** The phase prompt mentions "YAML spec files." The repo's non-negotiable invariant requires Go stdlib only; `encoding/yaml` is not in stdlib and there is no existing YAML parser in the codebase.

**Decision:** Spec files are JSON. The `--spec` flag accepts a path to a `.json` file parsed with `encoding/json`. This is fully stdlib-compatible and consistent with existing config files in the repo.

**Consequences:**
- Users must write specs in JSON (not YAML). This is a minor UX trade-off.
- No new dependency.
- If YAML is desired in a future phase, a lightweight single-file parser could be added or the gate could be revisited.

---

## ADR-002: OutTable is a flat NDJSON (not nested in `localRawRecord` envelope)

**Status:** Accepted

**Context:** InTable uses the `localRawRecord` envelope with a `record` field. OutTable could either: (a) wrap in the same envelope, or (b) write flat records with `_rlm_*` fields merged alongside source fields.

**Decision:** OutTable writes flat JSON objects. This makes the output immediately queryable with `pm query` and easy to inspect with standard tools without needing to unwrap a `record` field.

**Consequences:**
- OutTable is human-readable and directly usable as a source for `pm reverse`.
- OutTable cannot be used directly as InTable for another RLM run that expects `localRawRecord` format (but this use case is not needed in Phase 2).
- If re-ingestion into raw warehouse is needed, a future `pm query` step can reformat.

---

## ADR-003: Scoring is purely in-memory (no SQL engine)

**Status:** Accepted

**Context:** The phase prompt says "SQL/weighted-feature scoring." The repo has a DuckDB query engine (`internal/app/query_engine_duckdb.go`). We could implement scoring as a DuckDB SQL query.

**Decision:** Implement scoring in pure Go (iterate rows, apply feature rules, compute scores). DuckDB is an optional dependency (`query_engine_duckdb.go` is build-tag gated). Scoring in Go is simpler, fully offline without any optional dep, and easier to test deterministically.

**Consequences:**
- No dependency on DuckDB for RLM scoring.
- Scoring logic is ~50 lines of Go vs. a SQL string — more testable.
- If scoring needs to scale to millions of rows, a SQL-engine path can be added later as an opt-in backend without breaking the interface.

---

## ADR-004: LedgerAppender interface in `internal/rlm` (no direct `internal/ledger` import)

**Status:** Accepted

**Context:** `DeterministicAnalyzer` should write a ledger record. It could directly import `internal/ledger`, but that creates a coupling that makes testing harder (need a real ledger for unit tests).

**Decision:** Define a minimal `LedgerAppender` interface in `internal/rlm`. The CLI wiring in `internal/cli/rlm_cli.go` provides a `ledger.JSONLedger` (or `ledger.PostgresLedger`) as the concrete implementation. Tests can pass a no-op or in-memory implementation.

**Consequences:**
- `internal/rlm` has no import of `internal/ledger` — clean separation.
- Tests are simpler.
- One extra interface to define, but it is small (single method).

---

## ADR-005: Model backend is a compile-time stub only — no runtime selection path to real implementation

**Status:** Accepted

**Context:** Phase 4 will implement the real model backend. In Phase 2, `ModelAnalyzer` exists only so that the `Analyzer` interface and `NewAnalyzer("model")` compile and return a concrete type.

**Decision:** `ModelAnalyzer.Run` unconditionally returns `ErrNotImplemented`. No conditional code, no feature flag, no environment variable to enable it. The only way to enable real model calls is to write the implementation in Phase 4 after human gate approval.

**Consequences:**
- Zero risk of accidental model invocation in Phase 2 or Phase 3.
- Tested explicitly: `TestModelStubReturnsNotImplemented`.
- Phase 4 will replace the body of `ModelAnalyzer.Run` — the interface and type remain unchanged.
