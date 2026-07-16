'use client';

import { useEffect, useState } from 'react';
import {
  ExternalLink,
  FileCode2,
  GitCommitHorizontal,
  GitPullRequest,
  LoaderCircle,
  X,
} from 'lucide-react';
import type { BlogEvidence, BlogEvidenceKind } from '@/lib/blog';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@/components/ui/dialog';

type EvidenceView = BlogEvidence['snapshot'];
type FetchState = 'loading' | 'live' | 'snapshot';
type ResolvedEvidence = {
  evidenceId: string;
  fetchState: FetchState;
  view: EvidenceView;
};
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

export function GitHubEvidenceDialog({
  evidence,
  onClose,
}: {
  evidence: BlogEvidence | null;
  onClose: () => void;
}) {
  const [resolved, setResolved] = useState<ResolvedEvidence | null>(null);

  useEffect(() => {
    if (!evidence) return;

    const cached = liveEvidenceCache.get(evidence.id);
    if (cached) {
      setResolved({ evidenceId: evidence.id, fetchState: 'live', view: cached });
      return;
    }

    const controller = new AbortController();
    setResolved({ evidenceId: evidence.id, fetchState: 'loading', view: evidence.snapshot });

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
        setResolved({ evidenceId: evidence.id, fetchState: 'live', view: nextView });
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'AbortError') return;
        setResolved({
          evidenceId: evidence.id,
          fetchState: 'snapshot',
          view: evidence.snapshot,
        });
      });

    return () => controller.abort();
  }, [evidence]);

  if (!evidence) return null;

  const current = resolved?.evidenceId === evidence.id
    ? resolved
    : { evidenceId: evidence.id, fetchState: 'loading' as const, view: evidence.snapshot };
  const { fetchState, view } = current;

  const closedWithoutMerge = view.status === 'Closed, not merged';
  const statusClass = closedWithoutMerge
    ? 'border-red-300 bg-red-50 text-red-800'
    : 'border-line-cta bg-surface-cta-primary/20 text-line-cta';

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose(); }}>
      <DialogContent
        showCloseButton={false}
        className="corner-box-corners max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-[560px] gap-0 overflow-hidden rounded-none border border-line-cta bg-surface-bg p-0 text-text-primary shadow-[0_24px_80px_rgba(12,31,23,0.24)] ring-0 sm:max-w-[560px] motion-reduce:transition-none"
        aria-describedby="github-evidence-description"
        data-github-evidence-preview
      >
        <header className="relative border-b border-line-structure px-5 py-5 pr-14">
          <div className="mb-3 flex items-center gap-2 font-mono text-[9px] uppercase tracking-widest text-text-disabled">
            <EvidenceIcon kind={evidence.kind} className="size-3.5 text-line-cta" />
            Source record
          </div>
          <DialogTitle className="font-square text-[21px] font-semibold leading-tight text-text-primary">
            <span className="sr-only">GitHub evidence: </span>
            {evidence.label}
          </DialogTitle>
          <DialogDescription
            id="github-evidence-description"
            className="mt-2 text-[13px] leading-relaxed text-text-tertiary"
          >
            {view.title}
          </DialogDescription>
          <DialogClose asChild>
            <button
              type="button"
              aria-label="Close evidence"
              className="absolute right-4 top-4 flex size-8 items-center justify-center border border-line-structure bg-surface-bg text-text-tertiary transition-colors hover:border-line-cta hover:bg-surface-1 hover:text-text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
            >
              <X className="size-3.5" aria-hidden="true" />
            </button>
          </DialogClose>
        </header>

        <div className="min-h-0 overflow-y-auto px-5 py-5">
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

        <div className="border-t border-line-structure bg-surface-1 p-4">
          <a
            href={evidence.url}
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-10 w-full items-center justify-center gap-2 border border-line-cta bg-line-cta px-3 font-square text-[11px] font-semibold uppercase tracking-normal text-surface-bg transition-opacity hover:opacity-90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-line-cta"
          >
            {evidence.openLabel}
            <ExternalLink className="size-3.5" aria-hidden="true" />
          </a>
        </div>
      </DialogContent>
    </Dialog>
  );
}
