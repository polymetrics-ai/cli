import { NextResponse } from 'next/server';
import { CONNECTOR_CATALOG } from '@/lib/connectors.catalog.generated';
import { DOCS_PAGES } from '@/lib/docs.generated';

export const dynamic = 'force-static';

export async function GET() {
  const sections: string[] = [];

  // Concatenate all docs pages
  for (const page of DOCS_PAGES) {
    if (page.body) {
      sections.push(`# ${page.title}\n\n${page.body}`);
    } else {
      sections.push(`# ${page.title}\n\n_Content not available._`);
    }
  }

  // Compact connector index
  sections.push('# Connector catalog index\n');
  const connectorLines = CONNECTOR_CATALOG.map(
    (c) =>
      `- **${c.name}** (\`${c.slug}\`) — ${c.category} ${c.type}, ${c.releaseStage}, status: ${c.status}`,
  );
  sections.push(connectorLines.join('\n'));

  const text = sections.join('\n\n---\n\n');

  return new NextResponse(text, {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
}
