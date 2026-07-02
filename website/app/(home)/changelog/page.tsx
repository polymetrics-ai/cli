import type { Metadata } from 'next';
import { Download, ExternalLink, GitBranch as Github, PackageCheck, Radio } from 'lucide-react';
import { HomeSidebar } from '@/components/home/home-sidebar';
import { ReleaseDownloadPanel } from '@/components/home/release-download-panel';
import { OnPageTocAside } from '@/components/ui/on-page-toc';
import { getGithubReleases } from '@/lib/github-releases';
import {
  formatBytes,
  formatReleaseDate,
  releaseExcerpt,
  releaseTitle,
} from '@/lib/release-utils';

export const revalidate = 300;

export const metadata: Metadata = {
  title: 'Changelog',
  description: 'GitHub release notes, binary assets, and product changes for Polymetrics CLI.',
};

const tocItems = [
  { id: 'release-overview', label: 'Overview' },
  { id: 'downloads', label: 'Downloads' },
  { id: 'release-notes', label: 'Release notes' },
];

function totalDownloads(releaseAssets: { downloadCount: number }[]) {
  return releaseAssets.reduce((sum, asset) => sum + asset.downloadCount, 0);
}

export default async function ChangelogPage() {
  const releases = await getGithubReleases(12);
  const latestRelease = releases[0] ?? null;

  return (
    <div className="flex mx-auto w-full max-w-[95rem] overflow-clip">
      <HomeSidebar />

      <main className="flex-1 min-w-0 pattern-bg overflow-hidden pb-8 xl:px-5 2xl:px-10">
        <div className="relative z-[1] mx-auto w-full px-4 py-16 sm:px-8 md:px-0 md:max-w-[680px] md:py-24 xl:max-w-[760px]">
          <header id="release-overview" className="mb-10 grid gap-6 border-b border-line-structure pb-10 lg:grid-cols-[minmax(0,1fr)_18rem]">
            <div className="min-w-0">
              <span className="inline-flex items-center gap-2 font-mono text-[12px] uppercase tracking-widest text-text-disabled">
                <PackageCheck className="h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
                Product log
              </span>
              <h1 className="mt-4 max-w-[11ch] font-analog text-[44px] leading-[1] text-text-primary md:text-[68px]">
                Changes that ship the loop.
              </h1>
            </div>
            <div className="flex flex-col justify-end gap-4 text-[14px] leading-relaxed text-text-tertiary">
              <p>
                Release notes are read directly from GitHub releases, including attached
                binary assets and download counts.
              </p>
              <a
                href="https://github.com/polymetrics-ai/cli/releases"
                target="_blank"
                rel="noreferrer"
                className="inline-flex w-fit items-center gap-2 border border-line-structure bg-surface-1 px-3 py-2 font-square text-[12px] font-semibold text-text-secondary transition-colors hover:border-line-cta hover:text-text-primary"
              >
                <Github className="h-3.5 w-3.5" aria-hidden="true" />
                GitHub releases
                <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
              </a>
            </div>
          </header>

          <div className="mb-10">
            <ReleaseDownloadPanel release={latestRelease} />
          </div>

          <section id="release-notes" aria-label="Release notes" className="scroll-mt-24">
            <div className="mb-4 flex flex-wrap items-end justify-between gap-3 border-b border-line-structure pb-3">
              <div>
                <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                  GitHub release log
                </p>
                <h2 className="mt-1 font-square text-[20px] font-semibold text-text-primary">
                  {releases.length > 0 ? 'Latest releases' : 'Waiting for the first release'}
                </h2>
              </div>
              <span className="font-mono text-[11px] uppercase tracking-widest text-text-disabled">
                {releases.length} release{releases.length === 1 ? '' : 's'}
              </span>
            </div>

            {releases.length > 0 ? (
              <div className="divide-y divide-line-structure border-y border-line-structure">
                {releases.map((release) => (
                  <article key={release.id} className="bg-surface-bg py-7">
                    <div className="grid gap-5 lg:grid-cols-[8rem_minmax(0,1fr)]">
                      <aside className="font-mono text-[11px] uppercase tracking-widest text-text-disabled">
                        <div>{release.tagName}</div>
                        <div className="mt-1">{formatReleaseDate(release.publishedAt)}</div>
                        {release.prerelease ? (
                          <div className="mt-2 inline-flex border border-line-structure bg-surface-1 px-1.5 py-1 text-[10px] text-text-tertiary">
                            prerelease
                          </div>
                        ) : null}
                      </aside>

                      <div className="min-w-0">
                        <div className="flex flex-wrap items-start justify-between gap-3">
                          <div className="min-w-0">
                            <h3 className="font-square text-[24px] font-semibold leading-tight text-text-primary">
                              {releaseTitle(release)}
                            </h3>
                            <p className="mt-2 max-w-[74ch] whitespace-pre-line text-[14px] leading-relaxed text-text-tertiary">
                              {release.body || releaseExcerpt(release.body)}
                            </p>
                          </div>
                          <a
                            href={release.htmlUrl}
                            target="_blank"
                            rel="noreferrer"
                            className="inline-flex items-center gap-1.5 border border-line-structure bg-surface-1 px-2.5 py-1.5 font-mono text-[10px] uppercase tracking-wider text-text-secondary transition-colors hover:border-line-cta hover:text-text-primary"
                          >
                            Open
                            <ExternalLink className="h-3 w-3" aria-hidden="true" />
                          </a>
                        </div>

                        <div className="mt-5 grid gap-2 md:grid-cols-3">
                          <div className="border border-line-structure bg-surface-1 px-3 py-2">
                            <span className="block font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                              Assets
                            </span>
                            <span className="mt-1 block font-square text-[18px] font-semibold text-text-primary">
                              {release.assets.length}
                            </span>
                          </div>
                          <div className="border border-line-structure bg-surface-1 px-3 py-2">
                            <span className="block font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                              Downloads
                            </span>
                            <span className="mt-1 block font-square text-[18px] font-semibold text-text-primary">
                              {totalDownloads(release.assets).toLocaleString()}
                            </span>
                          </div>
                          <div className="border border-line-structure bg-surface-1 px-3 py-2">
                            <span className="block font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                              Largest asset
                            </span>
                            <span className="mt-1 block font-square text-[18px] font-semibold text-text-primary">
                              {release.assets.length > 0
                                ? formatBytes(Math.max(...release.assets.map((asset) => asset.size)))
                                : '0 B'}
                            </span>
                          </div>
                        </div>

                        {release.assets.length > 0 ? (
                          <div className="mt-4 divide-y divide-line-structure border border-line-structure bg-surface-1">
                            {release.assets.map((asset) => (
                              <a
                                key={asset.id}
                                href={asset.url}
                                className="grid gap-2 px-3 py-3 transition-colors hover:bg-surface-bg sm:grid-cols-[minmax(0,1fr)_7rem_7rem] sm:items-center"
                              >
                                <span className="min-w-0 truncate font-mono text-[12px] text-text-secondary">
                                  <Download className="mr-1.5 inline h-3.5 w-3.5 text-line-cta" aria-hidden="true" />
                                  {asset.name}
                                </span>
                                <span className="font-mono text-[11px] text-text-disabled">
                                  {formatBytes(asset.size)}
                                </span>
                                <span className="font-mono text-[11px] text-text-disabled">
                                  {asset.downloadCount.toLocaleString()} download{asset.downloadCount === 1 ? '' : 's'}
                                </span>
                              </a>
                            ))}
                          </div>
                        ) : null}
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <div className="border border-line-structure bg-surface-bg p-6">
                <div className="grid gap-4 sm:grid-cols-[2.5rem_minmax(0,1fr)]">
                  <span className="flex h-10 w-10 items-center justify-center border border-line-structure bg-surface-1 text-line-cta">
                    <Radio className="h-4 w-4" aria-hidden="true" />
                  </span>
                  <div className="min-w-0">
                    <h3 className="font-square text-[20px] font-semibold text-text-primary">
                      No GitHub releases published yet
                    </h3>
                    <p className="mt-2 max-w-[70ch] text-[14px] leading-relaxed text-text-tertiary">
                      The repository has no public releases right now. When the release workflow
                      creates the first tag, this page will show the release body, attached binary
                      files, file sizes, and download counts from GitHub.
                    </p>
                  </div>
                </div>
              </div>
            )}
          </section>
        </div>
      </main>

      <OnPageTocAside className="home-aside-panel" items={tocItems} />
    </div>
  );
}
