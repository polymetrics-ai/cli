'use client';

import * as React from 'react';
import { Slot } from 'radix-ui';
import { cn } from '@/lib/utils';

type ButtonVariant = 'default' | 'outline' | 'ghost' | 'quiet';
type ButtonSize = 'default' | 'sm' | 'lg' | 'icon';

const variantClasses: Record<ButtonVariant, string> = {
  default:
    'border-line-cta bg-line-cta text-surface-bg hover:[background:#123f31]',
  outline:
    'border-line-structure bg-surface-bg text-text-secondary hover:border-line-cta hover:bg-surface-1 hover:text-text-primary',
  ghost:
    'border-transparent bg-transparent text-text-tertiary hover:bg-surface-1 hover:text-text-primary',
  quiet:
    'border-line-structure bg-surface-1 text-text-secondary hover:bg-surface-2 hover:text-text-primary',
};

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  asChild?: boolean;
}

const sizeClasses: Record<ButtonSize, string> = {
  default: 'h-8 gap-1.5 px-2.5 text-[11px]',
  sm: 'h-7 gap-1 px-2 text-[10px]',
  lg: 'h-9 gap-2 px-3 text-[12px]',
  icon: 'h-8 w-8 p-0',
};

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'outline', size = 'default', type = 'button', asChild, ...props }, ref) => {
    const Comp = asChild ? Slot.Root : 'button';

    return (
      <Comp
        ref={ref}
        data-slot="button"
        type={asChild ? undefined : type}
        className={cn(
          'inline-flex shrink-0 items-center justify-center border font-square font-semibold uppercase tracking-normal transition-colors outline-none',
          'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-surface-cta-primary',
          'disabled:pointer-events-none disabled:opacity-50',
          '[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*=size-])]:size-3.5',
          sizeClasses[size],
          variantClasses[variant],
          className,
        )}
        {...props}
      />
    );
  },
);
Button.displayName = 'Button';

export { Button };
