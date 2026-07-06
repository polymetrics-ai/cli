import * as React from 'react';
import { cn } from '@/lib/utils';

// ── Badge variants ─────────────────────────────────────────────────────────
// Square (border-radius:0 already enforced globally), emerald palette.
// Spec §4: translucent emerald/amber/red fills; never solid dark emerald.

type BadgeVariant =
  | 'status-enabled'
  | 'status-planned'
  | 'support-certified'
  | 'support-community'
  | 'type-source'
  | 'type-destination'
  | 'capability'
  | 'release-generally_available'
  | 'release-ga'
  | 'release-beta'
  | 'release-alpha'
  | 'release-custom'
  | 'category'
  | 'default';

const variantClasses: Record<BadgeVariant, string> = {
  /* Spec §4: stable/enabled → rgba(52,211,153,0.12) bg, rgba(52,211,153,0.4) border, #0f3d2e text */
  'status-enabled':
    '[background:rgba(52,211,153,0.12)] text-line-cta [border-color:rgba(52,211,153,0.4)] border',
  'status-planned':
    'bg-surface-1 text-text-secondary border border-line-structure',
  'support-certified':
    '[background:rgba(52,211,153,0.12)] text-line-cta [border-color:rgba(52,211,153,0.4)] border',
  'support-community':
    'bg-surface-2 text-text-secondary border border-line-structure',
  'type-source':
    'bg-surface-1 text-text-secondary border border-line-structure',
  'type-destination':
    'bg-surface-2 text-text-secondary border border-line-structure',
  capability:
    '[background:rgba(52,211,153,0.10)] text-line-cta [border-color:rgba(52,211,153,0.32)] border',
  'release-generally_available':
    '[background:rgba(52,211,153,0.12)] text-line-cta [border-color:rgba(52,211,153,0.4)] border',
  'release-ga':
    '[background:rgba(52,211,153,0.12)] text-line-cta [border-color:rgba(52,211,153,0.4)] border',
  /* Spec §4: beta → rgba(251,191,36,0.12) bg, rgba(251,191,36,0.4) border, #78350f text */
  'release-beta':
    '[background:rgba(251,191,36,0.12)] text-[#78350f] [border-color:rgba(251,191,36,0.4)] border',
  /* Spec §4: alpha → rgba(248,113,113,0.12) bg, rgba(248,113,113,0.4) border, #7f1d1d text */
  'release-alpha':
    '[background:rgba(248,113,113,0.12)] text-[#7f1d1d] [border-color:rgba(248,113,113,0.4)] border',
  'release-custom':
    'bg-surface-1 text-text-secondary border border-line-structure',
  category:
    'bg-surface-1 text-text-secondary border border-line-structure',
  default:
    'bg-surface-1 text-text-secondary border border-line-structure',
};

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ variant = 'default', className, ...props }, ref) => (
    <span
      ref={ref}
      className={cn(
        'inline-flex items-center px-2 py-0.5 text-[11px] font-mono font-medium leading-none uppercase tracking-wider',
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  ),
);
Badge.displayName = 'Badge';

// ── Helpers to resolve variant from data ──────────────────────────────────

export function statusVariant(status: string): BadgeVariant {
  return status === 'enabled' || status === 'available' ? 'status-enabled' : 'status-planned';
}

export function supportVariant(level: string): BadgeVariant {
  return level === 'certified' ? 'support-certified' : 'support-community';
}

export function typeVariant(type: 'source' | 'destination'): BadgeVariant {
  return type === 'source' ? 'type-source' : 'type-destination';
}

export function releaseVariant(stage: string): BadgeVariant {
  const key = `release-${stage}` as BadgeVariant;
  return key in variantClasses ? key : 'default';
}

export { Badge };
