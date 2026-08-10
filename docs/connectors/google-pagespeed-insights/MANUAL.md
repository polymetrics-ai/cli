# pm connectors inspect google-pagespeed-insights

```text
NAME
  pm connectors inspect google-pagespeed-insights - Google PageSpeed Insights connector manual

SYNOPSIS
  pm connectors inspect google-pagespeed-insights
  pm connectors inspect google-pagespeed-insights --json
  pm credentials add <name> --connector google-pagespeed-insights [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA scores) for the configured URLs and strategies via the PageSpeed Insights v5 API.

ICON
  id: google-pagespeed-insights
  asset: icons/google-pagespeed-insights.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/speed/docs/insights/v5/get-started

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

SECURITY
  read risk: connector-specific
  write risk: connector-specific
  approval: external mutations require preview and approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Google PageSpeed Insights's declared streams and reverse-ETL actions.
  Usage: pm google-pagespeed-insights <command> [flags]
  Read streams
  Other Commands
    api get pagespeedonline v5 runpagespeed - Documented GET /pagespeedonline/v5/runPagespeed (not implemented) [intent=direct_read availability=not_implemented operation=google-pagespeed-insights.get.pagespeedonline-v5-runpagespeed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    pagespeed reports list - Run the pagespeed reports ETL stream [intent=etl availability=implemented stream=pagespeed_reports]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-pagespeed-insights

  # Inspect as structured JSON
  pm connectors inspect google-pagespeed-insights --json

AGENT WORKFLOW
  - Run pm connectors inspect google-pagespeed-insights before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
