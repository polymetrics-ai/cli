import * as React from 'react';
import { Check, Minus } from 'lucide-react';
import type { ConnectorCapabilities } from '@/lib/connectors.types';
import { cn } from '@/lib/utils';

interface CapabilityMatrixProps {
  capabilities: ConnectorCapabilities;
}

interface CapabilityRow {
  label: string;
  key: keyof ConnectorCapabilities;
}

const CAPABILITY_ROWS: CapabilityRow[] = [
  { label: 'Check', key: 'check' },
  { label: 'Read', key: 'read' },
  { label: 'Write', key: 'write' },
  { label: 'Query', key: 'query' },
  { label: 'CDC', key: 'cdc' },
  { label: 'Dynamic schema', key: 'dynamicSchema' },
];

export function CapabilityMatrix({ capabilities }: CapabilityMatrixProps) {
  return (
    <div className="space-y-3">
      <div className="grid min-w-0 grid-cols-1 border-l border-t border-line-structure sm:grid-cols-2 lg:grid-cols-3">
        {CAPABILITY_ROWS.map(({ label, key }) => {
          const supported = capabilities[key];
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
    </div>
  );
}
