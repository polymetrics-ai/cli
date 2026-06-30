'use client';

import Link from 'next/link';
import {
  ArrowRight, Github, Download, Database, ArrowLeftRight,
  Plug, Bot, Lock, RefreshCw,
} from 'lucide-react';
import { useState, useCallback, type MouseEvent } from 'react';
import { HomeSection } from '@/components/home/home-section';
import { CornerBox } from '@/components/ui/corner-box';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { TextHighlight } from '@/components/ui/text-highlight';
import { TiltCard } from '@/components/home/tilt-card';
import { HomeSidebar } from '@/components/home/home-sidebar';
import { HomeAside } from '@/components/home/home-aside';
import { CodeTerminal } from '@/components/home/code-terminal';
import { FaqAccordion } from '@/components/home/faq-accordion';
import { ConnectorMarquee } from '@/components/home/connector-marquee';
import { SiteFooter } from '@/components/home/site-footer';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.generated';


/* ── Tool cards ───────────────────────────────────────────────────────── */
type Tool = {
  icon: React.ComponentType<{ size?: number; className?: string }>;
  title: string;
  description: string;
  href: string;
};

const toolsRow1: Tool[] = [
  {
    icon: Download,
    title: 'Extract (ETL)',
    description: `Pull from the ${CONNECTOR_CATALOG_COUNT}-connector catalog: GitHub, Stripe, HubSpot, Postgres, and more. Cursor-incremental and full-refresh sync modes.`,
    href: '/docs/connectors',
  },
  {
    icon: Database,
    title: 'Query (DuckDB SQL)',
    description: 'Run real analytical SQL over extracted data: joins, window functions, aggregations. No separate warehouse required.',
    href: '/docs/query',
  },
  {
    icon: ArrowLeftRight,
    title: 'Write Back (Reverse-ETL)',
    description: 'Push query results back to any destination. Create Jira issues, upsert HubSpot contacts, open GitHub PRs from your data.',
    href: '/docs/reverse-etl',
  },
];

const toolsRow2: Tool[] = [
  {
    icon: Plug,
    title: `${CONNECTOR_CATALOG_COUNT} Connectors`,
    description: 'Catalog metadata, docs, and native Go runtime coverage on a shared HTTP/DB toolkit.',
    href: '/docs/connectors',
  },
  {
    icon: Bot,
    title: 'Agent-native',
    description: 'Every command speaks --json. Stable exit codes. Writes are approval-gated.',
    href: '/docs/agent',
  },
  {
    icon: Lock,
    title: 'Local Vault',
    description: 'AES-GCM encrypted credential store. Your secrets never leave your machine.',
    href: '/docs/vault',
  },
  {
    icon: RefreshCw,
    title: 'Bidirectional',
    description: 'Sources and destinations are unified. Extract from GitHub, write back to it.',
    href: '/docs/bidirectional',
  },
];

function ToolCard({ tool, large = false }: { tool: Tool; large?: boolean }) {
  const Icon = tool.icon;
  return (
    <Link href={tool.href} className="block h-full group">
      <CornerBox
        hoverStripes
        className="h-full flex flex-col gap-3 p-5 transition-[background,box-shadow] duration-200 group-hover:shadow-[0_8px_24px_-4px_rgba(64,64,57,0.14)]"
      >
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center justify-center w-8 h-8 rounded-sm bg-surface-1 border border-line-structure text-text-secondary transition-colors group-hover:bg-surface-2">
            <Icon size={15} />
          </span>
          <Text size="m" className="text-left font-medium text-text-secondary">
            {tool.title}
          </Text>
        </div>
        <Text size="s" className={`text-left text-text-tertiary ${large ? 'max-w-[30ch]' : ''}`}>
          {tool.description}
        </Text>
        <span className="mt-auto flex items-center gap-1 text-[12px] text-text-tertiary opacity-0 group-hover:opacity-100 transition-opacity duration-200">
          Read more <ArrowRight size={11} />
        </span>
      </CornerBox>
    </Link>
  );
}


/* ── Page ─────────────────────────────────────────────────────────────── */
export default function HomePage() {
  // Hero mouse-tracking glow
  const [heroGlow, setHeroGlow] = useState({ x: 0, y: 400 });
  const handleHeroMove = useCallback((e: MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    setHeroGlow({ x: e.clientX - rect.left, y: e.clientY - rect.top });
  }, []);

  return (
    <div className="flex mx-auto w-full max-w-[95rem] overflow-clip">

      {/* Left sidebar */}
      <HomeSidebar />

      {/* Center content */}
      <main className="flex-1 min-w-0 pattern-bg flex flex-col gap-0 pb-8 overflow-hidden xl:px-5 2xl:px-10">

        {/* ── Hero ─────────────────────────────────────────────────────── */}
        <HomeSection id="overview" pattern="p001" className="pt-5 sm:pt-8 md:pt-[60px]">
          {/* Main hero — mouse-tracking ambient glow */}
          <CornerBox
            onMouseMove={handleHeroMove}
            className="flex flex-col gap-6 md:gap-8 items-center px-4 py-12 sm:px-8 sm:py-16 relative overflow-hidden"
            style={{
              background: `
                radial-gradient(600px circle at ${heroGlow.x}px ${heroGlow.y}px, rgba(52,211,153,0.10), transparent 60%),
                radial-gradient(ellipse 80% 50% at 50% 35%, rgba(52,211,153,0.12) 0%, transparent 70%),
                var(--surface-bg)
              `,
            }}
          >
            {/* Fixed ambient orb — sits behind content */}
            <span
              className="absolute top-[35%] left-1/2 w-[520px] h-[320px] rounded-full orb-breathe pointer-events-none select-none"
              style={{
                background: 'radial-gradient(ellipse, rgba(52,211,153,0.30) 0%, rgba(16,185,129,0.08) 50%, transparent 75%)',
                filter: 'blur(48px)',
                transform: 'translate(-50%,-50%)',
              }}
              aria-hidden
            />

            <Heading
              as="h1"
              size="big"
              className="flex flex-col items-center gap-1 md:gap-2 text-center font-medium leading-[105%] relative"
            >
              <span className="flex flex-wrap justify-center gap-x-2 gap-y-1">
                <TextHighlight highlightClassName="mix-blend-multiply">Extract,</TextHighlight>
                <TextHighlight highlightClassName="mix-blend-multiply">query,</TextHighlight>
              </span>
              <span className="flex flex-wrap justify-center gap-x-2 gap-y-1">
                <TextHighlight highlightClassName="mix-blend-multiply">act</TextHighlight>
                <span>– repeat.</span>
              </span>
            </Heading>

            <Text className="max-w-[42ch] relative">
              <code className="font-mono text-[14px] font-medium bg-surface-1 border border-line-structure rounded px-1.5 py-0.5">pm</code>
              {' '}is a local-first, single-binary data engine. Browse {CONNECTOR_CATALOG_COUNT} connectors, query with embedded DuckDB SQL, and write results back. No Docker, no servers, agent-native by design.
            </Text>

            <div className="flex flex-wrap gap-3 justify-center items-center relative">
              <Link
                href="/docs/quickstart"
                className="btn-shine inline-flex items-center gap-2 border border-emerald-900 bg-emerald-800 px-5 py-2.5 text-[14px] font-medium text-white transition-opacity hover:opacity-90"
              >
                Get started <ArrowRight size={14} />
              </Link>
              <a
                href="https://github.com/karthik-sivadas/polymetrics-cli"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 rounded-sm border border-line-structure bg-surface-bg px-5 py-2.5 text-[14px] font-medium text-text-secondary transition-colors hover:bg-surface-1"
              >
                <Github size={14} /> View on GitHub
              </a>
            </div>
          </CornerBox>

          {/* Connector marquee — every connector scrolls past */}
          <CornerBox className="-mt-px overflow-hidden">
            <ConnectorMarquee compact />
          </CornerBox>
        </HomeSection>

        {/* ── All the tools ──────────────────────────────────────────── */}
        <HomeSection id="tools" pattern="p011" className="pt-[120px]">
          <div className="flex flex-col gap-3 items-start mb-10">
            <Heading>
              All the tools,{' '}
              <TextHighlight className="pr-1.5">one</TextHighlight>
              <TextHighlight>binary.</TextHighlight>
            </Heading>
            <Text className="text-left max-w-[46ch]">
              Extract, query, and act: the full data loop without leaving your terminal. One binary replaces your ETL pipeline, your SQL warehouse, and your reverse-ETL tool.
            </Text>
          </div>

          <div className="flex flex-col gap-2">
            {/* Row 1: 3 equal cards — 3D mouse tilt */}
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
              {toolsRow1.map((tool) => (
                <TiltCard key={tool.title} className="h-full" intensity={7}>
                  <ToolCard tool={tool} large />
                </TiltCard>
              ))}
            </div>
            {/* Row 2: 4 equal cards — 3D mouse tilt */}
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              {toolsRow2.map((tool) => (
                <TiltCard key={tool.title} className="h-full" intensity={9}>
                  <ToolCard tool={tool} />
                </TiltCard>
              ))}
            </div>
          </div>
        </HomeSection>

        {/* ── The loop ───────────────────────────────────────────────── */}
        <HomeSection id="loop" pattern="p024" className="pt-[120px]">
          <div className="grid md:grid-cols-2 gap-8 md:gap-12 items-center">
            <div className="flex flex-col gap-4">
              <Heading className="text-left">
                <TextHighlight className="pr-1.5">Extract</TextHighlight>
                → query → act.<br />
                <span className="text-text-tertiary font-analog">Repeat.</span>
              </Heading>
              <Text className="text-left">
                Every data problem is a loop. Pull data from a source, shape it with SQL, push results to a destination, and run again. pm makes this loop a single binary invocation.
              </Text>
              <Text size="s" className="text-left">
                Real-world example: extract open GitHub issues every hour, run a SQL query to find stale ones, create Jira tickets for them. Fully local. Fully auditable. Zero cost.
              </Text>
              <div className="flex gap-3 flex-wrap">
                <Link
                  href="/docs/quickstart"
                  className="inline-flex items-center gap-1.5 text-[14px] text-text-secondary border-b border-line-structure hover:border-text-primary transition-colors pb-0.5"
                >
                  60-second quickstart <ArrowRight size={12} />
                </Link>
              </div>
            </div>

            <CodeTerminal />
          </div>
        </HomeSection>

        {/* ── Open source & local-first ──────────────────────────────── */}
        <HomeSection id="open-source" pattern="p025" className="pt-[120px]">
          <CornerBox withStripes className="px-8 py-10 md:py-14 relative overflow-hidden">
            {/* Accent glow in top-right corner */}
            <span
              className="absolute -top-8 -right-8 w-56 h-56 rounded-full pointer-events-none"
              style={{
                background: 'radial-gradient(circle, rgba(52,211,153,0.32) 0%, transparent 70%)',
                filter: 'blur(32px)',
              }}
              aria-hidden
            />
            <div className="flex flex-col md:flex-row gap-8 md:gap-16 items-start md:items-center relative">
              <div className="flex flex-col gap-3 flex-1">
                <Heading className="text-left">
                  Built in the open.<br />
                  <TextHighlight>No vendor lock-in.</TextHighlight>
                </Heading>
                <Text className="text-left">
                  pm is MIT-licensed, written in pure Go, and trivially cross-compiled. No CGO, no Docker, no 8 GB of services. Install with a single{' '}
                  <code className="font-mono text-[13px] bg-surface-2 rounded px-1">go install</code>{' '}
                  and run anywhere.
                </Text>
              </div>
            </div>
          </CornerBox>
        </HomeSection>

        {/* ── Connector count ────────────────────────────────────────── */}
        <HomeSection id="connectors" pattern="p026" className="pt-[120px]">
          <div className="flex flex-col gap-4 mb-8">
            <Heading>
              {CONNECTOR_CATALOG_COUNT} connectors,{' '}
              <TextHighlight>written</TextHighlight>{' '}
              in Go.
            </Heading>
            <Text className="max-w-[46ch]">
              Every connector is native Go on a shared HTTP and database toolkit. No Python, no Docker, no connector images to download. Install once, run anywhere.
            </Text>
          </div>
          <ConnectorMarquee />
        </HomeSection>

        {/* ── Why pm? FAQ ────────────────────────────────────────────── */}
        <HomeSection id="why" pattern="p013" className="pt-[120px]">
          <div className="flex flex-col gap-4 mb-8">
            <Heading>Why pm?</Heading>
            <Text className="max-w-[46ch]">Honest answers to honest questions.</Text>
          </div>
          <FaqAccordion />
        </HomeSection>

        {/* ── CTA ──────────────────────────────────────────────────── */}
        <HomeSection id="get-started" pattern="p027" className="pt-[120px]">
          <CornerBox className="flex flex-col items-center gap-6 px-8 py-14 text-center relative overflow-hidden">
            {/* Centered warm glow */}
            <span
              className="absolute top-1/2 left-1/2 w-[480px] h-[300px] rounded-full pointer-events-none"
              style={{
                background: 'radial-gradient(ellipse, rgba(52,211,153,0.24) 0%, transparent 70%)',
                filter: 'blur(56px)',
                transform: 'translate(-50%,-50%)',
              }}
              aria-hidden
            />

            <Heading className="relative">
              <TextHighlight highlightClassName="mix-blend-multiply">Start</TextHighlight>{' '}
              in 60 seconds.
            </Heading>
            <Text className="max-w-[38ch] relative">
              One binary. Your machine. Every connector. AI agents welcome.
            </Text>
            <div className="flex flex-wrap gap-3 justify-center relative">
              <Link
                href="/docs/quickstart"
                className="btn-shine inline-flex items-center gap-2 border border-emerald-900 bg-emerald-800 px-5 py-2.5 text-[14px] font-medium text-white transition-opacity hover:opacity-90"
              >
                Read the docs <ArrowRight size={14} />
              </Link>
              <a
                href="https://github.com/karthik-sivadas/polymetrics-cli"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 rounded-sm border border-line-structure bg-surface-bg px-5 py-2.5 text-[14px] font-medium text-text-secondary transition-colors hover:bg-surface-1"
              >
                <Github size={14} /> Star on GitHub
              </a>
            </div>
          </CornerBox>
        </HomeSection>

        {/* ── Footer — inside the center grid, between the sidebars ──── */}
        <SiteFooter />

      </main>

      {/* Right aside */}
      <HomeAside />

    </div>
  );
}
