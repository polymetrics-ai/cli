const trim = (value) => (typeof value === 'string' ? value.trim() : '');

const keyNames = (keyStyle) => {
  if (keyStyle === 'camel') {
    return {
      mapsTo: 'mapsTo',
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
    sourceCli: 'source_cli',
    sourceCliPath: 'source_cli_path',
    sourceUrl: 'source_url',
    outputPolicy: 'output_policy',
    globalFlags: 'global_flags',
    helpTopics: 'help_topics',
  };
};

export function mapFlags(flags, options = {}) {
  const keys = keyNames(options.keyStyle);
  return (Array.isArray(flags) ? flags : [])
    .map((flag) => ({
      name: trim(flag.name),
      type: trim(flag.type),
      summary: trim(flag.summary),
      values: Array.isArray(flag.values) ? flag.values.map((value) => trim(value)).filter(Boolean) : [],
      [keys.mapsTo]: trim(flag.maps_to),
    }))
    .filter((flag) => flag.name);
}

export function mapCLISurface(surface, options = {}) {
  const keys = keyNames(options.keyStyle);
  if (!surface || typeof surface !== 'object') return null;

  const commands = (Array.isArray(surface.commands) ? surface.commands : [])
    .map((command) => ({
      path: trim(command.path),
      summary: trim(command.summary),
      intent: trim(command.intent),
      availability: trim(command.availability),
      stream: trim(command.stream),
      write: trim(command.write),
      [keys.sourceCliPath]: trim(command.source_cli_path),
      [keys.sourceUrl]: trim(command.source_url),
      flags: mapFlags(command.flags, options),
      examples: Array.isArray(command.examples) ? command.examples.map((example) => trim(example)).filter(Boolean) : [],
      [keys.outputPolicy]: trim(command.output_policy),
      risk: trim(command.risk),
      approval: trim(command.approval),
      notes: trim(command.notes),
    }))
    .filter((command) => command.path);

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
    [keys.globalFlags]: mapFlags(surface.global_flags, options),
    commands,
    [keys.helpTopics]: (Array.isArray(surface.help_topics) ? surface.help_topics : [])
      .map((topic) => ({
        name: trim(topic.name),
        summary: trim(topic.summary),
      }))
      .filter((topic) => topic.name),
  };
}
