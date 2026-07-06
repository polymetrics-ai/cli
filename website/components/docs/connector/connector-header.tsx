import * as React from 'react';
import type { ConnectorMeta } from '@/lib/connectors.types';
import {
  Badge,
  releaseVariant,
} from '@/components/ui/badge';
import { ConnectorIcon } from '@/components/docs/connector/connector-icon';
import { Card } from '@/components/ui/card';

interface ConnectorHeaderProps {
  connector: ConnectorMeta;
}

// ── Human-readable label formatters ──────────────────────────────────────

function formatReleaseStage(stage: string): string {
  if (stage === 'ga') return 'GA';
  if (stage === 'generally_available') return 'GA';
  if (stage === 'beta') return 'Beta';
  if (stage === 'alpha') return 'Alpha';
  if (stage === 'custom') return 'Custom';
  return stage;
}

export function ConnectorHeader({ connector }: ConnectorHeaderProps) {
  const { name, categoryLabel, releaseStage, capabilityLabels, icon } = connector;
  const highlightedCapabilities = capabilityLabels.filter((capability) =>
    ['read', 'write', 'query', 'cdc', 'dynamic schema'].includes(capability),
  );

  return (
    <Card className="mb-5 p-5">
      <div className="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 items-start gap-4">
          <ConnectorIcon
            icon={icon}
            name={name}
            className="h-11 w-11 bg-surface-1 text-[13px]"
            imageClassName="h-7 w-7"
          />
          <div className="min-w-0 flex-1">
            <p className="mb-1 font-mono text-[11px] uppercase tracking-wider text-text-tertiary">
              Connector reference
            </p>
            <h1 className="break-words font-square text-[26px] font-bold leading-[1.15] text-text-primary sm:text-[30px]">
              {name} connector
            </h1>
            <div className="mt-3 flex max-w-full flex-wrap gap-2">
              <Badge variant="category">{categoryLabel}</Badge>
              <Badge variant={releaseVariant(releaseStage)}>
                {formatReleaseStage(releaseStage)}
              </Badge>
              {highlightedCapabilities.map((capability) => (
                <Badge key={capability} variant="capability">
                  {capability}
                </Badge>
              ))}
            </div>
          </div>
        </div>
        <div className="shrink-0 border border-line-structure bg-surface-1 px-3 py-2">
          <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
            Catalog slug
          </p>
          <code className="mt-1 block font-mono text-[12px] text-text-primary">
            {connector.slug}
          </code>
        </div>
      </div>
    </Card>
  );
}
