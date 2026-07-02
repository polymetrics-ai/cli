'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Download, ExternalLink, PackageCheck, Radio } from 'lucide-react';
import type { GithubRelease } from '@/lib/release-utils';
import { formatReleaseDate, releaseExcerpt, releaseTitle } from '@/lib/release-utils';

type ReleasesPayload = {
  releases: GithubRelease[];
};

function ShortcutBadge({ value }: { value: string }) {
  return (
    <kbd className="inline-flex h-5 min-w-5 items-center justify-center border border-[rgba(15,61,46,0.20)] bg-[rgba(15,61,46,0.08)] px-1 font-mono text-[10px] font-medium leading-none text-text-tertiary">
      {value}
    </kbd>
  );
}

export function ReleaseHighlights() {
  const router = useRouter();
  const [releases, setReleases] = useState<GithubRelease[] | null>(null);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.repeat || event.metaKey || event.ctrlKey || event.altKey) return;
      if (event.key.toLowerCase() !== 'r') return;
      if (event.target instanceof HTMLElement) {
        const tag = event.target.tagName;
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || event.target.isContentEditable) {
          return;
        }
      }

      event.preventDefault();
      router.push('/changelog');
    }

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [router]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const response = await fetch('/api/releases?limit=3');
        const payload = await response.json() as ReleasesPayload;
        if (!cancelled) setReleases(payload.releases ?? []);
      } catch {
        if (!cancelled) setReleases([]);
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  if (releases === null) {
    return (
      <div className="space-y-2">
        {[0, 1].map((index) => (
          <div key={index} className="grid grid-cols-[1.1rem_minmax(0,1fr)] gap-2">
            <span className="mt-0.5 h-4 w-4 animate-pulse border border-line-structure bg-surface-1" />
            <span className="min-w-0 space-y-1.5">
              <span className="block h-3 w-28 animate-pulse bg-surface-1" />
              <span className="block h-2.5 w-20 animate-pulse bg-surface-1" />
            </span>
          </div>
        ))}
      </div>
    );
  }

  if (releases.length === 0) {
    return (
      <div className="border border-line-structure bg-surface-1 p-2.5">
        <div className="flex items-start gap-2">
          <span className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center border border-line-structure bg-surface-bg text-line-cta">
            <Radio className="h-3 w-3" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <p className="text-[12px] font-medium leading-snug text-text-secondary">
              No GitHub releases yet
            </p>
            <p className="mt-1 text-[11px] leading-snug text-text-tertiary">
              Release notes and binary assets will stream here after the first tag.
            </p>
          </div>
        </div>
        <Link
          href="/changelog"
          className="mt-3 inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-secondary transition-colors hover:text-text-primary"
        >
          Open releases
          <ShortcutBadge value="R" />
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-2.5">
      {releases.map((release) => (
        <a
          key={release.id}
          href={release.htmlUrl}
          target="_blank"
          rel="noreferrer"
          className="link-box group relative block border border-transparent p-1.5 transition-colors hover:border-line-structure hover:bg-surface-1"
        >
          <span aria-hidden className="corner-box-hover-child" />
          <span className="grid grid-cols-[1.1rem_minmax(0,1fr)] gap-2">
            <span className="mt-0.5 flex h-4 w-4 items-center justify-center border border-line-structure bg-surface-1">
              <PackageCheck className="h-2.5 w-2.5 text-line-cta" aria-hidden="true" />
            </span>
            <span className="min-w-0">
              <span className="block truncate text-[12px] leading-snug text-text-secondary">
                {releaseTitle(release)}
              </span>
              <span className="mt-0.5 block font-mono text-[10px] text-text-disabled">
                {release.tagName} · {formatReleaseDate(release.publishedAt)}
              </span>
              <span className="mt-1 line-clamp-2 block text-[11px] leading-snug text-text-tertiary">
                {releaseExcerpt(release.body, 118)}
              </span>
              <span className="mt-1.5 inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
                <Download className="h-3 w-3 text-line-cta" aria-hidden="true" />
                {release.assets.length} asset{release.assets.length === 1 ? '' : 's'}
              </span>
            </span>
          </span>
        </a>
      ))}

      <Link
        href="/changelog"
        className="inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-secondary transition-colors hover:text-text-primary"
      >
        View all releases
        <ExternalLink className="h-3 w-3" aria-hidden="true" />
        <ShortcutBadge value="R" />
      </Link>
    </div>
  );
}
