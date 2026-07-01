'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MouseEvent } from 'react';
import { usePathname } from 'next/navigation';
import { AnchorProvider, TOCItem, useActiveAnchor } from 'fumadocs-core/toc';
import type { TOCItemType } from 'fumadocs-core/toc';
import { Github, Menu, MessageSquare } from 'lucide-react';
import {
  Sidebar,
  SidebarAccent,
  SidebarContent,
  SidebarFooter,
  SidebarInner,
} from '@/components/ui/sidebar';

interface TocHeading {
  id: string;
  text: string;
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
      text: el.textContent?.trim() ?? '',
      level: el.tagName === 'H2' ? (2 as const) : (3 as const),
    }))
    .filter((heading) => heading.id && heading.text);
}

function sameHeadings(a: TocHeading[], b: TocHeading[]) {
  return (
    a.length === b.length &&
    a.every((heading, index) => {
      const next = b[index];
      return (
        next &&
        heading.id === next.id &&
        heading.text === next.text &&
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

function DocsTocList({ headings }: { headings: TocHeading[] }) {
  const observedActive = useActiveAnchor();
  const [manualActive, setManualActive] = useState('');
  const active = manualActive || observedActive || headings[0]?.id || '';
  const listRef = useRef<HTMLDivElement>(null);
  const [indicator, setIndicator] = useState<{ top: number; height: number } | null>(null);

  useEffect(() => {
    if (!manualActive) return;
    const timer = window.setTimeout(() => setManualActive(''), 700);
    return () => window.clearTimeout(timer);
  }, [manualActive]);

  useEffect(() => {
    if (!active) {
      setIndicator(null);
      return;
    }

    let frame = 0;
    const update = () => {
      const el = listRef.current?.querySelector<HTMLElement>(`[data-id="${active}"]`);
      setIndicator(el ? { top: el.offsetTop, height: el.offsetHeight } : null);
    };

    frame = requestAnimationFrame(update);
    window.addEventListener('resize', update);

    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('resize', update);
    };
  }, [active, headings]);

  const handleAnchorClick = useCallback((event: MouseEvent<HTMLAnchorElement>, id: string) => {
    const target = document.getElementById(id);
    if (!target) return;

    event.preventDefault();
    setManualActive(id);

    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    target.scrollIntoView({
      behavior: prefersReducedMotion ? 'auto' : 'smooth',
      block: 'start',
    });

    window.history.pushState(null, '', `${window.location.pathname}${window.location.search}#${id}`);
  }, []);

  return (
    <div className="docs-toc-list" ref={listRef}>
      <span className="docs-toc-rail" aria-hidden="true" />
      {indicator ? (
        <span
          className="docs-toc-indicator"
          aria-hidden="true"
          style={{
            height: indicator.height,
            transform: `translateY(${indicator.top}px)`,
          }}
        />
      ) : null}

      {headings.map(({ id, text, level }) => (
        <TOCItem
          key={id}
          href={`#${id}`}
          data-id={id}
          data-level={level}
          data-active={active === id ? 'true' : 'false'}
          aria-current={active === id ? 'location' : undefined}
          className="docs-toc-link"
          onClick={(event) => handleAnchorClick(event, id)}
        >
          <span>{text}</span>
        </TOCItem>
      ))}
    </div>
  );
}

export function DocsTocAside() {
  const headings = useDOMHeadings();
  const toc = useMemo<TOCItemType[]>(
    () =>
      headings.map((heading) => ({
        title: heading.text,
        url: `#${heading.id}`,
        depth: heading.level,
      })),
    [headings],
  );

  return (
    <Sidebar className="docs-toc-panel" data-docs-toc>
      <SidebarInner className="docs-toc-inner" aria-label="On this page">
        <SidebarContent className="docs-toc-content">
          <div className="docs-toc-header">
            <span className="inline-flex min-w-0 items-center gap-2">
              <Menu className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              <span>On this page</span>
            </span>
            <span className="docs-toc-count" aria-label={`${headings.length} sections`}>
              {headings.length}
            </span>
          </div>

          {headings.length > 0 ? (
            <AnchorProvider toc={toc} single>
              <DocsTocList headings={headings} />
            </AnchorProvider>
          ) : (
            <p className="docs-toc-empty">No sections on this page.</p>
          )}
        </SidebarContent>

        <SidebarFooter className="docs-toc-footer">
          <a
            href="https://github.com/karthik-sivadas/polymetrics-cli/discussions"
            target="_blank"
            rel="noreferrer"
            className="docs-toc-footer-link docs-toc-footer-link-block"
          >
            <span className="flex items-center gap-2 text-[12px] font-medium text-text-secondary">
              <MessageSquare className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
              Join the discussion
            </span>
            <span className="mt-1 block text-[11px] leading-snug text-text-tertiary">
              Questions, ideas, and feedback.
            </span>
          </a>
          <a
            href="https://github.com/karthik-sivadas/polymetrics-cli"
            target="_blank"
            rel="noreferrer"
            className="docs-toc-footer-link flex items-center gap-2"
          >
            <Github className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
            <span className="min-w-0 truncate">GitHub repository</span>
          </a>
          <SidebarAccent />
        </SidebarFooter>
      </SidebarInner>
    </Sidebar>
  );
}
