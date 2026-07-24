# CLI Architecture v2 release split state

**State date:** 2026-07-24  
**Decision:** build a default-branch release candidate for the Cobra + typed Viper foundation before
TUI and OpenTelemetry  
**Release status:** candidate work only; not merged, tagged, or released  
**Default-branch base at reconstruction start:**
`873cd7b251f70c4a35a607a0d4e86051ea0fbd15`  
**Preserved Architecture v2 parent:** `feat/cli-architecture-v2` at
`0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20` when the split began; draft PR #438 remains
human-gated

## Why the split is dependency-closed

Typed Viper configuration is not an independent patch. Its minimum safe executable boundary is:

1. #399 / PR #439 — golden transcript and generated-doc safety;
2. #400 / PR #440 — Cobra strangler router with legacy flag ownership;
3. #401 / PR #441 — typed invocation-scoped Viper configuration;
4. #402 / PR #448 — config consumer and environment migration.

The candidate also includes #453 / PR #454 so the release smoke gate requires reverse ETL plan ->
preview -> approval -> execute. That safety test is not a Viper dependency, but it is a release gate
for this candidate.

The source squashes are patch provenance, not a historical branch to merge unchanged:

```text
379cb5015335ff7c9b20e5bb780952ead22c53b2
8900db141cc289b65491365d2ebcab490af57789
7683087d41646c92b2bd7f677f47cf2bc9d88462
cc2a90e918b2814a64516d6bad6d14462b3ac079
20475ddf8ae3486282ead4fc7d2129f2bd1129b3
```

They are reconstructed on latest `main` so current Gong dynamic-connector behavior remains present.

## Included behavior

- a fresh Cobra command tree per `cli.Run` invocation;
- legacy handler and dynamic connector passthrough through the strangler shell;
- existing error/output/manual ownership;
- typed non-secret project, warehouse, runtime, RLM, and schedule configuration;
- explicit config precedence:
  changed flag > `POLYMETRICS_*` > matching `PM_*` > effective-root config file > default;
- a fresh `viper.New()` per invocation, with explicit env bindings and no `AutomaticEnv` or watcher;
- config consumer migration that does not silently enable runtime/RLM workers;
- golden CLI compatibility coverage and reverse-smoke preview ordering.

## Explicitly excluded

This candidate does not contain:

```text
internal/events/**
internal/logging/**
internal/telemetry/**
internal/ui/**
docs/adr/0003-interactive-tui-layer.md
docs/adr/0004-opentelemetry-observability.md
docs/design/tui-ux-design.md
```

It also excludes Bubble Tea/Charm dependencies, `golang.org/x/term`, OpenTelemetry dependency
promotion, TUI/OTel phase material, native namespace migration phases, PR #493-owned delivery-skill
and routing changes, and PM review-system implementation.

## Compatibility boundary

Unchanged commands must preserve exit code and stdout/stderr bytes, including exact JSON envelopes,
bare namespace help, invalid-action errors, dynamic connector help/dispatch, in-process re-entrancy,
non-TTY execution, credential handling, and isolated workers.

The intentional configuration changes are:

- `.polymetrics/config.yaml` becomes active;
- malformed config fails through the existing validation category and exit code 3;
- a primary `POLYMETRICS_*` setting wins over its `PM_*` alias;
- `pm help config` documents the typed non-secret schema and precedence.

No migration command rewrites a user's config file. Secret values remain outside typed config.

## Independent successor streams

After this foundation reaches `main`, TUI and OpenTelemetry can proceed as sibling stacks:

- **TUI:** build from the released foundation and a pre-logging event bus; it must not import
  telemetry. Shared CLI/docs/golden/module files still require serialized integration.
- **OpenTelemetry:** build from the released foundation plus logging, with telemetry lifecycle at a
  neutral invocation seam; it must not import UI or events. Event/logging coupling in the published
  parent must be removed during reconstruction.

Existing parent and feature branches remain source/provenance records. This split does not authorize
rebasing, resetting, force-pushing, deleting, or rewriting them.

## Squash and parent consequences

A squash merge of the foundation PR gives `main` the candidate tree but does not make the historical
parent commits ancestors of `main`. PR #495's one-time squash authorization and review/Snyk deferrals
do not apply here.

After a human-authorized default-branch merge, parent reconciliation needs a separate additive
human decision. The non-rewriting choices are:

1. create a successor parent from post-release `main` and reconstruct only unshipped streams; or
2. explicitly authorize an ancestry-preserving merge of post-release `main` into a reconciliation
   branch for the existing parent, resolve conflicts, and run full review again.

This candidate performs neither action and does not authorize merge of PR #438.

## Delivery gates

Before any prerelease or default-branch merge:

- focused and full Go tests, race checks where practical, vet, build, formatting, module integrity,
  docs/golden freshness, and `make verify` pass;
- current-main/candidate compatibility comparisons pass;
- the diff and module graph prove TUI/OTel isolation;
- fresh Dependency Review, CodeQL/security checks, and Snyk pass;
- the canonical PM exact-version Codex packets and synthesis are clean, followed by independent
  Shepherd review;
- no-mistakes opens a green PR targeting `main`.

Release version remains a separate decision while release PR #16 (`1.0.0`) is open. No tag or
GitHub prerelease exists or is authorized by this state record.
