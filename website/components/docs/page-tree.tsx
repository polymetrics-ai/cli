'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Boxes, ChevronRight, FileText, FolderOpen } from 'lucide-react';
import { useState } from 'react';
import type { Item, Folder, Separator, Node } from 'fumadocs-core/page-tree';
import { cn } from '@/lib/utils';
import { documentationMetaFor } from '@/components/docs/doc-nav';

/* ── Single page item ────────────────────────────────────────────────── */
function PageItem({ node }: { node: Item }) {
  const pathname = usePathname();
  const isActive = pathname === node.url;
  const nodeName = typeof node.name === 'string' ? node.name : '';
  const meta = documentationMetaFor(node.url, nodeName);
  const Icon = meta?.icon ?? FileText;

  const cls = cn(
    'link-box group relative flex items-center gap-2 border border-transparent px-2 py-1.5 text-[13px] leading-snug transition-colors duration-150 hover:border-line-structure hover:bg-surface-bg',
    isActive
      ? 'sidebar-active-pulse border-line-structure bg-surface-bg font-medium text-text-primary'
      : 'text-text-tertiary hover:text-text-secondary',
  );

  const inner = (
    <>
      {/* corner-bracket hover */}
      <span aria-hidden className="corner-box-hover-child" />
      {/* Active marker — left edge tick */}
      {isActive && (
        <span
          aria-hidden
          className="absolute left-0 top-[20%] bottom-[20%] w-0.5 bg-surface-cta-primary"
        />
      )}
      <span
        className={cn(
          'relative z-[1] flex h-5 w-5 shrink-0 items-center justify-center border transition-colors',
          isActive
            ? 'border-line-structure bg-surface-1 text-line-cta'
            : 'border-transparent text-text-disabled group-hover:border-line-structure group-hover:bg-surface-1 group-hover:text-line-cta',
        )}
        aria-hidden="true"
      >
        <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      </span>
      <span className="relative z-[1] min-w-0">
        <span className="block truncate">{node.name}</span>
        {isActive && meta?.description ? (
          <span className="mt-0.5 hidden truncate text-[10px] font-normal leading-snug text-text-disabled min-[1100px]:block">
            {meta.description}
          </span>
        ) : null}
      </span>
    </>
  );

  if (node.external) {
    return (
      <a href={node.url} target="_blank" rel="noreferrer" className={cls}>
        {inner}
      </a>
    );
  }
  return (
    <Link href={node.url} className={cls}>
      {inner}
    </Link>
  );
}

/* ── Folder item (collapsible section) ──────────────────────────────── */
function FolderItem({ node }: { node: Folder }) {
  const pathname = usePathname();
  // Open by default, or if any child url matches current path
  const hasActive = (ns: Node[]): boolean =>
    ns.some((n) =>
      n.type === 'page'
        ? n.url === pathname
        : n.type === 'folder'
        ? hasActive(n.children)
        : false,
    );
  const [open, setOpen] = useState(() => node.defaultOpen !== false || hasActive(node.children));

  return (
    <div className="min-w-0">
      {/* Folder header — acts like a section header button */}
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        data-state={open ? 'open' : 'closed'}
        className="link-box group relative flex w-full items-center justify-between gap-2 border border-transparent px-2 py-1.5 text-left transition-colors duration-150 hover:border-line-structure hover:bg-surface-bg data-[state=open]:border-line-structure data-[state=open]:bg-surface-bg"
      >
        <span aria-hidden className="corner-box-hover-child" />
        <span className="relative z-[1] flex min-w-0 items-center gap-2">
          <FolderOpen className="h-3.5 w-3.5 shrink-0 text-text-disabled transition-colors group-hover:text-line-cta group-data-[state=open]:text-line-cta" aria-hidden="true" />
          <span className="truncate font-square text-[10px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
            {node.name}
          </span>
        </span>
        <ChevronRight
          size={11}
          className={cn(
            'relative z-[1] shrink-0 text-text-disabled transition-transform duration-150',
            open && 'rotate-90',
          )}
        />
      </button>

      {/* Children */}
      <div className="sidebar-tree-content ml-2" data-state={open ? 'open' : 'closed'}>
        <div>
          <div className="flex flex-col border-l border-line-structure pl-2">
            {node.index && <PageItem node={node.index} />}
            <PageTreeRenderer nodes={node.children} />
          </div>
        </div>
      </div>
    </div>
  );
}

/* ── Separator ───────────────────────────────────────────────────────── */
function SeparatorItem({ node }: { node: Separator }) {
  if (!node.name) {
    return <div className="my-2 border-t border-line-structure" />;
  }
  return (
    <div className="my-2 flex items-center gap-2 px-2">
      <span className="flex h-5 w-5 shrink-0 items-center justify-center border border-line-structure bg-surface-1">
        <Boxes className="h-3 w-3 text-line-cta" aria-hidden="true" />
      </span>
      <span className="shrink-0 font-square text-[10px] font-semibold uppercase tracking-[0.08em] text-text-tertiary">
        {node.name}
      </span>
      <span className="h-px min-w-4 flex-1 bg-line-structure" aria-hidden="true" />
    </div>
  );
}

/* ── Root renderer — exported ─────────────────────────────────────────── */
export function PageTreeRenderer({ nodes }: { nodes: Node[] }) {
  return (
    <>
      {nodes.map((node, i) => {
        if (node.type === 'page') {
          return <PageItem key={node.url ?? i} node={node} />;
        }
        if (node.type === 'folder') {
          return <FolderItem key={node.$id ?? String(i)} node={node} />;
        }
        if (node.type === 'separator') {
          return <SeparatorItem key={node.$id ?? String(i)} node={node} />;
        }
        return null;
      })}
    </>
  );
}
