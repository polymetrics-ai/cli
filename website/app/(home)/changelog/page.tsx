import type { Metadata } from 'next';
import Link from 'next/link';
import { Cable, Database, FileJson, Github, Lock } from 'lucide-react';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.generated';

export const metadata: Metadata = {
  title: 'Changelog',
  description: 'Product changelog for the Polymetrics CLI website and connector catalog.',
};

const entries = [
  {
    version: 'v0.1.0',
    date: 'today',
    title: `${CONNECTOR_CATALOG_COUNT} connector catalog pages`,
    body:
      'The website now exposes source and destination connector pages with generated JSON metadata for agents and documentation workflows.',
    icon: Cable,
    href: '/docs/connectors',
  },
  {
    version: 'v0.1.0',
    date: 'today',
    title: 'Embedded DuckDB SQL engine',
    body:
      'Query extracted connector data locally with DuckDB SQL before promoting the result into a reverse-ETL write.',
    icon: Database,
    href: '/docs/query',
  },
  {
    version: 'v0.1.0',
    date: 'today',
    title: 'Local AES-GCM vault',
    body:
      'Credentials stay local-first and encrypted so connector setup can be tested without passing raw secrets through documentation or CI.',
    icon: Lock,
    href: '/docs/architecture',
  },
  {
    version: 'v0.1.0',
    date: 'today',
    title: 'Agent-native JSON output',
    body:
      'CLI commands expose structured output for automation loops, validation agents, and downstream platform integrations.',
    icon: FileJson,
    href: '/docs/agent-guide',
  },
];

export default function ChangelogPage() {
  return (
    <main className="mx-auto w-full max-w-[95rem] px-6 py-16 md:py-24">
      <header className="mb-12 grid gap-6 border-b border-line-structure pb-10 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="min-w-0">
          <span className="font-mono text-[12px] uppercase tracking-widest text-text-disabled">
            Product log
          </span>
          <h1 className="mt-4 max-w-[11ch] font-analog text-[44px] leading-[1] text-text-primary md:text-[68px]">
            Changes that ship the loop.
          </h1>
        </div>
        <div className="flex flex-col justify-end gap-4 text-[14px] leading-relaxed text-text-tertiary">
          <p>
            Release notes for the local-first extract, query, and reverse-ETL workflow.
          </p>
          <a
            href="https://github.com/polymetrics-ai/cli/releases"
            target="_blank"
            rel="noreferrer"
            className="inline-flex w-fit items-center gap-2 border border-line-structure bg-surface-1 px-3 py-2 font-square text-[12px] font-semibold text-text-secondary transition-colors hover:border-line-cta hover:text-text-primary"
          >
            <Github className="h-3.5 w-3.5" aria-hidden="true" />
            GitHub releases
          </a>
        </div>
      </header>

      <section aria-label="Latest changes" className="divide-y divide-line-structure border-y border-line-structure">
        {entries.map(({ version, date, title, body, icon: Icon, href }) => (
          <Link
            key={title}
            href={href}
            className="grid gap-4 bg-surface-bg px-0 py-6 transition-colors hover:bg-surface-1 md:grid-cols-[8rem_2.75rem_minmax(0,1fr)_auto] md:items-start"
          >
            <span className="font-mono text-[11px] uppercase tracking-widest text-text-disabled">
              {version} / {date}
            </span>
            <span className="flex h-9 w-9 items-center justify-center border border-line-structure bg-surface-1 text-line-cta">
              <Icon className="h-4 w-4" aria-hidden="true" />
            </span>
            <span className="min-w-0">
              <span className="block font-square text-[18px] font-semibold text-text-primary">
                {title}
              </span>
              <span className="mt-2 block max-w-[74ch] text-[14px] leading-relaxed text-text-tertiary">
                {body}
              </span>
            </span>
            <span className="font-mono text-[11px] uppercase tracking-widest text-text-disabled md:pt-1">
              Read
            </span>
          </Link>
        ))}
      </section>
    </main>
  );
}
