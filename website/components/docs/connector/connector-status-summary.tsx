import * as React from 'react';
import type { ConnectorMeta } from '@/lib/connectors.catalog.generated';
import { cn } from '@/lib/utils';

interface ConnectorStatusSummaryProps {
  connector: ConnectorMeta;
}

function formatStatus(status: string): string {
  if (status === 'enabled') return 'Enabled';
  if (status === 'planned_native_port') return 'Catalog only';
  return status.replace(/_/g, ' ');
}

function formatValue(value: string): string {
  return value.replace(/_/g, ' ');
}

function Stat({
  label,
  value,
  positive,
}: {
  label: string;
  value: string;
  positive?: boolean;
}) {
  return (
    <div className="min-h-[68px] border-b border-r border-line-structure bg-surface-bg p-3">
      <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
        {label}
      </p>
      <p
        className={cn(
          'mt-2 flex items-center gap-1.5 text-[13px] font-medium capitalize text-text-primary',
          positive && 'text-line-cta',
        )}
      >
        {positive && <span className="h-1.5 w-1.5 bg-line-cta" aria-hidden="true" />}
        {value}
      </p>
    </div>
  );
}

export function ConnectorStatusSummary({ connector }: ConnectorStatusSummaryProps) {
  return (
    <div className="grid min-w-0 grid-cols-1 border-l border-t border-line-structure sm:grid-cols-3">
      <Stat label="Runtime" value={formatStatus(connector.status)} positive={connector.status === 'enabled'} />
      <Stat label="Support" value={formatValue(connector.supportLevel)} positive={connector.supportLevel === 'certified'} />
      <Stat label="Release" value={formatValue(connector.releaseStage)} positive={connector.releaseStage === 'generally_available'} />
      <Stat label="Type" value={connector.type} />
      <Stat label="Category" value={formatValue(connector.category)} />
      <Stat label="Incremental" value={connector.incremental ? 'Supported' : 'Unavailable'} positive={connector.incremental} />
    </div>
  );
}
