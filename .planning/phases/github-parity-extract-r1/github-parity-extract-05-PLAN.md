---
phase: github-parity-extract-r1
plan: "05"
type: tdd
wave: 5
depends_on: ["01", "02", "03", "04"]
files_modified:
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - cmd/connectorgen/validate.go
  - cmd/connectorgen/main_test.go
  - scripts/gen-github-parity.py
  - cmd/connectorgen/github_api_surface_test.go
  - internal/connectors/certify/stages_surface_inventory_internal_test.go
  - internal/cli/reverse_cli_test.go
  - internal/cli/testdata/golden_transcripts.json
  - docs/connectors/github/MANUAL.md
  - docs/skills/pm-github/SKILL.md
  - docs/connectors/catalog/all-connectors.json
  - website/data/connectors.generated.json
  - website/lib/connectors.catalog.data.generated.json
  - internal/connectors/defs/github/api_surface.json
  - internal/connectors/defs/github/operations.json
  - internal/connectors/defs/github/writes.json
  - internal/connectors/defs/github/cli_surface.json
  - internal/connectors/defs/operation_endpoint_ledger.json
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
  - .planning/phases/github-parity-extract-r1/VERIFICATION.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "Eight documented GitHub endpoints with disjoint root oneOf request arms are promoted as 19 distinct, bounded named write actions; no executable write action retains a root union."
    - "Each promoted command maps every required record field to a typed scalar, bounded string-array, or declared structured-JSON CLI flag; structured values are parsed before plan creation and validated by the action's closed record schema."
    - "The two bulk attestation delete-request endpoints are destructive and require caller-supplied --confirm destructive plus the existing plan, preview, approval, expiry, request-binding, and single-use-grant flow. All other actions in this slice remain approval-only creates."
    - "GitHub declarations are regenerated from scripts/gen-github-parity.py; shared ledger and generated catalog deltas remain confined to github."
    - "No OAuth application endpoint, credential-minting endpoint, duplicate disposition, root-polymorphic body, or anyOf-at-least-one body is promoted by this slice."
    - "D-13 inline GSD fallback and actual RED-before-GREEN evidence are recorded before behavior changes."
  artifacts:
    - path: "scripts/gen-github-parity.py"
      provides: "Explicit source-owned table for the 19 concrete GitHub oneOf write contracts."
    - path: "internal/connectors/defs/github/writes.json"
      provides: "Separate closed record schemas for every documented oneOf arm."
    - path: "internal/connectors/defs/github/api_surface.json"
      provides: "GitHub-only covered_by.writes links from one provider endpoint to each concrete action."
  key_links:
    - from: "scripts/gen-github-parity.py"
      to: "internal/connectors/defs/github/writes.json"
      via: "explicit oneOf-arm contracts with typed record schemas and CLI mappings"
      pattern: "EXPLICIT_ONE_OF_WRITE_CONTRACTS"
    - from: "internal/connectors/defs/github/api_surface.json"
      to: "internal/connectors/defs/github/writes.json"
      via: "covered_by.writes plural endpoint coverage"
      pattern: "writes"
    - from: "cmd/connectorgen/github_api_surface_test.go"
      to: "internal/connectors/commandrunner/runner.go"
      via: "embedded bundle plus real reverse-ETL preflight"
      pattern: "Preflight"
---

<objective>
Promote the smallest unambiguous remainder of GitHub's classified request-body gaps: the eight
documented endpoints whose top-level `oneOf` represents distinct request contracts. Model every
documented arm as a separate named reverse-ETL write action rather than flattening or accepting a
union at the command boundary.

Purpose: turn the engine's existing root-union promotion refusal into explicit, executable GitHub
contracts. The only shared foundation is declared structured JSON for a named `record.*` field:
it is parsed and schema-validated before planning, not a generic transport or raw HTTP body.
Output: 19 generated, preflight-tested GitHub write commands and a GitHub-only ledger delta.
</objective>

<scope>

| Endpoint family | Actions | Gate |
|---|---:|---|
| org/user bulk attestation delete requests | 4 | destructive typed acknowledgement |
| organization campaigns | 2 | approval-only create |
| org/user ProjectsV2 fields | 7 | approval-only create |
| org/user ProjectsV2 items | 4 | approval-only create |
| authenticated-user Codespaces | 2 | approval-only create |

The contract source is `.planning/phases/github-parity-extract-r1/BLOCKED-98-ARMS.json`, derived
from GitHub's recorded OpenAPI artifact. Each action uses the documented arm's root properties and
required set; shared endpoint fields are intentionally repeated in each action because each action
is a separately preflightable command contract.

</scope>

<non_goals>

- Do not add a generic union, generic raw HTTP body, or second rate-limiter facility. A `json`
  flag is allowed only as a structured value for a declared `record.*` field whose closed schema
  admits an object or array; it cannot target a path, query, header, or literal request body.
- Do not promote the OAuth application endpoints: D-15 still requires declared secret-backed
  Basic authentication and token withholding, never fallback to the ordinary bearer credential.
- Do not promote the two `anyOf` "at least one" custom-pattern updates or the root-polymorphic
  `/user/emails` endpoints in this slice; their body modelling has different semantics.
- Do not convert duplicate/deprecated rows or credential-minting `/app/installations/.../access_tokens`.
- Do not run a credentialed write while declaring these contracts; the full live harness remains
  the later proof and retains its dedicated private-repository boundary.

</non_goals>

<threat_model>

| Threat | Mitigation | Verification |
|---|---|---|
| One `oneOf` command accepts a partial/ambiguous payload | Emit one concrete action per recorded arm; no generated action has `oneOf`/`anyOf` at its record-schema root. | Contract table verifies every action's required set and `ValidatePromotableRecordSchema` path through real preflight. |
| Several actions behind one provider URL lose coverage | Use `covered_by.writes`, never synthetic endpoint paths or a single arbitrary representative action. | Generator test asserts exact action set per endpoint and `connectorgen validate` resolves each name. |
| A destructive POST is treated as a safe create because method alone is insufficient | Mark bulk attestation deletion actions `confirm: destructive`; generated help derives the real `--confirm destructive` requirement. | Commandrunner test checks both endpoints' actions resolve `ConfirmationKindDestructive`; safe actions resolve no typed challenge. |
| Generated JSON is hand-curated or unrelated connector data moves | Change only the generator source, run it and `surface-sync`, then structurally compare shared ledger keys. | Generated ownership and confinement script reports only `github`. |
| CLI flags claim a body the runtime cannot build | Generate typed `record.*` mappings for all required arm fields. A structured JSON flag is parsed before it is put in the record, allowed only for a declared object/array schema node, and the complete action record is then validated before a plan exists. | Parser/type-mismatch/no-dispatch tests plus the new contract test and `TestEveryImplementedCommandPassesRuntimePreflight`. |

</threat_model>

<tasks>
  <task type="tdd">
    <name>RED — specify each concrete oneOf write contract</name>
    <read_first>
      - .planning/phases/github-parity-extract-r1/BLOCKED-98-ARMS.json
      - internal/connectors/engine/record_schema_promotion.go
      - cmd/connectorgen/github_api_surface_test.go
      - internal/connectors/commandrunner/github_write_contract_test.go
    </read_first>
    <action>
      Add two focused tests before implementation: (1) a commandrunner test that proves a declared
      `json` record flag is currently rejected, and that malformed/type-mismatched structured
      values cannot reach a write plan; (2) a concrete table-driven GitHub bundle test naming all
      19 actions, endpoints, required fields, CLI mappings, expected confirmation class, and
      endpoint-level plural coverage. Run both against the current bundle and record their genuine
      failures in `TDD-LEDGER.md`.
    </action>
    <acceptance_criteria>
      - The red run fails because `json` has no runtime parser and none of the arm-specific
        commands/actions are declared.
      - The test asserts semantics through the loaded embedded bundle and commandrunner preflight,
        not merely source text or a hand-copied validator rule.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>GREEN — generate 19 closed arm-specific GitHub write actions</name>
    <read_first>
      - scripts/gen-github-parity.py
      - internal/connectors/defs/github/api_surface.json
      - internal/connectors/defs/github/writes.json
      - internal/connectors/defs/github/cli_surface.json
      - cmd/connectorgen/validate.go
    </read_first>
    <action>
      First add the narrow `json` flag parser: valid JSON only, bounded, accepted only for a
      declared `record.*` object/array schema node and still subjected to complete write-schema
      validation before planning. Then add an explicit `EXPLICIT_ONE_OF_WRITE_CONTRACTS` source
      table and narrow generator branch. It emits unique readable command/action/operation names,
      documented typed root fields, exact required flags, and plural `covered_by.writes`. Mark only
      the four bulk-attestation deletion actions destructive; all other contracts use the
      already-enforced approval-only write path. Regenerate the GitHub bundle and run
      `surface-sync`; do not hand edit generated JSON.
    </action>
    <acceptance_criteria>
      - Exactly eight `operation` rows become covered, with 19 newly declared actions.
      - No newly generated record schema has root `oneOf`/`anyOf` or an empty-object-only contract.
      - Every newly declared command passes real `commandrunner.Preflight`; a valid structured
        flag can build a plan record while malformed or wrong-shaped input fails before dispatch.
      - The red table becomes green with no changed expectation hiding a missing arm.
    </acceptance_criteria>
  </task>
  <task type="verification">
    <name>Verify generated ownership, safety, CLI discovery, and confined artifacts</name>
    <action>
      Run focused generator, commandrunner, engine/certification, and CLI-help checks; build the
      binary; use the generator-owned docs/manual and website catalog commands as applicable;
      compare the shared ledger structurally before and after to prove its changed connector key is
      only `github`. Update phase verification with the exact commands/results and retain the
      full-live-sweep requirement as unfinished.
    </action>
    <acceptance_criteria>
      - `go run ./cmd/connectorgen validate internal/connectors/defs` and
        `go run ./cmd/connectorgen surface-sync --check` pass.
      - `pm help github`, `pm github <destructive-command> --help`, and a bare GitHub namespace
        render generated discovery without exposing a nonexistent flag or calling confirmation
        human approval.
      - Docs/website generated changes are GitHub-only or are intentionally postponed to the
        final combined regeneration with a recorded reason.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>Post-rebase GREEN — repair the generated CLI contract and rate-limit fixture</name>
    <action>
      The required post-rebase `internal/cli` gate must first fail on the stale GitHub golden
      transcript and on the token-authenticated reverse-ETL fixture missing its declared,
      non-secret `rate_limit_account` coordination subject. Add only the fixture's opaque
      non-secret subject, then regenerate the golden transcript and the source-owned GitHub
      documentation/catalog artifacts. Do not relax the rate-limit policy or make the runtime
      infer a subject from a secret or target repository.
    </action>
    <acceptance_criteria>
      - `TestReverseETLToGitHubCreatesPullRequestAfterApproval` reaches the mock provider through
        the declared token rate-limit policy with its explicit non-secret account subject.
      - `TestGoldenTranscripts` is green only after regeneration; generated docs/catalog diffs are
        restricted to GitHub entries.
      - The repair does not alter the 19 action contracts, classifications, or rate-limit policy.
    </acceptance_criteria>
  </task>
</tasks>

<verification>

- `go test -timeout 20m ./cmd/connectorgen/ -run 'TestGitHub(Alternative|APISurfaceOperationLedgerMetrics|DocumentedRESTSurfaceIsComplete)' -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner/ -run 'Test(EveryImplementedCommandPassesRuntimePreflight|GitHub.*Write)' -count=1`
- `go test -timeout 20m ./internal/connectors/certify/ -run 'TestSurfaceInventoryForGitHub|TestGithubWriteActionInventory' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `go vet ./...` and `go build ./cmd/pm`
- `go test -timeout 20m ./internal/cli/ -run 'Test(GoldenTranscripts|ReverseETLToGitHubCreatesPullRequestAfterApproval)$' -count=1`

</verification>

<success_criteria>

- Each of the 19 provider-documentation arms is an independently executable command contract.
- The only destructive actions in this slice accurately require caller-supplied intent
  acknowledgement; no safe create receives an invented typed challenge.
- Generated GitHub files, shared ledger, help/manual, and catalog remain source-owned and confined
  to this connector lane.
</success_criteria>
