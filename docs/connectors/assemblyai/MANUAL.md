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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  api_key (secret)

ETL STREAMS
  transcript:
    primary key: id
    cursor: created
    fields: audio_url(), completed(), created(), error(), id(), resource_url(), status()
  transcript_sentences:
    primary key: id
    cursor: created
    fields: created(), id(), resource_url(), status()
  transcript_paragraphs:
    primary key: id
    cursor: created
    fields: created(), id(), resource_url(), status()
  transcript_subtitle:
    primary key: id
    cursor: created
    fields: created(), id(), resource_url(), status()
  transcript_detail:
    primary key: id
    fields: audio_channels(), audio_duration(), audio_url(), auto_chapters(), auto_highlights(), confidence(), entity_detection(), error(), format_text(), id(), language_code(), language_confidence(), punctuate(), redact_pii(), sentiment_analysis(), speaker_labels(), status(), summarization(), summary(), text(), webhook_status_code(), webhook_url()
  transcript_sentence_items:
    primary key: transcript_id, start
    fields: channel(), confidence(), end(), speaker(), start(), text(), transcript_id()
  transcript_paragraph_items:
    primary key: transcript_id, start
    fields: channel(), confidence(), end(), speaker(), start(), text(), transcript_id()
  transcript_word_search_matches:
    primary key: transcript_id, text
    fields: count(), indexes(), text(), timestamps(), transcript_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_transcript:
    endpoint: POST /v2/transcript
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
