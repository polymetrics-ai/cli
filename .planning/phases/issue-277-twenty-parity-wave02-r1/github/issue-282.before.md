Part of #277 (S5). Write scope: `writes.json` (destructive rows). Deps: #281 (serializes shared writes.json).

- 28 `destructive_admin` delete actions (`DELETE /rest/{objects}/{id}`), `risk: destructive`, typed-confirmation required; blocked by default outside plan/approval/execute.

Acceptance: destructive rows gated; `covered_by` mapping complete.
Refs #277