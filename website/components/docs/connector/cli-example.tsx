'use client';

import { useMemo, useState } from 'react';
import { AlertTriangle, Check, Clipboard, Terminal } from 'lucide-react';
import type { ConnectorMeta } from '@/lib/connectors.catalog.generated';
import { Badge, statusVariant } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface CliExampleProps {
  connector: ConnectorMeta;
}

interface CommandRow {
  label: string;
  command?: string;
  note?: string;
}

function envName(slug: string, field: string): string {
  const prefix = slug
    .replace(/^(source|destination)-/, '')
    .replace(/[^a-z0-9]+/gi, '_')
    .replace(/^_+|_+$/g, '')
    .toUpperCase();
  const suffix = field
    .replace(/[^a-z0-9]+/gi, '_')
    .replace(/^_+|_+$/g, '')
    .toUpperCase();

  return `PM_${prefix}_${suffix}`;
}

function connectionName(connector: ConnectorMeta): string {
  return `${connector.pmName || connector.slug}-connection`;
}

function inspectName(connector: ConnectorMeta): string {
  return connector.pmName || connector.slug;
}

function buildCredentialArgs(connector: ConnectorMeta): string {
  const secretArgs = connector.secrets
    .slice(0, 3)
    .map((field) => `--from-env ${field}=${envName(connector.slug, field)}`);
  const configArgs = connector.config
    .filter((field) => field.required && !field.secret && !connector.secrets.includes(field.name))
    .slice(0, 2)
    .map((field) => `--config ${field.name}=<value>`);

  const args = [...secretArgs, ...configArgs];
  return args.length > 0 ? args.join(' ') : '--config <field>=<value>';
}

function buildRows(connector: ConnectorMeta): CommandRow[] {
  const name = inspectName(connector);

  if (connector.status !== 'enabled' || !connector.capabilities.etl) {
    return [
      {
        label: 'Inspect catalog metadata',
        command: `pm connectors inspect ${connector.slug} --json`,
      },
      {
        label: 'Review native-port plan',
        command: `pm connectors port-plan ${connector.slug} --json`,
      },
      {
        label: 'Runtime note',
        note: 'ETL and credential checks are unavailable until this catalog connector is enabled.',
      },
    ];
  }

  const conn = connectionName(connector);
  return [
    {
      label: 'Inspect connector metadata',
      command: `pm connectors inspect ${name} --json`,
    },
    {
      label: 'Store config and secrets',
      command: `pm credentials add ${conn} --connector ${name} ${buildCredentialArgs(connector)}`,
    },
    {
      label: 'Validate stored credentials',
      command: `pm credentials test ${conn} --json`,
    },
    {
      label: 'Run bounded ETL',
      command: `pm etl run --connection ${conn} --stream <stream_name> --batch-size 500 --json`,
    },
  ];
}

function rowsToText(rows: CommandRow[]): string {
  return rows
    .map((row) => {
      if (row.command) return `# ${row.label}\n${row.command}`;
      return `# ${row.label}\n# ${row.note}`;
    })
    .join('\n\n');
}

export function CliExample({ connector }: CliExampleProps) {
  const rows = useMemo(() => buildRows(connector), [connector]);
  const copyText = useMemo(() => rowsToText(rows), [rows]);
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(copyText);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  }

  return (
    <Card className="overflow-hidden">
      <div className="flex flex-col gap-3 border-b border-line-structure bg-surface-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-2">
          <Terminal className="h-4 w-4 shrink-0 text-line-cta" aria-hidden="true" />
          <div className="min-w-0">
            <p className="font-square text-[15px] font-semibold leading-tight text-text-primary">
              pm CLI
            </p>
            <p className="mt-0.5 text-[12px] leading-snug text-text-tertiary">
              Commands reflect this connector&apos;s current runtime status.
            </p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Badge variant={statusVariant(connector.status)}>
            {connector.status === 'enabled' ? 'Enabled' : 'Catalog only'}
          </Badge>
          <Button aria-label="Copy CLI commands" variant="quiet" onClick={handleCopy}>
            {copied ? (
              <Check className="h-3.5 w-3.5" aria-hidden="true" />
            ) : (
              <Clipboard className="h-3.5 w-3.5" aria-hidden="true" />
            )}
            {copied ? 'Copied' : 'Copy'}
          </Button>
        </div>
      </div>

      <div className="max-h-[260px] overflow-auto p-2">
        <div className="grid gap-2">
          {rows.map((row, index) => (
            <div
              key={`${row.label}-${index}`}
              className={cn(
                'grid grid-cols-[2rem_minmax(0,1fr)] border border-line-structure bg-surface-bg',
                row.note && 'bg-surface-1',
              )}
            >
              <span className="flex items-start justify-center border-r border-line-structure px-2 py-2 font-mono text-[11px] text-text-disabled">
                {String(index + 1).padStart(2, '0')}
              </span>
              <div className="min-w-0 px-3 py-2">
                <p className="mb-1 font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
                  {row.label}
                </p>
                {row.command ? (
                  <pre className="overflow-x-auto whitespace-pre-wrap break-words font-mono text-[12px] leading-[1.55] text-text-primary">
                    <code>{row.command}</code>
                  </pre>
                ) : (
                  <p className="flex items-start gap-2 text-[12px] leading-relaxed text-text-secondary">
                    <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-[#78350f]" aria-hidden="true" />
                    <span>{row.note}</span>
                  </p>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </Card>
  );
}
