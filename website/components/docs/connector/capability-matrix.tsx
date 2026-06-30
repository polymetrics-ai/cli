import * as React from 'react';
import { Check, Minus } from 'lucide-react';
import type { ConnectorCapabilities } from '@/lib/connectors.catalog.generated';
import { cn } from '@/lib/utils';

interface CapabilityMatrixProps {
  capabilities: ConnectorCapabilities;
}

interface CapabilityRow {
  label: string;
  key: keyof Omit<ConnectorCapabilities, 'unsupportedReason'>;
}

const CAPABILITY_ROWS: CapabilityRow[] = [
  { label: 'Metadata', key: 'metadata' },
  { label: 'Check', key: 'check' },
  { label: 'Catalog', key: 'catalog' },
  { label: 'Read', key: 'read' },
  { label: 'Write', key: 'write' },
  { label: 'Query', key: 'query' },
  { label: 'ETL', key: 'etl' },
  { label: 'Reverse ETL', key: 'reverseEtl' },
];

export function CapabilityMatrix({ capabilities }: CapabilityMatrixProps) {
  const { unsupportedReason, ...caps } = capabilities;

  return (
    <div className="space-y-3">
      <div className="grid min-w-0 grid-cols-1 border-l border-t border-line-structure sm:grid-cols-2 lg:grid-cols-4">
        {CAPABILITY_ROWS.map(({ label, key }) => {
          const supported = caps[key as keyof typeof caps];
          return (
            <div
              key={key}
              className={cn(
                'min-h-[72px] border-b border-r border-line-structure bg-surface-bg p-3',
                supported && '[background:rgba(52,211,153,0.08)]',
              )}
            >
              <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                {label}
              </p>
              <p
                className={cn(
                  'mt-2 inline-flex items-center gap-1.5 text-[13px] font-semibold',
                  supported ? 'text-line-cta' : 'text-text-disabled',
                )}
              >
                {supported ? (
                  <Check className="h-3.5 w-3.5" aria-hidden="true" />
                ) : (
                  <Minus className="h-3.5 w-3.5" aria-hidden="true" />
                )}
                {supported ? 'Supported' : 'Unavailable'}
              </p>
            </div>
          );
        })}
      </div>
      {unsupportedReason ? (
        <p className="border border-l-[3px] border-line-structure border-l-surface-cta-primary bg-surface-1 px-3 py-2.5 text-[13px] leading-relaxed text-text-tertiary">
          <span className="font-semibold text-text-secondary">Runtime note: </span>
          {unsupportedReason}
        </p>
      ) : null}
    </div>
  );
}
