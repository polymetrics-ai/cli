import { NextResponse } from 'next/server';
import { NextRequest } from 'next/server';
import { connectorBySlug } from '@/lib/connectors.catalog.generated';
import { docsPageByUrl } from '@/lib/docs.generated';

/** Build synthetic Markdown for a connector slug. */
function connectorMarkdown(slug: string): string | null {
  const c = connectorBySlug(slug);
  if (!c) return null;

  const lines: string[] = [];

  lines.push(`# ${c.name} connector`);
  lines.push('');
  lines.push(c.description);
  lines.push('');
  lines.push(`**Category:** ${c.category}  |  **Release stage:** ${c.releaseStage}`);
  lines.push('');

  // Capabilities
  lines.push('## Capabilities');
  lines.push('');
  const caps = c.capabilities;
  const capList = [
    ['Check', caps.check],
    ['Read', caps.read],
    ['Write', caps.write],
    ['Query', caps.query],
    ['CDC', caps.cdc],
    ['Dynamic schema', caps.dynamicSchema],
  ] as [string, boolean][];
  for (const [name, val] of capList) {
    lines.push(`- **${name}:** ${val ? 'Yes' : 'No'}`);
  }
  lines.push('');

  if (c.streams.length > 0) {
    lines.push('## ETL Streams');
    lines.push('');
    lines.push('| Stream | Primary key | Cursor | Incremental |');
    lines.push('|---|---|---|---|');
    for (const stream of c.streams) {
      lines.push(
        `| \`${stream.name}\` | ${stream.primaryKey.join(', ') || '-'} | ${stream.cursor || '-'} | ${stream.incremental ? 'Yes' : 'No'} |`,
      );
    }
    lines.push('');
  }

  if (c.writeActions.length > 0) {
    lines.push('## Reverse-ETL Write Actions');
    lines.push('');
    lines.push('| Action | Method | Kind |');
    lines.push('|---|---|---|');
    for (const action of c.writeActions) {
      lines.push(`| \`${action.name}\` | ${action.method || '-'} | ${action.kind || '-'} |`);
    }
    lines.push('');
  }

  if (c.cliSurface) {
    lines.push('## Command Surface');
    lines.push('');
    lines.push(`Usage: \`${c.cliSurface.usage}\``);
    lines.push('');
    lines.push('| Command | Intent | Availability | Mapping |');
    lines.push('|---|---|---|---|');
    for (const command of c.cliSurface.commands) {
      const mapping = command.stream
        ? `stream:${command.stream}`
        : command.write
          ? `write:${command.write}`
          : '-';
      lines.push(`| \`${command.path}\` | ${command.intent || '-'} | ${command.availability || '-'} | ${mapping} |`);
    }
    lines.push('');
  }

  if (c.docUrl) {
    lines.push('## Service API documentation');
    lines.push('');
    lines.push(c.docUrl);
    lines.push('');
  }

  lines.push(c.docsMd);
  lines.push('');

  return lines.join('\n');
}

function safeDocsSlug(parts: string[]): string | null {
  if (
    parts.some((part) =>
      part === '.' ||
      part === '..' ||
      /[\\/]|[\u0000-\u001f\u007f]/.test(part)
    )
  ) {
    return null;
  }

  if (parts.length === 0 || (parts.length === 1 && parts[0] === 'index')) {
    return '/docs';
  }

  return `/docs/${parts.join('/')}`;
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ slug: string[] }> },
) {
  const { slug } = await params;

  // Check if this is a connector slug (path starts with "connectors/")
  if (slug[0] === 'connectors' && slug.length === 2) {
    const connectorSlug = slug[1];
    const md = connectorMarkdown(connectorSlug);
    if (!md) {
      return new NextResponse('Not found', { status: 404 });
    }
    return new NextResponse(md, {
      headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
    });
  }

  // Regular docs page.
  const url = safeDocsSlug(slug);
  if (!url) {
    return new NextResponse('Not found', { status: 404 });
  }

  const page = docsPageByUrl(url);
  if (!page?.body) {
    return new NextResponse('Not found', { status: 404 });
  }

  return new NextResponse(page.body, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
}
