import * as React from 'react';
import {
  AlertTriangle,
  Database,
  ExternalLink,
  FileJson,
  FileText,
  KeyRound,
  ListChecks,
  PencilLine,
  ShieldCheck,
  Terminal,
} from 'lucide-react';
import type { ConnectorMeta } from '@/lib/connectors.types';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/table';
import { cn } from '@/lib/utils';

type SectionKey = 'overview' | 'auth' | 'streams' | 'writes' | 'limits';

type Block =
  | { type: 'paragraph'; lines: string[] }
  | { type: 'list'; ordered: boolean; items: string[] }
  | { type: 'code'; language: string; code: string }
  | { type: 'table'; rows: string[][] };

interface DocSections {
  overview: string;
  auth: string;
  streams: string;
  writes: string;
  limits: string;
}

interface ConnectionField {
  name: string;
  type: string;
  required: boolean;
  secret: boolean;
  defaultValue: string;
  format: string;
  description: string;
}

interface DetailItem {
  name: string;
  method: string;
  path: string;
  detail: string;
  raw: string;
}

const EMPTY_SECTIONS: DocSections = {
  overview: '',
  auth: '',
  streams: '',
  writes: '',
  limits: '',
};

const SECTION_KEYS: Record<string, SectionKey> = {
  overview: 'overview',
  'auth setup': 'auth',
  'streams notes': 'streams',
  'write actions & risks': 'writes',
  'known limits': 'limits',
};

function normalizeHeading(value: string): string {
  return value.trim().toLowerCase();
}

function splitSections(markdown: string): DocSections {
  const sections = { ...EMPTY_SECTIONS };
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  let active: SectionKey | null = null;
  let buffer: string[] = [];

  function flush() {
    if (active) {
      sections[active] = buffer.join('\n').trim();
    }
    buffer = [];
  }

  for (const line of lines) {
    const heading = line.match(/^#{1,2}\s+(.+)$/);
    if (heading) {
      flush();
      active = SECTION_KEYS[normalizeHeading(heading[1])] ?? null;
      continue;
    }

    if (active) buffer.push(line);
  }

  flush();
  return sections;
}

function parseTableRow(line: string): string[] {
  return line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((cell) => cell.trim());
}

function isTableSeparator(line: string): boolean {
  return /^\s*\|?[\s:-]+\|[\s|:-]+\|?\s*$/.test(line);
}

function parseBlocks(markdown: string): Block[] {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const blocks: Block[] = [];
  let index = 0;

  while (index < lines.length) {
    const line = lines[index];

    if (!line.trim()) {
      index += 1;
      continue;
    }

    const fence = line.match(/^```(\w+)?\s*$/);
    if (fence) {
      const codeLines: string[] = [];
      index += 1;
      while (index < lines.length && !/^```\s*$/.test(lines[index])) {
        codeLines.push(lines[index]);
        index += 1;
      }
      if (index < lines.length) index += 1;
      blocks.push({ type: 'code', language: fence[1] ?? '', code: codeLines.join('\n') });
      continue;
    }

    if (
      line.includes('|') &&
      index + 1 < lines.length &&
      isTableSeparator(lines[index + 1])
    ) {
      const rows: string[][] = [parseTableRow(line)];
      index += 2;
      while (index < lines.length && lines[index].includes('|') && lines[index].trim()) {
        rows.push(parseTableRow(lines[index]));
        index += 1;
      }
      blocks.push({ type: 'table', rows });
      continue;
    }

    if (/^\s*[-*]\s+/.test(line)) {
      const items: string[] = [];
      let current = '';

      while (index < lines.length) {
        const item = lines[index].match(/^\s*[-*]\s+(.+)$/);
        if (item) {
          if (current) items.push(current.trim());
          current = item[1].trim();
          index += 1;
          continue;
        }

        if (current && /^\s+\S/.test(lines[index])) {
          current += ` ${lines[index].trim()}`;
          index += 1;
          continue;
        }

        break;
      }

      if (current) items.push(current.trim());
      blocks.push({ type: 'list', ordered: false, items });
      continue;
    }

    if (/^\s*\d+\.\s+/.test(line)) {
      const items: string[] = [];
      let current = '';

      while (index < lines.length) {
        const item = lines[index].match(/^\s*\d+\.\s+(.+)$/);
        if (item) {
          if (current) items.push(current.trim());
          current = item[1].trim();
          index += 1;
          continue;
        }

        if (current && /^\s+\S/.test(lines[index])) {
          current += ` ${lines[index].trim()}`;
          index += 1;
          continue;
        }

        break;
      }

      if (current) items.push(current.trim());
      blocks.push({ type: 'list', ordered: true, items });
      continue;
    }

    const paragraph: string[] = [];
    while (
      index < lines.length &&
      lines[index].trim() &&
      !/^```/.test(lines[index]) &&
      !/^\s*[-*]\s+/.test(lines[index]) &&
      !/^\s*\d+\.\s+/.test(lines[index]) &&
      !(lines[index].includes('|') && index + 1 < lines.length && isTableSeparator(lines[index + 1]))
    ) {
      paragraph.push(lines[index].trim());
      index += 1;
    }
    blocks.push({ type: 'paragraph', lines: paragraph });
  }

  return blocks;
}

function parseListItems(markdown: string): string[] {
  return parseBlocks(markdown)
    .filter((block): block is Extract<Block, { type: 'list' }> => block.type === 'list')
    .flatMap((block) => block.items);
}

function contentBeforeFirstList(markdown: string): string {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const firstList = lines.findIndex((line) => /^\s*[-*]\s+/.test(line));
  return (firstList === -1 ? lines : lines.slice(0, firstList)).join('\n').trim();
}

function collectListAfterLabel(markdown: string, label: string): string[] {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const start = lines.findIndex((line) => line.trim() === label);
  if (start === -1) return [];

  const collected: string[] = [];
  let seenList = false;
  for (let index = start + 1; index < lines.length; index += 1) {
    const line = lines[index];
    if (/^\s*[-*]\s+/.test(line)) {
      seenList = true;
      collected.push(line);
      continue;
    }
    if (seenList && /^\s+\S/.test(line)) {
      collected.push(line);
      continue;
    }
    if (!seenList && !line.trim()) continue;
    if (seenList && !line.trim()) continue;
    if (seenList) break;
  }

  return parseListItems(collected.join('\n'));
}

function stripInlineCode(value: string): string {
  return value.replace(/`([^`]+)`/g, '$1').trim();
}

function parseConnectionFields(auth: string): ConnectionField[] {
  return collectListAfterLabel(auth, 'Connection fields:').map((item) => {
    const match = item.match(/^`([^`]+)`\s+\(([^)]*)\);?\s*(.*)$/);
    if (!match) {
      return {
        name: stripInlineCode(item),
        type: '',
        required: false,
        secret: false,
        defaultValue: '',
        format: '',
        description: item,
      };
    }

    const meta = match[2].split(',').map((part) => part.trim()).filter(Boolean);
    const description = match[3].trim().replace(/^\.$/, '');
    const defaultValue = description.match(/default `([^`]+)`/)?.[1] ?? '';
    const format = description.match(/format `([^`]+)`/)?.[1] ?? '';
    const type = meta.find((part) => !['required', 'optional', 'secret'].includes(part)) ?? '';

    return {
      name: match[1],
      type,
      required: meta.includes('required'),
      secret: meta.includes('secret'),
      defaultValue,
      format,
      description,
    };
  });
}

function extractParagraph(markdown: string, startsWith: string): string {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n');
  const start = lines.findIndex((line) => line.trim().startsWith(startsWith));
  if (start === -1) return '';

  const paragraph: string[] = [];
  for (let index = start; index < lines.length; index += 1) {
    const line = lines[index];
    if (!line.trim() && paragraph.length > 0) break;
    if (!line.trim()) continue;
    if (index !== start && /^\S[^:]+:$/.test(line.trim())) break;
    if (index !== start && /^\s*[-*]\s+/.test(line)) break;
    paragraph.push(line.trim());
  }

  return paragraph.join(' ');
}

function extractUrl(text: string): string {
  return text.match(/https?:\/\/\S+/)?.[0]?.replace(/[.,;:]$/, '') ?? '';
}

function parseDetailItems(markdown: string): DetailItem[] {
  return parseListItems(markdown).map((item) => {
    const match = item.match(/^`([^`]+)`:\s+([A-Z]+)\s+`([^`]+)`(?:\s+-\s+)?(.*)$/);
    if (!match) {
      return { name: '', method: '', path: '', detail: item, raw: item };
    }

    return {
      name: match[1],
      method: match[2],
      path: match[3],
      detail: match[4].trim(),
      raw: item,
    };
  });
}

function InlineMarkdown({ text }: { text: string }) {
  const parts = text.split(/(`[^`]+`|\[[^\]]+]\([^)]+\)|https?:\/\/\S+)/g);

  return (
    <>
      {parts.map((part, index) => {
        if (!part) return null;

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

        const markdownLink = part.match(/^\[([^\]]+)]\(([^)]+)\)$/);
        if (markdownLink) {
          return (
            <a
              key={`${part}-${index}`}
              href={markdownLink[2]}
              target="_blank"
              rel="noreferrer"
              className="text-line-cta underline underline-offset-2 hover:text-text-primary"
            >
              {markdownLink[1]}
            </a>
          );
        }

        if (/^https?:\/\//.test(part)) {
          const url = part.replace(/[.,;:]$/, '');
          const suffix = part.slice(url.length);
          return (
            <React.Fragment key={`${part}-${index}`}>
              <a
                href={url}
                target="_blank"
                rel="noreferrer"
                className="break-all text-line-cta underline underline-offset-2 hover:text-text-primary"
              >
                {url}
              </a>
              {suffix}
            </React.Fragment>
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
  action,
}: {
  icon: React.ReactNode;
  children: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="mb-3 flex items-center justify-between gap-3">
      <p className="flex min-w-0 items-center gap-2 font-mono text-[11px] uppercase tracking-wider text-text-tertiary">
        <span className="shrink-0 text-line-cta">{icon}</span>
        <span className="truncate">{children}</span>
      </p>
      {action}
    </div>
  );
}

function RichBlocks({ markdown, compact = false }: { markdown: string; compact?: boolean }) {
  const blocks = parseBlocks(markdown);

  if (blocks.length === 0) return null;

  return (
    <div className={cn('space-y-3', compact && 'space-y-2')}>
      {blocks.map((block, index) => {
        if (block.type === 'paragraph') {
          return (
            <p
              key={index}
              className={cn(
                'max-w-none text-[13px] leading-relaxed text-text-secondary',
                compact && 'text-[12px]',
              )}
            >
              <InlineMarkdown text={block.lines.join(' ')} />
            </p>
          );
        }

        if (block.type === 'list') {
          const List = block.ordered ? 'ol' : 'ul';
          return (
            <List
              key={index}
              className={cn(
                block.ordered ? 'ml-5 list-decimal' : 'ml-0 list-none',
                'space-y-2 text-[13px] leading-relaxed text-text-secondary',
                compact && 'text-[12px]',
              )}
            >
              {block.items.map((item, itemIndex) => (
                <li
                  key={`${index}-${itemIndex}`}
                  className={block.ordered ? undefined : 'grid grid-cols-[0.6rem_minmax(0,1fr)] gap-2'}
                >
                  {!block.ordered && (
                    <span className="mt-2 h-1.5 w-1.5 bg-line-cta" aria-hidden="true" />
                  )}
                  <span>
                    <InlineMarkdown text={item} />
                  </span>
                </li>
              ))}
            </List>
          );
        }

        if (block.type === 'code') {
          return (
            <pre
              key={index}
              className="overflow-x-auto border border-line-structure bg-surface-1 px-3 py-2 font-mono text-[12px] leading-relaxed text-text-primary"
            >
              <code>{block.code}</code>
            </pre>
          );
        }

        const [head = [], ...body] = block.rows;
        return (
          <Table key={index}>
            <THead>
              <TR>
                {head.map((cell, cellIndex) => (
                  <TH key={`${index}-head-${cellIndex}`}>
                    <InlineMarkdown text={cell} />
                  </TH>
                ))}
              </TR>
            </THead>
            <TBody>
              {body.map((row, rowIndex) => (
                <TR key={`${index}-row-${rowIndex}`}>
                  {row.map((cell, cellIndex) => (
                    <TD key={`${index}-cell-${rowIndex}-${cellIndex}`}>
                      <InlineMarkdown text={cell} />
                    </TD>
                  ))}
                </TR>
              ))}
            </TBody>
          </Table>
        );
      })}
    </div>
  );
}

function OverviewPanel({
  connector,
  overview,
}: {
  connector: ConnectorMeta;
  overview: string;
}) {
  const apiUrl = extractUrl(overview) || connector.docUrl;
  const summary = contentBeforeFirstList(overview)
    .split(/\n\s*\n/)
    .map((part) => part.trim())
    .filter((part) => {
      return (
        part &&
        !part.startsWith('Readable streams:') &&
        !part.startsWith('Write actions:') &&
        !part.startsWith('This connector is read-only') &&
        !part.startsWith('Service API documentation:')
      );
    })
    .join('\n\n');

  return (
    <Card className="overflow-hidden">
      <div className="flex flex-col gap-3 border-b border-line-structure bg-surface-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-2">
          <FileText className="h-4 w-4 shrink-0 text-line-cta" aria-hidden="true" />
          <div className="min-w-0">
            <p className="font-square text-[15px] font-semibold leading-tight text-text-primary">
              Generated bundle note
            </p>
            <p className="mt-0.5 text-[12px] leading-snug text-text-tertiary">
              Parsed from docs.md and kept complete for human review and agent ingestion.
            </p>
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap gap-1.5">
          <Badge variant="capability">docs.md</Badge>
          <Badge variant="category">{connector.streams.length} streams</Badge>
          <Badge variant={connector.writeActions.length > 0 ? 'capability' : 'category'}>
            {connector.writeActions.length} writes
          </Badge>
        </div>
      </div>
      <div className="grid gap-3 p-4 lg:grid-cols-[minmax(0,1fr)_220px]">
        <div className="min-w-0">
          {summary ? <RichBlocks markdown={summary} /> : null}
          <div className="mt-4 flex flex-wrap gap-1.5">
            {connector.capabilityLabels.map((label) => (
              <Badge key={label} variant="capability">
                {label}
              </Badge>
            ))}
          </div>
        </div>
        <div className="grid min-w-0 border-l border-t border-line-structure text-[12px]">
          <div className="border-b border-r border-line-structure bg-surface-bg p-3">
            <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
              Connector
            </p>
            <p className="mt-1 break-words font-mono text-[12px] text-text-primary">
              {connector.slug}
            </p>
          </div>
          <div className="border-b border-r border-line-structure bg-surface-bg p-3">
            <p className="font-mono text-[10px] uppercase tracking-wider text-text-tertiary">
              API docs
            </p>
            {apiUrl ? (
              <a
                href={apiUrl}
                target="_blank"
                rel="noreferrer"
                className="mt-1 flex min-w-0 items-center gap-1.5 break-all text-line-cta underline underline-offset-2 hover:text-text-primary"
              >
                Open source
                <ExternalLink className="h-3 w-3 shrink-0" aria-hidden="true" />
              </a>
            ) : (
              <p className="mt-1 text-text-disabled">Not declared</p>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
}

function FieldCard({ field }: { field: ConnectionField }) {
  return (
    <div className="border border-line-structure bg-surface-bg px-3 py-2.5">
      <div className="flex min-w-0 flex-wrap items-center gap-1.5">
        <code className="break-all font-mono text-[12px] font-semibold text-text-primary">
          {field.name}
        </code>
        <Badge variant={field.required ? 'status-enabled' : 'category'}>
          {field.required ? 'required' : 'optional'}
        </Badge>
        {field.secret && <Badge variant="release-beta">secret</Badge>}
        {field.type && <Badge variant="category">{field.type}</Badge>}
      </div>
      {field.description ? (
        <p className="mt-2 text-[12px] leading-relaxed text-text-secondary">
          <InlineMarkdown text={field.description} />
        </p>
      ) : null}
      {(field.defaultValue || field.format) && (
        <div className="mt-2 flex flex-wrap gap-1.5">
          {field.defaultValue && (
            <span className="border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[11px] text-text-tertiary">
              default: {field.defaultValue}
            </span>
          )}
          {field.format && (
            <span className="border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[11px] text-text-tertiary">
              format: {field.format}
            </span>
          )}
        </div>
      )}
    </div>
  );
}

function SetupStep({
  index,
  title,
  body,
  command,
  last,
}: {
  index: number;
  title: string;
  body: string;
  command?: string;
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
      <span className="z-10 flex h-8 w-8 items-center justify-center border font-mono text-[11px] font-semibold [background:rgba(52,211,153,0.12)] [border-color:rgba(52,211,153,0.4)] text-line-cta">
        {index}
      </span>
      <div className="min-w-0 pb-5">
        <p className="text-[13px] font-semibold text-text-primary">{title}</p>
        <p className="mt-1 text-[13px] leading-relaxed text-text-secondary">
          <InlineMarkdown text={body} />
        </p>
        {command ? (
          <pre className="mt-2 overflow-x-auto whitespace-pre-wrap break-words border border-line-structure bg-surface-1 px-2 py-1.5 font-mono text-[11px] leading-relaxed text-text-primary">
            <code>{command}</code>
          </pre>
        ) : null}
      </div>
    </div>
  );
}

function AuthPanel({
  connector,
  auth,
}: {
  connector: ConnectorMeta;
  auth: string;
}) {
  const fields = parseConnectionFields(auth);
  const authBehavior = collectListAfterLabel(auth, 'Authentication behavior:');
  const secretFields = extractParagraph(auth, 'Secret fields');
  const defaults = extractParagraph(auth, 'Default configuration values:');
  const requestDefaults = extractParagraph(auth, 'Requests use');
  const check = extractParagraph(auth, 'Connection checks call');
  const requiredFields = fields.filter((field) => field.required).map((field) => field.name);
  const secretNames = fields.filter((field) => field.secret).map((field) => field.name);
  const firstStream = connector.streams[0]?.name ?? '<stream_name>';
  const firstAction = connector.writeActions[0]?.name ?? '<action_name>';
  const steps = [
    {
      title: 'Inspect the bundle contract',
      body: 'Use the connector metadata before creating credentials or plans. This exposes capabilities, stream names, and write actions without reading secrets.',
      command: `pm connectors inspect ${connector.slug} --json`,
    },
    {
      title: 'Provide config and secrets safely',
      body:
        requiredFields.length > 0
          ? `Required fields: ${requiredFields.map((field) => `\`${field}\``).join(', ')}. Keep secret fields in environment variables or stdin.`
          : 'No required fields are declared. Keep any optional secret fields in environment variables or stdin.',
      command:
        secretNames.length > 0
          ? `pm credentials add ${connector.slug}-local --connector ${connector.slug} --from-env ${secretNames[0]}=${secretNames[0].toUpperCase()}`
          : `pm credentials add ${connector.slug}-local --connector ${connector.slug} --config <field>=<value>`,
    },
    {
      title: 'Check the connection',
      body: check || 'Run the connector check before reads or writes so failures stay isolated from data movement.',
      command: `pm etl check --connector ${connector.slug} --config <field>=<value> --json`,
    },
    {
      title: connector.writeActions.length > 0 ? 'Read with bounds, write with approval' : 'Read with explicit bounds',
      body:
        connector.writeActions.length > 0
          ? 'Use bounded reads for ETL. Reverse ETL must follow plan, preview, approval, and execute using the declared write actions.'
          : 'Use bounded reads for ETL and pick an explicit stream. This connector does not declare write actions.',
      command:
        connector.writeActions.length > 0
          ? `pm etl read --connector ${connector.slug} --stream ${firstStream} --limit 100 --json\npm reverse plan <plan_name> --destination ${connector.slug}:<credential_name> --action ${firstAction} --json`
          : `pm etl read --connector ${connector.slug} --stream ${firstStream} --limit 100 --json`,
    },
  ];

  if (!auth.trim()) return null;

  return (
    <div className="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
      <div className="min-w-0 space-y-4">
        <Card className="p-4">
          <BlockTitle
            icon={<ShieldCheck className="h-4 w-4" aria-hidden="true" />}
            action={<Badge variant="category">{fields.length} fields</Badge>}
          >
            Connection fields
          </BlockTitle>
          {fields.length > 0 ? (
            <div className="grid gap-2">
              {fields.map((field) => (
                <FieldCard key={field.name} field={field} />
              ))}
            </div>
          ) : (
            <p className="text-[13px] italic text-text-tertiary">
              No connection fields are declared in docs.md.
            </p>
          )}
        </Card>

        <Card muted className="p-4">
          <BlockTitle icon={<KeyRound className="h-4 w-4" aria-hidden="true" />}>
            Authentication
          </BlockTitle>
          <div className="space-y-3">
            {authBehavior.length > 0 && (
              <ul className="space-y-2">
                {authBehavior.map((item, index) => (
                  <li key={index} className="grid grid-cols-[0.6rem_minmax(0,1fr)] gap-2 text-[13px] leading-relaxed text-text-secondary">
                    <span className="mt-2 h-1.5 w-1.5 bg-line-cta" aria-hidden="true" />
                    <span>
                      <InlineMarkdown text={item} />
                    </span>
                  </li>
                ))}
              </ul>
            )}
            {[secretFields, defaults, requestDefaults, check].filter(Boolean).map((text) => (
              <p key={text} className="text-[13px] leading-relaxed text-text-secondary">
                <InlineMarkdown text={text} />
              </p>
            ))}
          </div>
        </Card>
      </div>

      <Card className="p-4">
        <BlockTitle icon={<ListChecks className="h-4 w-4" aria-hidden="true" />}>
          Agent setup path
        </BlockTitle>
        <div>
          {steps.map((step, index) => (
            <SetupStep
              key={step.title}
              index={index + 1}
              title={step.title}
              body={step.body}
              command={step.command}
              last={index === steps.length - 1}
            />
          ))}
        </div>
      </Card>
    </div>
  );
}

function DetailTable({
  title,
  icon,
  items,
  empty,
}: {
  title: string;
  icon: React.ReactNode;
  items: DetailItem[];
  empty: string;
}) {
  const parsed = items.filter((item) => item.name);
  const unparsed = items.filter((item) => !item.name);

  return (
    <Card className="p-4">
      <BlockTitle
        icon={icon}
        action={<Badge variant="category">{items.length} entries</Badge>}
      >
        {title}
      </BlockTitle>
      {parsed.length > 0 ? (
        <div className="max-h-[520px] overflow-auto">
          <Table>
            <THead className="sticky top-0 z-10">
              <TR>
                <TH>Name</TH>
                <TH>Request</TH>
                <TH>Notes</TH>
              </TR>
            </THead>
            <TBody>
              {parsed.map((item) => (
                <TR key={`${item.name}-${item.method}-${item.path}`}>
                  <TD>
                    <code className="break-words font-mono text-[12px] text-text-primary">
                      {item.name}
                    </code>
                  </TD>
                  <TD className="min-w-[180px]">
                    <div className="flex flex-wrap gap-1.5">
                      <Badge variant={item.method === 'GET' ? 'category' : 'capability'}>
                        {item.method}
                      </Badge>
                      <code className="break-all font-mono text-[12px] text-text-tertiary">
                        {item.path}
                      </code>
                    </div>
                  </TD>
                  <TD className="max-w-[54ch] leading-relaxed text-text-secondary">
                    <InlineMarkdown text={item.detail || item.raw} />
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        </div>
      ) : (
        <p className="text-[13px] italic text-text-tertiary">{empty}</p>
      )}
      {unparsed.length > 0 && (
        <div className="mt-3 border border-line-structure bg-surface-1 p-3">
          <RichBlocks markdown={unparsed.map((item) => `- ${item.raw}`).join('\n')} compact />
        </div>
      )}
    </Card>
  );
}

function BehaviorPanel({
  title,
  description,
  icon,
  section,
  tableTitle,
  tableIcon,
  empty,
}: {
  title: string;
  description: string;
  icon: React.ReactNode;
  section: string;
  tableTitle: string;
  tableIcon: React.ReactNode;
  empty: string;
}) {
  const general = contentBeforeFirstList(section);
  const items = parseDetailItems(section);

  if (!section.trim()) return null;

  return (
    <div className="space-y-3">
      <div>
        <p className="flex items-center gap-2 font-square text-[15px] font-semibold leading-tight text-text-primary">
          <span className="text-line-cta">{icon}</span>
          {title}
        </p>
        <p className="mt-1 text-[13px] leading-relaxed text-text-tertiary">
          {description}
        </p>
      </div>
      {general ? (
        <Card muted className="p-4">
          <RichBlocks markdown={general} />
        </Card>
      ) : null}
      <DetailTable title={tableTitle} icon={tableIcon} items={items} empty={empty} />
    </div>
  );
}

function LimitsPanel({ limits }: { limits: string }) {
  const items = parseListItems(limits);
  const general = contentBeforeFirstList(limits);

  if (!limits.trim()) return null;

  return (
    <Card muted className="p-4">
      <BlockTitle icon={<AlertTriangle className="h-4 w-4 text-[#78350f]" aria-hidden="true" />}>
        Known limits
      </BlockTitle>
      {general ? <RichBlocks markdown={general} /> : null}
      {items.length > 0 && (
        <ul className="mt-3 grid gap-2 sm:grid-cols-2">
          {items.map((item, index) => (
            <li
              key={index}
              className="grid grid-cols-[0.6rem_minmax(0,1fr)] gap-2 border border-line-structure bg-surface-bg px-3 py-2 text-[12px] leading-relaxed text-text-secondary"
            >
              <span className="mt-2 h-1.5 w-1.5 bg-line-cta" aria-hidden="true" />
              <span>
                <InlineMarkdown text={item} />
              </span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

function SourcesPanel({ connector }: { connector: ConnectorMeta }) {
  const sources = [
    ...connector.docs.map((doc) => ({
      title: doc.title,
      href: doc.url,
      external: true,
      note: doc.url,
    })),
    {
      title: 'Machine-readable connector data',
      href: `/docs/connectors/${connector.slug}/data.json`,
      external: false,
      note: `${connector.slug}/data.json`,
    },
  ].filter((source, index, arr) => {
    return source.href && arr.findIndex((item) => item.href === source.href) === index;
  });

  return (
    <Card muted className="p-4">
      <BlockTitle icon={<ExternalLink className="h-4 w-4" aria-hidden="true" />}>
        Verified sources
      </BlockTitle>
      <div className="grid min-w-0 gap-2 sm:grid-cols-2">
        {sources.map((source) => (
          <a
            key={source.href}
            href={source.href}
            target={source.external ? '_blank' : undefined}
            rel={source.external ? 'noreferrer' : undefined}
            className="link-box group relative min-w-0 border border-line-structure bg-surface-bg px-3 py-2 transition-colors hover:bg-surface-1"
          >
            <span aria-hidden className="corner-box-hover-child" />
            <span className="flex min-w-0 items-center gap-1.5 text-[13px] font-medium leading-snug text-text-primary group-hover:text-line-cta">
              <span className="truncate">{source.title}</span>
              {source.external ? <ExternalLink className="h-3 w-3 shrink-0" aria-hidden="true" /> : null}
            </span>
            <span className="mt-1 block min-w-0 truncate font-mono text-[11px] text-text-disabled">
              {source.note}
            </span>
          </a>
        ))}
      </div>
    </Card>
  );
}

function RawSource({ markdown }: { markdown: string }) {
  return (
    <details className="group border border-line-structure bg-surface-bg">
      <summary className="flex cursor-pointer list-none items-center justify-between gap-3 bg-surface-1 px-4 py-3">
        <span className="flex min-w-0 items-center gap-2">
          <FileJson className="h-4 w-4 shrink-0 text-line-cta" aria-hidden="true" />
          <span className="font-mono text-[11px] uppercase tracking-wider text-text-tertiary">
            Agent-readable docs.md source
          </span>
        </span>
        <Badge variant="category">verbatim</Badge>
      </summary>
      <div className="border-t border-line-structure p-3">
        <pre className="max-h-[420px] overflow-auto whitespace-pre-wrap break-words font-mono text-[11px] leading-relaxed text-text-secondary">
          <code>{markdown}</code>
        </pre>
      </div>
    </details>
  );
}

export function BundleMarkdown({
  connector,
  markdown,
}: {
  connector: ConnectorMeta;
  markdown: string;
}) {
  const sections = splitSections(markdown);

  if (!markdown.trim()) {
    return (
      <p className="text-[13px] italic text-text-tertiary">
        No connector documentation was generated for this bundle.
      </p>
    );
  }

  return (
    <div className="space-y-5">
      <OverviewPanel connector={connector} overview={sections.overview} />

      <div className="space-y-3">
        <div>
          <p className="flex items-center gap-2 font-square text-[15px] font-semibold leading-tight text-text-primary">
            <KeyRound className="h-4 w-4 text-line-cta" aria-hidden="true" />
            Setup and authentication
          </p>
          <p className="mt-1 text-[13px] leading-relaxed text-text-tertiary">
            Provider-side requirements, auth shape, safe credential handling, and an agent-ready setup path.
          </p>
        </div>
        <AuthPanel connector={connector} auth={sections.auth} />
      </div>

      <SourcesPanel connector={connector} />

      <BehaviorPanel
        title="Stream behavior"
        description="Pagination, endpoint shape, record paths, fan-out behavior, and incremental cursor notes from streams.json."
        icon={<Database className="h-4 w-4" aria-hidden="true" />}
        section={sections.streams}
        tableTitle="Stream endpoint details"
        tableIcon={<Database className="h-4 w-4" aria-hidden="true" />}
        empty="No stream endpoint notes are declared."
      />

      <BehaviorPanel
        title="Write behavior and risks"
        description="Write action request shape, mutation kind, required fields, and risk notes from writes.json."
        icon={<PencilLine className="h-4 w-4" aria-hidden="true" />}
        section={sections.writes}
        tableTitle="Write action details"
        tableIcon={<PencilLine className="h-4 w-4" aria-hidden="true" />}
        empty="This connector does not declare write action details."
      />

      <LimitsPanel limits={sections.limits} />

      <Card className="p-4">
        <BlockTitle icon={<Terminal className="h-4 w-4" aria-hidden="true" />}>
          Agent contract
        </BlockTitle>
        <p className="text-[13px] leading-relaxed text-text-secondary">
          This page keeps the rendered notes readable for humans, while the raw docs.md source and
          the connector data.json endpoint preserve the complete generated bundle contract for agents.
        </p>
      </Card>

      <RawSource markdown={markdown} />
    </div>
  );
}
