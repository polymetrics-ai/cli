'use client';

import { OnPageTocAside } from '@/components/ui/on-page-toc';
import type { OnPageTocItem } from '@/components/ui/on-page-toc';
import { BlogSidebarAuthCard } from '@/components/auth/blog-sidebar-auth-card';
import { githubDiscussionUrl } from '@/lib/discussions';

export function PageAside({
  items,
  discussionTitle,
}: {
  items: OnPageTocItem[];
  discussionTitle: string;
}) {
  const discussionHref = githubDiscussionUrl(discussionTitle);

  return (
    <OnPageTocAside
      className="page-aside-panel"
      items={items}
      discussionHref={discussionHref}
    >
      <BlogSidebarAuthCard />
    </OnPageTocAside>
  );
}
