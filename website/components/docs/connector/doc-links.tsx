import * as React from 'react';
import { ExternalLink } from 'lucide-react';
import type { ConnectorDocLink } from '@/lib/connectors.types';
import { Card } from '@/components/ui/card';

interface DocLinksProps {
  docs: ConnectorDocLink[];
  docUrl?: string;
  appDocUrl?: string;
}

// ── Doc type → short human label ─────────────────────────────────────────

const DOC_TYPE_LABELS: Record<string, string> = {
  api_reference: 'API',
  authentication_guide: 'Auth',
  rate_limits: 'Rate limits',
  status_page: 'Status',
  overview: 'Overview',
  changelog: 'Changelog',
  quickstart: 'Quickstart',
  sdk: 'SDK',
};

function docTypeLabel(type: string): string {
  return DOC_TYPE_LABELS[type] ?? type.replace(/_/g, ' ');
}

// ── Individual link card ──────────────────────────────────────────────────

interface LinkCardProps {
  title: string;
  label: string;
  url: string;
}

function LinkCard({ title, label, url }: LinkCardProps) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      className="link-box group block min-w-0 transition-colors"
    >
      <span aria-hidden className="corner-box-hover-child" />
      <Card muted className="flex h-full min-w-0 flex-col gap-2 p-3 group-hover:bg-surface-2">
        <span className="flex items-center justify-between gap-2">
          <span className="text-[11px] font-mono uppercase tracking-wider text-text-tertiary">
            {label}
          </span>
          <ExternalLink className="h-3.5 w-3.5 shrink-0 text-text-disabled group-hover:text-line-cta" aria-hidden="true" />
        </span>
        <span className="text-[13px] font-medium leading-snug text-text-primary transition-colors duration-150 group-hover:text-line-cta">
          {title}
        </span>
        <span className="min-w-0 truncate font-mono text-[11px] text-text-disabled">
          {url}
        </span>
      </Card>
    </a>
  );
}

// ── Main component ────────────────────────────────────────────────────────

export function DocLinks({ docs, appDocUrl }: DocLinksProps) {
  const hasLinks = docs.length > 0 || !!appDocUrl;

  if (!hasLinks) {
    return (
      <p className="text-[13px] text-text-tertiary italic">
        No documentation links available.
      </p>
    );
  }

  return (
    <div className="grid gap-3 sm:grid-cols-2">
      {/* Official application docs from catalog (vendor) */}
      {docs.map((doc) => (
        <LinkCard
          key={doc.url}
          title={doc.title}
          label={docTypeLabel(doc.type)}
          url={doc.url}
        />
      ))}
      {/* App docs (if distinct from the official docs) */}
      {appDocUrl && !docs.find((d) => d.url === appDocUrl) && (
        <LinkCard title="Application documentation" label="App docs" url={appDocUrl} />
      )}
    </div>
  );
}
