import { findNeighbour } from 'fumadocs-core/page-tree';
import type { Root } from 'fumadocs-core/page-tree';
import Link from 'next/link';
import { ArrowLeft, ArrowRight } from 'lucide-react';

interface PrevNextProps {
  url: string;
  tree: Root;
}

export function PrevNext({ url, tree }: PrevNextProps) {
  const { previous, next } = findNeighbour(tree, url);

  if (!previous && !next) return null;

  return (
    <nav
      aria-label="Previous and next pages"
      className="mt-12 grid gap-3 border-t border-line-structure pt-6 sm:grid-cols-2"
    >
      <div>
        {previous && (
          <Link
            href={previous.url}
            className="link-box group relative flex min-h-[88px] flex-col justify-between gap-3 border border-line-structure bg-surface-bg px-4 py-3 transition-colors hover:border-line-cta hover:bg-surface-1"
          >
            <span aria-hidden className="corner-box-hover-child" />
            <span className="inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-tertiary transition-colors group-hover:text-text-secondary">
              <ArrowLeft className="h-3 w-3 text-line-cta" aria-hidden="true" />
              Previous
            </span>
            <span className="text-[14px] font-medium leading-snug text-text-secondary transition-colors group-hover:text-text-primary">
              {previous.name}
            </span>
          </Link>
        )}
      </div>

      <div>
        {next && (
          <Link
            href={next.url}
            className="link-box group relative flex min-h-[88px] flex-col justify-between gap-3 border border-line-structure bg-surface-bg px-4 py-3 text-right transition-colors hover:border-line-cta hover:bg-surface-1"
          >
            <span aria-hidden className="corner-box-hover-child" />
            <span className="inline-flex items-center justify-end gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-tertiary transition-colors group-hover:text-text-secondary">
              Next
              <ArrowRight className="h-3 w-3 text-line-cta" aria-hidden="true" />
            </span>
            <span className="text-[14px] font-medium leading-snug text-text-secondary transition-colors group-hover:text-text-primary">
              {next.name}
            </span>
          </Link>
        )}
      </div>
    </nav>
  );
}
