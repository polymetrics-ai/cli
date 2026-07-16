'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  ExternalLink,
  Eye,
  FileCode2,
  GitCommitHorizontal,
  GitPullRequest,
  LoaderCircle,
} from 'lucide-react';
import type { BlogEvidence, BlogEvidenceKind } from '@/lib/blog';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';

type EvidenceView = BlogEvidence['snapshot'];
type FetchState = 'loading' | 'live' | 'snapshot';
const liveEvidenceCache = new Map<string, EvidenceView>();

function EvidenceIcon({ kind, className }: { kind: BlogEvidenceKind; className?: string }) {
  if (kind === 'pull_request') return <GitPullRequest className={className} aria-hidden="true" />;
  if (kind === 'commit') return <GitCommitHorizontal className={className} aria-hidden="true" />;
  return <FileCode2 className={className} aria-hidden="true" />;
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return value !== null && typeof value === 'object' ? (value as Record<string, unknown>) : undefined;
}

function stringValue(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined;
}

function numberValue(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(value);
}

function parsePullRequest(evidence: BlogEvidence, payload: Record<string, unknown>): EvidenceView {
  const additions = numberValue(payload.additions);
  const deletions = numberValue(payload.deletions);
  const changedFiles = numberValue(payload.changed_files);
  const head = asRecord(payload.head);
  const merged = Boolean(stringValue(payload.merged_at));
  const state = stringValue(payload.state);
  const draft = payload.draft === true;
  const status = merged ? 'Merged' : state === 'closed' ? 'Closed, not merged' : draft ? 'Draft' : 'Open';
  const stats = additions !== undefined && deletions !== undefined
    ? [
        { label: 'Changed lines', value: formatNumber(additions + deletions) },
        { label: 'Additions', value: formatNumber(additions) },
        { label: 'Deletions', value: formatNumber(deletions) },
        ...(changedFiles === undefined ? [] : [{ label: 'Files', value: formatNumber(changedFiles) }]),
      ]
    : evidence.snapshot.stats;

  return {
    ...evidence.snapshot,
    title: stringValue(payload.title) ?? evidence.snapshot.title,
    status,
    ref: stringValue(head?.sha) ?? evidence.snapshot.ref,
    stats,
  };
}

function parseCommit(evidence: BlogEvidence, payload: Record<string, unknown>): EvidenceView {
  const commit = asRecord(payload.commit);
  const stats = asRecord(payload.stats);
  const additions = numberValue(stats?.additions);
  const deletions = numberValue(stats?.deletions);
  const total = numberValue(stats?.total);
  const message = stringValue(commit?.message)?.split('\n')[0];

  return {
    ...evidence.snapshot,
    title: message ?? evidence.snapshot.title,
    ref: stringValue(payload.sha) ?? evidence.snapshot.ref,
    stats: total === undefined
      ? evidence.snapshot.stats
      : [
          { label: 'Changed lines', value: formatNumber(total) },
          ...(additions === undefined ? [] : [{ label: 'Additions', value: formatNumber(additions) }]),
          ...(deletions === undefined ? [] : [{ label: 'Deletions', value: formatNumber(deletions) }]),
          { label: 'SHA', value: (stringValue(payload.sha) ?? evidence.snapshot.ref).slice(0, 7) },
        ],
  };
}

function parseWorkflow(evidence: BlogEvidence, payload: Record<string, unknown>): EvidenceView {
  const size = numberValue(payload.size);
  const ref = stringValue(payload.sha) ?? evidence.snapshot.ref;

  return {
    ...evidence.snapshot,
    title: stringValue(payload.name) ?? evidence.snapshot.title,
    ref,
    stats: [
      { label: 'Branch', value: 'main' },
      ...(size === undefined ? [] : [{ label: 'File size', value: `${formatNumber(size)} B` }]),
      { label: 'Blob', value: ref.slice(0, 7) },
    ],
  };
}

function parseEvidence(evidence: BlogEvidence, payload: unknown): EvidenceView {
  const record = asRecord(payload);
  if (!record) return evidence.snapshot;
  if (evidence.kind === 'pull_request') return parsePullRequest(evidence, record);
  if (evidence.kind === 'commit') return parseCommit(evidence, record);
  return parseWorkflow(evidence, record);
}

export function GitHubEvidenceMarkers({
  evidence,
  evidenceIds,
  onOpen,
}: {
  evidence: BlogEvidence[];
  evidenceIds: string[];
  onOpen: (evidence: BlogEvidence) => void;
}) {
  const items = useMemo(() => {
    const byId = new Map(evidence.map((item) => [item.id, item]));
    return evidenceIds.flatMap((id) => {
      const item = byId.get(id);
      return item ? [item] : [];
    });
  }, [evidence, evidenceIds]);

  if (items.length === 0) return null;

  return (
    <div className="mt-5 border-y border-line-structure bg-surface-1/45 py-3" data-github-evidence-trail>
      <div className="mb-2 flex items-center justify-between gap-3 px-3">
        <p className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
          GitHub evidence
        </p>
        <span className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
          {items.length} {items.length === 1 ? 'source' : 'sources'}
        </span>
      </div>
      <div className="grid gap-px bg-line-structure sm:grid-cols-2">
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            aria-label={`Preview ${item.label} evidence`}
            data-evidence-marker={item.id}
            onClick={() => onOpen(item)}
            className="group flex min-h-14 items-center gap-3 bg-surface-bg px-3 py-2 text-left transition-colors hover:bg-surface-1 focus-visible:relative focus-visible:z-10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
          >
            <span className="flex size-8 shrink-0 items-center justify-center border border-line-structure bg-surface-1 text-line-cta transition-colors group-hover:border-line-cta">
              <EvidenceIcon kind={item.kind} className="size-3.5" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block truncate font-square text-[12px] font-semibold text-text-primary">
                {item.label}
              </span>
              <span className="mt-0.5 block truncate font-mono text-[9px] uppercase tracking-wider text-text-disabled">
                {item.snapshot.status}
              </span>
            </span>
            <Eye className="size-3.5 shrink-0 text-text-disabled transition-colors group-hover:text-line-cta" aria-hidden="true" />
          </button>
        ))}
      </div>
    </div>
  );
}

export function GitHubEvidenceSheet({
  evidence,
  onClose,
}: {
  evidence: BlogEvidence | null;
  onClose: () => void;
}) {
  const [view, setView] = useState<EvidenceView | null>(null);
  const [fetchState, setFetchState] = useState<FetchState>('snapshot');

  useEffect(() => {
    if (!evidence) return;

    const cached = liveEvidenceCache.get(evidence.id);
    if (cached) {
      setView(cached);
      setFetchState('live');
      return;
    }

    const controller = new AbortController();
    setView(evidence.snapshot);
    setFetchState('loading');

    fetch(evidence.apiUrl, {
      headers: { Accept: 'application/vnd.github+json' },
      signal: controller.signal,
    })
      .then((response) => {
        if (!response.ok) throw new Error(`GitHub returned ${response.status}`);
        return response.json() as Promise<unknown>;
      })
      .then((payload) => {
        const nextView = parseEvidence(evidence, payload);
        liveEvidenceCache.set(evidence.id, nextView);
        setView(nextView);
        setFetchState('live');
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'AbortError') return;
        setView(evidence.snapshot);
        setFetchState('snapshot');
      });

    return () => controller.abort();
  }, [evidence]);

  if (!evidence || !view) return null;

  const closedWithoutMerge = view.status === 'Closed, not merged';
  const statusClass = closedWithoutMerge
    ? 'border-red-300 bg-red-50 text-red-800'
    : 'border-line-cta bg-surface-cta-primary/20 text-line-cta';

  return (
    <Sheet open onOpenChange={(open) => { if (!open) onClose(); }}>
      <SheetContent
        side="right"
        className="w-full gap-0 border-line-structure bg-surface-bg p-0 sm:max-w-[460px]"
        aria-describedby="github-evidence-description"
        data-github-evidence-preview
      >
        <SheetHeader className="border-b border-line-structure px-5 py-5 pr-14">
          <div className="mb-3 flex items-center gap-2 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
            <EvidenceIcon kind={evidence.kind} className="size-3.5 text-line-cta" />
            GitHub evidence
          </div>
          <SheetTitle className="font-square text-[21px] font-semibold leading-tight text-text-primary">
            <span className="sr-only">GitHub evidence: </span>
            {evidence.label}
          </SheetTitle>
          <SheetDescription
            id="github-evidence-description"
            className="mt-2 text-[13px] leading-relaxed text-text-tertiary"
          >
            {view.title}
          </SheetDescription>
        </SheetHeader>

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
          <div className="flex flex-wrap items-center gap-2">
            <span className={`border px-2 py-1 font-mono text-[9px] uppercase tracking-widest ${statusClass}`}>
              {view.status}
            </span>
            <span className="inline-flex items-center gap-1.5 border border-line-structure bg-surface-1 px-2 py-1 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
              {fetchState === 'loading' ? (
                <LoaderCircle className="size-3 animate-spin motion-reduce:animate-none" aria-hidden="true" />
              ) : null}
              {fetchState === 'live' ? 'Live from GitHub' : 'Verified snapshot'}
            </span>
          </div>

          <p className="mt-5 text-[14px] leading-[1.7] text-text-tertiary">{view.summary}</p>

          <dl className="mt-5 grid grid-cols-2 gap-px bg-line-structure border border-line-structure">
            {view.stats.map((stat) => (
              <div key={stat.label} className="min-w-0 bg-surface-bg p-3">
                <dt className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
                  {stat.label}
                </dt>
                <dd className="mt-1 truncate font-square text-[15px] font-semibold tabular-nums text-text-primary" title={stat.value}>
                  {stat.value}
                </dd>
              </div>
            ))}
          </dl>

          <div className="mt-5 border-l-2 border-line-cta bg-surface-1 px-3 py-3">
            <p className="font-mono text-[9px] uppercase tracking-widest text-text-disabled">
              Exact reference
            </p>
            <code className="mt-1 block break-all text-[11px] leading-relaxed text-text-secondary">
              {view.ref}
            </code>
          </div>

          <p className="mt-3 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
            Snapshot verified {evidence.snapshot.verifiedAt}
          </p>
        </div>

        <SheetFooter className="border-t border-line-structure bg-surface-1 p-4">
          <a
            href={evidence.url}
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-10 w-full items-center justify-center gap-2 border border-line-cta bg-line-cta px-3 font-square text-[11px] font-semibold uppercase tracking-normal text-surface-bg transition-opacity hover:opacity-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
          >
            {evidence.openLabel}
            <ExternalLink className="size-3.5" aria-hidden="true" />
          </a>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
