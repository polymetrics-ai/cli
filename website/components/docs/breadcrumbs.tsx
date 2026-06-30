import { getBreadcrumbItems } from 'fumadocs-core/breadcrumb';
import type { Root } from 'fumadocs-core/page-tree';
import Link from 'next/link';

interface BreadcrumbsProps {
  url: string;
  tree: Root;
  fallbackName?: string;
}

export function Breadcrumbs({ url, tree, fallbackName }: BreadcrumbsProps) {
  const items = getBreadcrumbItems(url, tree, { includeRoot: true });

  if (items.length === 0 && fallbackName) {
    return (
      <nav aria-label="Breadcrumb" className="flex flex-wrap items-center gap-1.5">
        <Link
          href="/docs"
          className="font-mono text-[11px] uppercase tracking-widest text-text-tertiary transition-colors hover:text-text-secondary"
        >
          Docs
        </Link>
        <span className="select-none text-[10px] text-text-tertiary">/</span>
        <span className="font-mono text-[11px] uppercase tracking-widest text-text-secondary">
          {fallbackName}
        </span>
      </nav>
    );
  }

  if (items.length === 0) return null;

  return (
    <nav aria-label="Breadcrumb" className="flex flex-wrap items-center gap-1.5">
      {items.map((item, i) => {
        const isLast = i === items.length - 1;
        return (
          <span key={i} className="flex items-center gap-1.5">
            {i > 0 && (
              <span className="text-[10px] text-text-tertiary select-none">/</span>
            )}
            {item.url && !isLast ? (
              <Link
                href={item.url}
                className="font-mono text-[11px] uppercase tracking-widest text-text-tertiary hover:text-text-secondary transition-colors"
              >
                {item.name}
              </Link>
            ) : (
              <span
                className={
                  isLast
                    ? 'font-mono text-[11px] uppercase tracking-widest text-text-secondary'
                    : 'font-mono text-[11px] uppercase tracking-widest text-text-tertiary'
                }
              >
                {item.name}
              </span>
            )}
          </span>
        );
      })}
    </nav>
  );
}
