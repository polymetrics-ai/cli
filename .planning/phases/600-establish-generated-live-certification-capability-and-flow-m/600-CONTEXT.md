# Phase 600: Generated live certification capability and flow matrices - Context

**Gathered:** 2026-08-10  
**Status:** Ready for planning  
**Issue:** https://github.com/polymetrics-ai/cli/issues/3984

<domain>
## Phase Boundary

Build the generated, machine-readable baseline that makes a connector or a
source/destination flow certifiable only when repository-derived implementation,
fixture, and accepted live evidence all agree. This phase establishes reporting
and truth enforcement; it does not add provider credentials, execute live
providers, promote capability flags, add a database writer, or remove legacy
bundle `certification.json` files.

</domain>

<decisions>
## Implementation Decisions

### Runtime-derived function inventory

- **D-01:** The generator discovers function kinds from executable source and
  runtime contracts, rather than maintaining a copied list in the matrix tool.
  A newly recognized engine operation kind must appear automatically and be
  marked unsupported until a real executor path is discoverable.
- **D-02:** Per-connector declared state comes from the bundle/native contract;
  implemented state is established through the real registry/preflight path and
  must reject `ErrUnsupportedOperation` stubs. A metadata or reachable command
  alone is never implementation proof.
- **D-03:** Fixture coverage is derived only from recorded-response fixture and
  conformance evidence associated with that exact connector/function path. It
  is not inferred from operation counts or command resolution.

### Evidence and certification rule

- **D-04:** Generated output is authoritative and drift-checked. Existing
  bundle `certification.json` files and fixture filenames are ignored as live
  evidence; they remain untouched and are reported as legacy/non-evidence
  inputs.
- **D-05:** Live proof is false unless a strict accepted-evidence record embeds
  sanitized proof from a completed real-provider or real-instance run. For an
  operation that means the actual request and response representation; a
  pointer-only claim is invalid. Redaction uses actual prepared secret values
  on the way into storage and conservatively redacts any request or response
  value that cannot be proved safe, including response bodies. A missing,
  malformed, proofless, or unsafe record cannot make any cell true.
- **D-06:** A connector is certified if and only if every applicable function
  cell has declared, implemented, fixture-tested, and live-tested all true.
  Every non-applicable cell carries a non-empty, machine-checkable reason code
  and explanation; generic `n/a` and `blocked` markers are invalid.

### Pair-scoped end-to-end flows

- **D-07:** Flow certification is modeled from the start as a source connector
  plus the mandatory local Parquet warehouse mediator plus destination connector
  plus flow kind. Connector-local flow-role facts are generated to explain
  whether each endpoint may be an API/database source or destination; they are
  not manually omitted. Source and destination may be the same connector.
- **D-08:** A flow cell is true only with end-to-end evidence that independently
  reads back both the warehouse and destination data after one real `pm`
  round-trip. Individually certified endpoints or separately passing pipeline
  legs never imply a certified pair. The current absence of durable API sinks
  and database write paths must yield an honest red baseline.
- **D-11:** A successful round trip does not imply resumable, receipt-backed,
  checkpointed, replay-identifiable, or provider-idempotent delivery. Passed
  flow evidence must serialize these guarantees separately and name a specific
  limitation for each false guarantee, so one-shot GitHub reverse writes can be
  recorded as working without being advertised as replay-safe.

### Delivery shape

- **D-09:** Add a `connectorgen` generation/check command and committed JSON
  artifacts under a dedicated source-controlled certification directory. Add a
  focused make/verification gate, tests, and architecture documentation. This
  internal developer command has no `pm` public command/help/docs/website
  surface; that parity checklist is explicitly not applicable.
- **D-10:** Keep capability and flow layers in separate coherent commits on
  this branch. Record baseline totals separately and do not claim live
  certification for any connector or pair without accepted evidence.

### the agent's Discretion

- Choose deterministic JSON field ordering and compact pair representation that
  remains inspectable without causing an unnecessary all-pairs repository-size
  explosion.
- Reuse existing bundle loading, registry, preflight, fixture, and
  `connectorgen --check` patterns where they provide stronger evidence than a
  new parallel inventory.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Delivery and safety

- `AGENTS.md` — mandatory lifecycle, generated-surface, connector, and
  verification constraints.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue-first,
  TDD, PR, and review record requirements.
- `.agents/agentic-delivery/references/required-skills-routing.md` — required
  Go skill routing.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — adapter command
  provenance and inline fallback requirements.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` —
  explicit applicability assessment for the internal generator CLI.

### Runtime truth sources

- `internal/connectors/engine/bundle.go` — operation-kind validation and
  bundle declarations.
- `internal/connectors/engine/connector.go` — engine-backed runtime adapters
  and unsupported write behavior.
- `internal/connectors/engine/direct_read.go` — bounded direct-read executor
  eligibility.
- `internal/connectors/engine/direct_write.go` — typed direct-write executor
  eligibility.
- `internal/connectors/engine/binary_read.go` — binary-download executor.
- `internal/connectors/connectors.go` — public connector capability and
  optional runtime interfaces.
- `internal/connectors/bundleregistry/registry.go` — real bundled/native
  registry composition.
- `internal/connectors/native/postgres/connector.go` and
  `internal/connectors/native/mysql/connector.go` — known unsupported database
  write implementations that the generator must report false.

### Existing derivation and certification conventions

- `cmd/connectorgen/main.go` — generator CLI command-dispatch pattern.
- `cmd/connectorgen/surfacesync.go` — committed output plus `--check` drift
  gate pattern.
- `internal/connectors/conformance/` — recorded-response evidence conventions.
- `internal/connectors/certify/` — existing runtime report implementation that
  must not be confused with accepted source-controlled live proof.
- `docs/architecture/connector-certification-design.md` — existing
  certification architecture and artifact vocabulary.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `engine.LoadAll` and the bundle types load every connector consistently.
- `bundleregistry.New` composes declarative bundles with native overrides.
- Operation direct-read/write and binary preflight paths are the real
  no-network reachability contracts for operation-backed commands.
- `connectorgen surface-sync` provides the established deterministic write and
  `--check` drift pattern.

### Established Patterns

- Connector data is declarative-first; native implementations override only
  where a real protocol path exists.
- Generated metadata must derive from the runtime source of truth, not from a
  hand-authored copied validator.
- Fixture behavior is distinct from credentialed/live certification, and no
  secret may appear in an artifact or test fixture.

### Integration Points

- `cmd/connectorgen` is the natural regeneration and drift-check entry point.
- `Makefile` has focused generator verification gates used by `make verify`.
- `internal/connectors/certifications/` is deliberately absent at this
  baseline, so the new output/evidence namespace can distinguish accepted
  proof from legacy bundle contracts.

</code_context>

<specifics>
## Specific Ideas

The captain requires two separate baselines: per-connector function cells and
pair-scoped flow cells. The first baseline is expected to be overwhelmingly
false and must state that plainly rather than upgrading filename matches or
reachable commands to certification.

</specifics>

<deferred>
## Deferred Ideas

None — provider-specific live runs, durable API destinations, and database
write executors remain owned by their respective implementation/certification
issues.

</deferred>

---

*Phase: 600-establish-generated-live-certification-capability-and-flow-m*  
*Context gathered: 2026-08-10*
