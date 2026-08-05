# PLAN — clean project-local Claude workers

Issue: #3716. Parent: #3714. Branch: `fm/cli-agents-wave-claude-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed; `programming-loop` is absent.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review|ship`:
  passed.
- `scripts/gsd prompt discuss-phase issue-3716-claude-workers --auto`: generated and executed
  inline as the discussion record.
- `scripts/gsd prompt plan-phase issue-3716-claude-workers --tdd --skip-research`: generated and
  executed inline through the plan and RED/GREEN/REFACTOR slices below.
- Generated `execute-phase`, `verify-work`, and `code-review` prompts were executed inline. The
  Pi adapter cannot supply the workflow's reviewer/UAT subagents, so their prescribed artifacts
  are recorded manually below; no agent role was spawned because the canonical base worker
  declares delegation `none`.

## Required skills loaded

- `no-mistakes` v1.41.2 — exact child-gate authority and non-default-base workaround.
- `github-issue-first-delivery` — issue/stacked PR linkage and hard-stop rules.
- `golang-how-to` — task-specific Go skill router.
- `golang-testing` — isolated red/green frontmatter, drift, and ambient-delegation regression
  tests.
- `golang-cli` — generator command behavior and stdout/stderr boundaries.
- `golang-error-handling`, `golang-safety`, and `golang-security` — bounded projection paths,
  defensive errors, and no expansion across the repository boundary.
- `golang-lint` and `golang-documentation` — focused review and accurate generated/evidence text.
- `gsd-programming-loop` was read for its TDD lifecycle, but its absent adapter command is not
  invoked; this phase follows the canonical installed GSD sequence instead.

## Deliverable slices

### Slice 1 — RED: Claude projection and isolation contract

1. Add a focused test describing both Claude projections as required full files with parseable YAML
   frontmatter, exact base tools plus trusted plugin-qualified skill preloads,
   `permissionMode: default`, and no `Agent` or runtime `Skill` capability.
2. Model an unrelated ambient agent in the test fixture and make the assertion fail against the
   current generator, which has no Claude frontmatter projection and therefore cannot prove that
   the ambient agent is unreachable.
3. Record the failing focused test command and failure reason in `TDD-LEDGER.md` before production
   code or canonical source changes.

### Slice 2 — GREEN: canonical policy, renderer, generation, and drift enforcement

1. Add a validated Claude harness policy to
   `.agents/agentic-delivery/canonical/delivery-contract.json`, including the official source URL,
   project discovery/precedence evidence, exact tool allowlist, and default permission mode.
2. Extend `internal/agentcontract` to render complete Markdown/YAML Claude files from that policy,
   validate their frontmatter, create missing required targets only through the existing
   root-contained atomic writer, and compare full files for drift.
3. Make the two Claude projection targets required, generate them with
   `go run ./cmd/agentcontractgen sync`, and rerun the focused tests plus the canonical check.
4. Add regression coverage that adding `Agent` or `Skill`, removing a trusted preload, or
   weakening the `Agent`/`Task`/`Skill` denylist is rejected, while sync restores the
   no-delegation file.

### Slice 3 — REFACTOR: runtime smoke, documentation, and scope audit

1. Preserve the prior trusted-project smoke for direct `Agent` omission. Record the clean-home
   discovery smoke as **NOT PERFORMED** until a real authenticated Claude session can run from a
   clean trusted home containing unrelated global definitions; do not substitute static evidence
   for that runtime criterion.
2. Document the exact official behavior and caveats in generated projection metadata and phase
   verification: managed settings and CLI `--agents` override project definitions,
   plugin-qualified preloads avoid personal/project skill collisions, runtime `Skill` omission
   excludes `context: fork` routes, and `Agent` omission plus the denylist blocks direct invocation
   without erasing ambient definitions from their own scopes.
3. Run review, validation, diff, and path-scope checks. Do not touch installed plugins or any home
   directory.

### Slice 4 — REVIEW FIX: trusted skill origins, inventory, and portability

1. Replace collision-prone `Skill(name)` tool entries with Claude's documented `skills`
   frontmatter, using only plugin-qualified `cc-skills-golang:*` and
   `frontend-design:frontend-design` identifiers. Omit and deny the runtime `Skill` tool alongside
   `Agent` and `Task`; record the three repository-routed design skills that cannot be safely
   qualified and the website-work cost.
2. Inventory `.claude/agents` recursively during the canonical check, reject symlinks, unexpected
   Markdown definitions, duplicate names, and missing canonical paths before comparing generated
   content.
3. Validate slash-separated canonical paths with `io/fs` and `path`, and invoke the extensionless
   JavaScript GSD adapter through Node so checks remain executable on Windows.
4. Preserve the captain-authorized clean-home runtime smoke as **NOT PERFORMED**. Distinguish the
   official namespace guarantee and static generated-file checks from authenticated runtime
   discovery, source-version pinning, and managed/CLI override behavior.

### Slice 5 — REVIEW FIX: nested discovery scopes and line endings

1. Move Claude inventory enforcement from the root `.claude/agents` subtree to a repository-wide
   walk that recognizes every nested `.claude/agents` scope, skips Git metadata, and fails closed
   on scope symlinks, duplicate canonical names, or unexpected Markdown definitions.
2. Normalize CRLF to LF at the Claude projection parsing and whole-file comparison boundary so a
   Windows checkout remains semantically identical to the canonical renderer without weakening
   any frontmatter, body, inventory, or denylist check.
3. Add focused regressions for a duplicate and an unexpected definition under
   `website/.claude/agents`, plus CRLF parsing, drift checking, and no-op sync behavior.
4. Keep the canonical-source-only rule, the `Agent`/`Task`/`Skill` denylist, and the
   captain-authorized clean-home smoke wording unchanged.

### Slice 6 — REVIEW FIX: metadata identity and canonical EOL output

1. Restrict inventory pruning to the repository root's exact `.git` metadata directory so
   case-variant ordinary directories and nested project paths remain discoverable on
   case-sensitive filesystems.
2. Canonicalize Claude renderer output to LF, then normalize both expected and checked-out Claude
   whole-file bytes at check/sync boundaries so canonical strings containing CRLF cannot create a
   rewrite loop.
3. Add focused regressions for `.GIT` and `.Git` nested agent definitions, exact root `.git`
   metadata pruning, canonical CRLF rendering, successful checking, and no-op repeated sync.
4. Preserve canonical ownership, qualified skill preloads, runtime `Skill` denial, the full
   `Agent`/`Task`/`Skill` denylist, and the clean-home **NOT PERFORMED** wording.

### Slice 7 — REVIEW FIX: fixed-point Claude EOL normalization

1. Replace pairwise CRLF rewriting with a linear canonicalizer that collapses every carriage-return
   run immediately before LF in one pass while preserving bare carriage returns.
2. Add direct table-driven coverage for ordinary CRLF, repeated carriage-return runs, multiple
   runs, unchanged LF, and preserved bare carriage returns; assert a second normalization is
   byte-identical to the first.
3. Strengthen canonical-source render/check/sync coverage with repeated carriage returns before LF
   and prove rendering is LF-only and a second sync performs zero updates.
4. Preserve generator-only ownership of both Claude workers, qualified skill preloads, runtime
   `Skill` denial, the complete `Agent`/`Task`/`Skill` denylist, and the clean-home **NOT
   PERFORMED** wording.

## Verification

- RED/GREEN: `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- Canonical/drift: `go run ./cmd/agentcontractgen sync` and
  `go run ./cmd/agentcontractgen check`.
- Claude smoke: prior installed `claude` v2.1.222 trusted-project evidence remains bounded to direct
  `Agent` omission. Clean trusted-home discovery with unrelated global definitions is **NOT
  PERFORMED** and requires a real authenticated Claude session.
- Formatting/static: `gofmt -w cmd internal`, `go vet ./internal/agentcontract ./cmd/agentcontractgen`,
  and `make lint`.
- Repository gates applicable to this scope: `make tidy-check`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, `make connector-boundary`, and
  `make release-workflow-check`; connector validation/surface checks are run separately according
  to the repository timeout policy.
- Delivery: `no-mistakes axi run --intent <complete issue intent> --skip=push,pr,ci`, then use
  `gh-axi` to push/open a conventional-title sub-PR against
  `refactor/3714-canonical-delivery-flow`. Never use `--yes` or target `main`.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. Executed RED test checkpoint with recorded failure.
3. Green canonical policy/generator/generated files checkpoint.
4. Review/verification fixes checkpoint, if needed.
