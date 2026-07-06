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
  docsMd: string;
  docs: ConnectorDocLink[];
  docUrl: string;
  appDocUrl: string;
  icon: ConnectorIconMeta | null;
  featured: boolean;
};
