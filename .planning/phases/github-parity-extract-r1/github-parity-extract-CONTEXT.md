# Phase github-parity-extract: r1 — Context

**Gathered:** 2026-08-08  
**Status:** In progress — gap closure

<domain>
## Phase Boundary

Land GitHub's documented-operation parity as one GitHub-only PR, then replace its
fixture/reachability-only claim with a repeatable live proof. The live harness must
account for every GitHub command currently declared `implemented`: either prove it
against the real provider with a returned-data assertion, or record a concrete,
verifiable reason it cannot legitimately be exercised. It also declares GitHub's
existing rate-limit runtime policy, proves the existing limiter stops before GitHub
does, and keeps destructive-write safety intact.

</domain>

<decisions>
## Implementation Decisions

### Live proof and harness

- **D-01:** Commit a GitHub-specific, repeatable sweep harness. It must enumerate the
  full `implemented` surface rather than sample it, record redacted command identity,
  HTTP result, returned-data assertion, and one terminal state: `proven`,
  `untestable` with a concrete reason, or `failed`.
- **D-02:** The completion criterion is exactly `proven / untestable-with-reason /
  failed` for all 1,049 currently implemented GitHub commands. Fixture replay and
  dispatch-only reachability are supporting evidence, never a terminal outcome.
- **D-03:** Fix defects found by the sweep in this PR, including the known codeload
  redirect and unexplained live-read failures. Do not claim a command works from a
  sample or from a fixture.
- **D-04:** All writes are confined to the captain's dedicated private GitHub test
  repository. The harness must not print, persist, or summarize a credential,
  approval token, or token-derived value.

### Rate limits

- **D-05:** Do not build a second limiter. `Runtime.requesterFor` is the existing
  single request choke point; GitHub work is a `rate_limits.json` declaration and
  proof using its existing `Admit`/`Observe` wiring.
- **D-06:** GitHub policies use the existing schema's `selector.auth_types`, real
  provider-cited sources, and a non-secret account coordination scope. They must not
  key a limiter on a raw credential.
- **D-07:** The live sweep is the sustained-load proof. Deliberately exceed the
  configured local budget through `pm`, then use the same credential only through a
  redacted equivalent request to show GitHub still reports headroom. A GitHub 429 is
  a finding, not a pass. `GET /rate_limit` may measure headroom without consuming it.

### Destructive-write truthfulness and help

- **D-08:** Keep `repo delete` and `repo delete-2` destructive: plan, preview,
  caller-supplied intent acknowledgement (`--confirm destructive`), request-bound
  single-use grant, drift detection, expiry, and consumption all remain required.
  This is not human-presence or human approval.
- **D-09:** `repo create`, `repo archive`, `repo unarchive`, and `secret set` remain
  safe writes: approval applies where the runtime requires it, but they do not claim
  a destructive acknowledgement. `issue delete` remains blocked until a GraphQL
  mutation executor exists.
- **D-10:** Delete GitHub's eight false `--allow-destructive` notes and make the
  generator unable to recreate them. Generated point-of-use help must instead expose
  the real `--confirm <challenge>` and plan → preview → approve single-use grant
  sequence. The regression test is GitHub-scoped only.
- **D-11:** Count, but do not modify, equivalent phantom-flag debt in other
  connectors. The already observed count is 21 commands across five connectors;
  future parity lanes widen the check connector by connector.

### Delivery boundaries

- **D-12:** This is a GitHub-only connector PR. The paused consolidated sweep does
  not open or merge here and will rebase onto the resulting `main` afterward.
- **D-13:** The GSD adapter is available, but the runtime does not provide a
  compatible isolated worker for this session and delegation is not authorised.
  Run the generated GSD workflow inline, record that fallback in the plan and
  verification artifacts, and preserve red-before-green evidence.

### Remaining documented-operation parity

- **D-14:** Do not describe the extracted surface as complete or open the final
  GitHub PR while the 98 classified endpoint gaps remain. Close the real
  status-only/text-response/request-body foundations and their GitHub
  declarations in this same parent lane; do not create a separate task or PR.
  Each promotion must remain tied to a documented endpoint, pass the real
  command preflight, and be included in the eventual full live-proof sweep.
- **D-15:** A documented OAuth application endpoint that requires HTTP Basic
  client authentication must never silently reuse the connector's ordinary
  bearer credential. Its eventual declaration must select an explicitly
  declared, secret-backed Basic-auth contract and fail before dispatch if that
  contract is absent. Token-bearing request or response fields remain withheld
  from plans, reports, and persisted state.

### the agent's Discretion

- Select the smallest existing harness and test seams that prove returned data and
  rate-limit behavior without exposing secrets or broadening the connector scope.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Delivery and safety

- `AGENTS.md` — mandatory lifecycle, GSD evidence, command-surface and safety rules.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — one-issue/one-PR
  contract, inline fallback and review requirements.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` — required
  command help/manual/website parity checks.
- `.planning/phases/github-parity-extract-r1/PLAN.md` — prior extraction scope and
  regenerated-artifact evidence.
- `.planning/phases/github-parity-extract-r1/TDD-LEDGER.md` — prior red/green
  evidence and destructive-command corrections.
- `.planning/phases/github-parity-extract-r1/VERIFICATION.md` — existing binary
  reachability evidence, generated-artifact confinement, and local gates.

### Captain direction

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-parity-extract-r1/CAPTAIN-ORDER-prove-every-operation.md` — live-proof definition of done.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-parity-extract-r1/CAPTAIN-ORDER-github-rate-limits.md` — existing-limiter configuration and proof requirements.

### Architecture and connector contract

- `docs/migration/HANDOFF-CODEX.md` — connector migration ownership and collision rules.
- `docs/migration/conventions.md` — JSON-bundle authoring and generated-artifact rules.
- `docs/architecture/connector-architecture-v2-design.md` — declarative engine,
  embed, hook, and validation design.
- `internal/connectors/engine/rate_limit_runtime.go` — existing request choke point
  and policy resolution.
- `internal/connectors/connsdk/rate_limits.go` — `rate_limits.json` schema.
- `internal/connectors/defs/github/` — sole connector definition ownership.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/connectors/engine/rate_limit_runtime.go`: wraps every runtime requester
  with pre-request admission and response observation.
- `internal/connectors/coordination_identity.go`: derives a non-secret rate scope.
- `internal/connectors/engine/bundle_test.go`: exercises loading and validation of
  provider-cited rate-limit declarations.
- `internal/connectors/commandrunner/`: real runtime preflight and command behavior
  guard for every implemented bundle command.

### Established Patterns

- Connector definitions are declarative files embedded under
  `internal/connectors/defs/<connector>/`; shared generated artifacts must be
  regenerated by their owning commands rather than hand-edited.
- Reverse ETL is plan → preview → approval → execute. Confirmation provides
  caller-supplied intent acknowledgement and request binding, not human presence.
- CLI-visible connector changes derive help/docs/catalog/golden outputs from their
  owning generators and verify the confined diff.

### Integration Points

- A GitHub-only `rate_limits.json` is loaded through the engine's optional bundle
  loader and must be embedded with the connector definition.
- Live proof must run the built `pm` binary and use the real connector path, so it
  exercises credentials, command dispatch, runtime requesters, and rate limits.

</code_context>

<specifics>
## Specific Ideas

No generic live-provider harness should silently upgrade unrelated connectors. Make
the committed runner GitHub-specific now; extract shared machinery only when a
separate foundation issue establishes a second consumer.

</specifics>

<deferred>
## Deferred Ideas

- Fixing the 21 known phantom-flag notes in Ashby, YouTube Analytics, Recurly, Gong,
  and Gorgias is deliberately deferred to their own parity lanes.
- GitHub operations that require a missing runtime capability remain explicitly
  blocked rather than falsely classified as executable.

</deferred>

---

*Phase: github-parity-extract-r1*  
*Context gathered: 2026-08-08*
