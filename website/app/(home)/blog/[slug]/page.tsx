import type { Metadata } from 'next';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { ArrowLeft, ArrowRight, Calendar, Clock } from 'lucide-react';
import { BLOG_POSTS, blogUrl, getBlogPost } from '@/lib/blog';
import { CornerBox } from '@/components/ui/corner-box';
import { ArticleBody } from '@/components/blog/article-body';

interface Props {
  params: Promise<{ slug: string }>;
}

export function generateStaticParams() {
  return BLOG_POSTS.map((post) => ({ slug: post.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = getBlogPost(slug);
  if (!post) return { title: 'Article not found' };

  return {
    title: post.title,
    description: post.description,
    alternates: {
      canonical: blogUrl(post.slug),
    },
    openGraph: {
      title: post.title,
      description: post.description,
      url: blogUrl(post.slug),
      type: 'article',
      publishedTime: post.publishedAt,
      modifiedTime: post.updatedAt,
      tags: post.tags,
    },
  };
}

function sectionId(index: number): string {
  return `section-${index + 1}`;
}

export default async function BlogPostPage({ params }: Props) {
  const { slug } = await params;
  const post = getBlogPost(slug);
  if (!post) notFound();

  const related = BLOG_POSTS.filter((item) => item.slug !== post.slug).slice(0, 2);
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: post.title,
    description: post.description,
    datePublished: post.publishedAt,
    dateModified: post.updatedAt,
    author: {
      '@type': 'Organization',
      name: 'Polymetrics AI',
      url: 'https://cli.polymetrics.ai',
    },
    publisher: {
      '@type': 'Organization',
      name: 'Polymetrics AI',
      url: 'https://cli.polymetrics.ai',
    },
    mainEntityOfPage: `https://cli.polymetrics.ai${blogUrl(post.slug)}`,
    keywords: post.tags.join(', '),
  };

  return (
    <main className="mx-auto w-full max-w-[95rem] px-6 py-16 md:py-20">
      <article className="mx-auto w-full max-w-[840px]">
          <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
          />

          <Link
            href="/blog"
            className="mb-8 inline-flex items-center gap-2 font-square text-[11px] font-semibold uppercase tracking-normal text-text-tertiary transition-colors hover:text-text-primary"
          >
            <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
            Blog
          </Link>

          <header id="article-overview" className="mb-10 border-b border-line-structure pb-10">
            <div className="mb-5 flex flex-wrap items-center gap-3 font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              <span className="border border-line-structure bg-surface-1 px-2 py-1">
                {post.category}
              </span>
              <span className="inline-flex items-center gap-1.5">
                <Calendar className="h-3 w-3 text-line-cta" aria-hidden="true" />
                {post.publishedAt}
              </span>
              <span className="inline-flex items-center gap-1.5">
                <Clock className="h-3 w-3 text-line-cta" aria-hidden="true" />
                {post.readingTime}
              </span>
            </div>
            <h1 className="max-w-[13ch] font-analog text-[48px] leading-[1] text-text-primary md:text-[76px]">
              {post.title}
            </h1>
            <p className="mt-6 max-w-[72ch] text-[17px] leading-relaxed text-text-tertiary">
              {post.description}
            </p>
            <div className="mt-6 flex flex-wrap gap-2">
              {post.tags.map((tag) => (
                <span
                  key={tag}
                  className="border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-text-tertiary"
                >
                  {tag}
                </span>
              ))}
            </div>
          </header>

          <ArticleBody slug={post.slug} sections={post.sections} summary={post.summary} />

          <footer id="read-next" className="mt-8 border-t border-line-structure pt-8">
            <h2 className="font-square text-[16px] font-semibold text-text-primary">
              Read next
            </h2>
            <div className="mt-4 grid gap-3 md:grid-cols-2">
              {related.map((item) => (
                <Link key={item.slug} href={blogUrl(item.slug)} className="block group">
                  <CornerBox
                    hoverStripes
                    className="flex h-full flex-col gap-3 p-4 transition-colors group-hover:bg-surface-1"
                  >
                    <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                      {item.category}
                    </span>
                    <span className="font-square text-[17px] font-semibold leading-[1.2] text-text-primary">
                      {item.title}
                    </span>
                    <span className="mt-auto inline-flex items-center gap-1.5 font-square text-[11px] font-semibold uppercase tracking-normal text-text-secondary">
                      Open
                      <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" aria-hidden="true" />
                    </span>
                  </CornerBox>
                </Link>
              ))}
            </div>
          </footer>
      </article>
    </main>
  );
}
