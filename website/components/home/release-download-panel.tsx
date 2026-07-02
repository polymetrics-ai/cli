'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { Download, Radio } from 'lucide-react';
import type { GithubRelease, ReleaseAsset } from '@/lib/release-utils';
import { formatBytes, selectPreferredAsset } from '@/lib/release-utils';

function shouldIgnoreShortcutTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
}

function ShortcutBadge({ value }: { value: string }) {
  return (
    <kbd className="inline-flex h-5 min-w-5 items-center justify-center border border-white/30 bg-white/20 px-1 font-mono text-[10px] font-medium leading-none text-white/85">
      {value}
    </kbd>
  );
}

function platformHint() {
  if (typeof navigator === 'undefined') return '';
  const nav = navigator as Navigator & {
    userAgentData?: { platform?: string };
  };
  return `${nav.userAgentData?.platform ?? navigator.platform} ${navigator.userAgent}`;
}

function assetDownloads(asset: ReleaseAsset) {
  return `${asset.downloadCount.toLocaleString()} download${asset.downloadCount === 1 ? '' : 's'}`;
}

export function ReleaseDownloadPanel({ release }: { release: GithubRelease | null }) {
  const downloadRef = useRef<HTMLAnchorElement>(null);
  const [hint, setHint] = useState('');
  const selectedAsset = useMemo(
    () => selectPreferredAsset(release?.assets ?? [], hint),
    [release?.assets, hint],
  );

  useEffect(() => {
    setHint(platformHint());
  }, []);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.repeat || event.metaKey || event.ctrlKey || event.altKey) return;
      if (event.key.toLowerCase() !== 'x') return;
      if (shouldIgnoreShortcutTarget(event.target)) return;
      if (!selectedAsset) return;

      event.preventDefault();
      downloadRef.current?.click();
    }

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [selectedAsset]);

  return (
    <section id="downloads" className="scroll-mt-24 border border-line-structure bg-surface-bg">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line-structure bg-surface-1 px-4 py-3">
        <span className="inline-flex items-center gap-2 font-square text-[13px] font-semibold text-text-secondary">
          <Download className="h-4 w-4 text-line-cta" aria-hidden="true" />
          Binary downloads
        </span>
        <span className="inline-flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-text-disabled">
          Download shortcut
          <span className="border border-[rgba(15,61,46,0.20)] bg-[rgba(15,61,46,0.08)] px-1.5 py-1 text-text-tertiary">
            X
          </span>
        </span>
      </div>

      {selectedAsset ? (
        <div className="grid gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_15rem]">
          <div className="min-w-0">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Recommended for this machine
            </p>
            <h2 className="mt-2 break-words font-square text-[20px] font-semibold leading-tight text-text-primary">
              {selectedAsset.name}
            </h2>
            <p className="mt-2 text-[13px] leading-relaxed text-text-tertiary">
              The shortcut selects the closest matching release asset by OS and architecture.
              All published assets remain available below.
            </p>
          </div>

          <a
            ref={downloadRef}
            href={selectedAsset.url}
            className="btn-shine inline-flex h-fit items-center justify-center gap-2 border border-emerald-900 bg-emerald-800 px-4 py-2.5 font-square text-[13px] font-semibold text-white transition-opacity hover:opacity-90"
          >
            Download binary
            <ShortcutBadge value="X" />
          </a>

          <div className="lg:col-span-2">
            <div className="divide-y divide-line-structure border border-line-structure bg-surface-1">
              {release?.assets.map((asset) => (
                <a
                  key={asset.id}
                  href={asset.url}
                  className="grid gap-2 px-3 py-3 transition-colors hover:bg-surface-bg sm:grid-cols-[minmax(0,1fr)_7rem_7rem] sm:items-center"
                >
                  <span className="min-w-0 truncate font-mono text-[12px] text-text-secondary">
                    {asset.name}
                  </span>
                  <span className="font-mono text-[11px] text-text-disabled">
                    {formatBytes(asset.size)}
                  </span>
                  <span className="font-mono text-[11px] text-text-disabled">
                    {assetDownloads(asset)}
                  </span>
                </a>
              ))}
            </div>
          </div>
        </div>
      ) : (
        <div className="grid gap-4 p-4 lg:grid-cols-[2.5rem_minmax(0,1fr)]">
          <span className="flex h-10 w-10 items-center justify-center border border-line-structure bg-surface-1 text-line-cta">
            <Radio className="h-4 w-4" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <h2 className="font-square text-[18px] font-semibold text-text-primary">
              No binary artifacts published yet
            </h2>
            <p className="mt-2 max-w-[66ch] text-[14px] leading-relaxed text-text-tertiary">
              GitHub has no release assets for polymetrics-ai/cli right now. The X shortcut
              is reserved and will download the latest matching binary as soon as the first
              release artifact is attached.
            </p>
          </div>
        </div>
      )}
    </section>
  );
}
