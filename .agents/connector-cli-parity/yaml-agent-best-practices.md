# YAML agent best practices

Accessed: 2026-07-06

## Source-backed findings

- GitHub Actions uses YAML for workflows and separates trigger, permissions, jobs, steps, and
  concurrency. Apply the same separation to agents: trigger conditions, permissions, workflow steps,
  and hard stops should be distinct fields.
- GitHub issue forms use YAML with typed form elements and validations. Agent specs should likewise
  use typed fields and explicit required inputs instead of free-form prompt blobs.
- The OpenAI Agents SDK describes agents as instructions plus tools plus optional handoffs,
  guardrails, and structured outputs. The neutral YAML shape should have those same concepts without
  requiring one SDK.
- OpenAI guardrails documentation distinguishes input, output, and tool guardrails, and notes that
  blocking guardrails prevent execution before side effects. Agent YAML should classify guardrails by
  phase and mark hard stops as blocking.
- MCP defines a vendor-neutral way to expose context and tools. Agent YAML should reference resources
  and tools by capability and policy, not by raw credentials or unrestricted shell/API access.

## Repository rules

- Keep YAML declarative. Do not embed long prompts when a referenced Markdown contract exists.
- Use stable `id`, `version`, `role`, `objective`, `inputs`, `outputs`, `workflow`, `guardrails`,
  `permissions`, `skills`, and `verification` fields.
- Quote glob patterns and values that start with YAML-special characters.
- Prefer lists of small steps over paragraph instructions.
- Declare allowed paths and denied paths.
- Declare human gates explicitly.
- Keep model/provider fields optional adapters, not core contract fields.
- Validate YAML syntax in CI before trusting an agent spec.
- Keep runtime-specific adapters in `adapters`, so the same core spec can be translated to Codex,
  Claude, OpenCode, GitHub Actions, or a custom runner.

## Official sources

- GitHub Actions workflow syntax: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax
- GitHub issue form schema: https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
- OpenAI Agents SDK agents: https://openai.github.io/openai-agents-python/agents/
- OpenAI Agents SDK guardrails: https://openai.github.io/openai-agents-python/guardrails/
- Model Context Protocol specification: https://modelcontextprotocol.io/specification/2025-06-18
