'use client';

import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Bot,
  Cable,
  ChevronDown,
  Database,
  ExternalLink,
  Menu,
  Newspaper,
  Repeat2,
  Rocket,
  X,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { DocsSearch } from '@/components/docs/docs-search';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { PmLogoMark } from '@/components/brand/pm-logo-mark';
import { CONNECTOR_CATALOG_COUNT } from '@/lib/connectors.generated';

/* ─── Product dropdown items ──────────────────────────────────────────── */
const PRODUCT_ITEMS = [
  { label: 'Connector Catalog', desc: `${CONNECTOR_CATALOG_COUNT} connector pages`, href: '/docs/connectors', icon: Cable, shortcutKey: '1' },
  { label: 'SQL Queries',       desc: 'Local SQL over warehouse data', href: '/docs/query', icon: Database, shortcutKey: '2' },
  { label: 'Reverse ETL',       desc: 'Plan, preview, approve, then write back', href: '/docs/reverse-etl', icon: Repeat2, shortcutKey: '3' },
  { label: 'Agent Mode',        desc: 'JSON contracts and approval gates', href: '/docs/agent-guide', icon: Bot, shortcutKey: '4' },
  { label: 'Quickstart',        desc: 'Install pm and run the loop in 60 seconds', href: '/docs/quickstart', icon: Rocket, shortcutKey: '5' },
  { label: 'Blog',              desc: 'Engineering notes on the data loop', href: '/blog', icon: Newspaper, shortcutKey: '6' },
];

/* ─── Hover corner brackets (from Langfuse corner-box.tsx) ──────────── */
function HoverCorners() {
  return <span aria-hidden className="corner-box-hover-child" />;
}

/* ─── Keyboard shortcut badge ────────────────────────────────────────── */
function Kbd({ k, variant }: { k: string; variant: 'primary' | 'secondary' }) {
  const cls =
    variant === 'primary'
      ? 'border border-white/30 bg-white/20 text-white/85'
      : 'border border-[rgba(15,61,46,0.20)] bg-[rgba(15,61,46,0.08)] text-text-tertiary';
  return (
    <kbd
      aria-hidden
      className={`flex justify-center items-center not-italic shrink-0 w-[20px] h-[20px] font-mono text-[11px] font-[450] leading-none tracking-normal ${cls}`}
    >
      {k.toUpperCase()}
    </kbd>
  );
}

function shouldIgnoreShortcutTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
}

function useShortcutAction(kbdKey: string | undefined, action: () => void) {
  const actionRef = useRef(action);

  useEffect(() => {
    actionRef.current = action;
  }, [action]);

  useEffect(() => {
    if (!kbdKey) return;
    const key = kbdKey.toLowerCase();

    function handler(e: KeyboardEvent) {
      if (e.repeat || e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key.toLowerCase() !== key) return;
      if (shouldIgnoreShortcutTarget(e.target)) return;
      e.preventDefault();
      actionRef.current();
    }

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [kbdKey]);
}

function useKeyboardShortcut<T extends HTMLElement>(kbdKey?: string) {
  const ref = useRef<T>(null);

  useEffect(() => {
    if (!kbdKey) return;
    const key = kbdKey.toLowerCase();

    function handler(e: KeyboardEvent) {
      if (e.repeat || e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key.toLowerCase() !== key) return;
      if (shouldIgnoreShortcutTarget(e.target)) return;
      e.preventDefault();
      ref.current?.click();
    }

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [kbdKey]);

  return ref;
}

/* ─── CTA button — exact Langfuse button.tsx structure ───────────────── */
const btnBase =
  'inline-flex w-full min-w-0 max-w-full items-center justify-start no-underline gap-[6px] overflow-hidden py-0.5 shadow-sm [box-shadow:0_4px_8px_0_rgba(0,0,0,0.05),0_4px_4px_0_rgba(0,0,0,0.03)] rounded-[2px] border transition-colors focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50 cursor-pointer font-square text-[12px] font-semibold leading-[150%] tracking-normal whitespace-nowrap';

const btnVariants = {
  primary:   'border-emerald-900 bg-emerald-800 text-white h-[26px] pl-[8px] pr-[3px]',
  secondary: 'border-line-structure bg-surface-bg text-text-secondary group-hover:border-line-cta h-[26px] pl-[8px] pr-[3px]',
};

const navChipCls =
  'inline-flex h-[26px] items-center justify-center gap-[6px] overflow-hidden border border-emerald-900 bg-emerald-800 py-0.5 pl-[8px] pr-[3px] font-square text-[12px] font-semibold leading-[150%] tracking-normal text-white shadow-sm [box-shadow:0_4px_8px_0_rgba(0,0,0,0.05),0_4px_4px_0_rgba(0,0,0,0.03)] transition-opacity group-hover:opacity-90 whitespace-nowrap';

function NavBtn({
  href,
  variant,
  kbdKey,
  children,
  external,
}: {
  href: string;
  variant: 'primary' | 'secondary';
  kbdKey?: string;
  children: React.ReactNode;
  external?: boolean;
}) {
  const btnRef = useKeyboardShortcut<HTMLAnchorElement>(kbdKey);

  const innerContent = (
    <>
      <span className="flex items-center min-w-0 truncate">{children}</span>
      {kbdKey && <Kbd k={kbdKey} variant={variant} />}
    </>
  );

  const anchorCls = `${btnBase} ${btnVariants[variant]}`;

  return (
    /* button-wrapper triggers .corner-box-hover-child::before on hover */
    <div className="button-wrapper relative flex items-center p-1 group max-h-[34px] cursor-pointer">
      <HoverCorners />
      {external ? (
        <a ref={btnRef} href={href} target="_blank" rel="noreferrer" className={anchorCls}>
          {innerContent}
        </a>
      ) : (
        <Link ref={btnRef} href={href} className={anchorCls}>
          {innerContent}
        </Link>
      )}
    </div>
  );
}

/* ─── Product dropdown panel item ────────────────────────────────────── */
function DropdownItem({ href, icon: Icon, label, desc, shortcutKey }: {
  href: string; icon: LucideIcon; label: string; desc: string; shortcutKey: string;
}) {
  return (
    <DropdownMenuItem
      asChild
      className="link-box cursor-pointer border border-transparent bg-transparent p-0 focus:border-line-cta focus:bg-surface-bg"
    >
      <Link
        href={href}
        className="group/link relative grid grid-cols-[2rem_minmax(0,1fr)_1.25rem] items-start gap-3 px-2.5 py-2 no-underline transition-colors hover:bg-surface-bg"
      >
        <HoverCorners />
        <span className="relative z-[1] mt-0.5 flex size-8 shrink-0 items-center justify-center border border-line-structure bg-surface-bg text-line-cta transition-colors group-hover/link:border-line-cta">
          <Icon className="size-4" aria-hidden="true" />
        </span>
        <div className="relative z-[1] min-w-0">
          <div className="font-square text-[13px] font-semibold leading-[1.2] text-text-secondary transition-colors group-hover/link:text-text-primary">
            {label}
          </div>
          <div className="mt-1 text-[12px] leading-snug text-text-tertiary">{desc}</div>
        </div>
        <Kbd k={shortcutKey} variant="secondary" />
      </Link>
    </DropdownMenuItem>
  );
}

/* ─── Product menu — shadcn primitive, Boxy surface ───────────────────── */
function ProductDropdown({ kbdKey }: { kbdKey?: string }) {
  const [open, setOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const router = useRouter();

  useShortcutAction(kbdKey, () => {
    setOpen(true);
    window.requestAnimationFrame(() => triggerRef.current?.focus());
  });

  useEffect(() => {
    if (!open) return;

    function onKeyDown(event: KeyboardEvent) {
      if (event.repeat || event.metaKey || event.ctrlKey || event.altKey) return;
      if (shouldIgnoreShortcutTarget(event.target)) return;

      const key = event.key.toLowerCase();
      const item = PRODUCT_ITEMS.find((candidate) => candidate.shortcutKey.toLowerCase() === key);
      const targetHref = item?.href ?? (key === 'd' ? '/docs' : '');
      if (!targetHref) return;

      event.preventDefault();
      event.stopPropagation();
      setOpen(false);
      router.push(targetHref);
    }

    window.addEventListener('keydown', onKeyDown, true);
    return () => window.removeEventListener('keydown', onKeyDown, true);
  }, [open, router]);

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <button
          ref={triggerRef}
          type="button"
          className="button-wrapper relative flex max-h-[34px] cursor-pointer items-center p-1 group focus:outline-none"
          aria-label="Open product navigation"
          aria-expanded={open}
        >
          <HoverCorners />
          <span className={navChipCls}>
            Product
            {kbdKey && <Kbd k={kbdKey} variant="primary" />}
            <ChevronDown className={`size-3.5 opacity-60 transition-transform duration-150 ease-in-out ${open ? 'rotate-180' : ''}`} />
          </span>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="center"
        sideOffset={18}
        className="w-[318px] border border-line-structure bg-surface-1 p-2 text-text-primary shadow-[0_18px_60px_rgba(12,31,23,0.16)] ring-0"
      >
        <DropdownMenuLabel className="px-2 pb-2 pt-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
          Product surfaces
        </DropdownMenuLabel>
        <DropdownMenuGroup className="flex flex-col gap-1">
          {PRODUCT_ITEMS.map(item => (
            <DropdownItem key={item.href} {...item} />
          ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator className="mx-0 my-2 bg-line-structure" />
        <DropdownMenuItem asChild className="cursor-pointer p-0 focus:bg-surface-bg">
          <Link
            href="/docs"
            className="flex items-center justify-between px-2.5 py-2 font-square text-[11px] font-semibold uppercase tracking-normal text-text-secondary hover:text-text-primary"
          >
            Browse all documentation
            <Kbd k="D" variant="secondary" />
            <ExternalLink className="size-3.5" aria-hidden="true" />
          </Link>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/* ─── Plain nav link ──────────────────────────────────────────────────── */
function NavLink({
  href,
  children,
  external,
  kbdKey,
}: {
  href: string;
  children: React.ReactNode;
  external?: boolean;
  kbdKey?: string;
}) {
  const linkRef = useKeyboardShortcut<HTMLAnchorElement>(kbdKey);
  const cls = 'button-wrapper relative flex max-h-[34px] cursor-pointer items-center p-1 group no-underline';
  const content = (
    <>
      <HoverCorners />
      <span className={navChipCls}>
        {children}
        {kbdKey && <Kbd k={kbdKey} variant="primary" />}
      </span>
    </>
  );
  return external
    ? <a ref={linkRef} href={href} target="_blank" rel="noreferrer" className={cls}>{content}</a>
    : <Link ref={linkRef} href={href} className={cls}>{content}</Link>;
}

/* ─── Mobile menu ─────────────────────────────────────────────────────── */
function MobileMenu({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="top"
        showCloseButton={false}
        className="max-h-[calc(100dvh-1rem)] overflow-y-auto border-b border-line-cta bg-surface-bg p-0 shadow-[0_20px_60px_rgba(12,31,23,0.22)]"
      >
        <SheetHeader className="border-b border-line-structure bg-surface-1 px-5 py-4 text-left">
          <div className="flex items-start justify-between gap-3">
            <div>
              <SheetTitle className="font-square text-[13px] uppercase tracking-wider text-text-secondary">
                PM navigation
              </SheetTitle>
              <SheetDescription className="mt-1 text-[12px] text-text-tertiary">
                Product surfaces, documentation, and source repository.
              </SheetDescription>
            </div>
            <SheetClose asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-8 border border-line-structure bg-surface-bg text-text-tertiary hover:bg-surface-2 hover:text-text-primary"
                aria-label="Close navigation menu"
              >
                <X className="size-4" aria-hidden="true" />
              </Button>
            </SheetClose>
          </div>
        </SheetHeader>

        <div className="px-5 py-4 flex flex-col gap-1">
        <p className="mb-2 text-[10px] font-semibold uppercase tracking-widest text-text-disabled">Product</p>
        {PRODUCT_ITEMS.map(item => (
          <SheetClose key={item.href} asChild>
            <Link
              href={item.href}
              className="flex items-start gap-3 border border-transparent px-2 py-2.5 font-square text-[14px] font-semibold text-text-secondary transition-colors hover:border-line-cta hover:bg-surface-1 hover:text-text-primary"
            >
              <span className="flex size-8 shrink-0 items-center justify-center border border-line-structure bg-surface-1 text-line-cta">
                <item.icon className="size-4" aria-hidden="true" />
              </span>
              <span className="min-w-0">
                <span className="block">{item.label}</span>
                <span className="mt-0.5 block text-[12px] font-normal leading-snug text-text-tertiary">{item.desc}</span>
              </span>
            </Link>
          </SheetClose>
        ))}
        <div className="mt-3 pt-3 border-t border-line-structure flex flex-col gap-0.5">
          <SheetClose asChild>
            <Link href="/docs" className="py-2 font-square text-[14px] font-semibold text-text-secondary hover:text-text-primary transition-colors">Docs</Link>
          </SheetClose>
          <SheetClose asChild>
            <Link href="/blog" className="py-2 font-square text-[14px] font-semibold text-text-secondary hover:text-text-primary transition-colors">Blog</Link>
          </SheetClose>
          <SheetClose asChild>
            <Link href="/changelog" className="py-2 font-square text-[14px] font-semibold text-text-secondary hover:text-text-primary transition-colors">Changelog</Link>
          </SheetClose>
          <a href="https://github.com/polymetrics-ai/cli" target="_blank" rel="noreferrer" className="py-2 font-square text-[14px] font-semibold text-text-secondary hover:text-text-primary transition-colors">GitHub</a>
        </div>
        <div className="mt-4 flex flex-col gap-2">
          <SheetClose asChild>
            <Link
              href="/docs"
              className="flex items-center justify-center border border-line-cta bg-line-cta px-4 py-2.5 font-square text-[13px] font-semibold text-surface-bg"
            >
              Get Started
            </Link>
          </SheetClose>
        </div>
      </div>
      </SheetContent>
    </Sheet>
  );
}

/* ─── Main navbar — three-panel layout exactly like Langfuse ──────────── */
/*
  Desktop: all three panels are flex-1; left/right capped at 256px.
  Mobile:  side panels shrink to content; spacer fills the gap.
  The 1px bg-line-structure showing through adjacent panel padding = divider.
*/
const outerPanel = 'flex items-stretch bg-line-structure p-px py-0';
const innerPanel = 'flex items-center w-full bg-surface-1 pl-3 pr-2.5';

export function SiteNavbar() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 h-[60px] bg-surface-1 border-b border-line-structure">
      <nav className="flex h-full w-full max-w-[95rem] mx-auto">

        {/* ── LEFT BOX: logo + "by Polymetrics" ── */}
        <div className={`${outerPanel} flex-shrink-0 lg:w-[256px] lg:pr-px`}>
          <div className={`${innerPanel} lg:rounded-sm lg:rounded-r-none`}>
            <Link
              href="/"
              className="flex items-center gap-2 group/logo shrink-0"
              aria-label="PM homepage"
            >
              <PmLogoMark className="h-[26px] w-[26px] shrink-0 select-none" />
              <span className="navbar-by-tag font-square text-[11px] font-light uppercase tracking-[0.14em] leading-none text-text-tertiary/70 whitespace-nowrap hover:text-text-tertiary transition-colors">
                command line interface
              </span>
            </Link>
          </div>
        </div>

        {/* ── CENTER: navigation links ── */}
        <div className={`${outerPanel} navbar-desktop-nav flex-1 px-0`}>
          <div className="flex flex-1 min-w-0 items-center justify-center gap-2 px-2 rounded-none bg-surface-1">
            <div className="flex min-w-0 shrink-0 items-center justify-center gap-0">
              <ProductDropdown kbdKey="P" />
              <NavLink href="/docs" kbdKey="D">Docs</NavLink>
              <NavLink href="/blog" kbdKey="B">Blog</NavLink>
              <NavLink href="/changelog" kbdKey="C">Changelog</NavLink>
              <NavLink href="https://github.com/polymetrics-ai/cli" kbdKey="H" external>GitHub</NavLink>
            </div>
            <div className="navbar-desktop-search min-w-0 shrink-0">
              <DocsSearch variant="navbar" />
            </div>
          </div>
        </div>

        {/* Flex spacer on mobile (hidden on desktop since all panels are flex-1) */}
        <div className="navbar-mobile-spacer" />

        {/* ── RIGHT BOX: CTA buttons ── */}
        <div className={`${outerPanel} flex-shrink-0 lg:w-[256px] lg:pl-px`}>
          <div className={`${innerPanel} justify-center rounded-none lg:rounded-sm lg:rounded-l-none gap-0`}>

            {/* Desktop buttons — primary first, like Langfuse */}
            <div className="navbar-desktop-cta items-center gap-0">
              <NavBtn href="/docs" variant="primary" kbdKey="L">
                Get Started
              </NavBtn>
              <NavBtn href="https://github.com/polymetrics-ai/cli" variant="secondary" kbdKey="G" external>
                Get Demo
              </NavBtn>
            </div>

            {/* Mobile hamburger */}
            <button
              className="navbar-mobile-btn p-2 text-text-secondary hover:text-text-primary transition-colors"
              onClick={() => setMobileOpen(o => !o)}
              aria-label="Toggle navigation menu"
            >
              {mobileOpen ? <X size={18} /> : <Menu size={18} />}
            </button>
          </div>
        </div>
      </nav>

      <MobileMenu open={mobileOpen} onOpenChange={setMobileOpen} />
    </header>
  );
}
