const GITHUB_DISCUSSIONS_URL = 'https://github.com/polymetrics-ai/cli/discussions';

export function githubDiscussionUrl(title: string): string {
  return `${GITHUB_DISCUSSIONS_URL}?discussions_q=${encodeURIComponent(`"${title}"`)}`;
}

export { GITHUB_DISCUSSIONS_URL };
