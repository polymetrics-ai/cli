import type { ReactNode } from 'react';
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  Info,
  Lightbulb,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

type DocsCalloutType =
  | 'note'
  | 'info'
  | 'tip'
  | 'warn'
  | 'warning'
  | 'caution'
  | 'danger'
  | 'error'
  | 'success';

interface DocsCalloutProps {
  type?: DocsCalloutType;
  title?: ReactNode;
  children: ReactNode;
}

const calloutMeta: Record<
  DocsCalloutType,
  {
    icon: LucideIcon;
    title: string;
    variant: 'info' | 'warning' | 'danger' | 'default';
  }
> = {
  note: { icon: Info, title: 'Note', variant: 'info' },
  info: { icon: Info, title: 'Note', variant: 'info' },
  tip: { icon: Lightbulb, title: 'Tip', variant: 'info' },
  warn: { icon: AlertTriangle, title: 'Watch this', variant: 'warning' },
  warning: { icon: AlertTriangle, title: 'Watch this', variant: 'warning' },
  caution: { icon: AlertTriangle, title: 'Caution', variant: 'warning' },
  danger: { icon: AlertCircle, title: 'Danger', variant: 'danger' },
  error: { icon: AlertCircle, title: 'Error', variant: 'danger' },
  success: { icon: CheckCircle2, title: 'Success', variant: 'default' },
};

export function DocsCallout({ type = 'info', title, children }: DocsCalloutProps) {
  const meta = calloutMeta[type] ?? calloutMeta.info;
  const Icon = meta.icon;

  return (
    <Alert
      variant={meta.variant}
      className="not-prose my-5 grid-cols-[auto_minmax(0,1fr)] gap-x-3"
      data-docs-callout=""
    >
      <Icon aria-hidden="true" />
      <AlertTitle className="font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
        {title ?? meta.title}
      </AlertTitle>
      <AlertDescription className="max-w-none text-[14px] leading-relaxed text-text-secondary">
        {children}
      </AlertDescription>
    </Alert>
  );
}
