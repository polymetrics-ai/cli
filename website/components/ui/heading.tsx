import * as React from 'react';
import { cn } from '@/lib/utils';

export type HeadingSize = 'big' | 'large' | 'normal' | 'small';
export type HeadingLevel = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

export interface HeadingProps extends React.HTMLAttributes<HTMLHeadingElement> {
  as?: HeadingLevel;
  size?: HeadingSize;
}

const sizeClasses: Record<HeadingSize, string> = {
  big: 'max-[390px]:text-[35px] text-[44px] md:text-[54px] xl:text-[68px] leading-[105%] text-center',
  large: 'text-[32px] sm:text-[44px] leading-[100%] md:text-[50px]',
  normal: 'text-[32px] leading-[115%]',
  small: 'text-[15px] leading-[115%]',
};

const baseClasses =
  'text-text-primary font-square font-semibold not-italic tracking-normal';

const Heading = React.forwardRef<HTMLHeadingElement, HeadingProps>(
  ({ as: Tag = 'h2', size = 'normal', className, ...props }, ref) => (
    <Tag ref={ref} className={cn(baseClasses, sizeClasses[size], className)} {...props} />
  ),
);
Heading.displayName = 'Heading';

export { Heading };
