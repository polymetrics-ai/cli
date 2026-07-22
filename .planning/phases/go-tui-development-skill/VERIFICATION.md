# Verification Checklist

## Baseline And Scope

- [x] Isolation paths resolve to the disposable worktree.
- [x] Feature branch created before edits.
- [x] `git fetch origin main --prune` completed and HEAD matched current `origin/main`.
- [x] Analyzed base SHA recorded.
- [x] `AGENTS.md`, skill routing, matrix, schema, conventions, GSD adapter, and required skills read.
- [x] GSD adapter checked; missing `programming-loop` fallback recorded.
- [ ] No `cmd/`, product `internal/` source, CLI help, website, dependency, or runtime behavior change.

## Research Evidence

- [ ] Required library/framework surfaces evaluated from authoritative sources.
- [ ] Representative production Go TUIs evaluated from current code/docs.
- [ ] Stable source URLs and access date recorded.
- [ ] Current Polymetrics code/dependency/output baseline recorded.
- [ ] Relevant open and closed issue/PR inventory captured with read-only `gh-axi`.
- [ ] Actual issues distinguished from speculative future ideas.
- [ ] Each mapped issue includes evidence, gaps, choices, acceptance tests, risks, and sequence.
- [ ] Artifact states what this PR does not implement.

## Skill Contract

- [ ] `SKILL.md` follows repository frontmatter conventions and force-triggers all required terms.
- [ ] Skill starts with repository/issue/code inspection.
- [ ] Library selection matrix exists.
- [ ] UX/accessibility checklist exists.
- [ ] Architecture/concurrency/performance guide exists.
- [ ] Testing/terminal compatibility guide exists.
- [ ] Dated source ledger exists.
- [ ] MUST/SHOULD/MAY and definition of done are explicit.
- [ ] Non-interactive and machine-readable behavior is preserved.
- [ ] Accessibility, restoration, cancellation, deterministic tests, safe output, and UX evidence are
  hard gates.
- [ ] Relevant Go skills are routed rather than duplicated.

## Automated Validation

- [ ] Red test fails for the missing skill/routing contract before implementation.
- [ ] `go test ./internal/agentdocs`
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check`
- [ ] YAML parses and Markdown/local links resolve through the focused contract test.
- [ ] `go.mod` and `go.sum` unchanged.

## Safety And Delivery

- [ ] No secret value read, printed, stored, or summarized.
- [ ] No credentialed connector, reverse ETL, generic write tool, or runtime lifecycle action.
- [ ] GitHub research remained read-only through `gh-axi`.
- [ ] Browser research used `chrome-devtools-axi`.
- [ ] No Claude or GitHub Copilot review requested.
- [ ] Branch changes committed before firstmate handoff.
- [ ] no-mistakes is deferred until firstmate instructs the shipping stage.
