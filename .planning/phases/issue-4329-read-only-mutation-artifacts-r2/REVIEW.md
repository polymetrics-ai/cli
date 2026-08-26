# Code Review — issue #4329, r2

## Inline review

- Reviewed the final local diff for scope, source provenance, action precedence,
  failure behavior, and generated-surface drift.
- Independent audit of `42fd5d48f78f03bfa713dfd30de8059a2d663e99` reported:
  - **Medium, accepted and fixed:** a copied automatic artifact could waive
    coverage in a `write:true` bundle. Validation and projection now require
    explicit `metadata.capabilities.write=false`; a recorded red-to-green test
    proves it.
  - **Acceptance gap, strengthened:** test inputs had constructed citations.
    Tests now load byte-identical preserved Sentry/Vercel full source locks and
    exercise their full mutation inventories. The built-binary credential
    boundary remains a named downstream source-bound-read foundation gap; no
    user command is claimed here.
- No other local findings. The helper is fail-closed without a retained provider
  citation, does not alter a manual disposition, and does not downgrade an
  executable or implemented action claim.
- Current-main integration: merged `origin/main` at
  `1324c52bab0b224ed8958858af7676b8b8e191b4` with no conflicts. The added
  source-lock regression checks actual Sentry and Vercel delete endpoints and
  their declared `reverse_etl` command/action relationship, rather than a
  hand-constructed citation. It confirms the artifact rule never serves as a
  policy gate for a complete executable delete.

## Required independent review

After the current-main integration commit is pushed, obtain the Captain-requested
fresh independent Codex audit of that exact SHA. Record its result in the PR
review coverage before requesting merge.

## Automated review route

- Primary: `claude_auto` on the non-draft main-targeted PR.
- Fallback: only if Claude fails, is skipped by trust, or exhausts quota, follow
  the recorded routing loop; do not issue repeated manual requests.
