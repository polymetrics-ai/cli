```
NAME
  pm credentials - manage encrypted connector credentials

SYNOPSIS
  pm credentials add <name> --connector <connector> [--provider-family family] [--auth-profile profile] [--link-credential credential] [--from-env field=ENV] [--value-stdin field] [--config key=value]
  pm credentials link <name> --to <credential> [--json]
  pm credentials list [--json]
  pm credentials inspect <name> [--json]
  pm credentials test <name> [--json]
  pm credentials remove <name>

DESCRIPTION
  Credentials combine non-secret connector config with encrypted secret fields.
  Secrets should be supplied through environment variables or stdin, not shell
  arguments. Use --from-env field=ENV for non-interactive setup. Use
  --value-stdin field for multiline secrets such as GitHub App PEM keys.

  Provider family defaults to the connector name and auth profile to default
  when omitted. Existing credentials receive the same defaults when their
  project is opened. Each unlinked credential receives an isolated protected
  binding. Links require matching effective declarations. For a cross-connector
  link, every credential in the resulting cohort must have both declarations
  supplied explicitly; matching defaults alone are not enough.

OPTIONS
  --connector name       connector that owns the credential
  --provider-family id   non-secret provider family declaration
  --auth-profile id      non-secret authentication compatibility declaration
  --link-credential id   join a compatible credential's binding on add
  --to credential        join a compatible credential's binding
  --from-env field=ENV   read one secret field from an environment variable
  --value-stdin field    read one secret field from standard input
  --config key=value     store non-secret connector config
  --root path            project root containing .polymetrics
  --json                 render machine-readable JSON

SECURITY
  Secret values are encrypted with AES-GCM in .polymetrics/vault and are not
  stored in state.json. Inspection output shows only secret field names.
  Provider family and auth profile are non-secret credential metadata. Credential
  bindings are protected project state and are never shown in credential output;
  internal coordination receives only opaque projections. Linking records
  identity metadata only: it does not change connector authentication, rate
  limits, or transport behavior.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
