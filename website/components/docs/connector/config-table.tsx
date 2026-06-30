import * as React from 'react';
import type { ConnectorConfigField } from '@/lib/connectors.catalog.generated';
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';

interface ConfigTableProps {
  config: ConnectorConfigField[];
  secrets?: string[];
}

export function ConfigTable({ config, secrets = [] }: ConfigTableProps) {
  if (config.length === 0) {
    return (
      <p className="text-[13px] text-text-tertiary italic">
        No configuration fields advertised.
      </p>
    );
  }

  return (
    <Table>
      <THead>
        <TR>
          <TH>Name</TH>
          <TH>Type</TH>
          <TH>Required</TH>
          <TH>Secret</TH>
          <TH>Description</TH>
        </TR>
      </THead>
      <TBody>
        {config.map((field) => {
          const isSecret = field.secret || secrets.includes(field.name);
          return (
            <TR key={field.name}>
              <TD>
                <span className="break-words font-mono text-[12px] text-text-primary">
                  {field.name}
                </span>
              </TD>
              <TD className="font-mono text-[12px] text-text-tertiary">
                {field.type}
              </TD>
              <TD>
                <Badge variant={field.required ? 'status-enabled' : 'category'}>
                  {field.required ? 'Required' : 'Optional'}
                </Badge>
              </TD>
              <TD>
                {isSecret ? (
                  <span
                    className="inline-flex items-center border px-2 py-0.5 font-mono text-[11px] font-medium uppercase tracking-wider text-[#78350f] [background:rgba(251,191,36,0.12)] [border-color:rgba(251,191,36,0.4)]"
                    title="Secret field; do not log or expose values"
                  >
                    Secret
                  </span>
                ) : (
                  <Badge variant="category">Plain</Badge>
                )}
              </TD>
              <TD className="max-w-[46ch] break-words text-text-secondary leading-relaxed">
                {field.description || (
                  <span className="italic text-text-disabled">—</span>
                )}
              </TD>
            </TR>
          );
        })}
      </TBody>
    </Table>
  );
}
