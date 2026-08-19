# pm connectors inspect hugging-face-datasets

```text
NAME
  pm connectors inspect hugging-face-datasets - Hugging Face - Datasets connector manual

SYNOPSIS
  pm connectors inspect hugging-face-datasets
  pm connectors inspect hugging-face-datasets --json
  pm credentials add <name> --connector hugging-face-datasets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads dataset splits and per-split sizes from the Hugging Face dataset-viewer REST API. Read-only; an optional user access token unlocks gated and private datasets.

ICON
  id: simple-icons-huggingface
  asset: icons/simple-icons/huggingface.svg
  title: Hugging Face
  simple_icon_slug: huggingface
  simple_icon_hex: FFD21E
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Hugging%20Face
  match: curated-alias
  matched_by: huggingface

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  dataset_name (required)
  access_token (secret)
  api_token (secret)
  token (secret)

ETL STREAMS
  splits:
    primary key: dataset, config, split
    fields: config(string), dataset(string), split(string)
  sizes:
    primary key: dataset, config, split
    fields: config(string), dataset(string), num_bytes_memory(integer), num_bytes_parquet_files(integer), num_columns(integer), num_rows(integer), split(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Hugging Face dataset-viewer API read of dataset split/size metadata; an optional access token unlocks gated/private dataset reads
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hugging-face-datasets

  # Inspect as structured JSON
  pm connectors inspect hugging-face-datasets --json

AGENT WORKFLOW
  - Run pm connectors inspect hugging-face-datasets before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
