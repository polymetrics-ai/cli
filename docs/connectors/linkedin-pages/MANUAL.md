# pm connectors inspect linkedin-pages

```text
NAME
  pm connectors inspect linkedin-pages - LinkedIn Pages connector manual

SYNOPSIS
  pm connectors inspect linkedin-pages
  pm connectors inspect linkedin-pages --json
  pm credentials add <name> --connector linkedin-pages [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads LinkedIn organization (company page) profile, follower statistics, share statistics, and total follower count through the LinkedIn Community Management REST API.

ICON
  id: linkedin
  asset: icons/linkedin.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/community-management/organizations

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  linkedin_version
  mode
  org_id (required)
  access_token (secret) (required)

ETL STREAMS
  follower_statistics:
    primary key: org_id
    fields: followerCountsByAssociationType(array), followerCountsByCountry(array), followerCountsByFunction(array), followerCountsByIndustry(array), followerCountsByRegion(array), followerCountsBySeniority(array), followerCountsByStaffCountRange(array), followerGains(object), org_id(string), organizationalEntity(string)
  share_statistics:
    primary key: org_id
    fields: org_id(string), organizationalEntity(string), shareStatisticsByPost(array), totalShareStatistics(object)
  organizations:
    primary key: id
    fields: id(integer), industries(array), localized_name(string), localized_website(string), locations(array), name(object), org_id(string), organization_type(string), primary_organization_type(string), staff_count_range(string), urn(string), vanity_name(string), version_tag(string)
  total_follower_count:
    primary key: org_id
    fields: first_degree_size(integer), org_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external LinkedIn Community Management API read of company page profile and lifetime statistics
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect linkedin-pages

  # Inspect as structured JSON
  pm connectors inspect linkedin-pages --json

AGENT WORKFLOW
  - Run pm connectors inspect linkedin-pages before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
