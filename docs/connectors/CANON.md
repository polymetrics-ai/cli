# Connector Delivery Status

For the current implementation procedure, binding source reports, archive
index, and clean-environment limitations, read
[the connector delivery canon](../connector-canon/INDEX.md).

Important current facts:

- The warehouse is always the mediator: API → warehouse → API, API → warehouse
  → database, database → warehouse → API, and database → warehouse → database.
- `availability: implemented` requires real command-runner preflight, not only
  a bundle declaration or generated manual.
- There are zero accepted live certifications. A `certification.json`, fixture,
  or matching filename is not certification.
- Inspect runtime-visible connector metadata with
  `pm connectors inspect <name> --json`; do not infer support from a catalog
  filename.
