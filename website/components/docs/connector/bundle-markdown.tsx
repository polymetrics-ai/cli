import * as React from 'react';
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/table';

type Block =
  | { type: 'heading'; level: number; text: string }
  | { type: 'paragraph'; lines: string[] }
  | { type: 'list'; ordered: boolean; items: string[] }
  | { type: 'code'; language: string; code: string }
  | { type: 'table'; rows: string[][] };

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

    const heading = line.match(/^(#{1,4})\s+(.+)$/);
    if (heading) {
      blocks.push({ type: 'heading', level: heading[1].length, text: heading[2].trim() });
      index += 1;
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
      while (index < lines.length && /^\s*[-*]\s+/.test(lines[index])) {
        items.push(lines[index].replace(/^\s*[-*]\s+/, '').trim());
        index += 1;
      }
      blocks.push({ type: 'list', ordered: false, items });
      continue;
    }

    if (/^\s*\d+\.\s+/.test(line)) {
      const items: string[] = [];
      while (index < lines.length && /^\s*\d+\.\s+/.test(lines[index])) {
        items.push(lines[index].replace(/^\s*\d+\.\s+/, '').trim());
        index += 1;
      }
      blocks.push({ type: 'list', ordered: true, items });
      continue;
    }

    const paragraph: string[] = [];
    while (
      index < lines.length &&
      lines[index].trim() &&
      !/^```/.test(lines[index]) &&
      !/^(#{1,4})\s+/.test(lines[index]) &&
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
                className="text-line-cta underline underline-offset-2 hover:text-text-primary"
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

function HeadingBlock({ level, text }: { level: number; text: string }) {
  const className =
    level <= 2
      ? 'mt-7 font-square text-[15px] font-semibold leading-[1.3] text-text-primary'
      : 'mt-5 font-square text-[13px] font-semibold uppercase tracking-wider text-text-secondary';

  return (
    <h3 className={className}>
      <InlineMarkdown text={text} />
    </h3>
  );
}

export function BundleMarkdown({ markdown }: { markdown: string }) {
  const blocks = parseBlocks(markdown);

  if (blocks.length === 0) {
    return (
      <p className="text-[13px] italic text-text-tertiary">
        No connector documentation was generated for this bundle.
      </p>
    );
  }

  return (
    <div className="space-y-3 border border-line-structure bg-surface-bg px-4 py-4">
      {blocks.map((block, index) => {
        if (block.type === 'heading') {
          return <HeadingBlock key={index} level={block.level} text={block.text} />;
        }

        if (block.type === 'paragraph') {
          return (
            <p key={index} className="max-w-none text-[13px] leading-relaxed text-text-secondary">
              <InlineMarkdown text={block.lines.join(' ')} />
            </p>
          );
        }

        if (block.type === 'list') {
          const List = block.ordered ? 'ol' : 'ul';
          return (
            <List
              key={index}
              className={
                block.ordered
                  ? 'ml-5 list-decimal space-y-2 text-[13px] leading-relaxed text-text-secondary'
                  : 'ml-5 list-disc space-y-2 text-[13px] leading-relaxed text-text-secondary'
              }
            >
              {block.items.map((item, itemIndex) => (
                <li key={`${index}-${itemIndex}`}>
                  <InlineMarkdown text={item} />
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
