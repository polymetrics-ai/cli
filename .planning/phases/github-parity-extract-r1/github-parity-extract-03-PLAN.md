---
phase: github-parity-extract-r1
plan: "03"
type: execute
wave: 3
depends_on: ["01", "02"]
files_modified:
  - .planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json
  - .planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.md
  - .planning/phases/github-parity-extract-r1/VERIFICATION.md
  - .planning/phases/github-parity-extract-r1/PR-BODY.md
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "D-07: The sustained live sweep crosses the declared local budget, records local admission before a provider 429, and shows redacted nonzero GitHub headroom through GET /rate_limit."
    - "D-08: repo delete and repo delete-2 retain the destructive plan, preview, caller-supplied --confirm destructive, request-bound single-use grant, drift, expiry, and consumption flow; no text calls that human approval."
    - "D-09: repo create, archive, unarchive, and secret set remain approval-only; issue delete remains blocked for its missing GraphQL mutation executor."
    - "D-10: GitHub's eight false phantom notes stay removed, generator-owned help advertises --confirm <challenge> and plan/preview/approve semantics, and the regression is GitHub-scoped."
    - "D-11: The unchanged other-connector phantom-flag debt is counted and reported only, without a repository-wide blocking test."
    - "D-12: Generated/shared delta is confined to GitHub and the PR tells the paused consolidated sweep to rebase after this lands on main."
    - "D-13: This session records the inline/manual GSD fallback and red/green evidence because compatible isolated workers are unavailable and delegation is unauthorised."
  artifacts:
    - path: ".planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json"
      provides: "Full redacted per-command live proof accounting."
    - path: ".planning/phases/github-parity-extract-r1/PR-BODY.md"
      provides: "Reviewer-facing safety classifications, rate proof, debt count, and rebase notice."
  key_links:
    - from: "scripts/github-live-proof-sweep.mjs"
      to: ".planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json"
      via: "validated terminal proof records"
      pattern: "untestable"
---

<objective>
Run the complete GitHub live proof through the committed runner, use it as the sustained-load
rate-limit test, repair every reproducible implementation gap, and publish a redacted complete
accounting report and PR body.

Purpose: prove parity against the provider rather than fixtures.
Output: all implemented commands end as proven or concrete-untestable; final failed count is zero.
</objective>

<must_haves>
<truths>
- D-07: The sustained live sweep crosses the declared local budget, records local admission before a provider 429, and shows redacted nonzero GitHub headroom through `GET /rate_limit`.
- D-08: `repo delete` and `repo delete-2` retain the destructive plan, preview, caller-supplied `--confirm destructive`, request-bound single-use grant, drift, expiry, and consumption flow; no text calls that human approval.
- D-09: `repo create`, archive, unarchive, and secret set remain approval-only; issue delete remains blocked for its missing GraphQL mutation executor.
- D-10: GitHub's eight false phantom notes stay removed, generator-owned help advertises `--confirm <challenge>` and plan/preview/approve semantics, and the regression is GitHub-scoped.
- D-11: The unchanged other-connector phantom-flag debt is counted and reported only, without a repository-wide blocking test.
- D-12: Generated/shared delta is confined to GitHub and the PR tells the paused consolidated sweep to rebase after this lands on `main`.
- D-13: This session records the inline/manual GSD fallback and red/green evidence because compatible isolated workers are unavailable and delegation is unauthorised.
</truths>
</must_haves>

<threat_model>
| Threat | Mitigation | Verification |
|---|---|---|
| External write outside test repository | Create/identify the dedicated private test repository through `pm`; validate every write target before execution. | Report records repository alias only and rejects a mismatched config. |
| Destructive safety weakened for coverage | Run existing plan → preview → acknowledgement → approval flow unchanged; never invent a bypass. | Tests and report show destructive command help/gates remain intact. |
| Limiter reacts after GitHub throttles | Drive workload through `pm`, observe a local stop, then check `/rate_limit` with equivalent credential without logging it. | Report records local issued count/wait plus provider remaining headroom and zero 429s. |
| Ambiguous untestable claim | Require a per-command concrete reason tied to plan, permission, provider edition, or legitimately unavailable state. | Report validation rejects blank or generic reasons. |
</threat_model>

<tasks>
  <task type="execute">
    <name>Validate generated GitHub help and false-flag scope</name>
    <read_first>
      - internal/connectors/defs/github/cli_surface.json
      - internal/cli/cli.go
      - .planning/phases/github-parity-extract-r1/PR-BODY.md
    </read_first>
    <action>
      Confirm the existing derived help announces `--confirm destructive` for every destructive
      GitHub command including `repo delete-2`; confirm GitHub has no `--allow-destructive` note;
      count the unchanged non-GitHub debt and record `21 commands / 5 connectors`. Do not change
      another connector or blanket a repository-wide test.
    </action>
    <acceptance_criteria>
      - Targeted GitHub help/notes tests pass.
      - `pm github repo delete --help` exposes the actual confirmation mechanism, not a phantom flag.
      - The PR body gives exact destructive/approval-only classifications and the outside-debt count.
    </acceptance_criteria>
  </task>
  <task type="execute">
    <name>Execute and repair the complete live sweep</name>
    <read_first>
      - scripts/github-live-proof-sweep.mjs
      - .planning/phases/github-parity-extract-r1/github-parity-extract-01-PLAN.md
      - .planning/phases/github-parity-extract-r1/github-parity-extract-02-PLAN.md
      - docs/cli/reverse.md
    </read_first>
    <action>
      Build `pm`, run the harness against the dedicated private test repository with its configured
      credential name, and re-run only failed commands after each red/green repair. Capture only
      redacted report fields. Use `GET /rate_limit` to measure headroom and deliberately cross the
      declared local budget through `pm`; record that local admission stopped/waited before GitHub
      returned a 429 and that the equivalent provider request still had headroom.
    </action>
    <acceptance_criteria>
      - Every command declared implemented receives one terminal result in `LIVE-PROOF-REPORT.json`.
      - Every `proven` result has a status and returned-data/state assertion.
      - Every `untestable` result has a concrete, nonempty, verifiable reason.
      - `failed == 0`; no GitHub response in the sweep is 429.
      - The rate proof records local stop/wait, issued count, and nonzero GitHub headroom without a credential value.
    </acceptance_criteria>
  </task>
  <task type="execute">
    <name>Regenerate, confine, and publish evidence</name>
    <read_first>
      - .planning/phases/github-parity-extract-r1/VERIFICATION.md
      - .planning/phases/github-parity-extract-r1/PR-BODY.md
      - .planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json
      - .agents/agentic-delivery/references/cli-help-docs-website-parity.md
    </read_first>
    <action>
      Regenerate every owned shared artifact after final bundle changes and object-diff the ledger,
      website catalogs, generated GitHub docs, and golden transcripts. Update verification and PR
      body with the exact live tally, untestable reasons, rate proof, classification table, scope
      statement, used skills, GSD inline fallback, and notice that the paused sweep must rebase onto
      this PR's resulting main.
    </action>
    <acceptance_criteria>
      - Shared artifact delta is proven confined to GitHub, except the existing one-time embed code.
      - PR body contains exact proven / untestable-with-reason / failed counts, per-command classification, and sweep rebase notice.
      - CLI help/manual/website parity commands are recorded with results.
    </acceptance_criteria>
  </task>
</tasks>

<verification>
- targeted tests from plans 01 and 02
- `go vet ./...`
- `go test -timeout 20m ./internal/connectors/engine/ ./internal/connectors/commandrunner/ ./internal/cli/`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- individual `make` gates from `AGENTS.md`
- `go build ./cmd/pm`
- `node scripts/github-live-proof-sweep.mjs ...` live run and redacted report validation
</verification>

<success_criteria>
- The final report has no third state and no failed implemented command.
- The limiter stops local traffic before GitHub does under the real sweep.
- Destructive command safety and truthful help remain intact.
</success_criteria>
