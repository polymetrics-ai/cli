import * as React from 'react';
import type { ConnectorMeta } from '@/lib/connectors.types';
import { cn } from '@/lib/utils';

interface ConnectorStatusSummaryProps {
  connector: ConnectorMeta;
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
  const incrementalStreams = connector.streams.filter((stream) => stream.incremental).length;

  return (
    <div className="grid min-w-0 grid-cols-1 border-l border-t border-line-structure sm:grid-cols-3">
      <Stat label="Definition" value="Bundle" positive />
      <Stat
        label="Release"
        value={connector.releaseStage === 'ga' ? 'GA' : formatValue(connector.releaseStage)}
        positive={connector.releaseStage === 'ga'}
      />
      <Stat label="Integration" value={connector.categoryLabel} />
      <Stat label="ETL streams" value={String(connector.streams.length)} positive={connector.streams.length > 0} />
      <Stat label="Incremental" value={String(incrementalStreams)} positive={incrementalStreams > 0} />
      <Stat label="Write actions" value={String(connector.writeActions.length)} positive={connector.writeActions.length > 0} />
    </div>
  );
}
