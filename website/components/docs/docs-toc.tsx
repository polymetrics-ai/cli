'use client';

import { useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';
import { OnPageTocAside } from '@/components/ui/on-page-toc';
import type { OnPageTocItem } from '@/components/ui/on-page-toc';

interface TocHeading extends OnPageTocItem {
  id: string;
  label: string;
  level: 2 | 3;
}

function readHeadings(): TocHeading[] {
  const article =
    document.querySelector('#nd-docs-layout article') ??
    document.querySelector('#nd-docs-layout main') ??
    document.querySelector('#nd-docs-layout');

  if (!article) return [];

  return Array.from(article.querySelectorAll<HTMLHeadingElement>('h2[id], h3[id]'))
    .map((el) => ({
      id: el.id,
      label: el.textContent?.trim() ?? '',
      level: el.tagName === 'H2' ? (2 as const) : (3 as const),
    }))
    .filter((heading) => heading.id && heading.label);
}

function sameHeadings(a: TocHeading[], b: TocHeading[]) {
  return (
    a.length === b.length &&
    a.every((heading, index) => {
      const next = b[index];
      return (
        next &&
        heading.id === next.id &&
        heading.label === next.label &&
        heading.level === next.level
      );
    })
  );
}

function useDOMHeadings(): TocHeading[] {
  const pathname = usePathname();
  const [headings, setHeadings] = useState<TocHeading[]>([]);

  useEffect(() => {
    let frame = 0;

    const update = () => {
      frame = 0;
      const next = readHeadings();
      setHeadings((current) => (sameHeadings(current, next) ? current : next));
    };

    const schedule = () => {
      if (frame) cancelAnimationFrame(frame);
      frame = requestAnimationFrame(update);
    };

    const timer = window.setTimeout(schedule, 80);
    const root = document.querySelector('#nd-docs-layout') ?? document.body;
    const observer = new MutationObserver(schedule);
    observer.observe(root, {
      childList: true,
      subtree: true,
      characterData: true,
    });

    return () => {
      window.clearTimeout(timer);
      if (frame) cancelAnimationFrame(frame);
      observer.disconnect();
    };
  }, [pathname]);

  return headings;
}

export function DocsTocAside() {
  const headings = useDOMHeadings();
  return <OnPageTocAside className="docs-toc-panel" items={headings} />;
}
