# Worker and review prompts

## Implementation

Implement only the acceptance checks in `PLAN.md`, test first. Use existing #474-#478 ports where
they directly fit. Prefer a small injected port over completing a speculative production adapter.
Do not broaden into recovery/quorum/security edge matrices once the vertical trajectory is green.

## Blocker-only review

Review the exact MVP head once. Report only defects that prevent the accepted end-to-end trajectory,
command loading, safe stop/resume, durable human wait, or the prohibition on parent-to-main merge.
Record other observations as deferred backlog; do not request another hardening cycle.
