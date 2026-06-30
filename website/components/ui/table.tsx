import * as React from 'react';
import { cn } from '@/lib/utils';

// ── Plain-Tailwind table primitives ───────────────────────────────────────
// Square (no radius — enforced globally), emerald palette.
// header bg: surface-1, borders: line-structure.

const Table = React.forwardRef<HTMLTableElement, React.HTMLAttributes<HTMLTableElement>>(
  ({ className, ...props }, ref) => (
    <div className="w-full min-w-0 overflow-x-auto border border-line-structure">
      <table
        ref={ref}
        className={cn('w-full border-collapse text-sm', className)}
        {...props}
      />
    </div>
  ),
);
Table.displayName = 'Table';

const THead = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
  ({ className, ...props }, ref) => (
    <thead
      ref={ref}
      className={cn('bg-surface-1 border-b border-line-structure', className)}
      {...props}
    />
  ),
);
THead.displayName = 'THead';

const TBody = React.forwardRef<HTMLTableSectionElement, React.HTMLAttributes<HTMLTableSectionElement>>(
  ({ className, ...props }, ref) => (
    <tbody
      ref={ref}
      className={cn('divide-y divide-line-structure', className)}
      {...props}
    />
  ),
);
TBody.displayName = 'TBody';

const TR = React.forwardRef<HTMLTableRowElement, React.HTMLAttributes<HTMLTableRowElement>>(
  ({ className, ...props }, ref) => (
    <tr
      ref={ref}
      className={cn('even:bg-surface-1/40 hover:bg-surface-1 transition-colors', className)}
      {...props}
    />
  ),
);
TR.displayName = 'TR';

const TH = React.forwardRef<HTMLTableCellElement, React.ThHTMLAttributes<HTMLTableCellElement>>(
  ({ className, ...props }, ref) => (
    <th
      ref={ref}
      className={cn(
        'px-3 py-2 text-left text-[12px] font-square font-semibold uppercase tracking-[0.04em] text-text-tertiary',
        className,
      )}
      {...props}
    />
  ),
);
TH.displayName = 'TH';

const TD = React.forwardRef<HTMLTableCellElement, React.TdHTMLAttributes<HTMLTableCellElement>>(
  ({ className, ...props }, ref) => (
    <td
      ref={ref}
      className={cn('px-3 py-2.5 text-[13px] text-text-primary align-top', className)}
      {...props}
    />
  ),
);
TD.displayName = 'TD';

export { Table, THead, TBody, TR, TH, TD };
