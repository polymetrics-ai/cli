# DISCUSSION LOG — issue #3716

Mode: `discuss-phase --auto` with the authoritative issue and parent contract.

No product choices remain open. Issue #3716 fixes the target paths, project-local scope,
canonical-only generation, required Claude format, minimum tools, `Agent` omission, and stacked PR
topology. The captain subsequently fixed the skill boundary: personal same-name skills must not
shadow an allowed source, and runtime skill access must be removed if trusted qualification cannot
hold. Official Claude documentation confirms plugin-qualified identifiers cannot collide with
personal or project skills and recommends `skills` frontmatter for subagent preloads. The canonical
policy therefore preloads only qualified `cc-skills-golang:*` and
`frontend-design:frontend-design` identifiers, omits and denies `Skill`, and denies both `Agent`
and its `Task` alias. The three unqualified website/design skills are unavailable with an explicit
handoff cost. Managed definitions and CLI `--agents` remain documented higher-precedence
exceptions.
