import { describe, expect, it } from 'vitest';

import { mapGithubRelease } from '../../lib/github-releases';

describe('GitHub release mapping', () => {
  it('normalizes GitHub release API payloads for the website', () => {
    const release = mapGithubRelease({
      id: 42,
      tag_name: 'v0.2.0',
      name: null,
      body: 'Release body',
      html_url: 'https://github.com/polymetrics-ai/cli/releases/tag/v0.2.0',
      published_at: '2026-07-02T08:00:00Z',
      prerelease: false,
      draft: false,
      assets: [
        {
          id: 7,
          name: 'pm-linux-amd64.tar.gz',
          browser_download_url: 'https://github.com/download/linux',
          download_count: 5,
          size: 18_000_000,
          content_type: 'application/gzip',
        },
      ],
    });

    expect(release).toEqual({
      id: 42,
      tagName: 'v0.2.0',
      name: 'v0.2.0',
      body: 'Release body',
      htmlUrl: 'https://github.com/polymetrics-ai/cli/releases/tag/v0.2.0',
      publishedAt: '2026-07-02T08:00:00Z',
      prerelease: false,
      draft: false,
      assets: [
        {
          id: 7,
          name: 'pm-linux-amd64.tar.gz',
          url: 'https://github.com/download/linux',
          downloadCount: 5,
          size: 18_000_000,
          contentType: 'application/gzip',
        },
      ],
    });
  });
});
