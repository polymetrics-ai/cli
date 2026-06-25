# Threat Model

## Threats

- Agent treats catalog-only connector as runnable and asks for secrets.
- CLI output leaks secret values from config schemas.
- Future generator changes omit capability metadata.
- Upstream image references are mistaken for executable runtime support.

## Mitigations

- Planned connectors set all data-plane capability booleans to false.
- Unsupported reason is required for non-enabled connectors.
- Manual and skill text says upstream images are metadata only.
- Tests require capability coverage for every connector.
- Existing docs validation continues to require security sections.
