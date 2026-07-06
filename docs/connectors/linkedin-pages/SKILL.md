---
name: pm-linkedin-pages
description: LinkedIn Pages connector knowledge and safe action guide.
---

# pm-linkedin-pages

## Purpose

Reads LinkedIn organization (company page) profile, follower statistics, share statistics, and total follower count through the LinkedIn Community Management REST API.

## Icon

- asset: icons/linkedin.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/community-management/organizations

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- linkedin_version
- mode
- org_id
- access_token (secret)

## ETL Streams

- follower_statistics:
  - primary key: org_id
  - fields: followerCountsByAssociationType(), followerCountsByCountry(), followerCountsByFunction(), followerCountsByIndustry(), followerCountsByRegion(), followerCountsBySeniority(), followerCountsByStaffCountRange(), followerGains(), org_id(), organizationalEntity()
- share_statistics:
  - primary key: org_id
  - fields: org_id(), organizationalEntity(), shareStatisticsByPost(), totalShareStatistics()
- organizations:
  - primary key: id
  - fields: id(), industries(), localized_name(), localized_website(), locations(), name(), org_id(), organization_type(), primary_organization_type(), staff_count_range(), urn(), vanity_name(), version_tag()
- total_follower_count:
  - primary key: org_id
  - fields: first_degree_size(), org_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external LinkedIn Community Management API read of company page profile and lifetime statistics
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect linkedin-pages
```

### Inspect as structured JSON

```bash
pm connectors inspect linkedin-pages --json
```

## Agent Rules

- Run pm connectors inspect linkedin-pages before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
