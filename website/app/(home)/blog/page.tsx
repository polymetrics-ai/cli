import type { Metadata } from 'next';
import Link from 'next/link';
import { ArrowRight, BookOpen, Clock, Rss } from 'lucide-react';
import { BLOG_POSTS, blogUrl } from '@/lib/blog';
import { HomeSidebar } from '@/components/home/home-sidebar';
import { OnPageTocAside } from '@/components/ui/on-page-toc';
import { CornerBox } from '@/components/ui/corner-box';

export const metadata: Metadata = {
  title: 'Blog',
  description:
    'Engineering essays from Polymetrics on local-first ETL, DuckDB SQL, reverse ETL, AI agent data workflows, and one-binary automation.',
  alternates: {
    canonical: '/blog',
  },
  openGraph: {
    title: 'Polymetrics Blog',
    description:
      'Local-first data automation, agent-native CLI design, connector engineering, and reverse ETL notes from Polymetrics.',
    url: '/blog',
    type: 'website',
  },
};

const featured = BLOG_POSTS[0];
const rest = BLOG_POSTS.slice(1);
const tocItems = [
  { id: 'blog-overview', label: 'Overview' },
  { id: 'featured-article', label: 'Featured' },
  { id: 'all-articles', label: 'All articles' },
];

export default function BlogPage() {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Blog',
    name: 'Polymetrics Blog',
    url: 'https://cli.polymetrics.ai/blog',
    description: metadata.description,
    blogPost: BLOG_POSTS.map((post) => ({
      '@type': 'BlogPosting',
      headline: post.title,
      description: post.description,
      datePublished: post.publishedAt,
      dateModified: post.updatedAt,
      url: `https://cli.polymetrics.ai${blogUrl(post.slug)}`,
    })),
  };

  return (
    <div className="flex mx-auto w-full max-w-[95rem] overflow-clip">
      <HomeSidebar />

      <main className="flex-1 min-w-0 pattern-bg overflow-hidden pb-8 xl:px-5 2xl:px-10">
        <div className="relative z-[1] mx-auto w-full px-4 py-16 sm:px-8 md:px-0 md:max-w-[680px] md:py-24 xl:max-w-[840px]">
          <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
          />

          <header id="blog-overview" className="mb-12 grid gap-6 border-b border-line-structure pb-10 lg:grid-cols-[minmax(0,1fr)_18rem]">
            <div className="min-w-0">
              <span className="inline-flex items-center gap-2 font-mono text-[12px] uppercase tracking-widest text-text-disabled">
                <Rss className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
                Polymetrics writing
              </span>
              <h1 className="mt-4 max-w-[12ch] font-analog text-[44px] leading-[1] text-text-primary md:text-[68px]">
                Notes for the data loop.
              </h1>
            </div>
            <div className="flex flex-col justify-end gap-4 text-[14px] leading-relaxed text-text-tertiary">
              <p>
                Research-backed product notes on local-first ETL, embedded DuckDB SQL,
                reverse ETL, connector design, and agent-safe automation.
              </p>
              <Link
                href="/docs"
                className="inline-flex w-fit items-center gap-2 border border-line-structure bg-surface-1 px-3 py-2 font-square text-[12px] font-semibold text-text-secondary transition-colors hover:border-line-cta hover:text-text-primary"
              >
                Documentation
                <ArrowRight className="h-3.5 w-3.5" aria-hidden="true" />
              </Link>
            </div>
          </header>

          <section id="featured-article" aria-label="Featured article" className="mb-12">
            <Link href={blogUrl(featured.slug)} className="block group">
              <CornerBox
                hoverStripes
                className="grid gap-8 p-6 transition-colors group-hover:bg-surface-1 md:grid-cols-[minmax(0,1fr)_14rem] md:p-8"
              >
                <div className="min-w-0">
                  <div className="mb-5 flex flex-wrap items-center gap-2 font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                    <span className="border border-line-structure bg-surface-1 px-2 py-1">
                      Featured
                    </span>
                    <span>{featured.category}</span>
                    <span>{featured.publishedAt}</span>
                  </div>
                  <h2 className="max-w-[14ch] font-analog text-[40px] leading-[1.02] text-text-primary md:text-[58px]">
                    {featured.title}
                  </h2>
                  <p className="mt-5 max-w-[68ch] text-[15px] leading-relaxed text-text-tertiary">
                    {featured.description}
                  </p>
                </div>

                <div className="flex flex-col justify-between gap-6 pt-1">
                  <p className="text-[13px] leading-relaxed text-text-tertiary">
                    {featured.summary}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {featured.tags.map((tag) => (
                      <span
                        key={tag}
                        className="border border-line-structure bg-surface-bg px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-text-tertiary"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                  <span className="inline-flex items-center gap-2 font-mono text-[11px] uppercase tracking-widest text-text-secondary">
                    Read article
                    <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" aria-hidden="true" />
                  </span>
                </div>
              </CornerBox>
            </Link>
          </section>

          <section id="all-articles" aria-label="All articles" className="grid gap-3 md:grid-cols-2">
            {rest.map((post) => (
              <Link key={post.slug} href={blogUrl(post.slug)} className="block h-full group">
                <CornerBox
                  hoverStripes
                  className="flex h-full flex-col gap-5 p-5 transition-colors group-hover:bg-surface-1"
                >
                  <div className="flex flex-wrap items-center gap-3 font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                    <span className="inline-flex items-center gap-1.5">
                      <BookOpen className="h-3 w-3 text-line-cta" aria-hidden="true" />
                      {post.category}
                    </span>
                    <span className="inline-flex items-center gap-1.5">
                      <Clock className="h-3 w-3 text-line-cta" aria-hidden="true" />
                      {post.readingTime}
                    </span>
                  </div>
                  <div className="min-w-0">
                    <h2 className="font-square text-[22px] font-semibold leading-[1.15] text-text-primary">
                      {post.title}
                    </h2>
                    <p className="mt-3 text-[14px] leading-relaxed text-text-tertiary">
                      {post.description}
                    </p>
                  </div>
                  <div className="mt-auto flex items-center justify-between gap-4 border-t border-line-structure pt-4">
                    <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                      {post.publishedAt}
                    </span>
                    <span className="inline-flex items-center gap-1.5 font-mono text-[11px] uppercase tracking-widest text-text-secondary">
                      Read
                      <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" aria-hidden="true" />
                    </span>
                  </div>
                </CornerBox>
              </Link>
            ))}
          </section>
        </div>
      </main>

      <OnPageTocAside className="home-aside-panel" items={tocItems} />
    </div>
  );
}
