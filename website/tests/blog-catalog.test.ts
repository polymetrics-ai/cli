import { describe, expect, it } from 'vitest';

import { getBlogPost } from '@/lib/blog';

describe('blog catalog', () => {
  it('publishes the human harnesses essay with verified repository evidence', () => {
    const post = getBlogPost('human-harnesses');

    expect(post).toBeDefined();
    if (!post) throw new Error('Expected human-harnesses blog post');
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
    expect(articleText).toContain('PR #27 and 1,961,878 changed lines');
    expect(articleText).toContain('PR #29 eventually landed 2,792,444 changed lines');
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

    const interactivePost = post as unknown as {
      repositoryCta?: { label: string; href: string };
      evidence?: Array<{
        id: string;
        url: string;
        apiUrl?: string;
        snapshot: { status: string; stats: Array<{ label: string; value: string }> };
      }>;
      sections: Array<{
        body: string[];
        evidenceRefs?: Array<{ blockIndex: number; evidenceId: string; text: string }>;
        evidenceIds?: string[];
      }>;
    };
    const evidenceById = new Map(
      interactivePost.evidence?.map((evidence) => [evidence.id, evidence]) ?? [],
    );

    expect(interactivePost.repositoryCta).toEqual({
      label: 'Star Polymetrics on GitHub',
      href: 'https://github.com/polymetrics-ai/cli',
    });
    expect(evidenceById.get('precursor-pr-27')).toMatchObject({
      url: 'https://github.com/polymetrics-ai/cli/pull/27',
      apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/27',
      snapshot: { status: 'Closed, not merged' },
    });
    expect(evidenceById.get('migration-pr-29')).toMatchObject({
      url: 'https://github.com/polymetrics-ai/cli/pull/29',
      apiUrl: 'https://api.github.com/repos/polymetrics-ai/cli/pulls/29',
      snapshot: { status: 'Merged' },
    });
    expect(evidenceById.get('migration-merge-commit')).toMatchObject({
      url: 'https://github.com/polymetrics-ai/cli/commit/605b006e5aa1adae697d5b7dd26ec485c570c250',
      apiUrl:
        'https://api.github.com/repos/polymetrics-ai/cli/commits/605b006e5aa1adae697d5b7dd26ec485c570c250',
    });
    expect(evidenceById.get('migration-merge-commit')?.snapshot.stats).toEqual(
      expect.arrayContaining([{ label: 'Changed lines', value: '2,792,444' }]),
    );
    expect(evidenceById.get('issue-first-pr-47')?.url).toBe(
      'https://github.com/polymetrics-ai/cli/pull/47',
    );
    expect(evidenceById.get('parent-orchestrator-pr-51')?.url).toBe(
      'https://github.com/polymetrics-ai/cli/pull/51',
    );

    expect(interactivePost.sections.every((section) => section.evidenceIds === undefined)).toBe(true);

    const references = interactivePost.sections.flatMap((section) =>
      (section.evidenceRefs ?? []).map((reference) => ({ section, reference })),
    );
    const referencedEvidence = new Set(references.map(({ reference }) => reference.evidenceId));
    for (const id of evidenceById.keys()) {
      expect(referencedEvidence).toContain(id);
    }
    for (const { section, reference } of references) {
      expect(evidenceById.has(reference.evidenceId)).toBe(true);
      expect(section.body[reference.blockIndex]).toContain(reference.text);
    }

    const reviewSection = post.sections.find(
      (section) => section.heading === 'Review, fix, repeat, locally',
    ) as typeof post.sections[number] & {
      image?: {
        src: string;
        alt: string;
        caption: string;
        width: number;
        height: number;
        afterBlock: number;
      };
    };
    expect(reviewSection.code).toBeUndefined();
    expect(reviewSection.image).toEqual({
      src: '/blog/human-harnesses/04-review-repair-loop.webp',
      alt: 'Separate review and repair stations passing the same verified artifact around a loop.',
      caption:
        'Review and repair stay separate while verification sends the same artifact back around the loop.',
      width: 1536,
      height: 1024,
      afterBlock: 2,
    });
  });
});
