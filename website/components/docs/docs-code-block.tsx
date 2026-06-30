'use client';

import { useMemo, useState } from 'react';
import { Check, Clipboard, FileCode2, Terminal } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';

interface DocsCodeBlockProps {
  code: string;
  language?: string;
}

function labelForLanguage(language?: string): string {
  if (!language) return 'text';
  if (language === 'sh' || language === 'shell') return 'bash';
  if (language === 'text' || language === 'txt') return 'text';
  return language;
}

export function DocsCodeBlock({ code, language }: DocsCodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const label = labelForLanguage(language);
  const lines = useMemo(() => code.replace(/\n$/, '').split('\n'), [code]);
  const Icon = label === 'bash' ? Terminal : FileCode2;

  async function handleCopy() {
    await navigator.clipboard.writeText(code.replace(/\n$/, ''));
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  }

  return (
    <figure
      className="not-prose my-5 overflow-hidden border border-line-structure bg-surface-bg shadow-[0_14px_36px_rgba(12,31,23,0.07)]"
      data-docs-code-block=""
    >
      <figcaption className="flex flex-wrap items-center justify-between gap-2 border-b border-line-structure bg-surface-1 px-3 py-2">
        <span className="flex min-w-0 items-center gap-2">
          <span className="flex size-7 shrink-0 items-center justify-center border border-line-structure bg-surface-bg text-line-cta">
            <Icon aria-hidden="true" />
          </span>
          <span className="min-w-0">
            <span className="block font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
              Code example
            </span>
            <span className="block font-mono text-[10px] uppercase tracking-wider text-text-disabled">
              {lines.length} {lines.length === 1 ? 'line' : 'lines'}
            </span>
          </span>
        </span>
        <span className="flex shrink-0 items-center gap-2">
          <Badge variant="category">{label}</Badge>
          <Button
            type="button"
            variant="quiet"
            size="sm"
            onClick={handleCopy}
            className="bg-surface-bg"
            aria-label="Copy code"
          >
            {copied ? (
              <Check data-icon="inline-start" aria-hidden="true" />
            ) : (
              <Clipboard data-icon="inline-start" aria-hidden="true" />
            )}
            <span aria-live="polite">{copied ? 'Copied' : 'Copy'}</span>
          </Button>
        </span>
      </figcaption>
      <ScrollArea className="max-h-[420px] bg-surface-bg">
        <pre className="m-0 overflow-x-auto border-0 bg-transparent p-0 shadow-none">
          <code className="grid min-w-max bg-transparent py-3 font-mono text-[12px] leading-[1.65] text-text-primary">
            {lines.map((line, index) => (
              <span
                key={`${index}-${line}`}
                className="grid grid-cols-[2.75rem_minmax(0,1fr)] border-b border-line-structure/50 last:border-b-0"
              >
                <span className="select-none border-r border-line-structure bg-surface-1 px-2 py-0.5 text-right font-mono text-[10px] text-text-disabled">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <span className="px-3 py-0.5">{line || ' '}</span>
              </span>
            ))}
          </code>
        </pre>
      </ScrollArea>
    </figure>
  );
}
