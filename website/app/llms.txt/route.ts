import { NextResponse } from 'next/server';
import { BLOG_POSTS, blogUrl } from '@/lib/blog';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.catalog.generated';
import { DOCS_PAGES } from '@/lib/docs.generated';

export const dynamic = 'force-static';

export function GET() {
  const docLines = DOCS_PAGES
    .map((p) => {
      const desc = p.description ? `: ${p.description}` : '';
      return `- [${p.title}](${p.url})${desc}`;
    })
    .join('\n');

  const blogLines = BLOG_POSTS
    .map((post) => `- [${post.title}](${blogUrl(post.slug)}): ${post.description}`)
    .join('\n');

  const text = `\
# pm — local-first data engine for ETL, SQL analytics, and reverse-ETL

> pm is a single Go binary for connector-backed ETL, embedded DuckDB queries, and reverse-ETL. The website exposes ${CONNECTOR_CATALOG_COUNT} bundle-generated connector pages without a vendor cloud, Docker, or Kubernetes. Every command is accessible to both human operators and AI agents via identical CLI interfaces.

## Documentation

${docLines}

## Blog

${blogLines}

## Connectors

- [Connector catalog](/docs/connectors): Browse all ${CONNECTOR_CATALOG_COUNT} connectors by capability, ETL stream, write action, and integration type. Each connector page exposes a machine-readable \`data.json\` endpoint at \`/docs/connectors/<slug>/data.json\`.

## Full content

For the complete raw Markdown of every documentation page concatenated into a single document, see [/llms-full.txt](/llms-full.txt).
`;

  return new NextResponse(text, {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
}
