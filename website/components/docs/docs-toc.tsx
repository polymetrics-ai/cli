'use client';

import { useEffect, useRef, useState } from 'react';
import { usePathname } from 'next/navigation';
import { Github, MessageSquare, Navigation } from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  Sidebar,
  SidebarAccent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupHeader,
  SidebarGroupLabel,
  SidebarInner,
} from '@/components/ui/sidebar';

interface TocHeading {
  id: string;
  text: string;
  level: 2 | 3;
}

function useDOMHeadings(): TocHeading[] {
  const pathname = usePathname();
  const [headings, setHeadings] = useState<TocHeading[]>([]);

  useEffect(() => {
    const timer = setTimeout(() => {
      const article =
        document.querySelector('#nd-docs-layout article') ??
        document.querySelector('#nd-docs-layout main') ??
        document.querySelector('#nd-docs-layout');

      if (!article) return;

      const els = Array.from(
        article.querySelectorAll<HTMLHeadingElement>('h2[id], h3[id]'),
      );

      setHeadings(
        els.map((el) => ({
          id: el.id,
          text: el.textContent ?? '',
          level: el.tagName === 'H2' ? 2 : 3,
        })),
      );
    }, 80);

    return () => clearTimeout(timer);
  }, [pathname]);

  return headings;
}

export function DocsTocAside() {
  const pathname = usePathname();
  const headings = useDOMHeadings();
  const [active, setActive] = useState<string>('');
  const listRef = useRef<HTMLDivElement>(null);
  const [indicator, setIndicator] = useState<{ top: number; height: number } | null>(null);

  useEffect(() => {
    if (headings.length === 0) return;
    setActive(headings[0].id);

    const els = headings
      .map(({ id }) => document.getElementById(id))
      .filter(Boolean) as HTMLElement[];

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActive(entry.target.id);
            break;
          }
        }
      },
      { rootMargin: '-20% 0px -70% 0px', threshold: 0 },
    );

    els.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [headings, pathname]);

  useEffect(() => {
    if (!active) return;
    const el = listRef.current?.querySelector(`[data-id="${active}"]`) as HTMLElement | null;
    if (el) setIndicator({ top: el.offsetTop, height: el.offsetHeight });
  }, [active]);

  return (
    <Sidebar className="docs-toc-panel">
      <SidebarInner>
        <SidebarContent className="space-y-2 pt-3">
          <SidebarGroup>
            <SidebarGroupHeader>
              <SidebarGroupLabel>On this page</SidebarGroupLabel>
              <span className="inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
                <Navigation className="h-3 w-3" aria-hidden="true" />
                {headings.length}
              </span>
            </SidebarGroupHeader>
            <SidebarGroupContent>
              {headings.length > 0 ? (
                <div className="relative">
                  <div className="absolute bottom-0 left-0 top-0 w-px bg-line-structure" />
                  {indicator && (
                    <div
                      className="absolute left-0 w-0.5 bg-surface-cta-primary"
                      style={{
                        top: indicator.top,
                        height: indicator.height,
                        transition:
                          'top 150ms cubic-bezier(0.23,1,0.32,1), height 150ms cubic-bezier(0.23,1,0.32,1)',
                      }}
                    />
                  )}

                  <div ref={listRef} className="flex flex-col">
                    {headings.map(({ id, text, level }) => (
                      <a
                        key={id}
                        href={`#${id}`}
                        data-id={id}
                        className={cn(
                          'link-box group relative block border border-transparent py-1.5 text-[13px] leading-snug transition-colors duration-150 hover:border-line-structure hover:bg-surface-bg',
                          level === 2 ? 'ps-3 pe-2' : 'ps-5 pe-2 text-[12px]',
                          active === id
                            ? 'border-line-structure bg-surface-bg text-text-primary'
                            : 'text-text-tertiary hover:text-text-secondary',
                        )}
                      >
                        <span aria-hidden className="corner-box-hover-child" />
                        <span className="block truncate">{text}</span>
                      </a>
                    ))}
                  </div>
                </div>
              ) : (
                <p className="text-[12px] leading-relaxed text-text-disabled">
                  No sections on this page.
                </p>
              )}
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <a
            href="https://github.com/karthik-sivadas/polymetrics-cli/discussions"
            target="_blank"
            rel="noreferrer"
            className="link-box group relative block border-b border-line-structure px-4 py-3 transition-colors hover:bg-surface-bg"
          >
            <span aria-hidden className="corner-box-hover-child" />
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
            className="link-box group relative flex items-center gap-2 px-4 py-3 text-[12px] text-text-tertiary transition-colors hover:bg-surface-bg hover:text-text-secondary"
          >
            <span aria-hidden className="corner-box-hover-child" />
            <Github className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
            <span className="min-w-0 truncate">GitHub repository</span>
          </a>
          <SidebarAccent />
        </SidebarFooter>
      </SidebarInner>
    </Sidebar>
  );
}
