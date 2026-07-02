import type { GithubRelease, ReleaseAsset } from '@/lib/release-utils';

type GithubApiAsset = {
  id: number;
  name: string;
  browser_download_url: string;
  download_count: number;
  size: number;
  content_type: string;
};

type GithubApiRelease = {
  id: number;
  tag_name: string;
  name: string | null;
  body: string | null;
  html_url: string;
  published_at: string | null;
  prerelease: boolean;
  draft: boolean;
  assets: GithubApiAsset[];
};

const RELEASES_URL = 'https://api.github.com/repos/polymetrics-ai/cli/releases';

function githubHeaders(): HeadersInit {
  const headers: HeadersInit = {
    Accept: 'application/vnd.github+json',
    'User-Agent': 'polymetrics-cli-website',
    'X-GitHub-Api-Version': '2022-11-28',
  };
  const token = process.env.GITHUB_TOKEN ?? process.env.GH_TOKEN;
  if (token) headers.Authorization = `Bearer ${token}`;
  return headers;
}

export function mapGithubRelease(release: GithubApiRelease): GithubRelease {
  return {
    id: release.id,
    tagName: release.tag_name,
    name: release.name ?? release.tag_name,
    body: release.body ?? '',
    htmlUrl: release.html_url,
    publishedAt: release.published_at ?? '',
    prerelease: release.prerelease,
    draft: release.draft,
    assets: release.assets.map((asset): ReleaseAsset => ({
      id: asset.id,
      name: asset.name,
      url: asset.browser_download_url,
      downloadCount: asset.download_count,
      size: asset.size,
      contentType: asset.content_type,
    })),
  };
}

export async function getGithubReleases(limit = 10): Promise<GithubRelease[]> {
  const perPage = Math.min(Math.max(limit, 1), 20);

  try {
    const response = await fetch(`${RELEASES_URL}?per_page=${perPage}`, {
      headers: githubHeaders(),
      next: { revalidate: 300 },
    });

    if (!response.ok) return [];

    const releases = await response.json() as GithubApiRelease[];
    return releases.filter((release) => !release.draft).map(mapGithubRelease);
  } catch {
    return [];
  }
}
