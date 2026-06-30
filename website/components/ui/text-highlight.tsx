'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

export interface TextHighlightProps extends React.HTMLAttributes<HTMLSpanElement> {
  highlightClassName?: string;
}

const TextHighlight = React.forwardRef<HTMLSpanElement, TextHighlightProps>(
  ({ className, highlightClassName, children, ...props }, ref) => (
    <span ref={ref} className={cn('inline-flex relative items-center', className)} {...props}>
      <span
        className={cn(
          'absolute inset-x-0 top-1/2 h-[0.76em] -translate-y-[52%]',
          'bg-emerald-300',
          highlightClassName,
        )}
        aria-hidden
      />
      <span className="relative">{children}</span>
    </span>
  ),
);
TextHighlight.displayName = 'TextHighlight';

export { TextHighlight };
