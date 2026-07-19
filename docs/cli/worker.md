```
NAME
  pm worker - serve and inspect the optional RLM Temporal worker

SYNOPSIS
  pm worker status [--json]
  pm worker serve [--json]

DESCRIPTION
  The worker serves the typed RLM Temporal workflow and its Podman activity on
  the polymetrics-rlm task queue. It is optional and used only by RLM agent
  mode; dependency-free CLI and RLM modes do not require it.

  status reports whether the explicitly configured Temporal endpoint is
  reachable. serve starts the long-lived worker and reports readiness only
  after the Temporal client and worker have started successfully.

CONFIGURATION
  Set POLYMETRICS_TEMPORAL_ADDR or its legacy PM_TEMPORAL_ADDR alias, or set
  runtime.temporal_addr in .polymetrics/config.yaml. Worker execution also uses
  rlm.image and rlm.podman_bin from typed configuration.

SECURITY
  This is not a generic remote command runner. It registers only the typed RLM
  workflow and activity. Output includes status and endpoint metadata, never
  credential values or workflow payloads.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
  3 validation error when no Temporal endpoint is explicitly configured

```
