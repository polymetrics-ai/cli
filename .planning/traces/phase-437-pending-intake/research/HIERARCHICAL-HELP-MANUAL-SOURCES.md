# Hierarchical Help and Manual Research Sources

**Research date:** 2026-07-20

**Policy:** Prefer primary project documentation, package documentation, and source code. Links are
recorded here so the Phase 19 / issue #417 worker can revalidate behavior against the pinned
dependency before implementation.

## Cobra: command tree, help, and document generation

1. [Cobra user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
   - Every command automatically receives help support.
   - `app help <path>` resolves nested commands.
   - Cobra permits a custom help command, function, or template.
   - Command groups organize large parent help output.

2. [Cobra working with commands](https://cobra.dev/docs/how-to-guides/working-with-commands/)
   - Commands are attached to parents to form the authoritative tree.
   - Modular command constructors are recommended for large applications.
   - Grouped immediate children improve parent discoverability.

3. [Cobra enterprise guide](https://cobra.dev/docs/explanations/enterprise-guide/)
   - Command descriptions, flag documentation, and examples should live with the command they
     document.
   - Cobra help integrates with man-page generation.

4. [`github.com/spf13/cobra/doc` v1.10.2 package documentation](https://pkg.go.dev/github.com/spf13/cobra@v1.10.2/doc)
   - `GenManTreeFromOpts` and `GenMarkdownTreeCustom` walk a command and all descendants.
   - The package explicitly warns about flattened filename ambiguity when command names contain
     hyphens.

5. [Cobra v1.10.2 man generator source](https://github.com/spf13/cobra/blob/v1.10.2/doc/man_docs.go)
   - Man generation renders the command path, local flags, inherited flags, examples, and
     parent/child `SEE ALSO` references.
   - `SOURCE_DATE_EPOCH` can control the generated header date.

6. [Cobra v1.10.2 Markdown generator source](https://github.com/spf13/cobra/blob/v1.10.2/doc/md_docs.go)
   - Markdown pages are generated recursively and link to immediate parents/children.
   - Output is primarily derived from `Short`, `Long`, `UseLine`, `Example`, and flags.

## Mature CLI behavior and information architecture

7. [Git `git-help` manual](https://git-scm.com/docs/git-help)
   - Root help and `--all` are discovery surfaces.
   - `git help <command>` opens the relevant individual command or guide manual.
   - `git <command> --help` and command-specific help converge on the command manual.

8. [GitHub CLI `gh help`](https://cli.github.com/manual/gh_help)
   - `gh help [path to command]` resolves full details for a specific nested command path.

9. [Docker CLI root reference](https://docs.docker.com/reference/cli/docker/)
   - Root documentation describes global behavior and lists management commands.
   - Every command accepts its own `--help`.

10. [Docker `container` parent reference](https://docs.docker.com/reference/cli/docker/container/)
    - A parent namespace has its own description and a table of immediate subcommands.

11. [Docker `container exec` leaf reference](https://docs.docker.com/reference/cli/docker/container/exec/)
    - A leaf page owns its detailed usage, options, constraints, and examples.

12. [Generated kubectl reference](https://kubernetes.io/docs/reference/kubectl/generated/)
    - The website exposes the full command hierarchy with focused references and expandable tree
      presentation.

13. [Kubernetes kubectl reference-generation process](https://kubernetes.io/docs/contribute/generate-ref-docs/kubectl/)
    - Generated command reference is treated as a reproducible build artifact and reviewed through
      generated-file diffs.

14. [Go command reference](https://go.dev/cmd/go/)
    - The root lists commands and conceptual topics, then directs users to `go help <command>` or
      `go help <topic>` for focused documentation.

15. [Command Line Interface Guidelines](https://clig.dev/)
    - Explicit `--help` should be extensive and available on subcommands.
    - Incomplete parent invocations should provide concise help.
    - Related operations benefit from consistent subcommand hierarchy.

## Manual-page conventions

16. [Linux `man-pages(7)` conventions](https://man7.org/linux/man-pages/man7/man-pages.7.html)
    - Section 1 is for user commands.
    - Conventional ordered sections include `NAME`, `SYNOPSIS`, `DESCRIPTION`, `OPTIONS`,
      `EXIT STATUS`, `ENVIRONMENT`, `FILES`, `EXAMPLES`, and `SEE ALSO`.
    - Predictable headings make manuals easier to navigate and process.

17. [Linux `man(1)` manual](https://man7.org/linux/man-pages/man1/man.1.html)
    - Documents conventional manual sections and command-synopsis notation.

## Repository evidence

18. `internal/cli/docs.go`
    - Current flat `map[string]string` source with namespace-level raw manuals.

19. `internal/cli/cobra_router.go`
    - `setManualHelp` binds a Cobra command to a free-form flat topic.

20. `internal/cli/certify_cli.go`
    - `newCertifyCobraCommand` currently binds to `"connectors"`, producing the full parent manual.

21. `internal/cli/cli.go`
    - `runDocs("generate", ...)` loops over the flat topic map and writes one Markdown file per topic.

22. `internal/cli/golden_transcript_test.go`
    - Runtime and generated manual bytes are protected by golden and docs-drift tests.

23. `website/content/docs/cli-reference.mdx` and `website/scripts/gen-docs-data.mjs`
    - Website reference prose is maintained separately, then copied into generated website data.

24. [Issue #417: deepen help tree and generate man pages](https://github.com/polymetrics-ai/cli/issues/417)
    - Owns per-subcommand help, `pm docs man`, generated manuals/goldens, CLI docs, website parity,
      and help tests.

25. `docs/plans/cli-architecture-v2-improvement-plan.md`
    - Defines Phase 19 as the deliberate help-churn phase.

## Revalidation checklist for the implementation worker

- Confirm the pinned Cobra version in `go.mod` before relying on generator APIs.
- Re-read the exact pinned `man_docs.go` and `md_docs.go` sources.
- Re-run the current command/manual inventory because TUI and completion phases may add commands
  before issue #417 becomes unblocked.
- Re-check issue #417 dependencies and the parent branch head.
- Confirm whether the website information architecture changed before generating new routes.
- Do not copy prose from third-party manuals; use their information architecture as design evidence.
