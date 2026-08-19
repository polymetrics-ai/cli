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

- **D-16:** The first remaining request-body gap is GitHub's eight top-level
  `oneOf` endpoints. Their 19 disjoint documented arms are individual write
  contracts, not a generic union: use the existing plural `covered_by.writes`
  relationship and generate one named action per arm. Bulk attestation delete
  requests are destructive despite using `POST`; their command contracts use
  the existing caller-supplied `--confirm destructive` acknowledgement. The
  campaigns, ProjectsV2, and Codespaces arms are approval-only creates.

- **D-17:** Several of those documented arms contain required object or
  array-of-object fields. The command surface already has a `json` flag type,
  but commandrunner currently fails closed because it does not parse one. The
  authorized request-body foundation is narrowly scoped: only a declared
  `record.*` field whose action schema admits an object or array may receive
  valid, bounded JSON; it is decoded then validated by the existing closed
  action schema before a plan exists. This is not a generic request body,
  path/query/header input, or raw HTTP escape hatch.

### Captain-authorized 957-case live-proof expansion

- **D-18:** The 957 historical `untestable_reason` rows are planning input,
  not terminal proof. Generate a source-derived, machine-readable lab manifest
  with exactly one row for each of those cases. Each row records the command,
  API target, credential class, plan/feature, PM-only setup/test/read-back/
  cleanup templates, required target-allowlist entry, destructive acknowledgement,
  residual-state check, cohort, and the earliest divergence from an existing
  proven path. The generator must prove the row count and command set against
  `LIVE-PROOF-CASES.json`; no hand-maintained count is acceptable.
- **D-19:** The factual cohorts are mutually exclusive: retained personal lab
  repository, GitHub Free sandbox organization, GitHub App or draft Marketplace
  listing, and unavailable Team/Enterprise/Codespaces/Advanced Security (or
  other named entitlement). A missing entitlement is recorded as an exact
  external prerequisite, never as an implementation failure or a blanket skip.
- **D-20:** The live-lab boundary is fail closed. It defaults to deny, explicitly
  denies `polymetrics-ai` and the worktree repositories, and permits writes only
  when both a run-owned slug and immutable GitHub ID match one exact allowlist
  entry. Ambiguous target resolution, missing IDs after resolution, an owner-only
  match, or a protected target is a pre-dispatch failure. Prefix matching alone
  is never enough.
- **D-21:** Every provider fixture lifecycle uses the built `pm github` connector
  only: creation, plan, preview, approval, execute, independent read-back,
  neutralization, and cleanup. `gh`, `gh-axi`, raw REST/GraphQL, browser controls,
  SDKs, and the GitHub UI are forbidden for proof fixtures. A missing `pm` surface
  is an explicit bootstrap impossibility with its affected command count and
  least-privilege prerequisite; it is never bypassed.
- **D-22:** The retained `karthik-sivadas/pm-live-test-direct-read-20260808081515`
  repository is not eligible for a mutation until the plan demonstrates its
  proof-program provenance and records it in the append-only cleanup ledger.
  If that cannot be established, create a fresh private lab repository through
  `pm github repo create` after the boundary tests are green. Organization creation
  and, last, deletion are real connector tests; organization deletion requires an
  exact ID, a typed destructive acknowledgement, and a ledger proof that no
  non-lab resource is referenced.
- **D-23:** Provider writes retain the existing plan → preview → approval →
  execute flow, typed destructive acknowledgement where declared, independent
  non-secret read-back, idempotent cleanup, and one terminal record per command.
  Dummy secret values remain process-only and are redacted before any report,
  manifest, ledger, argument rendering, or commit.
- **D-24:** Inline/manual GSD remains the required fallback for this session:
  the repository adapter is healthy, but isolated GSD-role delegation is not
  authorized. The planning and red/green evidence are recorded in this phase
  directory before a provider write; no new no-mistakes run starts until the
  framework and next coherent live-proof increment are committed.
- **D-25:** `repo view` remains the preserved credential-pinned PM control and
  has no command-specific target flags. Keep the rejected owner/repo form only
  as sanitized harness-regression evidence; do not invent a new `repo view`
  argument form or rewrite stored credential scope. A fresh bootstrap target is
  resolved solely through `pm github repos list-for-authenticated-user`, filtered
  in memory for exactly one private generated slug under authenticated user ID
  `6113982`, then bound by its immutable repository ID before normal fixture use.
- **D-26:** External bootstrap remains PM-only and default-deny. Current GitHub
  artifacts expose typed-confirmed `orgs delete` but no PM organization-create
  command, so no organization can be created or deleted until a PM-only creator
  can bind its immutable run-owned target. `apps create-from-manifest` consumes
  a required conversion code but PM exposes no command to issue it; no App is
  created. The only targetless probes are fixed PM direct reads: App
  authentication returned sanitized HTTP 401 with the user credential, whereas
  Marketplace-user subscriptions returned HTTP 200. The latter is a proven read,
  not evidence that a Marketplace fixture exists.
- **D-27:** The next personal-repository family is a generated label lifecycle:
  baseline PM `label list` absence; PM create; immutable-ID/color read-back; PM
  edit with same-ID edited-color assertion; and typed-confirmed PM delete followed
  by absence read-back. The resolver keeps the provider record in memory and
  records only the generated label ID and sanitized lifecycle events. This family
  is complete and establishes the reusable pattern for further reversible
  repository resources.
- **D-28:** Editable-issue proof uses a fresh generated issue, never a retained
  prior fixture. The PM lab runner must re-supply its already-validated record
  flags to both preview and execution, because the runtime may withhold declared
  fields from persisted plans. Independent PM issue-list read-back retries only a
  successful stale assertion, at most six attempts with five one-second waits;
  credential, entitlement, scope, and provider errors propagate immediately.
  The proven edit/comment lifecycle retained the same immutable issue identity,
  observed exactly one comment, and PM-closed/retained the run-owned issue under
  the existing `issue delete` safety decision. No generated title/body/comment
  content or provider payload is persisted in evidence.

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
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-parity-extract-r1/CAPTAIN-ORDER-exhaustive-ops-proof-and-boundary-20260809.md` — frozen inventory, provider-boundary, and live-proof rules.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-parity-extract-r1/brief.md` — captain-authorized 957-case expansion and final PM-only provider-lab constraint.

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
- `scripts/github-live-cases.mjs` is the current historical pre-skip classifier;
  `scripts/github-live-proof-sweep.mjs` is the current terminal-record runner.
  The new lab generator and boundary wrapper extend these GitHub-only seams rather
  than changing shared runtime behavior.

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
