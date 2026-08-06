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

OPTIONS
  --connector name       connector that owns the credential
  --provider-family id   non-secret provider family for an explicit binding
  --auth-profile id      non-secret compatible authentication profile
  --link-credential id   explicitly share a compatible credential binding on add
  --to credential        explicitly share a compatible credential binding
  --from-env field=ENV   read one secret field from an environment variable
  --value-stdin field    read one secret field from standard input
  --config key=value     store non-secret connector config
  --root path            project root containing .polymetrics
  --json                 render machine-readable JSON

SECURITY
  Secret values are encrypted with AES-GCM in .polymetrics/vault and are not
  stored in state.json. Inspection output shows only secret field names.
  Credential bindings are protected project state and are never shown in
  credential output; runtime coordination receives only opaque projections.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
