'use client';

import { OnPageTocAside } from '@/components/ui/on-page-toc';

const sections = [
  { id: 'overview', label: 'Overview' },
  { id: 'tools', label: 'All the tools' },
  { id: 'loop', label: 'The loop' },
  { id: 'public-source', label: 'Public source' },
  { id: 'connectors', label: 'Connectors' },
  { id: 'why', label: 'Why pm?' },
  { id: 'get-started', label: 'Get started' },
];

export function HomeAside() {
  return <OnPageTocAside className="home-aside-panel" items={sections} />;
}
