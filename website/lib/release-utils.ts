export type ReleaseAsset = {
  id: number;
  name: string;
  url: string;
  downloadCount: number;
  size: number;
  contentType: string;
};

export type GithubRelease = {
  id: number;
  tagName: string;
  name: string;
  body: string;
  htmlUrl: string;
  publishedAt: string;
  prerelease: boolean;
  draft: boolean;
  assets: ReleaseAsset[];
};

export function releaseTitle(release: GithubRelease): string {
  return release.name || release.tagName;
}

export function formatReleaseDate(value: string | undefined): string {
  if (!value) return 'unpublished';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'unpublished';
  return new Intl.DateTimeFormat('en', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  }).format(date);
}

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }

  const precision = size >= 10 || unitIndex === 0 ? 0 : 1;
  return `${size.toFixed(precision)} ${units[unitIndex]}`;
}

export function releaseExcerpt(body: string, maxLength = 220): string {
  const normalized = body
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/[#*_>`-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();

  if (!normalized) return 'Release notes will appear here once the release is published.';
  if (normalized.length <= maxLength) return normalized;
  if (maxLength <= 3) return '.'.repeat(Math.max(maxLength, 0));
  return `${normalized.slice(0, maxLength - 3).trim()}...`;
}

export function normalizePlatformHint(platform: string): string {
  const hint = platform.toLowerCase();
  if (hint.includes('darwin') || hint.includes('mac')) return 'darwin';
  if (hint.includes('win')) return 'windows';
  if (hint.includes('linux')) return 'linux';
  return hint;
}

export function selectPreferredAsset(
  assets: ReleaseAsset[],
  platformHint = '',
): ReleaseAsset | null {
  if (assets.length === 0) return null;

  const normalizedHint = normalizePlatformHint(platformHint);
  const architectureHint = platformHint.toLowerCase().includes('arm')
    ? 'arm64'
    : platformHint.toLowerCase().includes('aarch64')
      ? 'arm64'
      : platformHint.toLowerCase().includes('64')
        ? 'amd64'
        : '';

  const byPlatform = normalizedHint
    ? assets.filter((asset) => asset.name.toLowerCase().includes(normalizedHint))
    : assets;

  const byArchitecture = architectureHint
    ? byPlatform.find((asset) => {
        const name = asset.name.toLowerCase();
        return name.includes(architectureHint) || (architectureHint === 'amd64' && name.includes('x86_64'));
      })
    : undefined;

  return byArchitecture ?? byPlatform[0] ?? assets[0];
}
