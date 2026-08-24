# Review — issue-4336 request-contract execution envelopes

Inline code-review fallback: the repository contract forbids spawning a reviewer
role in this lane. Review scope was the complete `origin/main...HEAD` diff, with
special attention to schema truth, finite-resource enforcement, request authority,
numeric compatibility, generated-surface drift, and provider-neutral behavior.

## Findings and dispositions

| Severity | Finding | Disposition | Evidence |
| --- | --- | --- | --- |
| warning | Body-envelope validation checked only positive shape, so an altered but plausible effective limit could project. | fixed | Canonical full-envelope comparison and red/green altered-limit test; commit `8a5bcfcdd`. |
| warning | `json.Valid` narrowed the previous CLI numeric grammar and would reject already accepted spellings such as `+01` and hexadecimal floats. | fixed | Preserve the prior `strconv` grammar with exact `big.Int`/`big.Rat` arithmetic; four red/green compatibility cases; commit `8e368b5ac`. |
| warning | A numeric header with finite provider range still lacks a finite textual byte bound and could evade the uncensused-header quarantine. | fixed | Such headers now produce the shared merge-blocking gap; booleans derive a five-byte bound; commit `8e368b5ac`. |
| info | Runtime inspection reports the effective bound as PM policy even when an immutable source envelope records a tighter provider-derived component. | accepted for this slice | The plan deliberately avoids a fleet-wide JSON rewrite; source descriptors retain `provider_and_pm_policy`, while help/inspection truthfully describe the enforced effective value as PM policy. |

No critical or unresolved warning finding remains. The final full generator and
engine suites, lint, vet, builds, generated parity, connector boundary, and
release workflow checks passed after the fixes. External Claude/Copilot coverage
will be recorded after the direct PR is opened; it is not claimed here.

## Post-merge CI disposition

Final-head Verify run `32679838571` found tracked skill drift after main's
#4341 integration. Canonical regeneration changed only the Twenty skill by
adding the same PM execution-policy notice already emitted for command-bearing
connectors. The generator was byte-stable on a second pass and the exact failing
test passed. This is a generated integration closure, not a hand-authored
connector exception. Automated Claude/Copilot review remained unavailable due
fleet defect `fm-gh-axi-projectcards-deprecation-r1`; Firstmate explicitly
confirmed it is not a required check, so no automated review pass is claimed.

The later #4346 main integration overlapped `sourceprojection.go` but merged
without conflict. Full generator tests plus source-derived surface and skill
checks passed on the combined tree; no hand resolution or projection rewrite
was necessary.

Verify subsequently found the corresponding tracked Twenty connector manual
and connector-local skill stale. The whole Go tree passed; individual later
gates isolated `docs-check`. Canonical docs regeneration changed only that
manual/skill pair, was byte-stable, and passed exact generated-surface tests.
