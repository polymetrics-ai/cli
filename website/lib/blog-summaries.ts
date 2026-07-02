export type BlogSummary = {
  slug: string;
  title: string;
  description: string;
  category: string;
  readingTime: string;
  publishedAt: string;
};

export const BLOG_SUMMARIES: BlogSummary[] = [
  {
    slug: 'one-cli-to-rule-them-all',
    title: 'One CLI To Rule Them All',
    description:
      'Why Polymetrics puts ETL, DuckDB SQL, reverse ETL, scheduling, and agent-safe JSON contracts behind one local binary.',
    category: 'Product essay',
    readingTime: '5 min read',
    publishedAt: '2026-07-02',
  },
  {
    slug: 'agent-native-data-workflows',
    title: 'Agent-Native Data Workflows Need Boring Contracts',
    description:
      'How JSON envelopes, stable exit codes, local credentials, and approval-gated writes make CLI automation safer for LLM agents.',
    category: 'Engineering note',
    readingTime: '6 min read',
    publishedAt: '2026-07-02',
  },
  {
    slug: 'local-first-data-engine',
    title: 'Local-First Data Pipelines Without Warehouse Sprawl',
    description:
      'A practical case for keeping extraction, local storage, analytical SQL, and write-back automation close to the developer.',
    category: 'Architecture',
    readingTime: '5 min read',
    publishedAt: '2026-07-02',
  },
];

export function blogUrl(slug: string): string {
  return `/blog/${slug}`;
}
