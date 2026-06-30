import * as React from 'react';
import {
  Alert as ShadcnAlert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from '@/components/shadcn/ui/alert';
import { cn } from '@/lib/utils';

type AlertVariant = 'default' | 'info' | 'warning' | 'danger' | 'destructive';

const variantClasses: Record<AlertVariant, string> = {
  default: 'border-line-structure bg-surface-1 text-text-primary',
  info: 'border-line-structure bg-surface-1 text-text-primary [--alert-accent:var(--surface-cta-primary)]',
  warning: 'border-line-structure bg-surface-1 text-text-primary [--alert-accent:var(--warning)]',
  danger: 'border-line-structure bg-surface-1 text-text-primary [--alert-accent:var(--destructive)]',
  destructive: 'border-line-structure bg-surface-1 text-text-primary [--alert-accent:var(--destructive)]',
};

function Alert({
  className,
  variant = 'default',
  ...props
}: Omit<React.ComponentProps<typeof ShadcnAlert>, 'variant'> & {
  variant?: AlertVariant;
}) {
  return (
    <ShadcnAlert
      className={cn(
        'relative border-l-[3px] border-l-[var(--alert-accent,var(--line-structure))] px-4 py-3 shadow-[0_10px_24px_rgba(12,31,23,0.06)]',
        '[&>svg]:mt-0.5 [&>svg]:text-[var(--alert-accent,var(--line-cta))]',
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  );
}

export { Alert, AlertAction, AlertDescription, AlertTitle };
