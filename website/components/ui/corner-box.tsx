import * as React from 'react';
import { cn } from '@/lib/utils';

export interface CornerBoxProps extends React.HTMLAttributes<HTMLDivElement> {
  hoverStripes?: boolean;
  noBorder?: boolean;
  withStripes?: boolean;
}

const CornerBox = React.forwardRef<HTMLDivElement, CornerBoxProps>(
  ({ className, hoverStripes, noBorder, withStripes, style, children, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        'relative bg-surface-bg',
        noBorder ? 'corner-box-corners-flush' : 'border border-line-structure corner-box-corners',
        withStripes && 'with-stripes',
        hoverStripes && 'corner-box-hover-stripes transition-[background] duration-180 ease-out',
        className,
      )}
      style={style}
      {...props}
    >
      {children}
    </div>
  ),
);
CornerBox.displayName = 'CornerBox';

export { CornerBox };
