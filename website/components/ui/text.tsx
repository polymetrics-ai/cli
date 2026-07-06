import * as React from 'react';
import { cn } from '@/lib/utils';

export type TextSize = 'm' | 's' | 'xs';

const sizeClasses: Record<TextSize, string> = {
  m: 'text-center font-sans text-[15px] font-normal leading-[150%] tracking-[-0.075px] text-text-tertiary',
  s: 'text-center font-sans text-[14px] font-[430] leading-[150%] lg:leading-[120%] tracking-[-0.26px] text-text-tertiary',
  xs: 'text-center font-mono text-[10px] font-normal tracking-[-0.2px] text-text-tertiary underline',
};

export interface TextProps<T extends React.ElementType = 'p'> extends React.HTMLAttributes<HTMLElement> {
  size?: TextSize;
  as?: T;
}

const Text = React.forwardRef<HTMLElement, TextProps>(
  ({ className, size = 'm', as: Comp = 'p', ...props }, ref) => (
    // @ts-expect-error polymorphic ref
    <Comp ref={ref} className={cn(sizeClasses[size], className)} {...props} />
  ),
);
Text.displayName = 'Text';

export { Text };
