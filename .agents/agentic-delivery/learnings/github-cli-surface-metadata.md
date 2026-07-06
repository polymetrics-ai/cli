# GitHub CLI Surface Metadata Learnings

Date: 2026-07-06

Scope: issue #34, first slice under the GitHub CLI parity parent issue #44.

## What worked

- Treating the issue as the prompt kept the implementation scoped to a single green slice:
  optional metadata loading, validation, GitHub metadata, docs, website notes, and learning capture.
- A docs-first `cli_surface.json` lets the connector describe gh-inspired commands without enabling
  command dispatch before the executor and approval model exist.
- The validator should fail closed on references. A command mapped to a stream, write action, or
  API endpoint must point at an existing bundle object; otherwise the metadata is stale.
- Test-first evidence was useful. The first red tests caught the missing `Bundle.CLISurface` model,
  missing validator rules, unsupported schema keywords, and incomplete secret-looking token
  detection before the GitHub bundle was added.
- The official `gh` surface must be checked command by command. For example, current `gh ruleset`
  documents `check`, `list`, and `view`; repository ruleset create/update/delete are connector-native
  write actions, not current gh command paths.

## Reuse For Future Connectors

- Start each provider with a current provider CLI/API inventory from primary sources, then separate
  provider CLI parity from connector-native commands.
- Use `availability` for executable readiness and `intent` for execution model. Do not infer safety
  from a command name alone.
- Keep sensitive or local commands explicit:
  `unsupported_local` for browser, shell, git, completion, extension, and local config behavior;
  `unsafe_or_disallowed` for generic raw API writes, secret output, or mutation models without an
  approval gate.
- Prefer metadata and docs changes before runtime dispatch. Runtime aliases should be thin wrappers
  around validated metadata, not hand-coded command trees for one provider.
- Do not hand-edit generated connector manuals or website data when a generator owns them. Update
  the source bundle or website content, run the generator, and review the resulting diff.

## Agent Guardrails

- Use the GSD programming loop before production edits. If the local GSD scripts are unavailable,
  record the manual fallback in `.planning/phases/<phase>/` and still keep PRD, plan, TDD ledger,
  sources, verification, and run state current.
- Red-green-refactor is mandatory for code slices. Capture at least one failing test before the
  implementation when the task changes Go behavior.
- Commit coherent green slices regularly when repository push policy permits. In this repository,
  pushes are coordinator-gated unless explicitly delegated.
- CodeRabbit review is a gate after implementation. Each review item needs a disposition comment:
  fixing now, deferring with reason, or declining with reason.
- Never include real tokens, private keys, authorization headers, encrypted secret payloads, or
  secret-looking fixtures in docs, tests, examples, or agent notes.

## Follow-Up For Later Slices

- Add a renderer for `cli_surface.json` so connector help can show command groups, examples, risk,
  approval, and known gaps offline.
- Add endpoint coverage metrics from `api_surface.json` to generated docs.
- Add GraphQL operation metadata before modeling Projects v2, discussions, search, status, and
  other gh surfaces that are not REST-only.
- Design constrained raw API support as a separate reviewed phase with host allowlists, method risk
  labels, mutation approval, and secret redaction.
