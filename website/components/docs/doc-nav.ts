import {
  Bot,
  Cable,
  Cpu,
  Database,
  FileJson,
  GitBranch,
  ListTree,
  PackageCheck,
  Rocket,
  Route,
  type LucideIcon,
} from 'lucide-react';

export type DocumentationNavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
  description: string;
};

export const DOCUMENTATION_NAV: DocumentationNavItem[] = [
  {
    label: 'Introduction',
    href: '/docs',
    icon: Route,
    description: 'Map the local extract-query-act loop.',
  },
  {
    label: 'Quickstart',
    href: '/docs/quickstart',
    icon: Rocket,
    description: 'Run the first local workflow.',
  },
  {
    label: 'Installation',
    href: '/docs/installation',
    icon: PackageCheck,
    description: 'Install or build the pm binary.',
  },
  {
    label: 'Connectors',
    href: '/docs/connectors',
    icon: Cable,
    description: 'Browse capabilities, streams, and actions.',
  },
  {
    label: 'ETL',
    href: '/docs/etl',
    icon: GitBranch,
    description: 'Extract connector data into the warehouse.',
  },
  {
    label: 'Query',
    href: '/docs/query',
    icon: Database,
    description: 'Run SQL over local data.',
  },
  {
    label: 'Reverse ETL',
    href: '/docs/reverse-etl',
    icon: FileJson,
    description: 'Plan, preview, approve, and write back.',
  },
  {
    label: 'CLI Reference',
    href: '/docs/cli-reference',
    icon: ListTree,
    description: 'Every command and subcommand.',
  },
  {
    label: 'Architecture',
    href: '/docs/architecture',
    icon: Cpu,
    description: 'Runtime, flows, schedules, and RLM.',
  },
  {
    label: 'Agent Guide',
    href: '/docs/agent-guide',
    icon: Bot,
    description: 'Use JSON output and safe agent gates.',
  },
];

const DOC_META_BY_HREF = new Map(DOCUMENTATION_NAV.map((item) => [item.href, item]));
const DOC_META_BY_LABEL = new Map(
  DOCUMENTATION_NAV.map((item) => [item.label.toLowerCase(), item]),
);

export function documentationMetaFor(url: string | undefined, name: string) {
  const fromUrl = url ? DOC_META_BY_HREF.get(url) : undefined;
  return fromUrl ?? DOC_META_BY_LABEL.get(name.toLowerCase());
}
