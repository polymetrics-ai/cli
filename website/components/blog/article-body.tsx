'use client';

import { Fragment, useMemo, useRef, useState } from 'react';
import { CornerBox } from '@/components/ui/corner-box';
import type { BlogEvidence, BlogSection } from '@/lib/blog';
import { AnnotationsProvider } from '@/components/blog/annotations-provider';
import { HighlightedBlock, HoverPreview } from '@/components/blog/highlight-text';
import { SelectionPopover } from '@/components/blog/selection-popover';
import { CommentComposer } from '@/components/blog/comment-composer';
import { MarginNotesRail } from '@/components/blog/margin-notes-rail';
import { CommentsSheet } from '@/components/blog/comments-sheet';
import { GitHubEvidenceDialog } from '@/components/blog/github-evidence';
import { ArticleFigure } from '@/components/blog/article-figure';

function sectionId(index: number): string {
  return `section-${index + 1}`;
}

/**
 * Client-side extraction of the blog article body. Renders the exact
 * markup the server component used to emit (so prerendered pages are
 * unchanged) plus the annotation layer: text selection → popover →
 * comment/bookmark, and highlight rendering for existing notes.
 */
export function ArticleBody({
  slug,
  sections,
  summary,
  evidence,
}: {
  slug: string;
  sections: BlogSection[];
  summary: string;
  evidence: BlogEvidence[];
}) {
  const gridRef = useRef<HTMLDivElement>(null);
  const evidenceTriggerRef = useRef<HTMLElement | null>(null);
  const [activeEvidence, setActiveEvidence] = useState<BlogEvidence | null>(null);
  const evidenceById = useMemo(
    () => new Map(evidence.map((item, index) => [item.id, { evidence: item, number: index + 1 }])),
    [evidence],
  );

  const closeEvidence = () => {
    const trigger = evidenceTriggerRef.current;
    setActiveEvidence(null);
    window.requestAnimationFrame(() => trigger?.focus());
  };

  return (
    <AnnotationsProvider slug={slug}>
      <div ref={gridRef} className="blog-article-grid relative grid gap-10">
        <div className="min-w-0" data-annotation-root>
          {sections.map((section, index) => (
            <section key={section.heading} id={sectionId(index)} className="mb-12 scroll-mt-24">
              {(section.images ?? [])
                .filter((image) => image.beforeHeading)
                .map((image) => (
                  <ArticleFigure key={image.src} image={image} className="mb-5" />
                ))}
              <div className="mb-4 flex items-baseline gap-3 border-b border-line-structure pb-2">
                <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h2 className="font-square text-[20px] font-semibold leading-[1.25] text-text-primary">
                  {section.heading}
                </h2>
              </div>

              <div className="flow-root">
                {section.body.map((paragraph, blockIndex) => (
                  <Fragment key={`${section.heading}-${blockIndex}`}>
                    <p
                      data-annotation-block
                      data-section-index={index}
                      data-block-type="body"
                      data-block-index={blockIndex}
                      className="mb-4 text-[15px] leading-[1.75] text-text-tertiary"
                    >
                      <HighlightedBlock
                        text={paragraph}
                        sectionIndex={index}
                        blockType="body"
                        blockIndex={blockIndex}
                        evidenceReferences={(section.evidenceRefs ?? []).flatMap((reference) => {
                          if (reference.blockIndex !== blockIndex) return [];
                          const item = evidenceById.get(reference.evidenceId);
                          return item ? [{ ...item, text: reference.text }] : [];
                        })}
                        onEvidenceOpen={(item, trigger) => {
                          evidenceTriggerRef.current = trigger;
                          setActiveEvidence(item);
                        }}
                      />
                    </p>
                    {(section.images ?? [])
                      .filter((image) => image.afterBlock === blockIndex)
                      .map((image) => <ArticleFigure key={image.src} image={image} />)}
                  </Fragment>
                ))}
              </div>

              {section.points ? (
                <ul className="mt-5 grid gap-2">
                  {section.points.map((point, blockIndex) => (
                    <li
                      key={point}
                      data-annotation-block
                      data-section-index={index}
                      data-block-type="point"
                      data-block-index={blockIndex}
                      className="border-l border-line-cta bg-surface-1 px-4 py-2 text-[14px] leading-relaxed text-text-secondary"
                    >
                      <HighlightedBlock
                        text={point}
                        sectionIndex={index}
                        blockType="point"
                        blockIndex={blockIndex}
                      />
                    </li>
                  ))}
                </ul>
              ) : null}

              {section.code ? (
                <pre className="mt-5 overflow-x-auto border border-line-structure bg-[#101713] p-4 text-[13px] leading-relaxed text-emerald-100">
                  <code>{section.code}</code>
                </pre>
              ) : null}
            </section>
          ))}
        </div>

        <aside className="blog-article-aside">
          <CornerBox className="p-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Summary
            </p>
            <p className="mt-3 text-[13px] leading-relaxed text-text-tertiary">{summary}</p>
          </CornerBox>
          <MarginNotesRail containerRef={gridRef} />
        </aside>
      </div>

      <SelectionPopover />
      <CommentComposer />
      <HoverPreview />
      <CommentsSheet sections={sections} />
      <GitHubEvidenceDialog evidence={activeEvidence} onClose={closeEvidence} />
    </AnnotationsProvider>
  );
}
