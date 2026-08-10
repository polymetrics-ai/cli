# pm connectors inspect assemblyai

```text
NAME
  pm connectors inspect assemblyai - AssemblyAI connector manual

SYNOPSIS
  pm connectors inspect assemblyai
  pm connectors inspect assemblyai --json
  pm credentials add <name> --connector assemblyai [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AssemblyAI transcripts, per-transcript detail, sentences, paragraphs, and word-search matches, and submits new transcription jobs, through the AssemblyAI REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  word_search_terms
  api_key (secret) (required)

ETL STREAMS
  transcript:
    primary key: id
    cursor: created
    fields: audio_url(string), completed(string), created(string), error(string), id(string), resource_url(string), status(string)
  transcript_sentences:
    primary key: id
    cursor: created
    fields: created(string), id(string), resource_url(string), status(string)
  transcript_paragraphs:
    primary key: id
    cursor: created
    fields: created(string), id(string), resource_url(string), status(string)
  transcript_subtitle:
    primary key: id
    cursor: created
    fields: created(string), id(string), resource_url(string), status(string)
  transcript_detail:
    primary key: id
    fields: audio_channels(integer), audio_duration(number), audio_url(string), auto_chapters(boolean), auto_highlights(boolean), confidence(number), entity_detection(boolean), error(string), format_text(boolean), id(string), language_code(string), language_confidence(number), punctuate(boolean), redact_pii(boolean), sentiment_analysis(boolean), speaker_labels(boolean), status(string), summarization(boolean), summary(string), text(string), webhook_status_code(integer), webhook_url(string)
  transcript_sentence_items:
    primary key: transcript_id, start
    fields: channel(string), confidence(number), end(integer), speaker(string), start(integer), text(string), transcript_id(string)
  transcript_paragraph_items:
    primary key: transcript_id, start
    fields: channel(string), confidence(number), end(integer), speaker(string), start(integer), text(string), transcript_id(string)
  transcript_word_search_matches:
    primary key: transcript_id, text
    fields: count(integer), indexes(array), text(string), timestamps(array), transcript_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_transcript:
    endpoint: POST /v2/transcript
    required fields: audio_url
    risk: external mutation; submits a new transcription job against a caller-supplied audio_url and consumes AssemblyAI account balance/quota; approval required

SECURITY
  read risk: external AssemblyAI API read of transcript metadata, per-transcript detail/sentences/paragraphs/word-search matches, fanned out over every listed transcript
  write risk: submits a new transcription job (POST /v2/transcript) against a caller-supplied audio_url; consumes account balance/quota
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect assemblyai

  # Inspect as structured JSON
  pm connectors inspect assemblyai --json

AGENT WORKFLOW
  - Run pm connectors inspect assemblyai before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
