'use client';

import type { ReactNode } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';

function SidebarShortcutBadge({ value }: { value: string }) {
  return (
    <kbd className="absolute right-2 top-1/2 z-[2] inline-flex h-5 min-w-5 -translate-y-1/2 items-center justify-center border border-[rgba(15,61,46,0.20)] bg-[rgba(15,61,46,0.08)] px-1 font-mono text-[10px] font-medium leading-none text-text-tertiary transition-colors group-hover:border-line-cta group-hover:text-text-secondary">
      {value.toUpperCase()}
    </kbd>
  );
}

/**
 * Sidebar row that mirrors Langfuse's LinkBox: the `.link-box` class triggers
 * the corner-bracket hover (the "mathematical" L-marks, square edges), a subtle
 * surface tint on hover, and an optional dark tooltip below the row. No Radix —
 * a lightweight CSS group-hover tooltip keeps it dependency-free.
 */
export function SidebarLink({
  href,
  tooltip,
  className,
  children,
  active,
  kbdKey,
}: {
  href: string;
  tooltip?: ReactNode;
  className?: string;
  children: ReactNode;
  active?: boolean;
  kbdKey?: string;
}) {
  const pathname = usePathname();
  const external = href.startsWith('http');
  const isActive =
    active ??
    (!external &&
      (pathname === href ||
        (href !== '/' && href !== '/docs' && pathname.startsWith(`${href}/`))));

  const cls = cn(
    'link-box group relative block border border-transparent px-2 py-1.5 transition-colors hover:border-line-structure hover:bg-surface-bg',
    isActive && 'sidebar-active-pulse border-line-structure bg-surface-bg text-text-primary',
    kbdKey && 'pr-8',
    className,
  );
  const inner = (
    <>
      {/* corner-bracket hover (see .corner-box-hover-child in globals.css) */}
      <span aria-hidden className="corner-box-hover-child" />
      {isActive && (
        <span
          aria-hidden="true"
          className="absolute left-0 top-[20%] bottom-[20%] w-0.5 bg-surface-cta-primary"
        />
      )}
      {children}
      {kbdKey && <SidebarShortcutBadge value={kbdKey} />}
      {tooltip && (
        <span
          role="tooltip"
          className="pointer-events-none absolute left-1/2 top-full z-30 -translate-x-1/2 translate-y-[-4px] whitespace-nowrap border border-line-cta bg-text-primary px-2 py-1 text-[11px] font-medium text-surface-bg opacity-0 shadow-sm transition-opacity duration-150 group-hover:opacity-100"
        >
          {tooltip}
        </span>
      )}
    </>
  );
  return external ? (
    <a href={href} target="_blank" rel="noreferrer" className={cls}>
      {inner}
    </a>
  ) : (
    <Link href={href} className={cls}>
      {inner}
    </Link>
  );
}
