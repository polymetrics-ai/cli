# THREAT MODEL: Agentic ETL Platform

## Threats

- Agent passes terminal escape sequences to spoof output.
- Agent passes path traversal values to read/write outside the project.
- Agent requests broad/generic mutation tools.
- API responses contain prompt injection or terminal control text.
- Credentials leak through JSON output, logs, docs, skills, or tests.
- Large ETL run exhausts memory by buffering all records.

## Mitigations

- Shared sanitizer for terminal output.
- Shared validators for identifiers, URLs, and paths.
- No generic shell, HTTP write, or SQL write tools.
- Preview/approval boundary for reverse ETL.
- Secret redaction tests.
- Bounded ETL batches and checkpoint metadata.
