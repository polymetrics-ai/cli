import * as React from 'react';
import { AlertTriangle, ExternalLink, KeyRound, ListChecks, ShieldCheck } from 'lucide-react';
import type { ConnectorMeta } from '@/lib/connectors.catalog.generated';
import { connectorEnrichment } from '@/lib/connectors.enrichment.generated';
import { Card } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface ConnectorSetupProps {
  connector: ConnectorMeta;
}

function InlineText({ text }: { text: string }) {
  const parts = text.split(/(`[^`]+`)/g);

  return (
    <>
      {parts.map((part, index) => {
        if (part.startsWith('`') && part.endsWith('`')) {
          return (
            <code
              key={`${part}-${index}`}
              className="border border-line-structure bg-surface-2 px-1 font-mono text-[12px] text-text-primary"
            >
              {part.slice(1, -1)}
            </code>
          );
        }

        return <React.Fragment key={`${part}-${index}`}>{part}</React.Fragment>;
      })}
    </>
  );
}

function BlockTitle({
  icon,
  children,
}: {
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <p className="mb-3 flex items-center gap-2 font-mono text-[11px] uppercase tracking-wider text-text-tertiary">
      <span className="text-line-cta">{icon}</span>
      {children}
    </p>
  );
}

function RuntimeCallout({ connector }: { connector: ConnectorMeta }) {
  if (connector.status === 'enabled') return null;

  return (
    <Card muted className="border-l-[3px] border-l-[#78350f] px-4 py-3">
      <p className="flex items-start gap-2 text-[13px] leading-relaxed text-text-secondary">
        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-[#78350f]" aria-hidden="true" />
        <span>
          <span className="font-semibold text-text-primary">Catalog-only availability. </span>
          Setup details describe provider-side requirements, but pm runtime checks and ETL are
          unavailable until this connector&apos;s native Go port is enabled.
        </span>
      </p>
    </Card>
  );
}

function SetupStep({
  index,
  title,
  body,
  last,
}: {
  index: number;
  title: string;
  body: string;
  last: boolean;
}) {
  return (
    <div className="relative grid grid-cols-[2rem_minmax(0,1fr)] gap-3">
      {!last && (
        <span
          aria-hidden="true"
          className="absolute left-[15px] top-8 h-[calc(100%-1.5rem)] w-px bg-line-structure"
        />
      )}
      <span
        className={cn(
          'z-10 flex h-8 w-8 items-center justify-center border font-mono text-[11px] font-semibold',
          '[background:rgba(52,211,153,0.12)] [border-color:rgba(52,211,153,0.4)] text-line-cta',
        )}
        aria-hidden="true"
      >
        {index}
      </span>
      <div className="min-w-0 pb-5">
        <p className="text-[13px] font-semibold text-text-primary">{title}</p>
        <p className="mt-1 text-[13px] leading-relaxed text-text-secondary">
          <InlineText text={body} />
        </p>
      </div>
    </div>
  );
}

export function ConnectorSetup({ connector }: ConnectorSetupProps) {
  const enrichment = connectorEnrichment(connector.slug);
  if (!enrichment) return null;

  const { prerequisites, authMethods, setupSteps, sources } = enrichment;

  return (
    <div className="space-y-4">
      <RuntimeCallout connector={connector} />

      <div className="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <div className="min-w-0 space-y-4">
          {prerequisites.length > 0 && (
            <Card className="p-4">
              <BlockTitle icon={<ShieldCheck className="h-4 w-4" aria-hidden="true" />}>
                Prerequisites
              </BlockTitle>
              <ul className="space-y-2">
                {prerequisites.map((item, i) => (
                  <li key={i} className="flex gap-2 text-[13px] leading-relaxed text-text-secondary">
                    <span className="mt-2 h-1.5 w-1.5 shrink-0 bg-line-cta" aria-hidden="true" />
                    <span>
                      <InlineText text={item} />
                    </span>
                  </li>
                ))}
              </ul>
            </Card>
          )}

          {authMethods.length > 0 && (
            <div className="space-y-2">
              <BlockTitle icon={<KeyRound className="h-4 w-4" aria-hidden="true" />}>
                Authentication
              </BlockTitle>
              {authMethods.map((method) => (
                <Card key={method.name} muted className="p-4">
                  <p className="text-[13px] font-semibold text-text-primary">{method.name}</p>
                  <p className="mt-1 text-[13px] leading-relaxed text-text-secondary">
                    <InlineText text={method.summary} />
                  </p>
                </Card>
              ))}
            </div>
          )}
        </div>

        {setupSteps.length > 0 && (
          <Card className="p-4">
            <BlockTitle icon={<ListChecks className="h-4 w-4" aria-hidden="true" />}>
              Setup steps
            </BlockTitle>
            <div>
              {setupSteps.map((step, i) => (
                <SetupStep
                  key={`${step.title}-${i}`}
                  index={i + 1}
                  title={step.title}
                  body={step.body}
                  last={i === setupSteps.length - 1}
                />
              ))}
            </div>
          </Card>
        )}
      </div>

      {sources.length > 0 && (
        <Card muted className="p-4">
          <BlockTitle icon={<ExternalLink className="h-4 w-4" aria-hidden="true" />}>
            Verified sources
          </BlockTitle>
          <div className="grid min-w-0 gap-2 sm:grid-cols-2">
            {sources.map((src) => (
              <a
                key={src.url}
                href={src.url}
                target="_blank"
                rel="noreferrer"
                className="link-box group relative min-w-0 border border-line-structure bg-surface-bg px-3 py-2 transition-colors hover:bg-surface-1"
              >
                <span aria-hidden className="corner-box-hover-child" />
                <span className="block text-[13px] font-medium leading-snug text-text-primary group-hover:text-line-cta">
                  {src.title}
                </span>
                <span className="mt-1 block min-w-0 truncate font-mono text-[11px] text-text-disabled">
                  {src.url}
                </span>
              </a>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}
