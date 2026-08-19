# Delivery summary — GitHub certification suite r1

## Delivered

- Added the connector-neutral `connectorgen certification-sweep` generator and
  its Makefile verification gate.
- Generated GitHub's 1,571-command ledger directly from `cli_surface.json`,
  retaining declared command path, summary, intent, availability, stream,
  typed/enum flags, and API surface.
- Accounted for every command exactly once with a concrete non-pass reason:
  20 `eligible_pending_live`, 1,408 `fixture_required`, 92
  `product_defect`, one `provider_refused`, and 50 `not_applicable`.
- Reported 92 declaration/runtime defects, led by `releases assets view` and
  its non-required `--asset-id` for required REST `asset_id`.
- Recorded the measured GitHub HTTP-422 refusal for
  `actions fork-pr-contributor-approval view` as a provider observation, not
  a product defect.
- Proved a declaration-owned produced-value assertion goes red after schema
  compilation, restored it, and passed all 23 existing direct-read candidates
  against the named disposable fixture.

## Boundary

No command is promoted to `pass` in the generated ledger. Accepted evidence
remains deferred because PR #4198 has not supplied required `http_exchanges`.
The generated candidates and live observations are ready to be wired to that
evidence path when it lands.

## Lifecycle

The required GSD verification and code-review prompts were resolved with the
documented inline/manual fallback because this autonomous delivery contract
prohibits role spawning. `VERIFICATION.md`, `UAT.md`, and `REVIEW.md` contain
the resulting evidence and disposition.
