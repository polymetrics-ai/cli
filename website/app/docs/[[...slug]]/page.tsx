import { source } from '@/lib/source';
import { DocsBody } from 'fumadocs-ui/layouts/docs/page';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import Link from 'next/link';
import { Bot, FileCode2, Route } from 'lucide-react';
import { getMDXComponents } from '@/components/mdx';
import { Breadcrumbs } from '@/components/docs/breadcrumbs';
import { PrevNext } from '@/components/docs/prev-next';
import { CopyMarkdown } from '@/components/docs/copy-markdown';
import { documentationMetaFor } from '@/components/docs/doc-nav';

interface Props {
  params: Promise<{ slug?: string[] }>;
}

export default async function Page({ params }: Props) {
  const { slug } = await params;
  const page = source.getPage(slug);
  if (!page) notFound();

  const MDX = page.data.body;
  const tree = source.getPageTree();
  const pageMeta = documentationMetaFor(page.url, page.data.title);
  const PageIcon = pageMeta?.icon ?? Route;

  return (
    <article className="mx-auto w-full max-w-[840px] px-6 py-10">
      <header className="mb-8 border border-line-structure bg-surface-bg">
        <div className="flex flex-col gap-3 border-b border-line-structure bg-surface-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
          <Breadcrumbs url={page.url} tree={tree} fallbackName={page.data.title} />
          <div className="flex shrink-0 flex-wrap items-center gap-2">
            <CopyMarkdown slug={slug} label="Copy Markdown" />
            <Link
              href="/llms.txt"
              className="inline-flex h-8 items-center gap-1.5 border border-line-structure bg-surface-bg px-2.5 font-mono text-[10px] uppercase tracking-wider text-text-tertiary transition-colors hover:border-line-cta hover:bg-surface-2 hover:text-text-primary"
            >
              <Bot className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
              llms.txt
            </Link>
          </div>
        </div>

        <div className="px-4 py-6 sm:px-5 sm:py-7">
          <div className="mb-3 inline-flex items-center gap-2 border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
            <PageIcon className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            {pageMeta?.label ?? 'Documentation'}
          </div>
          <h1 className="font-square text-[30px] font-bold leading-[1.15] text-text-primary sm:text-[34px]">
            {page.data.title}
          </h1>
          {page.data.description ? (
            <p className="mt-3 max-w-[68ch] text-[15px] leading-relaxed text-text-tertiary">
              {page.data.description}
            </p>
          ) : null}
          {pageMeta?.description ? (
            <p className="mt-3 inline-flex border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
              {pageMeta.description}
            </p>
          ) : null}
        </div>

        <div className="grid border-t border-line-structure bg-surface-1 text-[12px] text-text-tertiary sm:grid-cols-3">
          <div className="flex min-w-0 items-center gap-2 border-b border-line-structure px-4 py-3 sm:border-b-0 sm:border-r">
            <Route className="h-3.5 w-3.5 shrink-0 text-line-cta" aria-hidden="true" />
            <span className="min-w-0 truncate">Human: read, copy, run</span>
          </div>
          <div className="flex min-w-0 items-center gap-2 border-b border-line-structure px-4 py-3 sm:border-b-0 sm:border-r">
            <Bot className="h-3.5 w-3.5 shrink-0 text-line-cta" aria-hidden="true" />
            <span className="min-w-0 truncate">Agent: JSON and exits</span>
          </div>
          <div className="flex min-w-0 items-center gap-2 px-4 py-3">
            <FileCode2 className="h-3.5 w-3.5 shrink-0 text-line-cta" aria-hidden="true" />
            <span className="min-w-0 truncate">Raw Markdown for tools</span>
          </div>
        </div>
      </header>

      <DocsBody>
        <MDX components={getMDXComponents()} />
      </DocsBody>

      <PrevNext url={page.url} tree={tree} />
    </article>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const page = source.getPage(slug);
  if (!page) notFound();

  return {
    title: page.data.title,
    description: page.data.description,
  };
}
