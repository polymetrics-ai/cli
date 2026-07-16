import { describe, expect, it } from 'vitest';

import { getBlogPost } from '@/lib/blog';

describe('blog catalog', () => {
  it('publishes the human harnesses essay with verified repository evidence', () => {
    const post = getBlogPost('human-harnesses');

    expect(post).toBeDefined();
    expect(post).toMatchObject({
      title: 'Humans Need Harnesses Too',
      publishedAt: '2026-07-16',
      updatedAt: '2026-07-16',
      readingTime: '14 min read',
      category: 'Build in public',
    });
    expect(post?.tags).toEqual(
      expect.arrayContaining(['human harnesses', 'GitHub Actions', 'approval gates']),
    );

    const headings = post?.sections.map((section) => section.heading);
    expect(headings).toEqual(
      expect.arrayContaining([
        'The PR that ate the repository',
        'The repository became a harness',
        'Review, fix, repeat, locally',
        'What the harness still does not do',
      ]),
    );

    const articleText = post?.sections
      .flatMap((section) => [
        ...section.body,
        ...(section.points ?? []),
        section.code ?? '',
      ])
      .join(' ');

    expect(articleText).toContain('29,129');
    expect(articleText).toContain('14,780');
    expect(articleText).toContain('14,169');
    expect(articleText).toContain('177');
    expect(articleText).toContain('2,903');
    expect(articleText).toContain('7,088');
    expect(articleText).toContain('pm reverse run <plan-id> --approve <approval-token> --json');
    expect(articleText).not.toContain('pm reverse approve');
    expect(articleText).toContain('roughly a million changed lines');
    expect(articleText).toContain('isolated worktree');
    expect(articleText).toContain('exact candidate head SHA');
    expect(articleText).toContain('read-only reviewer');
    expect(articleText).toContain('isolated repair worker');
    expect(articleText).toContain('four correction rounds');
    expect(articleText).toContain('Remote PR-bot review can still run as supplemental');
    expect(articleText).toContain('Shepherd is the next story');
    expect(articleText).toContain('star the repository');
    expect(articleText).not.toContain('The Claude review workflow');
    expect(articleText).not.toContain('Shepherd independently checks whether');
    expect(articleText).not.toContain('Repository: github.com/polymetrics-ai/cli');
    expect(articleText).not.toContain('Documentation: cli.polymetrics.ai');
    expect(articleText).not.toContain('Inventory snapshot:');
  });
});
