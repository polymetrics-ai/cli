# pm intercom

## NAME

`pm intercom` - run Intercom connector commands from the declarative command surface.

## SYNOPSIS

```bash
pm intercom
pm help intercom
pm intercom <resource> <action> [flags]
pm intercom <resource> <action> --help
```

## DESCRIPTION

The Intercom connector covers all 149 official Intercom REST API 2.14 operations with one of:

- ETL stream commands for bounded collection/list/search reads;
- fixed direct-read commands with `json_response`, `text_response`, or `binary_metadata` output policies;
- typed reverse ETL write commands; or
- metadata-only binary/text policies that avoid arbitrary file writes.

## EXAMPLES

```bash
pm intercom
pm intercom contact list --credential intercom-local --limit 25 --json
pm intercom contact view --credential intercom-local --contact-id <id> --json
pm intercom contact create --credential intercom-local --email ada@example.test --preview --json
pm intercom contact create --plan <plan-id> --preview --json
pm intercom contact create --plan <plan-id> --approve <approval-token> --json
```

## SAFETY

Do not pass Intercom access tokens as flags or prompt text. Store `access_token` with `pm credentials add ... --from-env access_token=<env-var-name>` or stdin.

Reverse ETL commands create a plan before any live mutation. Execution requires preview, approval token, and execute; destructive/admin write actions also require the typed confirmation challenge declared by the connector.

## PARITY NOTES

`api_surface.json` is the operation ledger. `cli_surface.json` maps provider-style commands to streams, direct reads, bounded text/binary policies, and write actions. The CLI does not expose raw generic HTTP, SQL, shell, or arbitrary GraphQL write surfaces.
