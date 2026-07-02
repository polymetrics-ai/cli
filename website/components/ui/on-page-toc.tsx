'use client';

import { useCallback, useEffect, useId, useRef, useState } from 'react';
import type { CSSProperties, MouseEvent } from 'react';
import { Cable, Heart, MessageSquare } from 'lucide-react';
import {
  Sidebar,
  SidebarAccent,
  SidebarContent,
  SidebarFooter,
  SidebarInner,
} from '@/components/ui/sidebar';

export interface OnPageTocItem {
  id: string;
  label: string;
  level?: 2 | 3;
}

interface TocIndicator {
  path: string;
  width: number;
  height: number;
  pathLength: number;
  activeLength: number;
  activeEnd: {
    x: number;
    y: number;
  };
}

interface OnPageTocAsideProps {
  items: OnPageTocItem[];
  className: string;
}

interface TocPathPoint {
  id: string;
  top: number;
  bottom: number;
  x: number;
}

interface TocScrollPosition {
  activeId: string;
  activeIndex: number;
  ratio: number;
}

const TOC_NODE_SIZE = 8;
const TOC_NODE_HALF = TOC_NODE_SIZE / 2;
const TOC_NODE_MASK_PADDING = 3;
const TOC_NODE_MASK_SIZE = TOC_NODE_SIZE + (TOC_NODE_MASK_PADDING * 2);

const CREATOR_LINKS = [
  {
    label: 'GitHub',
    href: 'https://github.com/karthik-sivadas',
    iconSrc: '/connectors/icons/github.svg',
  },
  {
    label: 'LinkedIn',
    href: 'https://www.linkedin.com/in/karthiksivadas/',
    iconSrc: '/connectors/icons/linkedin.svg',
  },
  {
    label: 'X',
    href: 'https://x.com/karthik_sivadas',
    iconSrc: '/social/x.svg',
  },
] as const;

function findListItem(list: HTMLDivElement | null, id: string) {
  if (!list) return null;
  return Array.from(list.querySelectorAll<HTMLElement>('[data-id]')).find(
    (el) => el.dataset.id === id,
  ) ?? null;
}

function lineOffset(level: OnPageTocItem['level']) {
  return level === 3 ? 25 : 9;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function buildPath(points: TocPathPoint[]) {
  let d = '';

  points.forEach((point, index) => {
    if (index === 0) {
      d += `M ${point.x} ${point.top} L ${point.x} ${point.bottom}`;
      return;
    }

    const previous = points[index - 1];
    d += ` C ${previous.x} ${point.top - 7} ${point.x} ${previous.bottom + 7} ${point.x} ${point.top}`;
    d += ` L ${point.x} ${point.bottom}`;
  });

  return d;
}

function createSvgPath(path: string) {
  const el = document.createElementNS('http://www.w3.org/2000/svg', 'path');
  el.setAttribute('d', path);
  return el;
}

function getLengthAtY(el: SVGPathElement, targetY: number) {
  const totalLength = el.getTotalLength();
  const step = Math.max(totalLength / 320, 0.35);

  for (let length = 0; length <= totalLength; length += step) {
    const point = el.getPointAtLength(length);
    if (point.y >= targetY - 0.2) return length;
  }

  return totalLength;
}

function getScrollPosition(items: OnPageTocItem[]): TocScrollPosition {
  if (items.length === 0) {
    return { activeId: '', activeIndex: 0, ratio: 0 };
  }

  const offset = 132;
  const anchorY = window.scrollY + offset;
  const pageBottom = window.scrollY + window.innerHeight;
  const documentBottom = document.documentElement.scrollHeight;
  const starts = items.map((item) => {
    const section = document.getElementById(item.id);
    if (!section) return Number.POSITIVE_INFINITY;
    return section.getBoundingClientRect().top + window.scrollY;
  });
  let activeIndex = 0;

  for (let index = 0; index < starts.length; index += 1) {
    if (starts[index] <= anchorY) activeIndex = index;
  }

  if (pageBottom >= documentBottom - 4) {
    activeIndex = items.length - 1;
  }

  const currentStart = Number.isFinite(starts[activeIndex]) ? starts[activeIndex] : anchorY;
  const finalEnd = Math.max(currentStart + 1, documentBottom - window.innerHeight + offset);
  const nextStart = starts[activeIndex + 1] ?? finalEnd;
  const rangeEnd = Math.max(currentStart + 1, nextStart);
  const ratio = pageBottom >= documentBottom - 4
    ? 1
    : clamp((anchorY - currentStart) / (rangeEnd - currentStart), 0, 1);

  return {
    activeId: items[activeIndex]?.id ?? items[0]?.id ?? '',
    activeIndex,
    ratio,
  };
}

export function OnPageTocAside({ items, className }: OnPageTocAsideProps) {
  const rawId = useId();
  const idPrefix = rawId.replace(/:/g, '');
  const gradientId = `site-toc-active-${idPrefix}`;
  const maskId = `site-toc-mask-${idPrefix}`;
  const [observedActive, setObservedActive] = useState(items[0]?.id ?? '');
  const [manualActive, setManualActive] = useState('');
  const active = manualActive || observedActive || items[0]?.id || '';
  const listRef = useRef<HTMLDivElement>(null);
  const [indicator, setIndicator] = useState<TocIndicator | null>(null);

  useEffect(() => {
    if (!manualActive) return;
    const timer = window.setTimeout(() => setManualActive(''), 850);
    return () => window.clearTimeout(timer);
  }, [manualActive]);

  useEffect(() => {
    if (items.length === 0) {
      setObservedActive('');
      setIndicator(null);
      return;
    }

    let frame = 0;
    const update = () => {
      frame = 0;
      const scrollPosition = getScrollPosition(items);
      setObservedActive((previous) => (
        previous === scrollPosition.activeId ? previous : scrollPosition.activeId
      ));

      const list = listRef.current;
      if (!list) {
        setIndicator(null);
        return;
      }

      const points = items.flatMap((item) => {
        const el = findListItem(list, item.id);
        if (!el) return [];

        const styles = window.getComputedStyle(el);
        const top = el.offsetTop + Number.parseFloat(styles.paddingTop);
        const bottom = el.offsetTop + el.offsetHeight - Number.parseFloat(styles.paddingBottom);
        return [{
          id: item.id,
          top,
          bottom,
          x: lineOffset(item.level),
        }];
      });

      if (points.length === 0) {
        setIndicator(null);
        return;
      }

      const path = buildPath(points);
      const pathEl = createSvgPath(path);
      const pathLength = pathEl.getTotalLength();
      const endLengths = points.map((point) => getLengthAtY(pathEl, point.bottom));
      const measuredActiveIndex = points.findIndex((point) => point.id === scrollPosition.activeId);
      const activeIndex = measuredActiveIndex >= 0
        ? measuredActiveIndex
        : clamp(scrollPosition.activeIndex, 0, points.length - 1);
      const currentLength = endLengths[activeIndex] ?? 0;
      const nextLength = endLengths[activeIndex + 1] ?? pathLength;
      const activeLength = clamp(
        currentLength + ((nextLength - currentLength) * scrollPosition.ratio),
        0,
        pathLength,
      );
      const activeEnd = pathEl.getPointAtLength(activeLength);
      const maxX = Math.max(...points.map((point) => point.x));
      const maxY = Math.max(...points.map((point) => point.bottom));

      setIndicator({
        path,
        width: maxX + 12,
        height: maxY + 8,
        pathLength,
        activeLength,
        activeEnd: {
          x: activeEnd.x,
          y: activeEnd.y,
        },
      });
    };

    const schedule = () => {
      if (frame) return;
      frame = requestAnimationFrame(update);
    };

    schedule();
    const resizeObserver = new ResizeObserver(schedule);
    if (listRef.current) resizeObserver.observe(listRef.current);
    window.addEventListener('scroll', schedule, { passive: true });
    window.addEventListener('resize', schedule);

    return () => {
      if (frame) cancelAnimationFrame(frame);
      resizeObserver.disconnect();
      window.removeEventListener('scroll', schedule);
      window.removeEventListener('resize', schedule);
    };
  }, [items]);

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
    <Sidebar className={`${className} site-toc-panel`} data-site-toc>
      <SidebarInner className="site-toc-inner" aria-label="On this page">
        <SidebarContent className="site-toc-content">
          <div className="site-toc-header">
            <span className="inline-flex min-w-0 items-center gap-2">
              <Cable className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              <span>On this page</span>
            </span>
            <span className="site-toc-count" aria-label={`${items.length} sections`}>
              {items.length}
            </span>
          </div>

          {items.length > 0 ? (
            <div className="site-toc-list" ref={listRef}>
              {indicator ? (
                <svg
                  className="site-toc-svg"
                  width={indicator.width}
                  height={indicator.height}
                  viewBox={`0 0 ${indicator.width} ${indicator.height}`}
                  aria-hidden="true"
                  style={{
                    '--site-toc-active-length': `${indicator.activeLength}px`,
                    '--site-toc-path-length': `${indicator.pathLength + 1}px`,
                  } as CSSProperties}
                >
                  <defs>
                    <mask
                      id={maskId}
                      maskUnits="userSpaceOnUse"
                      x="0"
                      y="0"
                      width={indicator.width}
                      height={indicator.height}
                    >
                      <rect width={indicator.width} height={indicator.height} fill="white" />
                      <rect
                        x={indicator.activeEnd.x - TOC_NODE_HALF - TOC_NODE_MASK_PADDING}
                        y={indicator.activeEnd.y - TOC_NODE_HALF - TOC_NODE_MASK_PADDING}
                        width={TOC_NODE_MASK_SIZE}
                        height={TOC_NODE_MASK_SIZE}
                        fill="black"
                      />
                    </mask>
                    <linearGradient
                      id={gradientId}
                      x1="0"
                      y1="0"
                      x2="0"
                      y2={indicator.height}
                      gradientUnits="userSpaceOnUse"
                    >
                      <stop offset="0%" stopColor="var(--line-cta)" stopOpacity="0.42" />
                      <stop offset="68%" stopColor="var(--surface-cta-primary)" stopOpacity="0.85" />
                      <stop offset="100%" stopColor="var(--toc-terminal)" stopOpacity="1" />
                    </linearGradient>
                  </defs>
                  <path
                    className="site-toc-path site-toc-path-base"
                    d={indicator.path}
                    mask={`url(#${maskId})`}
                  />
                  <path className="site-toc-path site-toc-path-join" d={indicator.path} />
                  <path
                    className="site-toc-path site-toc-path-active"
                    d={indicator.path}
                    stroke={`url(#${gradientId})`}
                  />
                  <g
                    className="site-toc-node"
                    aria-hidden="true"
                    style={{
                      transform: `translate(${indicator.activeEnd.x}px, ${indicator.activeEnd.y}px)`,
                    }}
                  >
                    <rect
                      x={-TOC_NODE_HALF}
                      y={-TOC_NODE_HALF}
                      width={TOC_NODE_SIZE}
                      height={TOC_NODE_SIZE}
                    />
                  </g>
                </svg>
              ) : null}

              {items.map(({ id, label, level = 2 }) => (
                <a
                  key={id}
                  href={`#${id}`}
                  data-id={id}
                  data-level={level}
                  data-active={active === id ? 'true' : 'false'}
                  aria-current={active === id ? 'location' : undefined}
                  className="site-toc-link"
                  onClick={(event) => handleAnchorClick(event, id)}
                >
                  <span>{label}</span>
                </a>
              ))}
            </div>
          ) : (
            <p className="site-toc-empty">No sections on this page.</p>
          )}
        </SidebarContent>

        <SidebarFooter className="site-toc-footer">
          <a
            href="https://github.com/polymetrics-ai/cli/discussions"
            target="_blank"
            rel="noreferrer"
            className="site-toc-footer-link site-toc-footer-link-block"
          >
            <span className="flex items-center gap-2 text-[12px] font-medium text-text-secondary">
              <MessageSquare className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
              Join the discussion
            </span>
            <span className="mt-1 block text-[11px] leading-snug text-text-tertiary">
              Questions, ideas, and feedback.
            </span>
          </a>
          <div className="site-toc-footer-link site-toc-footer-link-block">
            <span className="flex items-center gap-2 text-[12px] font-medium text-text-secondary">
              <Heart className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
              Created with love by Karthik
            </span>
            <span className="site-toc-social-links" aria-label="Karthik social profiles">
              {CREATOR_LINKS.map((link) => (
                <a
                  key={link.label}
                  href={link.href}
                  target="_blank"
                  rel="noreferrer"
                  className="site-toc-social-link"
                >
                  <img
                    src={link.iconSrc}
                    alt=""
                    width={14}
                    height={14}
                    className="site-toc-social-icon"
                  />
                  <span>{link.label}</span>
                </a>
              ))}
            </span>
          </div>
          <SidebarAccent />
        </SidebarFooter>
      </SidebarInner>
    </Sidebar>
  );
}
