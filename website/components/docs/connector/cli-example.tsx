'use client';

import { useMemo, useState } from 'react';
import { AlertTriangle, Check, Clipboard, Terminal } from 'lucide-react';
import type { ConnectorMeta } from '@/lib/connectors.catalog.generated';
import { Badge } from '@/components/ui/badge';
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

function connectionName(connector: ConnectorMeta): string {
  return `${connector.slug}-connection`;
}

function buildRows(connector: ConnectorMeta): CommandRow[] {
  const conn = connectionName(connector);
  const rows: CommandRow[] = [
    {
      label: 'Inspect connector metadata',
      command: `pm connectors inspect ${connector.slug} --json`,
    },
  ];

  if (connector.capabilities.check) {
    rows.push({
      label: 'Run a credential check',
      command: `pm etl check --connector ${connector.slug} --config <field>=<value> --json`,
    });
  }

  if (connector.capabilities.read) {
    rows.push(
      {
        label: 'Read a bounded stream',
        command: `pm etl read --connector ${connector.slug} --stream <stream_name> --limit 100 --json`,
      },
      {
        label: 'Run bounded ETL',
        command: `pm etl run --connection ${conn} --stream <stream_name> --batch-size 500 --json`,
      },
    );
  }

  if (connector.capabilities.write && connector.writeActions.length > 0) {
    rows.push(
      {
        label: 'Plan a reverse-ETL write',
        command: `pm reverse plan <plan_name> --destination ${connector.slug}:<credential_name> --action <action_name> --map <column>:<field> --json`,
      },
      {
        label: 'Approval gate',
        note: 'Preview the plan and execute only with an explicit approval token.',
      },
    );
  }

  if (rows.length === 1) {
    rows.push({
      label: 'Runtime note',
      note: 'This bundle exposes metadata but no read or write capability.',
    });
  }

  return rows;
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
              Commands reflect this connector&apos;s declared bundle capabilities.
            </p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Badge variant="capability">Bundle</Badge>
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
