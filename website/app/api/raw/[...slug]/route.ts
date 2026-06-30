import { NextResponse } from 'next/server';
import { NextRequest } from 'next/server';
import { connectorBySlug } from '@/lib/connectors.catalog.generated';
import { connectorEnrichment } from '@/lib/connectors.enrichment.generated';
import { docsPageByUrl } from '@/lib/docs.generated';

/** Build synthetic Markdown for a connector slug. */
function connectorMarkdown(slug: string): string | null {
  const c = connectorBySlug(slug);
  if (!c) return null;

  const lines: string[] = [];

  lines.push(`# ${c.name} connector`);
  lines.push('');
  lines.push(
    `**Type:** ${c.type}  |  **Category:** ${c.category}  |  **Release stage:** ${c.releaseStage}  |  **Status:** ${c.status}`,
  );
  lines.push('');

  if (c.notes) {
    lines.push(`> ${c.notes}`);
    lines.push('');
  }

  // Capabilities
  lines.push('## Capabilities');
  lines.push('');
  const caps = c.capabilities;
  const capList = [
    ['Metadata', caps.metadata],
    ['Check', caps.check],
    ['Catalog', caps.catalog],
    ['Read', caps.read],
    ['Write', caps.write],
    ['Query', caps.query],
    ['ETL', caps.etl],
    ['Reverse ETL', caps.reverseEtl],
  ] as [string, boolean][];
  for (const [name, val] of capList) {
    lines.push(`- **${name}:** ${val ? 'Yes' : 'No'}`);
  }
  lines.push('');

  // Config table
  if (c.config.length > 0) {
    lines.push('## Configuration');
    lines.push('');
    lines.push('| Field | Type | Required | Secret | Description |');
    lines.push('|---|---|---|---|---|');
    for (const f of c.config) {
      lines.push(
        `| \`${f.name}\` | ${f.type} | ${f.required ? 'Yes' : 'No'} | ${f.secret ? 'Yes' : 'No'} | ${f.description.replace(/\|/g, '\\|')} |`,
      );
    }
    lines.push('');
  }

  // Sync modes
  if (c.syncModes.length > 0) {
    lines.push('## Sync modes');
    lines.push('');
    for (const m of c.syncModes) {
      lines.push(`- ${m.replace(/_/g, ' ')}`);
    }
    lines.push('');
    lines.push(`**Incremental syncs:** ${c.incremental ? 'Supported' : 'Not supported'}`);
    lines.push('');
  }

  // CLI examples
  lines.push('## CLI usage');
  lines.push('');
  if (c.status === 'enabled' && c.pmName) {
    lines.push('```bash');
    lines.push(`# Inspect connector metadata`);
    lines.push(`pm connectors inspect ${c.pmName} --json`);
    lines.push('');
    lines.push(`# Run ETL through a configured connection`);
    lines.push(`pm etl run --connection <connection-name> --stream <stream-name> --json`);
    lines.push('```');
  } else {
    lines.push(
      `_ETL is not yet enabled for this connector. Use \`pm connectors list\` to browse enabled connectors._`,
    );
  }
  lines.push('');

  // Official docs
  if (c.docs.length > 0 || c.appDocUrl) {
    lines.push('## Official documentation');
    lines.push('');
    if (c.appDocUrl) lines.push(`- [App documentation](${c.appDocUrl})`);
    for (const d of c.docs) {
      lines.push(`- [${d.title}](${d.url})`);
    }
    lines.push('');
  }

  // Setup & auth from enrichment
  const enrich = connectorEnrichment(slug);
  if (enrich) {
    lines.push('## Setup & authentication');
    lines.push('');

    if (enrich.prerequisites.length > 0) {
      lines.push('### Prerequisites');
      lines.push('');
      for (const p of enrich.prerequisites) {
        lines.push(`- ${p}`);
      }
      lines.push('');
    }

    if (enrich.authMethods.length > 0) {
      lines.push('### Authentication methods');
      lines.push('');
      for (const a of enrich.authMethods) {
        lines.push(`**${a.name}:** ${a.summary}`);
        lines.push('');
      }
    }

    if (enrich.setupSteps.length > 0) {
      lines.push('### Setup steps');
      lines.push('');
      for (let i = 0; i < enrich.setupSteps.length; i++) {
        const s = enrich.setupSteps[i];
        lines.push(`${i + 1}. **${s.title}** — ${s.body}`);
        lines.push('');
      }
    }

    if (enrich.sources.length > 0) {
      lines.push('### Sources');
      lines.push('');
      for (const s of enrich.sources) {
        lines.push(`- [${s.title}](${s.url})`);
      }
      lines.push('');
    }
  }

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
