```
NAME
  pm context - manage PM Broker Organization, Workspace, and Environment context

SYNOPSIS
  pm context create <name> --organization <org-id> --organization-name <name> --workspace <workspace-id> --workspace-name <name> --environment <environment-id> --environment-name <name> --environment-type development|staging|production|ephemeral --broker-profile <profile-id> --broker-profile-name <name> [--runtime-mode remote|local|hybrid] [--hybrid-policy policy-id] [--json]
  pm context use <name> [--json]
  pm context show [--context <name>] [--json]
  pm context list [--json]

DESCRIPTION
  Contexts are named safe metadata tuples for PM Broker Organization,
  Workspace, Environment, and BrokerProfile references. The CLI stores this
  user-level metadata in the operating system's standard Polymetrics user
  config directory, not in the project vault. Context state contains immutable
  broker IDs and display names only; it never stores tokens, provider
  credentials, secret locators, or raw secret values.

  Live PM Broker provider operations are not enabled in this foundation slice.
  The metadata shown by these commands is cached from explicit context entries
  and is a narrow seam for the future fake-client/contract-fixture package.

RUNTIME MODES
  remote
    Use the broker as the authority. This is the default for production.

  local
    Use local-only evaluation for non-production or legacy-local reads. Local
    fallback is forbidden for production writes and scheduled production jobs.

  hybrid
    Policy-bound mode that requires --hybrid-policy or broker.hybrid_policy.
    Hybrid cannot provide local fallback for production writes or scheduled
    production jobs.

CONTEXT RESOLUTION
  pm first honors future approval-bound flow/sync requirements. Explicit
  --context, project broker.required_context, active user context, and
  broker.default_context must match the approval-bound identity or resolution
  stops safely. Without an approval-bound requirement, explicit --context wins
  before project broker.required_context, active user context,
  broker.default_context, and finally a synthesized legacy-local context for
  unmigrated projects.

SECURITY
  Context commands do not read credentials, legacy vault entries, provider
  secrets, service-account JSON keys, or live broker resources. They cannot
  grant Organization membership or change internal partition identity.

EXIT STATUS
  0 success
  2 usage error
  3 validation error

```
