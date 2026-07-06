export type ConnectorCapabilities = {
  check: boolean;
  read: boolean;
  write: boolean;
  query: boolean;
  cdc: boolean;
  dynamicSchema: boolean;
};

export type ConnectorConfigField = {
  name: string;
  type: string;
  description: string;
  required: boolean;
  secret: boolean;
};

export type ConnectorStream = {
  name: string;
  primaryKey: string[];
  cursor: string;
  incremental: boolean;
};

export type ConnectorWriteAction = {
  name: string;
  method: string;
  kind: string;
};

export type ConnectorDocLink = {
  title: string;
  type: string;
  url: string;
};

export type ConnectorCliFlag = {
  name: string;
  type: string;
  summary: string;
  values: string[];
  mapsTo: string;
};

export type ConnectorCliSource = {
  name: string;
  docs: string;
  reference: string;
  source: string;
};

export type ConnectorCliGroup = {
  id: string;
  title: string;
  commands: string[];
};

export type ConnectorCliCommand = {
  path: string;
  summary: string;
  intent: string;
  availability: string;
  stream: string;
  write: string;
  sourceCliPath: string;
  sourceUrl: string;
  flags: ConnectorCliFlag[];
  examples: string[];
  risk: string;
  approval: string;
  notes: string;
};

export type ConnectorCliHelpTopic = {
  name: string;
  summary: string;
};

export type ConnectorCliSurface = {
  tagline: string;
  usage: string;
  sourceCli: ConnectorCliSource | null;
  groups: ConnectorCliGroup[];
  globalFlags: ConnectorCliFlag[];
  commands: ConnectorCliCommand[];
  helpTopics: ConnectorCliHelpTopic[];
};

export type ConnectorIconMeta = {
  id: string;
  path: string;
  publicPath: string;
  source: string;
  reviewStatus: string;
  reviewUrl: string;
};

export type ConnectorMeta = {
  slug: string;
  name: string;
  description: string;
  category: string;
  categoryLabel: string;
  releaseStage: string;
  status: 'available';
  capabilities: ConnectorCapabilities;
  capabilityLabels: string[];
  streams: ConnectorStream[];
  writeActions: ConnectorWriteAction[];
  cliSurface: ConnectorCliSurface | null;
  docsMd: string;
  docs: ConnectorDocLink[];
  docUrl: string;
  appDocUrl: string;
  icon: ConnectorIconMeta | null;
  featured: boolean;
};
