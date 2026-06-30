'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { ArrowRight, Cable, Database, FileJson, Filter, Search } from 'lucide-react';
import {
  CONNECTOR_CATALOG,
  CONNECTOR_CATALOG_COUNT,
  CONNECTOR_SOURCES,
  CONNECTOR_DESTINATIONS,
  CONNECTOR_CATEGORY_COUNTS,
  type ConnectorMeta,
} from '@/lib/connectors.catalog.generated';
import { Badge, statusVariant, typeVariant } from '@/components/ui/badge';
import { ConnectorIcon } from '@/components/docs/connector/connector-icon';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';

const CATEGORY_LABELS: Record<string, string> = {
  api: 'API',
  database: 'Database',
  file: 'File',
  vectorstore: 'Vector store',
  message_queue: 'Message queue',
  other: 'Other',
};

type TypeFilter = 'all' | 'source' | 'destination';

function ConnectorCard({ c }: { c: ConnectorMeta }) {
  return (
    <Link
      href={`/docs/connectors/${c.slug}`}
      className="link-box group relative flex min-h-[138px] flex-col justify-between gap-3 border border-line-structure bg-surface-bg p-4 transition-colors hover:border-line-cta hover:bg-surface-1"
    >
      <span aria-hidden className="corner-box-hover-child" />
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-3">
          <ConnectorIcon icon={c.icon} name={c.name} />
          <span className="min-w-0 truncate font-medium text-text-primary">{c.name}</span>
        </div>
        {c.featured && (
          <span className="font-mono text-[10px] uppercase tracking-wider text-text-disabled">
            featured
          </span>
        )}
      </div>
      <p className="line-clamp-2 text-[12px] leading-relaxed text-text-tertiary">
        {c.notes ||
          `${c.name} exposes ${c.config.length} config fields, ${c.secrets.length} secret fields, and ${c.docs.length} catalog doc links.`}
      </p>
      <div className="grid grid-cols-3 border border-line-structure bg-line-structure text-[10px] text-text-tertiary">
        <span className="bg-surface-bg px-2 py-1 font-mono">
          {c.config.length} config
        </span>
        <span className="bg-surface-bg px-2 py-1 font-mono">
          {c.secrets.length} secret
        </span>
        <span className="bg-surface-bg px-2 py-1 font-mono">
          {c.docs.length} docs
        </span>
      </div>
      <div className="flex flex-wrap gap-1.5">
        <Badge variant={typeVariant(c.type)}>{c.type}</Badge>
        <Badge variant="category">{CATEGORY_LABELS[c.category] ?? c.category}</Badge>
        <Badge variant={statusVariant(c.status)}>
          {c.status === 'enabled' ? 'enabled' : 'planned'}
        </Badge>
      </div>
      <span className="mt-1 inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-disabled transition-colors group-hover:text-line-cta">
        Open connector
        <ArrowRight className="h-3 w-3" aria-hidden="true" />
      </span>
    </Link>
  );
}

export function ConnectorBrowser() {
  const [query, setQuery] = useState('');
  const [type, setType] = useState<TypeFilter>('all');
  const [category, setCategory] = useState<string>('all');

  const categories = useMemo(
    () => Object.keys(CONNECTOR_CATEGORY_COUNTS).sort(),
    [],
  );

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return CONNECTOR_CATALOG.filter((c) => {
      if (type !== 'all' && c.type !== type) return false;
      if (category !== 'all' && c.category !== category) return false;
      if (!q) return true;
      return (
        c.name.toLowerCase().includes(q) ||
        c.slug.toLowerCase().includes(q) ||
        c.category.toLowerCase().includes(q)
      );
    });
  }, [query, type, category]);

  return (
    <div className="flex flex-col gap-5">
      <header className="border border-line-structure bg-surface-bg">
        <div className="border-b border-line-structure bg-surface-1 px-4 py-3">
          <div className="inline-flex items-center gap-2 border border-line-structure bg-surface-bg px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
            <Cable className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            Connector catalog
          </div>
          <h1 className="mt-4 font-square text-[30px] font-bold leading-[1.15] text-text-primary sm:text-[34px]">
            Search every connector page
          </h1>
          <p className="mt-3 max-w-[70ch] text-[14px] leading-relaxed text-text-tertiary">
            Browse generated docs for {CONNECTOR_CATALOG_COUNT} source and destination connectors,
            including setup notes, auth fields, official docs links, and machine-readable data.
          </p>
        </div>
        <div className="grid border-b border-line-structure text-[12px] text-text-tertiary sm:grid-cols-3">
          <div className="flex items-center gap-2 border-b border-line-structure px-4 py-3 sm:border-b-0 sm:border-r">
            <Cable className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            {CONNECTOR_SOURCES} sources
          </div>
          <div className="flex items-center gap-2 border-b border-line-structure px-4 py-3 sm:border-b-0 sm:border-r">
            <Database className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            {CONNECTOR_DESTINATIONS} destinations
          </div>
          <div className="flex items-center gap-2 px-4 py-3">
            <FileJson className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
            data.json on each page
          </div>
        </div>

        <div className="flex flex-col gap-3 px-4 py-4">
          <InputGroup className="h-11 border-line-structure bg-surface-1 focus-within:border-line-cta focus-within:ring-2 focus-within:ring-surface-cta-primary/20">
            <InputGroupAddon>
              <Search className="h-4 w-4 shrink-0 text-line-cta" aria-hidden="true" />
            </InputGroupAddon>
            <InputGroupInput
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search names, slugs, categories, or auth fields..."
              aria-label="Search connectors"
              className="font-sans text-[14px] text-text-primary placeholder:text-text-disabled"
            />
          </InputGroup>

          <div className="grid gap-3 border border-line-structure bg-surface-1 p-3 md:grid-cols-[minmax(0,1fr)_220px]">
            <div className="md:col-span-2 flex items-center gap-2 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
              <Filter className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
              Filters
            </div>

            <ToggleGroup
              type="single"
              value={type}
              onValueChange={(value) => {
                if (value) setType(value as TypeFilter);
              }}
              variant="outline"
              spacing={1}
              className="flex w-full flex-wrap"
            >
              {(['all', 'source', 'destination'] as TypeFilter[]).map((t) => (
                <ToggleGroupItem
                  key={t}
                  value={t}
                  className="border-line-structure bg-surface-bg px-3 font-sans text-[12px] text-text-secondary data-[state=on]:border-line-cta data-[state=on]:bg-surface-2 data-[state=on]:text-line-cta"
                >
                  {t === 'all' ? 'All types' : t === 'source' ? 'Sources' : 'Destinations'}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>

            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="h-8 w-full border-line-structure bg-surface-bg font-sans text-[12px] text-text-secondary focus-visible:border-line-cta focus-visible:ring-surface-cta-primary/20">
                <SelectValue placeholder="All categories" />
              </SelectTrigger>
              <SelectContent className="border border-line-structure bg-surface-bg text-text-secondary ring-0">
                <SelectItem value="all">All categories</SelectItem>
                {categories.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {CATEGORY_LABELS[cat] ?? cat} ({CONNECTOR_CATEGORY_COUNTS[cat]})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </header>

      <div className="flex flex-col gap-2 border border-line-structure bg-surface-1 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="font-mono text-[11px] uppercase tracking-wider text-text-tertiary">
          Showing {filtered.length} of {CONNECTOR_CATALOG_COUNT}
        </p>
        <Button
          variant="quiet"
          onClick={() => {
            setQuery('');
            setType('all');
            setCategory('all');
          }}
          className="w-fit"
        >
          Reset filters
        </Button>
      </div>

      {/* Grid */}
      {filtered.length === 0 ? (
        <Empty className="border border-line-structure bg-surface-1 px-4 py-12">
          <EmptyHeader>
            <EmptyMedia variant="icon" className="border border-line-structure bg-surface-bg text-line-cta">
              <Search className="h-4 w-4" aria-hidden="true" />
            </EmptyMedia>
            <EmptyTitle className="font-square text-[14px] font-semibold uppercase tracking-wider text-text-secondary">
              No connectors match
            </EmptyTitle>
            <EmptyDescription className="text-[14px] text-text-tertiary">
              Try a vendor name, connector slug, auth field, type, or category.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((c) => (
            <ConnectorCard key={c.slug} c={c} />
          ))}
        </div>
      )}
    </div>
  );
}
