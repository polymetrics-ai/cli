# pm connectors inspect amazon-sqs

```text
NAME
  pm connectors inspect amazon-sqs - Amazon SQS connector manual

SYNOPSIS
  pm connectors inspect amazon-sqs
  pm connectors inspect amazon-sqs --json
  pm credentials add <name> --connector amazon-sqs [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Retains evidence-backed visibility for DeleteQueue and PurgeQueue while their configured-queue actions lack a schema-4 typed record binding. No Amazon SQS provider action is executable from this bundle.

ICON
  id: amazon-sqs
  asset: icons/amazon-sqs.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=false catalog=true read=false write=false query=false
  Integration type: queue

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  queue_url

SECURITY
  read risk: No provider read is executable from this bounded disposition bundle.
  write risk: DeleteQueue and PurgeQueue remain visible with evidence-backed unsupported typed-binding gaps; the runtime never sends a request for either command.
  approval: No write can reach planning, approval, or provider I/O from this bundle.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Amazon SQS queue actions with explicit runtime dispositions.
  Usage: pm amazon-sqs <command>
  Queue actions
  Other Commands
    queue delete - DeleteQueue is visible but unsupported without a schema-4 typed configured-queue binding. [intent=docs_only availability=unsupported_with_provider_evidence]; approval: not executable; risk: deletes the configured SQS queue
    queue purge - PurgeQueue is visible but unsupported without a schema-4 typed configured-queue binding. [intent=docs_only availability=unsupported_with_provider_evidence]; approval: not executable; risk: purges all available messages from the configured queue

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amazon-sqs

  # Inspect as structured JSON
  pm connectors inspect amazon-sqs --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-sqs before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
