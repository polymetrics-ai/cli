# Pending User Requests — Phase 437

## Aggregate connectors help should present GitHub only as an example

**Status:** Pending explicit user authorization to implement.

**Recorded:** 2026-07-20

**Applies to:** PR #466, branch `refactor/437-connectors-certify-native-cobra`

### Problem

`pm help connectors` currently includes detailed sections titled `GITHUB AUTHENTICATION` and
`GITHUB ETL STREAMS`. In the aggregate connectors namespace manual, that level of GitHub-specific
detail can look like general connector requirements or live GitHub data instead of documentation
for one example connector.

### Desired result

- Keep `pm help connectors` focused on connector discovery, inspection, certification, and the
  shared ETL/reverse-ETL model.
- Replace the detailed `GITHUB AUTHENTICATION` and `GITHUB ETL STREAMS` sections with one concise,
  clearly labeled `CONNECTOR EXAMPLE` section.
- State explicitly that GitHub is an illustrative connector example, not a requirement and not
  live data loaded by the help command.
- Direct users to `pm connectors inspect github` or
  `pm connectors inspect github --json` for GitHub-specific authentication, streams, write
  actions, fields, and risk notes.
- Remove GitHub credential-creation commands from the aggregate examples. Keep the generic
  discovery and inspection examples, including `pm connectors inspect github` as an explicitly
  labeled example.
- Leave the connector-specific GitHub manual and GitHub connector behavior unchanged.

### Proposed aggregate-help wording

```text
CONNECTOR EXAMPLE
  GitHub is one example of a connector with authenticated reads, ETL streams,
  and approval-gated reverse ETL actions. It is shown only to illustrate the
  shared connector model; this help command does not read credentials or load
  live GitHub data.

  Run pm connectors inspect github for GitHub-specific authentication,
  streams, configuration fields, write actions, and risk notes. Use --json for
  structured metadata.
```

The final wording may be tightened during implementation, but it must preserve all of the above
meaning.

### Red → green → refactor implementation plan

1. **RED — help contract test**
   - Update the connectors manual test to require `CONNECTOR EXAMPLE` and the explicit
     illustrative/no-live-data explanation.
   - Require the aggregate output not to contain `GITHUB AUTHENTICATION`, `GITHUB ETL STREAMS`, or
     GitHub credential-creation commands.
   - Run the focused test and record the expected failure before editing production help.

2. **GREEN — canonical help**
   - Update the canonical embedded connectors manual in `internal/cli/docs.go`.
   - Remove only the detailed GitHub-specific aggregate sections and credential examples.
   - Preserve general connector discovery, inspect, certification, safety, and exit-status help.

3. **REFACTOR — parity regeneration**
   - Regenerate `docs/cli/connectors.md`.
   - Regenerate affected CLI golden transcripts.
   - Regenerate the website CLI-reference data/page through existing repository generators.
   - Do not hand-edit generated outputs when a repository generator owns them.

4. **VERIFY**
   - Confirm `pm help connectors`, bare `pm connectors`, and `pm connectors --help` are
     byte-identical and exit successfully.
   - Confirm `pm connectors inspect github` still provides connector-specific metadata without
     reading credentials.
   - Run focused CLI help/manual/golden tests, docs and website drift checks, `go vet ./...`,
     `go test ./...`, `go build ./cmd/pm`, and `make verify` as applicable.

### Explicit non-goals

- Do not change GitHub authentication, connector streams, write actions, certification behavior,
  or credential handling.
- Do not access a PAT, macOS Keychain, live GitHub API, or any stored credential.
- Do not perform reverse ETL, external writes, sweeps, or live certification.
- Do not add dependencies or merge PR #466.

### Start condition

Do not begin RED, implementation, regeneration, commit, or push work for this request until the
user explicitly says to implement it.

## Connector catalog value errors must distinguish missing and invalid values

**Status:** Validated and pending explicit user authorization to implement.

**Recorded:** 2026-07-20

**Validated against:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

**Primary owner:** Issue #437 / PR #466 native Cobra connector-command contract. This is command
parsing and error-schema work, not Bubble Tea rendering work.

### Observed behavior

Running a value-taking flag without its value leaks pflag's internal boolean sentinel into the
user-facing error:

```text
$ pm connectors catalog --capability
error: invalid --capability "true", want read|write|cdc|query
$ echo $?
3
```

The adjacent empty and stage cases are also incorrect:

- `pm connectors catalog --capability=` silently disables the filter, prints the entire catalog,
  and exits `0`.
- `pm connectors catalog --stage` is interpreted as stage `true`, prints no rows, and exits `0`.
- `pm connectors catalog --capability --json` emits a structured envelope, but classifies the
  syntax mistake as `validation` / `validation_error` and repeats the misleading value `"true"`.
- An explicit unsupported value such as `--capability nope` is a genuinely invalid value and must
  remain distinguishable from an omitted value.

### Validation evidence

A fresh binary was built from the recorded PR head and the focused native-connectors/golden suite
was run before this request was added to the improvement backlog:

```text
HEAD: 26f98a72419010b961b5b8378ef4a695b0c0a06f
go build -o /tmp/pm-pr466-current ./cmd/pm: PASS
go test ./internal/cli -run 'TestNativeConnectors|TestGoldenTranscripts' -count=1: PASS
```

The passing suite does not currently cover the broken contract. Direct invocation produced this
matrix:

| Invocation | Exit | Observed result |
|---|---:|---|
| `catalog --capability` | 3 | validation error exposes `"true"` |
| `catalog --capability=` | 0 | all 551 connectors |
| `catalog --capability=<spaces>` | 0 | all 551 connectors |
| `catalog --capability nope` | 3 | genuine invalid-value error |
| `catalog --capability read --json` | 0 | valid catalog envelope |
| `catalog --capability --stage ga` | 3 | validation error exposes `"true"` |
| `catalog --stage` | 0 | empty human output |
| `catalog --stage=` | 0 | all 551 connectors |
| `catalog --stage nope --json` | 0 | successful `ConnectorCatalog` with `count: 0` |
| `catalog --stage alpha --json` | 0 | valid catalog envelope |
| `catalog --type` | 3 | useful legacy-removal guidance |
| `list --all --json` | 0 | valid bare boolean behavior |
| `catalog --capability --json` | 3 | one `Error` envelope plus stderr, but wrong category/message |

This establishes that the fix must cover missing, explicitly blank, whitespace-only, and invalid
values; changing only the displayed `"true"` text would leave silent-success defects in place.

### Root cause

`addConnectorsStringArrayFlag` registers `--all`, `--capability`, `--stage`, and the removed
`--type` through one `StringArrayVar` helper, then assigns `NoOptDefVal = "true"` to every flag.
Bare flags therefore parse as the string `"true"`. `connectorCatalogEntries` later treats that
sentinel as user input. It also trims an explicit empty capability to the same empty value used for
"filter not provided," while stage values have no allowlist validation at all.

`--all` is a real boolean control and may be bare. `--capability` and `--stage` require non-empty
values and must not share the boolean parsing contract. The removed `--type` flag should continue
to report its migration guidance whenever it is present.

### Desired human contract

For a missing or blank capability value, fail before catalog work with usage exit `2` and an
actionable diagnostic that never exposes `"true"`:

```text
error: --capability requires a value; allowed values: read, write, cdc, query; example: pm connectors catalog --capability read
```

For an explicitly invalid value, preserve validation exit `3` but improve the wording:

```text
error: invalid value "nope" for --capability; allowed values: read, write, cdc, query
```

Apply the same missing/blank distinction to `--stage`, with `alpha`, `beta`, and `ga` as the
documented values. An explicit unsupported stage must produce a validation error instead of a
successful empty result. Focused help should remain discoverable through
`pm connectors catalog --help` once the hierarchical help work recorded in this file is
implemented.

### Desired agent/JSON contract

With `--json`, stdout must remain exactly one `Error` envelope and stderr may retain the concise
human diagnostic. At minimum, use the existing stable fields consistently:

```json
{
  "api_version": "polymetrics.ai/v1",
  "kind": "Error",
  "error": {
    "category": "usage",
    "code": "usage_error",
    "message": "--capability requires a value; allowed values: read, write, cdc, query; example: pm connectors catalog --capability read"
  }
}
```

Do not require an agent to infer that `"true"` means "missing." Preserve the existing envelope and
generic `usage_error` compatibility. If the central error contract accepts backward-compatible
optional detail fields in this delivery, include `reason: "missing_flag_value"`,
`flag: "capability"`, `allowed_values: ["read", "write", "cdc", "query"]`, and the example as
structured values. Otherwise, keep the complete actionable message above and create a separate
cross-command typed-error issue; do not invent an incompatible JSON shape only for this command.

### Red → green → refactor implementation plan

1. **RED — pin syntax and semantic errors separately**
   - Add table-driven human and JSON tests for bare, `--flag=`, and whitespace values for
     `--capability` and `--stage`.
   - Require missing/blank values to be `usage` / `usage_error`, exit `2`, with allowed values and
     a valid example; prohibit the misleading value `"true"`.
   - Require explicit unsupported values to be `validation` / `validation_error`, exit `3`.
   - Protect valid `--capability read|write|cdc|query`, valid `--stage alpha|beta|ga`, and the
     removed `--type` migration diagnostic.
   - Protect the one-JSON-envelope contract and prove no catalog rows are emitted on an error.

2. **GREEN — separate boolean and required-value parsing**
   - Register `--all` as a real boolean flag.
   - Register `--capability` and `--stage` as value-required flags without `NoOptDefVal`.
   - Reject missing and blank values at the command boundary as usage errors before loading or
     filtering catalog entries.
   - Validate normalized explicit capability and stage values against their documented allowlists.
   - Preserve JSON success schemas, deterministic catalog membership, and removed-`--type`
     guidance.

3. **REFACTOR — share the required-value convention**
   - Reuse or extract the strict required-value validation already used by connector
     certification rather than introducing another sentinel workaround.
   - Keep presentation/error construction at the CLI boundary and catalog filtering in its domain
     helper.
   - Update runtime help, generated `docs/cli/**`, website CLI reference, and golden transcripts
     through their existing generators.

4. **VERIFY**
   - Exercise space and equals forms, human and `--json` output, pipes, and no-match filters.
   - Run focused connectors native-Cobra, error-envelope, help/manual, golden, and docs-parity
     tests, then the repository Go and `make verify` gates required by the phase.

### Explicit non-goals

- Do not change connector definitions, catalog contents, credentials, networking, or TUI behavior.
- Do not add a connector-specific incompatible JSON error schema.
- Do not access a PAT or invoke any live connector.
- Do not merge PR #466.

### Start condition

Do not begin RED, production implementation, generated-file updates, commit, push, GitHub issue
editing, or PR changes for this request until the user explicitly says to implement it.

## GitHub credential creation must validate required owner/repo configuration

**Status:** Root cause confirmed; user recovery is available; product correction is pending explicit
authorization to implement.

**Recorded:** 2026-07-20

**Validated against:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

**Diagnosis trace:**
[resolved GSD debug session](../../debug/resolved/github-credential-missing-owner.md)

### Confirmed behavior

The GitHub bundle requires separate non-secret `owner` and `repo` keys, and its check path is
`/repos/{{ config.owner }}/{{ config.repo }}`. However, the connector-specific manual generated by
`pm connectors inspect github`, `docs/GUIDE.md`, `docs/skills/pm-github/SKILL.md`, and
`internal/connectors/guide.go` still contain credential examples using the removed legacy shape
`--config repository=OWNER/REPO`.

A disposable, secret-free root reproduced the complete failure chain:

1. `credentials add` accepted `repository=octocat/Hello-World` and exited `0` even though the
   required `owner` and `repo` keys were absent.
2. `credentials inspect` showed the persisted `repository` key.
3. `credentials test` failed before network IO with
   `resolve check path: interpolate: unresolved key "owner" in config`, emitted
   `internal` / `internal_error`, and exited `1`.

This is not a PAT validity, GitHub permission, rate-limit, or API-availability error: interpolation
fails locally before authentication or HTTP. The immediate credential data is incomplete, but the
product caused or permitted it through stale examples and missing validation; it is not solely a
user mistake.

### Safe user recovery

Create a replacement credential first, using separate owner and repository-name values and an
environment reference for the token. Test the replacement before removing or recreating the old
name. Do not edit `.polymetrics` state or vault files directly.

### Desired human and agent contract

- Connector manuals and every generated/docs/skill example must use
  `--config owner=OWNER --config repo=REPO` and must identify both as required.
- `credentials add` must reject missing required connector config before reading secrets, writing
  the vault, or persisting credential metadata.
- `credentials test` must defensively revalidate old/incomplete stored credentials before connector
  effects.
- Missing config is a validation error (exit `3`), not an internal error (exit `1`). A suitable
  diagnostic is:

  ```text
  error: credential "github-pr466" is missing required config: owner, repo; recreate it with --config owner=OWNER --config repo=REPO
  ```

- JSON must remain one `Error` envelope with `category: "validation"` and a stable actionable
  message. If the central error schema supports backward-compatible details, expose the connector,
  credential name, missing keys, and recovery example as structured fields.
- `pm connectors inspect github` should distinguish required/optional/defaulted config fields; its
  examples must be executable against its own advertised schema.

### Red → green → refactor implementation plan

1. **RED — reproduce at public boundaries**
   - Add a table-driven, network-free credential lifecycle test covering absent `owner`, absent
     `repo`, the stale `repository`-only shape, and valid separate fields.
   - Prove invalid creation returns validation exit `3` with one JSON error, and records no vault,
     state, connector, or network effect.
   - Seed an incomplete legacy credential fixture and prove `credentials test` fails as validation
     before connector `Check` or HTTP.
   - Add manual/generator tests requiring separate owner/repo examples and required markers; reject
     `--config repository=` in GitHub credential guidance.

2. **GREEN — validate from the connector schema**
   - Validate required non-secret config at the app/service boundary used by every caller, before
     vault persistence and connector execution; do not rely only on CLI parsing.
   - Preserve connector defaults and secret-field separation. Do not put secret values in config or
     error output.
   - Return a typed configuration/validation error that the CLI maps to exit `3` and the existing
     JSON envelope.
   - Update canonical GitHub guide examples to `owner` plus `repo`; regenerate owned manual/docs,
     skill, golden, and website artifacts.

3. **REFACTOR — prevent schema/help drift**
   - Derive required/optional/default markers and basic credential examples from the connector
     definition where practical instead of maintaining contradictory hard-coded field names.
   - Audit connector-specific examples for keys absent from their current specs and report drift in
     the existing connector/docs validation gate.

4. **VERIFY**
   - Run the focused credentials/app/engine/manual/generator tests with a local HTTP fixture only,
     followed by CLI docs/website/golden drift checks and the phase's Go/`make verify` gates.
   - Do not use a real PAT or contact GitHub for regression tests.

### Start condition

Do not begin RED, implementation, regeneration, commit, push, GitHub issue editing, or PR changes
for this correction until the user explicitly authorizes implementation.

## Duplicate connection names must return an actionable conflict

**Status:** Validated during GitHub PR #466 testing; pending explicit implementation authorization.

**Recorded:** 2026-07-20

### Confirmed behavior

Re-running `pm connections create github-pr466-repository ... --json` after that connection was
already persisted returns exit `1` with `internal` / `internal_error` and only
`connection "github-pr466-repository" already exists`.

The duplicate is a normal user/automation state conflict, not an internal program failure and not a
GitHub error. The current namespace provides `connections list` but no focused `inspect`, `update`,
or `remove`, so users must inspect the full list and either reuse the connection or choose another
name.

### Desired contract

- Keep `create` fail-closed when a name already exists unless an explicit future idempotency or
  replacement contract is designed; never silently overwrite source credentials or stream/table
  configuration.
- Classify a duplicate as a typed conflict/validation result rather than `internal_error`.
- Human error should say the connection already exists and offer the two valid next actions:
  inspect/reuse it or choose another name.
- JSON should retain one `Error` envelope and, when the central schema supports details, identify
  `reason: "resource_already_exists"`, `resource: "connection"`, and the conflicting name.
- Add `pm connections inspect <name> --json` or an equally focused safe read surface before adding
  any update/remove behavior. Any future replacement must be explicit and must not mutate an
  existing connection accidentally.

### RED → green → refactor plan

1. Add human/JSON tests for duplicate names, exact-existing definitions, and different definitions;
   require no state mutation and no connector effects.
2. Return a typed conflict/validation error from the app boundary and map it consistently through
   the CLI error envelope and exit contract.
3. Add focused connection inspection and help/docs/website/generated parity in the owning
   connections phase; do not hand-edit generated artifacts.
4. Verify create/list/inspect under human and JSON output, then run focused app/CLI and repository
   gates.

### Start condition

Do not begin RED, production implementation, regeneration, commit, push, issue editing, or PR
changes until the user explicitly authorizes implementation.

## Connector list should render a readable capability table

**Status:** Covered in part by open Bubble Tea issue #411; stronger plain-table acceptance recorded
and pending explicit user authorization to implement.

**Recorded:** 2026-07-20

**Observed against:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

**Complete-catalog research:**

- [Connector catalog table research](research/CONNECTOR-CATALOG-TABLE-RESEARCH.md)

### Observed behavior

Human output from `pm connectors list` is an unheaded tab-separated stream whose capability value is
formatted as a Go struct:

```text
100ms  api  {Check:true Catalog:true Read:true Write:true Query:false}
activecampaign  api  {Check:true Catalog:true Read:true Write:false Query:false}
```

The current implementation writes this value with `%+v` in `internal/cli/connectors_cli.go`. It is
not a stable or readable human presentation contract.

### Existing Bubble Tea ownership

Open issue #411, `feat(ui): add connector browser and query grid`, owns the Phase 13 connector
browser. Its acceptance criteria already require filtering and manual preview without raw `%+v`
dumps. The accepted TUI design specifies a separate `pm connectors browse` Bubble Tea split view
with fuzzy filtering, capability/stage filters, Vim and arrow-key navigation, and a sanitized manual
preview.

The design also assigns the plain `pm connectors list` cleanup to the same phase, but currently says
only to reuse the `read=... write=... query=...` output used by `--all`. That does not fully capture
the user's requested headed, aligned table and omits the visible `check` and `catalog` capabilities
from the proposed human columns.

### Desired result

Keep the interactive browser and the deterministic list as complementary surfaces:

```text
CONNECTOR                 TYPE  CHECK  CATALOG  READ  WRITE  QUERY
100ms                     api   yes    yes      yes   yes    no
7shifts                   api   yes    yes      yes   yes    no
activecampaign            api   yes    yes      yes   no     no
```

- `pm connectors browse` is the Bubble Tea interactive filter/list/manual-preview experience on an
  eligible TTY.
- `pm connectors list` remains non-interactive and renders a headed, aligned table with one explicit
  column per capability: `CHECK`, `CATALOG`, `READ`, `WRITE`, and `QUERY`.
- Use semantic `yes`/`no` text in the portable plain table; do not depend on color or expose Go struct
  formatting. A richer TTY renderer may pair glyphs with words, but accessibility cannot be
  color-only.
- Preserve deterministic connector ordering and one-record-per-row behavior.
- Piped output, `--plain`, `--no-input`, CI, and `TERM=dumb` must remain non-interactive and contain no
  ANSI control sequences.
- `pm connectors list --json` keeps its existing schema and values unchanged.
- `pm connectors list --all` and `pm connectors catalog` should use the same shared table-rendering
  convention for the capability columns they expose, avoiding three drifting formats.

### Complete catalog: `connectors list --all`

The `--all` surface needs its own explicit column contract rather than inheriting only the ordinary
list columns. It currently emits 551 unheaded rows (approximately 26 KB) and shows only name, type,
and `read=true write=true query=false`, even though each catalog definition already includes release
stage, streams, and write actions.

Use the following catalog-oriented table for both unfiltered `connectors list --all` and
`connectors catalog`:

```text
CONNECTOR                 TYPE      STAGE  READ  WRITE  QUERY  STREAMS  ACTIONS
100ms                     api       ga     yes   yes    no           8        7
7shifts                   api       ga     yes   yes    no          65       76
activecampaign            api       ga     yes   no     no          11        0
```

- Keep exact connector slugs and deterministic ordering.
- Show `-` for an unspecified release stage; do not invent one.
- Right-align stream/action counts and left-align names/types/stages.
- Keep `CHECK` and `CATALOG` in the ordinary runtime-list table; prioritize stage and counts in the
  complete-catalog table because they help users compare 551 connectors.
- With no filters, `connectors catalog` and `connectors list --all` should have identical human rows.
  Catalog filters change membership, not formatting.
- Default list and `--all` currently both contain 551 connectors, so help must explain the semantic
  distinction: runtime capability summary versus full catalog definition summary. Do not claim they
  have different membership until the underlying data actually differs.
- Keep `connectors list --all --json` as the existing complete `ConnectorCatalog` envelope.
- `pm connectors browse` may add fuzzy description/name search and stage/capability filters, but
  `list --all` remains deterministic and non-interactive.

### Red → green → refactor implementation plan

1. **RED — output contracts**
   - Add focused/golden tests requiring the headings and aligned `yes`/`no` capability cells.
   - Require the output not to contain `{Check:`, `%+v`-style braces, or ANSI on a piped/plain path.
   - Protect the exact existing JSON envelope/schema with a regression test.
   - For `--all`/`catalog`, require stage plus right-aligned stream/action counts, representative
     release/capability/count cases, equal unfiltered rows, and deterministic ordering.
   - Add browser tests under issue #411 for filtering, selection, manual preview, resize, empty
     state, accessibility fallback, and Vim/arrow key parity.

2. **GREEN — shared plain renderer and browser**
   - Replace the ad hoc `fmt.Fprintf` row dumps with a small shared, sanitized table renderer.
   - Render all five implemented-connector capabilities in `connectors list`; reuse the renderer for
     catalog surfaces with the catalog column set (`STAGE`, `READ`, `WRITE`, `QUERY`, `STREAMS`,
     `ACTIONS`).
   - Implement the separate Bubble Tea browser only when issue #411's dependency gates are ready.

3. **REFACTOR — parity and ownership**
   - Keep rendering at the CLI/UI boundary; do not alter connector domain models or JSON values.
   - Update runtime help, `docs/cli/**`, website CLI reference, generated manual/golden artifacts,
     and completion/discovery metadata as applicable.

4. **VERIFY**
   - Exercise terminal, piped, `--plain`, `--no-input`, `--json`, CI, and narrow-terminal cases.
   - Verify `connectors list`, `connectors list --all`, `connectors catalog`, and the Bubble Tea
     browser against the CLI help/docs/website parity checklist.
   - Run focused tests followed by `gofmt`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and
     `make verify` when implementation is authorized.

### Start condition

Do not begin RED, production implementation, generated-file updates, commit, push, GitHub issue
editing, or PR changes for this request until the user explicitly says to implement it. When
authorized, issue #411 is the correct implementation owner; first strengthen its acceptance criteria
with this plain-table contract.

## Connector inspect should use progressive disclosure

**Status:** Existing Bubble Tea coverage is partial; deep research and a cross-phase design are
recorded, pending explicit user authorization to implement or edit GitHub issues.

**Recorded:** 2026-07-20

**Validated against:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

**Research artifacts:**

- [Connector inspect experience and proposed architecture](research/CONNECTOR-INSPECT-EXPERIENCE.md)
- [Primary sources and repository evidence](research/CONNECTOR-INSPECT-PRIMARY-SOURCES.md)

### Observed problem

`pm connectors inspect github` emits 1,583 lines and approximately 150 KB because the generic manual
renderer eagerly expands 37 streams and 231 write actions. Generated connector manuals range from 48
to 3,039 lines, with a median of 81, so this must work for every connector rather than special-case
GitHub.

### Existing plan and gap

- Issue #411 already plans `pm connectors browse` with fuzzy connector filtering, a manual preview,
  Vim/arrow navigation, and full-screen viewport promotion.
- Issue #412 already plans a searchable Glamour/viewport reader for full command and connector
  manuals.
- Neither issue explicitly makes direct `pm connectors inspect <name>` concise or provides focused
  section, stream, and action inspection.

### Recorded direction

Use one canonical connector presentation model for three layers:

1. `pm connectors inspect <name>`: one-screen overview with capability/auth/config/stream/action
   counts, safety, representative names, and next commands.
2. `--section`, `--stream`, and `--action`: focused detail without unrelated content.
3. `--full`, `connectors man`, the #411 detail view, and the #412 pager: exhaustive searchable
   reference.

Keep `pm connectors inspect <name> --json` unchanged. Inspection must stay offline, sanitized, and
credential-free. Do not add charts or a shell-backed preview.

### Ownership

Issue #411 should own the shared presentation model, concise inspector/focused selectors, and browser
preview. If that makes #411 too large, create one child issue that lands immediately before its
connector-browser slice. Issue #412 consumes the full-reference model for the pager; #417 performs
the final help/man/docs/website consolidation.

### Start condition

Do not implement, add RED tests, regenerate docs, commit, push, or edit issues #411/#412/#417 until
the user explicitly authorizes implementation or planning synchronization. The detailed research
contains the proposed acceptance criteria and RED -> GREEN -> REFACTOR sequence for that point.

## One individual manual per command, composable by tree

**Status:** Deep research and implementation planning recorded; pending explicit user authorization
and the Phase 19 / issue #417 dependency gate.

**Recorded:** 2026-07-20

**Research artifacts:**

- [Hierarchical help/manual architecture](research/HIERARCHICAL-HELP-MANUAL-ARCHITECTURE.md)
- [Primary sources and repository evidence](research/HIERARCHICAL-HELP-MANUAL-SOURCES.md)
- [Pending RED → GREEN → REFACTOR plan](PENDING-HELP-MANUAL-TREE-PLAN.md)

### Request

Create an individual focused manual for every root, parent, and leaf command. Parent commands should
have their own useful manuals and list immediate children. Nested subcommands should show their own
manuals. The system should be able to assemble a concise tree or full subtree/reference explicitly
when needed.

### Research conclusion

Use Cobra as the authority for the command tree, path, usage, aliases, flags, inheritance, groups,
and child relationships. Add typed Polymetrics manual specifications for output/JSON contracts,
security, credential behavior, approval gates, examples, exit meanings, and related guides. Compile
both into one validated manual tree consumed by terminal text, JSON, Markdown, Section 1 man pages,
website data, and the future TUI docs viewer.

Do not concatenate every descendant's full manual during ordinary `--help`. Use four explicit
rendering modes:

- one focused node;
- a concise descendant outline;
- an explicit full subtree;
- an offline/generated all-command reference.

### Ownership

This is the intended scope of open issue #417, the deliberate Phase 19 help-churn phase. The issue
is currently blocked by #411, #412, #413, #414, and #416. Central help/compiler/generator writes
should have one issue worker; parallel sessions may perform read-only inventory or review but should
not concurrently edit the registry, generated docs, goldens, or website reference.

### Start condition

Do not implement, regenerate, commit, push, update issue #417, or merge any PR for this request until
the user explicitly authorizes implementation and the parent orchestrator marks the dependency gate
ready.

## Certify help should be scoped to certification

**Status:** Validated and pending explicit user authorization to implement.

**Recorded:** 2026-07-20

**Validated against:** PR #466 head `26f98a72419010b961b5b8378ef4a695b0c0a06f`

### Observed behavior

```bash
"$PM" connectors certify --help
```

exits successfully but prints the complete `pm connectors` namespace manual. The output is 8,391
bytes and is byte-for-byte identical to `pm help connectors`, including the catalog, GitHub,
inspection, certification, examples, security, and exit-status sections.

### Why it happens

This is explicitly wired in the current native Cobra implementation:

- `newCertifyCobraCommand` calls `setManualHelp(cmd, "connectors", ...)` in
  `internal/cli/certify_cli.go`.
- `setManualHelp` renders the named embedded manual topic, so Cobra renders the `connectors` topic
  even when help was requested from the nested `certify` command.
- `TestNativeConnectorsAndCertifyHelpDiscoveryGlobalsAndMalformedInputs` currently treats
  `docs["connectors"]` as canonical for certify help too, so the oversized output is protected by
  a test rather than being an accidental terminal artifact.

### Desired result

- `pm connectors certify --help` should print a certification-specific manual, not the entire
  connectors namespace manual.
- The certify manual should explain the three supported modes:
  - single connector: `pm connectors certify <name>`;
  - batch: `pm connectors certify --all --credentials-file <file>`;
  - cleanup recovery: `pm connectors certify --sweep`.
- It should list only certification flags, validation rules, credential-reference safety,
  plan/preview/approval/cleanup guarantees, report kinds, and certification exit meanings.
- `pm connectors certify <name> --help`, `pm connectors certify --all --help`, and
  `pm connectors certify --sweep --help` should resolve to the certification manual without
  starting telemetry, loading credentials, creating workspaces, running certification, or
  sweeping resources.
- `pm help connectors`, `pm connectors`, and `pm connectors --help` should continue to render only
  the aggregate connectors namespace manual.
- The aggregate manual may summarize and link to certify help, but it should not duplicate the
  complete certification reference.

### Red → green → refactor implementation plan

1. **RED — contextual help contract**
   - Add or change focused tests so certify help must contain a certification-specific name,
     synopsis, modes, options, safety notes, and exit meanings.
   - Require certify help not to contain aggregate-only sections such as `CATALOG`, connector
     discovery actions, or connector-specific example details.
   - Prove all certify help forms exit 0 with empty stderr and zero runtime/credential/sweep
     effects.
   - Record the expected failure before changing production help wiring.

2. **GREEN — dedicated certify manual topic**
   - Add a canonical embedded certify help topic, using the existing documentation mechanism.
   - Point `newCertifyCobraCommand` at that topic instead of `connectors`.
   - Preserve current parser, flag validation, stdout/stderr, JSON help envelope, and execution
     behavior.

3. **REFACTOR — aggregate/manual parity**
   - Keep the aggregate connectors page concise and route users to the certify-specific help.
   - Generate the corresponding CLI manual page, golden transcripts, discovery/help indexes, and
     website CLI reference through existing generators.
   - Avoid duplicating authoritative certification text across hand-maintained sources.

4. **VERIFY**
   - Compare `pm connectors certify --help` across single, batch, sweep, short-help, and JSON help
     forms.
   - Confirm help produces no credential read, telemetry, filesystem workspace, runner, or sweep
     effects.
   - Confirm invalid certify actions and flags still fail as usage errors rather than falling back
     to help.
   - Run focused native-Cobra/help/manual/golden tests, generated-doc and website drift checks,
     `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify` as applicable.

### Explicit non-goals

- Do not change certification execution, supported flags, credential handling, report schemas,
  cleanup behavior, or exit-code semantics as part of this help split.
- Do not access credentials, invoke a live connector, execute reverse ETL, or perform a sweep.
- Do not add dependencies or merge PR #466.

### Start condition

Do not begin RED, implementation, regeneration, commit, or push work for this request until the
user explicitly says to implement it.
