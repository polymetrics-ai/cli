# DISCUSSION LOG — issue #3716

Mode: `discuss-phase --auto` with the authoritative issue and parent contract.

No product choices remain open. Issue #3716 fixes the target paths, project-local scope,
canonical-only generation, required Claude format, minimum tools, `Agent` omission, and stacked PR
topology. The captain subsequently fixed the skill boundary: the worker must have skill access,
but bare `Skill` would also expose unrelated skills that can use `context: fork`. The canonical
policy therefore renders scoped `Skill(name)` rules for the repository-required Go/design skills,
omits bare `Skill`, and denies both `Agent` and its `Task` alias. Managed definitions and CLI
`--agents` remain documented higher-precedence exceptions.
