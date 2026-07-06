'use client';

import { useRouter } from 'next/navigation';
import { useEffect, useMemo, useState } from 'react';
import {
  ArrowRight,
  Bot,
  Cable,
  Command as CommandIcon,
  FileText,
  Search,
  X,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty';
import { Kbd, KbdGroup } from '@/components/ui/kbd';
import { Spinner } from '@/components/ui/spinner';
import { cn } from '@/lib/utils';

type SearchKind = 'doc' | 'connector';

type SearchResult = {
  title: string;
  description: string;
  url: string;
  kind: SearchKind;
  section: string;
  snippet: string;
  score: number;
};

type SearchResponse = {
  query: string;
  total: number;
  results: SearchResult[];
};

type DocsSearchProps = {
  variant?: 'sidebar' | 'navbar';
};

function ResultIcon({ kind }: { kind: SearchKind }) {
  const Icon = kind === 'connector' ? Cable : FileText;

  return (
    <span className="flex h-9 w-9 shrink-0 items-center justify-center border border-line-structure bg-surface-1">
      <Icon className="size-4 text-line-cta" aria-hidden="true" />
    </span>
  );
}

function SearchHint() {
  return (
    <KbdGroup className="shrink-0">
      <Kbd className="h-[18px] min-w-[22px] border border-line-structure bg-surface-1 px-1 font-mono text-[10px] text-text-disabled">
        Cmd
      </Kbd>
      <Kbd className="h-[18px] min-w-[18px] border border-line-structure bg-surface-1 px-1 font-mono text-[10px] text-text-disabled">
        K
      </Kbd>
    </KbdGroup>
  );
}

function SearchEmptyState({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: React.ReactNode;
}) {
  return (
    <Empty className="min-h-[250px] border border-line-structure bg-surface-1 p-8">
      <EmptyHeader>
        <EmptyMedia
          variant="icon"
          className="border border-line-structure bg-surface-bg text-line-cta"
        >
          {icon}
        </EmptyMedia>
        <EmptyTitle className="font-square text-[13px] font-semibold uppercase tracking-wider text-text-secondary">
          {title}
        </EmptyTitle>
        <EmptyDescription className="max-w-[44ch] text-[13px] leading-relaxed text-text-tertiary">
          {description}
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}

export function DocsSearch({ variant = 'sidebar' }: DocsSearchProps) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const target = event.target as HTMLElement | null;
      const typingInField =
        target?.tagName === 'INPUT' ||
        target?.tagName === 'TEXTAREA' ||
        target?.isContentEditable;

      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        event.stopPropagation();
        setOpen(true);
        return;
      }

      if (!typingInField && event.key === '/') {
        event.preventDefault();
        setOpen(true);
      }
    }

    window.addEventListener('keydown', onKeyDown, true);
    return () => window.removeEventListener('keydown', onKeyDown, true);
  }, []);

  useEffect(() => {
    if (!open) return;

    const controller = new AbortController();
    const timer = window.setTimeout(async () => {
      setLoading(true);
      setError(false);

      try {
        const params = new URLSearchParams({ q: query, limit: '18' });
        const response = await fetch(`/api/search?${params.toString()}`, {
          signal: controller.signal,
        });

        if (!response.ok) throw new Error('Search failed');

        const data = (await response.json()) as SearchResponse;
        setResults(data.results);
      } catch {
        if (!controller.signal.aborted) {
          setResults([]);
          setError(true);
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }, 120);

    return () => {
      window.clearTimeout(timer);
      controller.abort();
    };
  }, [open, query]);

  const grouped = useMemo(() => {
    const docs = results.filter((result) => result.kind === 'doc');
    const connectors = results.filter((result) => result.kind === 'connector');

    return [
      { label: 'Documentation', results: docs },
      { label: 'Connectors', results: connectors },
    ].filter((group) => group.results.length > 0);
  }, [results]);

  function onOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setQuery('');
      setError(false);
    }
  }

  function openResult(url: string) {
    onOpenChange(false);
    router.push(url);
  }

  const isNavbar = variant === 'navbar';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Button
        type="button"
        variant="ghost"
        onClick={() => setOpen(true)}
        className={cn(
          'group relative overflow-hidden border border-line-structure text-left font-sans normal-case tracking-normal',
          'focus-visible:outline-surface-cta-primary',
          isNavbar
            ? 'docs-search-navbar-trigger h-9 w-[clamp(250px,22vw,360px)] max-w-full justify-start bg-surface-bg px-2.5 text-[12px] shadow-[inset_0_1px_0_rgba(255,255,255,0.58)] hover:border-line-cta hover:bg-surface-2'
            : 'link-box h-auto w-full justify-start bg-surface-bg px-2.5 py-2 text-[12px] hover:bg-surface-2',
        )}
        aria-label="Search documentation"
      >
        <span aria-hidden className="corner-box-hover-child" />
        <span
          aria-hidden="true"
          className={cn(
            'flex shrink-0 items-center justify-center border border-line-structure bg-surface-1 text-line-cta transition-colors group-hover:border-line-cta',
            isNavbar ? 'size-6' : 'size-5 border-0 bg-transparent',
          )}
        >
          <Search className="size-3.5" />
        </span>
        <span className="min-w-0 flex-1 truncate text-text-tertiary transition-colors group-hover:text-text-secondary">
          Search docs and connectors
        </span>
        <SearchHint />
      </Button>

      <DialogContent
        showCloseButton={false}
        className="docs-command-panel top-[45%] flex max-h-[min(760px,calc(100vh-2rem))] w-[min(820px,calc(100vw-1.5rem))] max-w-none translate-y-[-45%] grid-rows-none flex-col gap-0 overflow-hidden border border-line-cta bg-surface-bg p-0 text-text-primary shadow-[0_24px_90px_rgba(12,31,23,0.26)] ring-0"
      >
        <div className="flex items-start gap-3 border-b border-line-structure bg-surface-1 px-3 py-3">
          <span className="flex size-9 shrink-0 items-center justify-center border border-line-structure bg-surface-bg">
            <CommandIcon className="size-4 text-line-cta" aria-hidden="true" />
          </span>
          <DialogHeader className="min-w-0 flex-1 gap-1">
            <DialogTitle className="font-square text-[12px] font-semibold uppercase tracking-wider text-text-secondary">
              Search pm documentation
            </DialogTitle>
            <DialogDescription className="text-[11px] leading-relaxed text-text-tertiary">
              Search the docs index, connector catalog pages, setup fields, commands, and agent-readable metadata.
            </DialogDescription>
          </DialogHeader>
          <DialogClose asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 border border-line-structure bg-surface-bg text-text-tertiary hover:bg-surface-2 hover:text-text-primary"
              aria-label="Close search"
            >
              <X className="size-4" aria-hidden="true" />
            </Button>
          </DialogClose>
        </div>

        <Command shouldFilter={false} loop className="min-h-0 bg-transparent p-0 text-text-primary">
          <div className="flex items-center gap-2 border-b border-line-structure px-3 py-2.5">
            <div className="min-w-0 flex-1">
              <CommandInput
                autoFocus
                value={query}
                onValueChange={setQuery}
                placeholder="Search 646 connectors, CLI commands, setup fields..."
                className="h-9 bg-transparent font-sans text-[15px] text-text-primary placeholder:text-text-disabled"
              />
            </div>
            {loading ? (
              <Spinner className="size-4 shrink-0 text-text-disabled" />
            ) : (
              <span className="hidden items-center gap-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled sm:flex">
                <Kbd className="h-[18px] min-w-[18px] border border-line-structure bg-surface-1 px-1 font-mono text-[10px] text-text-disabled">
                  /
                </Kbd>
                Focus
              </span>
            )}
          </div>

          <div className="grid grid-cols-1 border-b border-line-structure bg-surface-1 text-[11px] text-text-tertiary sm:grid-cols-3">
            <div className="flex items-center gap-2 border-b border-line-structure px-4 py-2 sm:border-b-0 sm:border-r">
              <Cable className="size-3.5 text-line-cta" aria-hidden="true" />
              Connector catalog indexed
            </div>
            <div className="flex items-center gap-2 border-b border-line-structure px-4 py-2 sm:border-b-0 sm:border-r">
              <Bot className="size-3.5 text-line-cta" aria-hidden="true" />
              Agent terms searchable
            </div>
            <div className="flex items-center gap-2 px-4 py-2">
              <FileText className="size-3.5 text-line-cta" aria-hidden="true" />
              Raw doc pages included
            </div>
          </div>

          <CommandList className="max-h-[min(520px,calc(100vh-18rem))] min-h-[260px] p-2">
            {loading && results.length === 0 ? (
              <div className="grid min-h-[250px] place-items-center border border-line-structure bg-surface-1">
                <div className="flex items-center gap-3 text-[13px] text-text-tertiary">
                  <Spinner className="size-4 text-line-cta" />
                  Searching the local docs index
                </div>
              </div>
            ) : null}

            {!loading && error ? (
              <CommandEmpty>
                <SearchEmptyState
                  icon={<Search className="size-4" aria-hidden="true" />}
                  title="Search is unavailable"
                  description="The local search endpoint did not respond. Try again after the dev server reloads."
                />
              </CommandEmpty>
            ) : null}

            {!loading && !error ? (
              <CommandEmpty>
                <SearchEmptyState
                  icon={<Search className="size-4" aria-hidden="true" />}
                  title="No results"
                  description={
                    <>
                      Try a connector name, command, config field, auth method, or route like{' '}
                      <code className="border border-line-structure bg-surface-bg px-1 font-mono">
                        management_token
                      </code>
                      .
                    </>
                  }
                />
              </CommandEmpty>
            ) : null}

            {grouped.map((group) => (
              <CommandGroup
                key={group.label}
                heading={`${group.label} · ${group.results.length}`}
                className="mb-2 last:mb-0 [&_[cmdk-group-heading]]:font-mono [&_[cmdk-group-heading]]:text-[10px] [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-wider [&_[cmdk-group-heading]]:text-text-disabled"
              >
                {group.results.map((result) => (
                  <CommandItem
                    key={result.url}
                    value={`${result.title} ${result.section} ${result.url}`}
                    onSelect={() => openResult(result.url)}
                    className="group/result mb-1 grid cursor-pointer grid-cols-[2.25rem_minmax(0,1fr)_1.5rem] items-start gap-3 border border-transparent px-3 py-2.5 text-text-primary transition-colors data-[selected=true]:border-line-cta data-[selected=true]:bg-surface-1 [&>svg:last-child]:hidden"
                  >
                    <ResultIcon kind={result.kind} />
                    <span className="min-w-0">
                      <span className="flex min-w-0 flex-wrap items-center gap-2">
                        <span className="truncate text-[14px] font-medium text-text-primary">
                          {result.title}
                        </span>
                        <span className="border border-line-structure bg-surface-bg px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
                          {result.section}
                        </span>
                      </span>
                      <span className="mt-1 block line-clamp-2 text-[12px] leading-relaxed text-text-tertiary">
                        {result.snippet || result.description}
                      </span>
                      <span className="mt-1 block truncate font-mono text-[10px] text-text-disabled">
                        {result.url}
                      </span>
                    </span>
                    <ArrowRight
                      className="mt-2 size-4 justify-self-end text-text-disabled transition-colors group-data-[selected=true]/result:text-line-cta"
                      aria-hidden="true"
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            ))}
          </CommandList>
        </Command>

        <div className="flex flex-wrap items-center justify-between gap-2 border-t border-line-structure bg-surface-1 px-4 py-2.5 text-[11px] text-text-tertiary">
          <span>Use arrow keys to move, Enter to open, Esc to close.</span>
          <span className="font-mono text-[10px] uppercase tracking-wider text-text-disabled">
            {results.length} indexed matches
          </span>
        </div>
      </DialogContent>
    </Dialog>
  );
}
