'use client';

import { useState } from 'react';
import { ChevronDown } from 'lucide-react';
import { CornerBox } from '@/components/ui/corner-box';
import { Text } from '@/components/ui/text';

const faqs = [
  {
    q: 'Why not a containerized ETL stack?',
    a: 'Most self-hosted ETL platforms need Docker and 8 GB+ of services to run locally, and they are ETL-only (no SQL engine, no reverse-ETL, no agent contract). pm runs as a single static binary, has embedded DuckDB, and supports bidirectional sync.',
  },
  {
    q: 'Why not Fivetran?',
    a: 'Fivetran is managed cloud-only, charges per row, and has no local mode. You have no control over where your data goes in transit. pm runs entirely on your machine, and credentials never leave your disk.',
  },
  {
    q: 'Why not dlt (data load tool)?',
    a: 'dlt is a Python library for ETL. It has no built-in SQL engine, no reverse-ETL, and is not designed for agent-safe operation. pm is a compiled Go binary with all three capabilities unified and a stable structured-output contract for LLM agents.',
  },
  {
    q: 'Who is pm for?',
    a: 'Developers and data engineers who want local-first data pipelines without infrastructure overhead. AI agents that need a stable, structured interface to extract data, query it, and act on it. Teams who want Fivetran-class connector quality without the cloud dependency.',
  },
  {
    q: 'Is pm open source?',
    a: 'Yes. pm is open source under the MIT license. The connector catalog, the DuckDB integration, and the CLI are all public. Contributions welcome.',
  },
];

export function FaqAccordion() {
  const [open, setOpen] = useState<number | null>(null);

  return (
    <div className="flex flex-col">
      {faqs.map((faq, i) => (
        <CornerBox
          key={faq.q}
          className="-mb-px cursor-pointer transition-[background] duration-200 hover:bg-surface-1/50 group relative"
          onClick={() => setOpen(open === i ? null : i)}
        >
          {/* Left accent bar on hover */}
          <span className="absolute left-0 top-0 bottom-0 w-0.5 bg-text-primary opacity-0 group-hover:opacity-100 transition-opacity duration-150 rounded-full" />

          {/* Question row */}
          <div className="flex items-center gap-3 px-6 py-4">
            <span className="shrink-0 w-7 text-[11px] font-mono text-text-disabled">
              {String(i + 1).padStart(2, '0')}
            </span>
            <Text
              size="m"
              className={`text-left font-medium flex-1 ${open === i ? 'text-text-primary' : 'text-text-secondary'}`}
            >
              {faq.q}
            </Text>
            <ChevronDown
              size={16}
              className={`shrink-0 text-text-tertiary transition-transform duration-300 ${open === i ? 'rotate-180' : ''}`}
            />
          </div>

          {/* Answer — always rendered for SEO, slides in/out via max-height */}
          <div
            style={{
              maxHeight: open === i ? '300px' : '0px',
              overflow: 'hidden',
              transition: 'max-height 0.32s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            <div className="px-6 pb-5 border-t border-line-structure">
              <Text size="s" className="text-left mt-3 text-text-tertiary">
                {faq.a}
              </Text>
            </div>
          </div>
        </CornerBox>
      ))}
    </div>
  );
}
