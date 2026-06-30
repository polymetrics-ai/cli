import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import Link from 'next/link';

import {
  CONNECTOR_CATALOG,
  connectorBySlug,
} from '@/lib/connectors.catalog.generated';

import { ConnectorHeader } from '@/components/docs/connector/connector-header';
import { CapabilityMatrix } from '@/components/docs/connector/capability-matrix';
import { ConfigTable } from '@/components/docs/connector/config-table';
import { DocLinks } from '@/components/docs/connector/doc-links';
import { CliExample } from '@/components/docs/connector/cli-example';
import { ConnectorSetup } from '@/components/docs/connector/connector-setup';
import { ConnectorStatusSummary } from '@/components/docs/connector/connector-status-summary';
import { connectorEnrichment } from '@/lib/connectors.enrichment.generated';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { CopyMarkdown } from '@/components/docs/copy-markdown';

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
    c.notes ||
    `${c.name} is a ${c.releaseStage} ${c.type} connector in the ${c.category} category.`;

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

// ── Sync mode label formatter ─────────────────────────────────────────────

function formatSyncMode(mode: string): string {
  const map: Record<string, string> = {
    full_refresh: 'Full refresh',
    incremental: 'Incremental',
    append: 'Append',
    overwrite: 'Overwrite',
    deduped_history: 'Deduped history',
  };
  return map[mode] ?? mode.replace(/_/g, ' ');
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
            description="Runtime availability and catalog metadata for this connector."
          >
            Status
          </SectionHeading>
          <ConnectorStatusSummary connector={c} />
          <div className="mt-3 grid min-w-0 gap-3 sm:grid-cols-2">
            <Card muted className="px-3 py-2.5">
              <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                Runtime status
              </p>
              <p className="mt-1 text-[13px] leading-relaxed text-text-secondary">
                {c.status === 'enabled' ? (
                  <span className="font-medium text-line-cta">Enabled; ETL is available.</span>
                ) : (
                  <span>
                    Catalog metadata is available; ETL is not enabled until the native Go port
                    passes conformance.
                  </span>
                )}
              </p>
            </Card>
            <Card muted className="px-3 py-2.5">
              <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                pm connector name
              </p>
              <code className="mt-1 block break-words font-mono text-[13px] text-text-primary">
                {c.pmName || c.slug}
              </code>
            </Card>
          </div>
          {c.notes && (
            <p className="mt-3 border border-line-structure bg-surface-1 px-3 py-2.5 text-[13px] leading-relaxed text-text-tertiary">
              {c.notes}
            </p>
          )}
        </section>

        {/* ③ Configuration */}
        <section className="min-w-0" aria-labelledby="section-configuration">
          <SectionHeading
            id="section-configuration"
            description="Field names, required flags, and secret boundaries from the connector catalog."
          >
            Configuration
          </SectionHeading>
          <ConfigTable config={c.config} secrets={c.secrets} />
          {c.secrets.length > 0 && (
            <p className="mt-3 text-[12px] font-mono leading-relaxed text-text-tertiary">
              Secret field names:{' '}
              {c.secrets.map((s, i) => (
                <span key={s}>
                  <code className="bg-surface-2 border border-line-structure px-1">{s}</code>
                  {i < c.secrets.length - 1 ? ', ' : ''}
                </span>
              ))}
            </p>
          )}
        </section>

        {/* ④ Setup & authentication — shown whenever verified enrichment exists */}
        {connectorEnrichment(c.slug) && (
          <section className="min-w-0" aria-labelledby="section-setup">
            <SectionHeading
              id="section-setup"
              description="Provider-side prerequisites, authentication shape, and verified source links."
            >
              Setup &amp; authentication
            </SectionHeading>
            <ConnectorSetup connector={c} />
          </section>
        )}

        {/* ⑤ CLI usage */}
        <section className="min-w-0" aria-labelledby="section-cli-usage">
          <SectionHeading
            id="section-cli-usage"
            description="Copy-ready commands generated from this connector's current runtime capabilities."
          >
            CLI usage
          </SectionHeading>
          <CliExample connector={c} />
        </section>

        {/* ⑥ Capabilities */}
        <section className="min-w-0" aria-labelledby="section-capabilities">
          <SectionHeading
            id="section-capabilities"
            description="Runtime operations currently exposed by pm for this connector."
          >
            Capabilities
          </SectionHeading>
          <CapabilityMatrix capabilities={c.capabilities} />
        </section>

        {/* ⑦ Sync modes */}
        <section className="min-w-0" aria-labelledby="section-sync-modes">
          <SectionHeading id="section-sync-modes">Sync modes</SectionHeading>
          <div className="flex flex-wrap gap-2">
            {c.syncModes.length > 0 ? (
              c.syncModes.map((mode) => (
                <Badge key={mode} variant="category">
                  {formatSyncMode(mode)}
                </Badge>
              ))
            ) : (
              <span className="text-[13px] text-text-tertiary italic">No sync modes listed.</span>
            )}
          </div>
          <p className="mt-3 text-[13px] text-text-secondary">
            <span className="font-semibold">Incremental syncs: </span>
            {c.incremental ? (
              <span className="text-line-cta">Supported</span>
            ) : (
              <span className="text-text-tertiary">Not supported</span>
            )}
          </p>
        </section>

        {/* ⑧ Official documentation */}
        <section className="min-w-0" aria-labelledby="section-documentation">
          <SectionHeading
            id="section-documentation"
            description="Vendor documentation links carried by the catalog."
          >
            Official documentation
          </SectionHeading>
          <DocLinks docs={c.docs} appDocUrl={c.appDocUrl} />
        </section>
      </div>
    </article>
  );
}
