import { NextResponse } from 'next/server';

import {
  CONNECTOR_CATALOG,
  CONNECTOR_CATALOG_COUNT,
  type ConnectorMeta,
} from '@/lib/connectors.catalog.generated';
import { DOCS_PAGES } from '@/lib/docs.generated';

export const dynamic = 'force-dynamic';

type SearchKind = 'doc' | 'connector';

type SearchRecord = {
  title: string;
  description: string;
  url: string;
  kind: SearchKind;
  section: string;
  keywords: string;
  content: string;
  priority: number;
};

type SearchResult = Omit<SearchRecord, 'keywords' | 'content' | 'priority'> & {
  snippet: string;
  score: number;
};

const DEFAULT_LIMIT = 18;
const MAX_LIMIT = 50;

function cleanText(value: string): string {
  return value
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/[{}[\]()*_>#|]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function docSection(url: string): string {
  if (url === '/docs') return 'Overview';
  const segment = url.split('/').filter(Boolean).at(-1) ?? 'docs';
  return segment
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

function docsRecords(): SearchRecord[] {
  const records = DOCS_PAGES.map((page) => {
    return {
      title: page.title || docSection(page.url),
      description: page.description,
      url: page.url,
      kind: 'doc' as const,
      section: docSection(page.url),
      keywords: `${page.title} ${page.description} ${page.url}`,
      content: cleanText(`${page.description} ${page.body}`),
      priority: page.url === '/docs' ? 10 : 6,
    };
  });

  records.push({
    title: 'Connector catalog',
    description: `Browse ${CONNECTOR_CATALOG_COUNT} connector bundles by capability, stream, action, and integration type.`,
    url: '/docs/connectors',
    kind: 'doc',
    section: 'Catalog',
    keywords: 'connectors catalog capabilities streams actions integrations browse search metadata',
    content:
      'Connector catalog capabilities ETL streams reverse ETL write actions setup authentication metadata data.json inspect credentials',
    priority: 9,
  });

  return records;
}

function connectorRecord(c: ConnectorMeta): SearchRecord {
  const streams = c.streams
    .map((stream) => `${stream.name} ${stream.primaryKey.join(' ')} ${stream.cursor} ${stream.incremental ? 'incremental' : ''}`)
    .join(' ');
  const writeActions = c.writeActions
    .map((action) => `${action.name} ${action.method} ${action.kind}`)
    .join(' ');
  const docs = c.docs.map((doc) => `${doc.title} ${doc.type} ${doc.url}`).join(' ');

  return {
    title: `${c.name} connector`,
    description:
      c.description ||
      `${c.name} exposes ${c.streams.length} ETL streams and ${c.writeActions.length} write actions.`,
    url: `/docs/connectors/${c.slug}`,
    kind: 'connector',
    section: 'Connector',
    keywords: [
      c.name,
      c.slug,
      c.category,
      c.categoryLabel,
      c.releaseStage,
      c.capabilityLabels.join(' '),
      streams,
      writeActions,
    ].join(' '),
    content: cleanText(`${c.description} ${streams} ${writeActions} ${docs} ${c.docsMd}`),
    priority: c.featured ? 8 : 4,
  };
}

const SEARCH_INDEX: SearchRecord[] = [
  ...docsRecords(),
  ...CONNECTOR_CATALOG.map(connectorRecord),
];

function tokensFor(query: string): string[] {
  return query
    .toLowerCase()
    .split(/\s+/)
    .map((token) => token.trim())
    .filter(Boolean);
}

function scoreRecord(record: SearchRecord, tokens: string[], query: string): number {
  if (tokens.length === 0) return record.priority;

  const title = record.title.toLowerCase();
  const description = record.description.toLowerCase();
  const section = record.section.toLowerCase();
  const keywords = record.keywords.toLowerCase();
  const content = record.content.toLowerCase();
  const url = record.url.toLowerCase();

  let score = record.priority;
  const normalizedQuery = query.toLowerCase();

  if (title === normalizedQuery) score += 140;
  if (title.startsWith(normalizedQuery)) score += 80;
  if (title.includes(normalizedQuery)) score += 50;
  if (url.includes(normalizedQuery)) score += 34;
  if (keywords.includes(normalizedQuery)) score += 26;
  if (description.includes(normalizedQuery)) score += 18;
  if (content.includes(normalizedQuery)) score += 8;

  for (const token of tokens) {
    if (title.startsWith(token)) score += 28;
    else if (title.includes(token)) score += 18;

    if (url.includes(token)) score += 10;
    if (section.includes(token)) score += 8;
    if (keywords.includes(token)) score += 7;
    if (description.includes(token)) score += 5;
    if (content.includes(token)) score += 2;

    const matched =
      title.includes(token) ||
      url.includes(token) ||
      section.includes(token) ||
      keywords.includes(token) ||
      description.includes(token) ||
      content.includes(token);

    if (!matched) score -= 90;
  }

  return score;
}

function snippetFor(record: SearchRecord, tokens: string[]): string {
  const fallback = record.description || record.content || record.section;
  if (tokens.length === 0) return fallback.slice(0, 220);

  const content = record.content || fallback;
  const lower = content.toLowerCase();
  const hit = tokens
    .map((token) => lower.indexOf(token))
    .filter((index) => index >= 0)
    .sort((a, b) => a - b)[0];

  if (hit === undefined) return fallback.slice(0, 220);

  const start = Math.max(0, hit - 70);
  const end = Math.min(content.length, hit + 170);
  const prefix = start > 0 ? '...' : '';
  const suffix = end < content.length ? '...' : '';
  return `${prefix}${content.slice(start, end).trim()}${suffix}`;
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const query =
    searchParams.get('q') ??
    searchParams.get('query') ??
    searchParams.get('search') ??
    '';
  const trimmedQuery = query.trim();
  const tokens = tokensFor(trimmedQuery);
  const limit = Math.min(
    MAX_LIMIT,
    Math.max(1, Number(searchParams.get('limit') ?? DEFAULT_LIMIT) || DEFAULT_LIMIT),
  );

  const results: SearchResult[] = SEARCH_INDEX.map((record) => ({
    title: record.title,
    description: record.description,
    url: record.url,
    kind: record.kind,
    section: record.section,
    score: scoreRecord(record, tokens, trimmedQuery),
    snippet: snippetFor(record, tokens),
  }))
    .filter((record) => record.score > 0)
    .sort((a, b) => b.score - a.score || a.title.localeCompare(b.title))
    .slice(0, limit);

  return NextResponse.json({
    query: trimmedQuery,
    total: results.length,
    results,
  });
}
