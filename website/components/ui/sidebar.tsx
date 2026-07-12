import * as React from 'react';
import { cn } from '@/lib/utils';

const Sidebar = React.forwardRef<HTMLElement, React.HTMLAttributes<HTMLElement>>(
  ({ className, style, ...props }, ref) => (
    <aside
      ref={ref}
      data-slot="sidebar"
      className={cn(
        'site-sidebar-panel sticky w-[256px] shrink-0 bg-line-structure p-px pt-0',
        className,
      )}
      style={{
        top: 'var(--fd-nav-height, 60px)',
        height: 'calc(100vh - var(--fd-nav-height, 60px))',
        ...style,
      }}
      {...props}
    />
  ),
);
Sidebar.displayName = 'Sidebar';

const SidebarInner = React.forwardRef<HTMLElement, React.HTMLAttributes<HTMLElement>>(
  ({ className, ...props }, ref) => (
    <nav
      ref={ref}
      data-slot="sidebar-inner"
      className={cn('flex min-h-0 min-w-0 flex-1 flex-col bg-surface-1', className)}
      {...props}
    />
  ),
);
SidebarInner.displayName = 'SidebarInner';

const SidebarHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="sidebar-header"
      className={cn('shrink-0 border-b border-line-structure p-2', className)}
      {...props}
    />
  ),
);
SidebarHeader.displayName = 'SidebarHeader';

const SidebarContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="sidebar-content"
      className={cn('min-h-0 flex-1 overflow-y-auto overflow-x-hidden p-2', className)}
      {...props}
    />
  ),
);
SidebarContent.displayName = 'SidebarContent';

const SidebarFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="sidebar-footer"
      className={cn('mt-auto w-full min-w-0 shrink-0 overflow-hidden border-t border-line-structure', className)}
      {...props}
    />
  ),
);
SidebarFooter.displayName = 'SidebarFooter';

const SidebarGroup = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <section
      ref={ref}
      data-slot="sidebar-group"
      className={cn('sidebar-reveal border border-line-structure bg-surface-bg', className)}
      {...props}
    />
  ),
);
SidebarGroup.displayName = 'SidebarGroup';

const SidebarGroupHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="sidebar-group-header"
      className={cn('flex items-center justify-between gap-2 border-b border-line-structure px-3 py-2', className)}
      {...props}
    />
  ),
);
SidebarGroupHeader.displayName = 'SidebarGroupHeader';

const SidebarGroupLabel = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p
      ref={ref}
      data-slot="sidebar-group-label"
      className={cn('font-square text-[11px] font-semibold uppercase tracking-wider text-text-secondary', className)}
      {...props}
    />
  ),
);
SidebarGroupLabel.displayName = 'SidebarGroupLabel';

const SidebarGroupContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      data-slot="sidebar-group-content"
      className={cn('p-2', className)}
      {...props}
    />
  ),
);
SidebarGroupContent.displayName = 'SidebarGroupContent';

const SidebarMenu = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} data-slot="sidebar-menu" className={cn('flex flex-col gap-1', className)} {...props} />
  ),
);
SidebarMenu.displayName = 'SidebarMenu';

const SidebarAccent = React.forwardRef<HTMLSpanElement, React.HTMLAttributes<HTMLSpanElement>>(
  ({ className, ...props }, ref) => (
    <span
      ref={ref}
      aria-hidden="true"
      data-slot="sidebar-accent"
      className={cn('block h-[2px] w-full bg-surface-cta-primary', className)}
      {...props}
    />
  ),
);
SidebarAccent.displayName = 'SidebarAccent';

const SidebarStat = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & {
  label: React.ReactNode;
  value: React.ReactNode;
}>(({ className, label, value, ...props }, ref) => (
  <div
    ref={ref}
    data-slot="sidebar-stat"
    className={cn(
      'grid grid-cols-[minmax(0,1fr)_minmax(0,max-content)] items-center gap-2 border border-line-structure bg-surface-1 px-2 py-1.5',
      className,
    )}
    {...props}
  >
    <span className="min-w-0 overflow-hidden text-ellipsis whitespace-nowrap text-[12px] text-text-tertiary">{label}</span>
    <span className="max-w-[6.5rem] overflow-hidden text-ellipsis whitespace-nowrap text-right font-mono text-[11px] font-medium tabular-nums text-text-secondary">{value}</span>
  </div>
));
SidebarStat.displayName = 'SidebarStat';

export {
  Sidebar,
  SidebarInner,
  SidebarHeader,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupHeader,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarAccent,
  SidebarStat,
};
