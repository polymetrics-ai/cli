# REVIEW — issue #3721 Codex project-local workers

Status: pass (inline manual review).

The generated GSD code-review prompt normally spawns a review role. That would conflict with the
canonical single-worker contract and this task's prohibition on additional roles, so the source,
tests, generated output, and verification results were reviewed inline instead.

## Findings

No critical, warning, or informational findings remain.

- The canonical source, rather than a hand-authored TOML file, owns every Codex-specific value.
- Required Codex projections use whole-file drift comparison; optional Markdown adapters retain
  their existing marked-block behavior.
- The TOML encoder and parser round-trip coverage preserve TOML-sensitive developer instructions,
  and projection I/O stays rooted in the selected worktree. The symlink escape test now exercises
  the required `.codex` ancestor.
- The generated files have the exact documented standalone fields and set
  `agents.enabled = false` in role configuration. Tests parse the TOML and exercise the
  corresponding drift regression.
- The trust prerequisite is stated without inventing a trust command, and same-filename
  user/project collision precedence is explicitly left undocumented.

## Review limitation

This review verifies the configuration and generated output statically. It does not claim that a
live Codex session was run under every trust state or that an undocumented user/project standalone
filename collision has a known winner.
