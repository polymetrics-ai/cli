# Discussion log — transport source eligibility club

- Command: `scripts/gsd prompt discuss-phase issue-4171-3976-3862-transport-eligibility-r1 --auto`
- Mode: inline/manual fallback; no roles spawned because the repository's canonical worker contract forbids them.
- Inputs read: issues #4171, #3976, #3862, #4093, #4090; PRs #4156, #4161, #4167; current connector canon and migration/design references; prior GSD phases for #3856/#3857/#3858/#3976/#4090/#4093 and PostgreSQL managed-target polling apply.
- User-locked choices: one PR for exactly three issues, positive allowlists only, production-entry proof, real PostgreSQL and GitHub evidence, exhaustive refusal/effect assertions, and explicit exclusion of #4125/#4158/#4169.
- Defaulted implementation choices: split GitHub's general source adapter from its issue-label destination semantics; retain bounded batch emission; reuse the shared polling source executor with a named native PostgreSQL binding; keep CDC separate; derive dynamic relation/cursor/key identity only after catalog discovery; report unavailable live infrastructure honestly.
- Deferred ideas: generic polling support for other databases, a generic transport-source generator, unrelated credential classification, and the known managed-target live test defect.

