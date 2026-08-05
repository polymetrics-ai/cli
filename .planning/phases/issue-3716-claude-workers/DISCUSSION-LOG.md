# DISCUSSION LOG — issue #3716

Mode: `discuss-phase --auto` with the authoritative issue and parent contract.

No product choices remain open. Issue #3716 fixes the target paths, project-local scope,
canonical-only generation, required Claude format, minimum tools, `Agent` omission, and stacked PR
topology. Official documentation resolves the implementation detail: `tools` is an allowlist and
an omitted `Agent` prevents the worker from spawning any agent. Managed definitions and CLI
`--agents` are documented higher-precedence exceptions, so the phase does not claim to override
them.
