import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const trim = (value) => (typeof value === 'string' ? value.trim() : '');

// The destination flags a binary_download command accepts are owned by the
// runtime, not by the bundle, so they appear in no cli_surface.json. They are
// read from the same file internal/connectors embeds, so the website cannot
// document a different flag surface than `pm <connector> <command> --help`.
const BINARY_DOWNLOAD_FLAGS = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../../internal/connectors/binary_download_flags.json',
    ),
    'utf8',
  ),
);

const DIRECT_READ_PAGE_FLAGS = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../../internal/connectors/direct_read_page_flags.json',
    ),
    'utf8',
  ),
);

const REVERSE_ETL_APPROVAL_FLAGS = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../../internal/connectors/reverse_etl_approval_flags.json',
    ),
    'utf8',
  ),
);

const keyNames = (keyStyle) => {
  if (keyStyle === 'camel') {
    return {
      mapsTo: 'mapsTo',
      allowEmpty: 'allowEmpty',
      leftFallback: 'leftFallback',
      rightFallback: 'rightFallback',
      valueType: 'valueType',
      sourceCli: 'sourceCli',
      sourceCliPath: 'sourceCliPath',
      sourceUrl: 'sourceUrl',
      outputPolicy: 'outputPolicy',
      globalFlags: 'globalFlags',
      helpTopics: 'helpTopics',
    };
  }
  return {
    mapsTo: 'maps_to',
    allowEmpty: 'allow_empty',
    leftFallback: 'left_fallback',
    rightFallback: 'right_fallback',
    valueType: 'value_type',
    sourceCli: 'source_cli',
    sourceCliPath: 'source_cli_path',
    sourceUrl: 'source_url',
    outputPolicy: 'output_policy',
    globalFlags: 'global_flags',
    helpTopics: 'help_topics',
  };
};

// withBinaryDownloadFlags appends the runtime's destination flags to a
// binary_download command, skipping any the command already lists.
//
// It must be idempotent: gen-connector-catalog.mjs re-maps the output of
// gen-connector-bundles.mjs, so an unconditional append would document
// --dest-root twice on the catalog page.
function withBinaryDownloadFlags(flags, intent) {
  const declared = Array.isArray(flags) ? flags : [];
  const runtimeFlags =
    intent === 'binary_download' || intent === 'text_export'
      ? BINARY_DOWNLOAD_FLAGS
      : intent === 'direct_read'
        ? DIRECT_READ_PAGE_FLAGS
        : null;
  if (!runtimeFlags) return declared;
  const names = new Set(declared.map((flag) => trim(flag?.name)));
  return [...declared, ...runtimeFlags.filter((flag) => !names.has(flag.name))];
}

function withReverseETLApprovalFlags(flags, commands) {
  const declared = Array.isArray(flags) ? flags : [];
  if (!commands.some((command) => command.intent === 'reverse_etl' || command.intent === 'direct_write')) {
    return declared;
  }
  const names = new Set(declared.map((flag) => trim(flag?.name)));
  return [...declared, ...REVERSE_ETL_APPROVAL_FLAGS.filter((flag) => !names.has(flag.name))];
}

export function mapFlags(flags, options = {}) {
  const keys = keyNames(options.keyStyle);
  return (Array.isArray(flags) ? flags : [])
    .map((flag) => {
      const out = {
        name: trim(flag.name),
        type: trim(flag.type),
        summary: trim(flag.summary),
        values: Array.isArray(flag.values) ? flag.values.map((value) => trim(value)).filter(Boolean) : [],
        [keys.mapsTo]: trim(flag.maps_to),
      };
      if (trim(flag.format)) out.format = trim(flag.format);
      if (typeof flag.allow_empty === 'boolean') out[keys.allowEmpty] = flag.allow_empty;
      if (typeof flag.minimum === 'number' && Number.isFinite(flag.minimum)) out.minimum = flag.minimum;
      if (typeof flag.required === 'boolean') out.required = flag.required;
      if (typeof flag.repeatable === 'boolean') out.repeatable = flag.repeatable;
      return out;
    })
    .filter((flag) => flag.name);
}

export function mapCLISurface(surface, options = {}) {
  const keys = keyNames(options.keyStyle);
  if (!surface || typeof surface !== 'object') return null;

  const commands = (Array.isArray(surface.commands) ? surface.commands : [])
    .map((command) => {
      const operation = trim(command.operation);
      const out = {
        path: trim(command.path),
        summary: trim(command.summary),
        intent: trim(command.intent),
        availability: trim(command.availability),
        stream: trim(command.stream),
        write: trim(command.write),
      };
      if (operation) out.operation = operation;
      Object.assign(out, {
        [keys.sourceCliPath]: trim(command.source_cli_path),
        [keys.sourceUrl]: trim(command.source_url),
        flags: mapFlags(withBinaryDownloadFlags(command.flags, out.intent), options),
        examples: Array.isArray(command.examples) ? command.examples.map((example) => trim(example)).filter(Boolean) : [],
        [keys.outputPolicy]: trim(command.output_policy),
        risk: trim(command.risk),
        approval: trim(command.approval),
        notes: trim(command.notes),
      });
      return out;
    })
    .filter((command) => command.path);

  for (const command of commands) {
    const constraints = (Array.isArray(surface.commands) ? surface.commands : [])
      .find((candidate) => trim(candidate.path) === command.path)?.constraints;
    const mappedConstraints = (Array.isArray(constraints) ? constraints : [])
      .map((constraint) => ({
        kind: trim(constraint.kind),
        left: trim(constraint.left),
        right: trim(constraint.right),
        op: trim(constraint.op),
        [keys.valueType]: trim(constraint.value_type),
        [keys.leftFallback]: trim(constraint.left_fallback),
        [keys.rightFallback]: trim(constraint.right_fallback),
        message: trim(constraint.message),
      }))
      .filter((constraint) => constraint.kind);
    if (mappedConstraints.length > 0) command.constraints = mappedConstraints;
  }

  if (!trim(surface.usage) && commands.length === 0) return null;

  return {
    tagline: trim(surface.tagline),
    usage: trim(surface.usage),
    [keys.sourceCli]: surface.source_cli
      ? {
          name: trim(surface.source_cli.name),
          docs: trim(surface.source_cli.docs),
          reference: trim(surface.source_cli.reference),
          source: trim(surface.source_cli.source),
        }
      : null,
    groups: (Array.isArray(surface.groups) ? surface.groups : [])
      .map((group) => ({
        id: trim(group.id),
        title: trim(group.title),
        commands: Array.isArray(group.commands) ? group.commands.map((command) => trim(command)).filter(Boolean) : [],
      }))
      .filter((group) => group.id || group.title || group.commands.length > 0),
    [keys.globalFlags]: mapFlags(withReverseETLApprovalFlags(surface.global_flags, commands), options),
    commands,
    [keys.helpTopics]: (Array.isArray(surface.help_topics) ? surface.help_topics : [])
      .map((topic) => ({
        name: trim(topic.name),
        summary: trim(topic.summary),
      }))
      .filter((topic) => topic.name),
  };
}
