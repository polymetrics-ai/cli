---
name: pm-faker
description: Sample Data connector knowledge and safe action guide.
---

# pm-faker

## Purpose

Generates deterministic sample users, purchases, and products without network access.

## Icon

- id: simple-icons-faker
- asset: icons/simple-icons/faker.svg
- title: Faker
- simple_icon_slug: faker
- simple_icon_hex: 779B2E
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Faker
- match: exact-name-or-slug
- matched_by: faker

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- count
- seed

## Security

- read risk: none; in-process synthetic data generation, no network access
- write risk: n/a (read-only source)
- approval: none required; no external data is read or written
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect faker
```

### Inspect as structured JSON

```bash
pm connectors inspect faker --json
```

## Agent Rules

- Run pm connectors inspect faker before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
