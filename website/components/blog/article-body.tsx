'use client';

import { CornerBox } from '@/components/ui/corner-box';
import type { BlogSection } from '@/lib/blog';
import { AnnotationsProvider } from '@/components/blog/annotations-provider';
import { HighlightedBlock, HoverPreview } from '@/components/blog/highlight-text';
import { SelectionPopover } from '@/components/blog/selection-popover';
import { CommentComposer } from '@/components/blog/comment-composer';

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
}: {
  slug: string;
  sections: BlogSection[];
  summary: string;
}) {
  return (
    <AnnotationsProvider slug={slug}>
      <div className="grid gap-10 lg:grid-cols-[minmax(0,1fr)_14rem]">
        <div className="min-w-0" data-annotation-root>
          {sections.map((section, index) => (
            <section key={section.heading} id={sectionId(index)} className="mb-12 scroll-mt-24">
              <div className="mb-4 flex items-baseline gap-3 border-b border-line-structure pb-2">
                <span className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h2 className="font-square text-[20px] font-semibold leading-[1.25] text-text-primary">
                  {section.heading}
                </h2>
              </div>

              <div className="flex flex-col gap-4">
                {section.body.map((paragraph, blockIndex) => (
                  <p
                    key={paragraph}
                    data-annotation-block
                    data-section-index={index}
                    data-block-type="body"
                    data-block-index={blockIndex}
                    className="text-[15px] leading-[1.75] text-text-tertiary"
                  >
                    <HighlightedBlock
                      text={paragraph}
                      sectionIndex={index}
                      blockType="body"
                      blockIndex={blockIndex}
                    />
                  </p>
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

        <aside className="lg:sticky lg:top-24 lg:self-start">
          <CornerBox className="p-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-text-disabled">
              Summary
            </p>
            <p className="mt-3 text-[13px] leading-relaxed text-text-tertiary">{summary}</p>
          </CornerBox>
        </aside>
      </div>

      <SelectionPopover />
      <CommentComposer />
      <HoverPreview />
    </AnnotationsProvider>
  );
}
