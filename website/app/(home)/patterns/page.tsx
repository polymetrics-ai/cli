import type { Metadata } from 'next';
import { HomeSidebar } from '@/components/home/home-sidebar';
import { PageAside } from '@/components/home/page-aside';
import { PATTERNS, PATTERN_COUNT, PATTERN_FAMILIES } from '@/lib/patterns.generated';

export const metadata: Metadata = {
  title: 'Pattern registry',
  description:
    'A registry of 100 generative mathematical patterns used across the Polymetrics site: phyllotaxis, quasicrystals, Penrose tilings, spirals, rose curves, spirographs, and more.',
};

function patternFamilyId(family: string) {
  return `pattern-family-${family.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`;
}

export default function PatternsPage() {
  const families = PATTERN_FAMILIES as string[];
  const patternTocItems = [
    { id: 'patterns-overview', label: 'Overview' },
    ...families.map((family) => ({
      id: patternFamilyId(family),
      label: family,
    })),
  ];

  return (
    <div className="mx-auto flex w-full max-w-[95rem] overflow-clip">
      <HomeSidebar />
      <main className="min-w-0 flex-1 pattern-bg px-6 py-16 md:py-24">
        <div className="mx-auto w-full max-w-[1040px]">
          <header id="patterns-overview" className="mb-12 flex scroll-mt-24 flex-col gap-3">
            <span className="font-mono text-[12px] uppercase tracking-widest text-text-disabled">
              {PATTERN_COUNT} patterns · {families.length} families
            </span>
            <h1 className="font-analog text-[40px] leading-[1.05] text-text-primary md:text-[56px]">
              The math behind the canvas.
            </h1>
            <p className="max-w-[60ch] text-[15px] leading-relaxed text-text-tertiary">
              Every surface on this site is quietly backed by a generative mathematical
              pattern. They are all deterministic, computed from first principles, and
              rendered as faint emerald line-art so the writing always stays in focus. Here
              is the full registry.
            </p>
          </header>

          {families.map((family) => {
            const items = PATTERNS.filter((p) => p.family === family);
            return (
              <section key={family} id={patternFamilyId(family)} className="mb-14 scroll-mt-24">
                <div className="mb-4 flex items-baseline gap-3 border-b border-line-structure pb-2">
                  <h2 className="text-[16px] font-medium text-text-secondary">{family}</h2>
                  <span className="text-[12px] text-text-disabled">{items.length}</span>
                </div>
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6">
                  {items.map((p) => (
                    <figure
                      key={p.id}
                      className="group flex flex-col border border-line-structure bg-surface-bg transition-colors hover:bg-surface-1"
                    >
                      <div className="aspect-square overflow-hidden bg-surface-1">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                          src={p.file}
                          alt={p.name}
                          loading="lazy"
                          width={1000}
                          height={1000}
                          className="h-full w-full object-cover opacity-90 transition-opacity group-hover:opacity-100"
                        />
                      </div>
                      <figcaption className="flex flex-col gap-0.5 p-2.5">
                        <span className="text-[12px] font-medium text-text-secondary">
                          {p.name}
                        </span>
                        <span className="font-mono text-[10px] leading-tight text-text-tertiary">
                          {p.formula}
                        </span>
                      </figcaption>
                    </figure>
                  ))}
                </div>
              </section>
            );
          })}
        </div>
      </main>
      <PageAside items={patternTocItems} discussionTitle="Pattern Registry" />
    </div>
  );
}
