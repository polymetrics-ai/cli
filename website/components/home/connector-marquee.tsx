'use client';

import Link from 'next/link';
import { ArrowRight, Cable } from 'lucide-react';
import {
  CONNECTORS,
  CONNECTOR_CATALOG_COUNT,
  type ConnectorListItem,
} from '@/lib/connectors.generated';
import { ConnectorIcon } from '@/components/docs/connector/connector-icon';

function Chip({ item }: { item: ConnectorListItem }) {
  return (
    <span className="mr-3 inline-flex h-12 shrink-0 items-center gap-3 border border-line-structure bg-surface-1 px-3.5 py-2">
      <ConnectorIcon
        icon={item.icon}
        name={item.name}
        className="h-8 w-8 min-w-8 bg-surface-2 text-[10px] [aspect-ratio:1/1]"
        imageClassName="h-5 w-5"
      />
      <span className="whitespace-nowrap text-[14px] font-medium text-text-secondary">{item.name}</span>
    </span>
  );
}

/* A single marquee row.
   One animated track holds the list TWICE; translateX(0 → -50%) moves exactly
   one copy width, so the loop is seamless. Spacing lives on the chips. */
function Row({
  items,
  reverse = false,
  duration = '80s',
}: {
  items: ConnectorListItem[];
  reverse?: boolean;
  duration?: string;
}) {
  const doubled = [...items, ...items];
  return (
    <div
      className={`flex w-max flex-nowrap ${reverse ? 'animate-marquee-reverse' : 'animate-marquee'}`}
      style={{ ['--marquee-dur' as string]: duration }}
      aria-hidden
    >
      {doubled.map((item, i) => (
        <Chip key={`${item.name}-${i}`} item={item} />
      ))}
    </div>
  );
}

/*
  Connector marquee. `compact` = single row (used in the hero strip).
  Otherwise two rows scrolling opposite directions with a count heading.
*/
export function ConnectorMarquee({ compact = false }: { compact?: boolean }) {
  // Famous-heavy first half leads row 1; remainder leads row 2.
  const mid = Math.ceil(CONNECTORS.length / 2);
  const rowTop = CONNECTORS;
  const rowBottom = [...CONNECTORS.slice(mid), ...CONNECTORS.slice(0, mid)];

  if (compact) {
    return (
      <div className="overflow-hidden">
        <div className="flex flex-col gap-3 border-b border-line-structure bg-surface-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex min-w-0 items-center gap-3">
            <span className="flex h-9 w-9 shrink-0 items-center justify-center border border-line-structure bg-surface-bg">
              <Cable className="h-4 w-4 text-line-cta" aria-hidden="true" />
            </span>
            <span className="min-w-0">
              <span className="block font-square text-[13px] font-semibold uppercase tracking-wider text-text-secondary">
                Connector catalog
              </span>
              <span className="mt-0.5 block text-[12px] leading-snug text-text-tertiary">
                Source and destination connector list, not customer logos.
              </span>
            </span>
          </div>
          <Link
            href="/docs/connectors"
            className="group inline-flex w-fit items-center gap-1.5 border border-line-structure bg-surface-bg px-2.5 py-1.5 font-mono text-[10px] uppercase tracking-wider text-text-tertiary transition-colors hover:border-line-cta hover:bg-surface-2 hover:text-text-primary"
          >
            Browse {CONNECTOR_CATALOG_COUNT}
            <ArrowRight className="h-3 w-3 transition-transform group-hover:translate-x-0.5" aria-hidden="true" />
          </Link>
        </div>
        <div className="marquee-track relative overflow-hidden py-5">
          <Row items={rowTop} duration="760s" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <span className="text-[10px] font-semibold uppercase tracking-widest text-text-disabled">
        {CONNECTOR_CATALOG_COUNT}+ connectors
      </span>
      <div className="marquee-track relative flex flex-col gap-3 overflow-hidden border border-line-structure bg-surface-bg py-5">
        <Row items={rowTop} duration="580s" />
        <Row items={rowBottom} reverse duration="640s" />
      </div>
      <p className="text-[12px] text-text-tertiary">
        {CONNECTOR_CATALOG_COUNT} catalog connectors share a Go HTTP &amp; database toolkit.
        Missing one?{' '}
        <a
          href="https://github.com/polymetrics-ai/cli/issues"
          target="_blank"
          rel="noreferrer"
          className="text-text-secondary underline hover:text-text-primary transition-colors"
        >
          Open an issue
        </a>
        .
      </p>
    </div>
  );
}
