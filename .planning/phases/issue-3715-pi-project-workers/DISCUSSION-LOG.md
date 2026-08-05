# DISCUSSION LOG — issue #3715 Pi clean project-only workers

## Resolved implementation choices

| Area | Decision | Evidence |
| --- | --- | --- |
| Discovery isolation | Add `clean-project`, derived from the canonical contract's two Pi roles; do not rely on precedence. | The current fork unconditionally adds extension roles before scope filtering. |
| Legacy project prompts | Retain them unchanged, but filter them out of clean-project discovery. | User scope fence assigns deletion to Wave 6. |
| Bundled prompts | Keep the files inert but remove all loader paths and UI/prompt claims that make them discoverable. | Issue requires remove/disable; isolation requires they are absent from discovery. |
| Delegation block | Encode no `subagent` in generated worker tools and retain extension-level stripping plus depth check. | Pi README documents `--tools` as an allowlist across built-in/extension/custom tools; official example registers `subagent` as an extension tool. |
| Trust | Treat `clean-project` as a project scope for confirmation and noninteractive safeguards. | Existing project scopes are trust-gated; this is a narrowing, not an override. |

## No open product decision

The task fixes the target roster, topology, and safety boundary. No captain decision is required
for the planned implementation.
