import type { MetadataRoute } from 'next';
import { BLOG_POSTS, blogUrl } from '@/lib/blog';
import { CONNECTOR_CATALOG } from '@/lib/connectors.catalog.generated';
import { DOCS_PAGES } from '@/lib/docs.generated';

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? 'https://cli.polymetrics.ai';
const DEFAULT_LAST_MODIFIED = new Date('2026-07-02');

function absoluteUrl(path: string): string {
  return new URL(path, SITE_URL).toString();
}

export default function sitemap(): MetadataRoute.Sitemap {
  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: absoluteUrl('/'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'weekly',
      priority: 1,
    },
    {
      url: absoluteUrl('/docs'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'weekly',
      priority: 0.9,
    },
    {
      url: absoluteUrl('/docs/connectors'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'weekly',
      priority: 0.9,
    },
    {
      url: absoluteUrl('/blog'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: absoluteUrl('/changelog'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'weekly',
      priority: 0.6,
    },
    {
      url: absoluteUrl('/patterns'),
      lastModified: DEFAULT_LAST_MODIFIED,
      changeFrequency: 'monthly',
      priority: 0.4,
    },
  ];

  const docsRoutes = DOCS_PAGES.map((page) => ({
    url: absoluteUrl(page.url),
    lastModified: DEFAULT_LAST_MODIFIED,
    changeFrequency: 'monthly' as const,
    priority: 0.7,
  }));

  const connectorRoutes = CONNECTOR_CATALOG.map((connector) => ({
    url: absoluteUrl(`/docs/connectors/${connector.slug}`),
    lastModified: DEFAULT_LAST_MODIFIED,
    changeFrequency: 'monthly' as const,
    priority: connector.featured ? 0.75 : 0.55,
  }));

  const blogRoutes = BLOG_POSTS.map((post) => ({
    url: absoluteUrl(blogUrl(post.slug)),
    lastModified: new Date(post.updatedAt),
    changeFrequency: 'monthly' as const,
    priority: 0.75,
  }));

  return [...staticRoutes, ...docsRoutes, ...connectorRoutes, ...blogRoutes];
}
