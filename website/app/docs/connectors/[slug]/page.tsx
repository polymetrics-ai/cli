import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import Link from 'next/link';

import {
  CONNECTOR_CATALOG,
  connectorBySlug,
} from '@/lib/connectors.catalog.generated';

import { ConnectorHeader } from '@/components/docs/connector/connector-header';
import { CapabilityMatrix } from '@/components/docs/connector/capability-matrix';
import { CliExample } from '@/components/docs/connector/cli-example';
import { ConnectorStatusSummary } from '@/components/docs/connector/connector-status-summary';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/table';
import { CopyMarkdown } from '@/components/docs/copy-markdown';
import { BundleMarkdown } from '@/components/docs/connector/bundle-markdown';

// ── Static params — all connector slugs ───────────────────────────────────

export function generateStaticParams() {
  return CONNECTOR_CATALOG.map((c) => ({ slug: c.slug }));
}

// ── Metadata ──────────────────────────────────────────────────────────────

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const c = connectorBySlug(slug);
  if (!c) return { title: 'Connector not found' };

  const description =
    c.description ||
    `${c.name} exposes ${c.streams.length} ETL streams and ${c.writeActions.length} write actions.`;

  return {
    title: `${c.name} connector`,
    description,
  };
}

// ── Section heading helper ────────────────────────────────────────────────

function SectionHeading({
  id,
  children,
  description,
}: {
  id: string;
  children: React.ReactNode;
  description?: string;
}) {
  return (
    <div className="mb-3 mt-8 scroll-mt-24">
      <h2
        id={id}
        className="font-square text-[16px] font-semibold leading-[1.3] text-text-primary"
      >
        {children}
      </h2>
      {description ? (
        <p className="mt-1 max-w-full text-[13px] leading-relaxed text-text-tertiary sm:max-w-[70ch]">
          {description}
        </p>
      ) : null}
    </div>
  );
}

// ── Connector data tables ─────────────────────────────────────────────────

function EmptyTableNote({ children }: { children: React.ReactNode }) {
  return (
    <p className="border border-line-structure bg-surface-1 px-3 py-2.5 text-[13px] italic text-text-tertiary">
      {children}
    </p>
  );
}

function StreamsTable({ connector }: { connector: NonNullable<ReturnType<typeof connectorBySlug>> }) {
  if (connector.streams.length === 0) {
    return (
      <EmptyTableNote>
        This bundle does not declare static streams. Dynamic schemas may be discovered at runtime.
      </EmptyTableNote>
    );
  }

  return (
    <Table>
      <THead>
        <TR>
          <TH>Stream</TH>
          <TH>Primary key</TH>
          <TH>Cursor</TH>
          <TH>Incremental</TH>
        </TR>
      </THead>
      <TBody>
        {connector.streams.map((stream) => (
          <TR key={stream.name}>
            <TD>
              <span className="break-words font-mono text-[12px] text-text-primary">
                {stream.name}
              </span>
            </TD>
            <TD className="font-mono text-[12px] text-text-tertiary">
              {stream.primaryKey.length > 0 ? stream.primaryKey.join(', ') : '—'}
            </TD>
            <TD className="font-mono text-[12px] text-text-tertiary">
              {stream.cursor || '—'}
            </TD>
            <TD>
              <Badge variant={stream.incremental ? 'status-enabled' : 'category'}>
                {stream.incremental ? 'Yes' : 'No'}
              </Badge>
            </TD>
          </TR>
        ))}
      </TBody>
    </Table>
  );
}

function WriteActionsTable({ connector }: { connector: NonNullable<ReturnType<typeof connectorBySlug>> }) {
  if (connector.writeActions.length === 0) return null;

  return (
    <Table>
      <THead>
        <TR>
          <TH>Action</TH>
          <TH>Method</TH>
          <TH>Kind</TH>
        </TR>
      </THead>
      <TBody>
        {connector.writeActions.map((action) => (
          <TR key={action.name}>
            <TD>
              <span className="break-words font-mono text-[12px] text-text-primary">
                {action.name}
              </span>
            </TD>
            <TD className="font-mono text-[12px] text-text-tertiary">
              {action.method || '—'}
            </TD>
            <TD>
              <Badge variant={action.kind === 'delete' ? 'release-beta' : 'category'}>
                {action.kind || 'action'}
              </Badge>
            </TD>
          </TR>
        ))}
      </TBody>
    </Table>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────

export default async function ConnectorPage({ params }: Props) {
  const { slug } = await params;
  const c = connectorBySlug(slug);
  if (!c) notFound();

  return (
    <article className="mx-auto w-full max-w-[980px] px-4 py-8 sm:px-6">
      {/* Header row: breadcrumb left, actions right */}
      <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <nav aria-label="Breadcrumb" className="flex items-center gap-1.5">
          <Link
            href="/docs"
            className="font-mono text-[11px] uppercase tracking-widest text-text-tertiary hover:text-text-secondary transition-colors"
          >
            Docs
          </Link>
          <span className="text-[10px] text-text-tertiary select-none">/</span>
          <Link
            href="/docs/connectors"
            className="font-mono text-[11px] uppercase tracking-widest text-text-tertiary hover:text-text-secondary transition-colors"
          >
            Connectors
          </Link>
          <span className="text-[10px] text-text-tertiary select-none">/</span>
          <span className="font-mono text-[11px] uppercase tracking-widest text-text-secondary">
            {c.name}
          </span>
        </nav>
        <div className="flex shrink-0 items-center gap-2">
          <Link
            href={`/docs/connectors/${slug}/data.json`}
            className="font-mono text-[11px] uppercase tracking-widest text-text-tertiary hover:text-text-secondary border border-line-structure px-2.5 py-1 transition-colors hover:bg-surface-1"
            target="_blank"
            rel="noopener noreferrer"
          >
            data.json
          </Link>
          <CopyMarkdown slug={['connectors', slug]} />
        </div>
      </div>

      <div className="grid min-w-0 gap-2">
        {/* ① Header — name + badge row */}
        <ConnectorHeader connector={c} />

        {/* ② Status */}
        <section className="min-w-0" aria-labelledby="section-status">
          <SectionHeading
            id="section-status"
            description="Bundle metadata generated from the connector definition files."
          >
            Status
          </SectionHeading>
          <ConnectorStatusSummary connector={c} />
          <div className="mt-3 grid min-w-0 gap-3 sm:grid-cols-2">
            <Card muted className="px-3 py-2.5">
              <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                Description
              </p>
              <p className="mt-1 text-[13px] leading-relaxed text-text-secondary">
                {c.description}
              </p>
            </Card>
            <Card muted className="px-3 py-2.5">
              <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                Connector name
              </p>
              <code className="mt-1 block break-words font-mono text-[13px] text-text-primary">
                {c.slug}
              </code>
            </Card>
          </div>
        </section>

        {/* ③ CLI usage */}
        <section className="min-w-0" aria-labelledby="section-cli-usage">
          <SectionHeading
            id="section-cli-usage"
            description="Copy-ready commands generated from this connector's current runtime capabilities."
          >
            CLI usage
          </SectionHeading>
          <CliExample connector={c} />
        </section>

        {/* ④ Capabilities */}
        <section className="min-w-0" aria-labelledby="section-capabilities">
          <SectionHeading
            id="section-capabilities"
            description="Operations declared by the connector bundle."
          >
            Capabilities
          </SectionHeading>
          <CapabilityMatrix capabilities={c.capabilities} />
        </section>

        {/* ⑤ ETL streams */}
        <section className="min-w-0" aria-labelledby="section-etl-streams">
          <SectionHeading
            id="section-etl-streams"
            description="Readable streams declared by streams.json, with schema-derived primary keys and cursors."
          >
            ETL Streams
          </SectionHeading>
          <StreamsTable connector={c} />
        </section>

        {/* ⑥ Reverse ETL write actions */}
        {c.writeActions.length > 0 && (
          <section className="min-w-0" aria-labelledby="section-write-actions">
            <SectionHeading
              id="section-write-actions"
              description="Reverse-ETL actions declared by writes.json."
            >
              Reverse-ETL Write Actions
            </SectionHeading>
            <WriteActionsTable connector={c} />
          </section>
        )}

        {/* ⑦ Bundle documentation */}
        <section className="min-w-0" aria-labelledby="section-bundle-docs">
          <SectionHeading
            id="section-bundle-docs"
            description="Connector documentation generated from the bundle docs.md file."
          >
            Bundle docs
          </SectionHeading>
          <BundleMarkdown markdown={c.docsMd} />
        </section>
      </div>
    </article>
  );
}
